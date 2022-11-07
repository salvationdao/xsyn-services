package api

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"xsyn-services/passport/db"
	"xsyn-services/passport/email"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
	"xsyn-services/types"

	"github.com/meehow/securebytes"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"go.uber.org/atomic"

	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"github.com/ninja-syndicate/ws"

	"errors"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type TwitchConfig struct {
	ExtensionSecret []byte
	ClientID        string
	ClientSecret    string
}

// API server
type API struct {
	SupremacyController *SupremacyControllerWS
	State               *types.State
	SupUSD              decimal.Decimal
	Log                 *zerolog.Logger
	Addr                string
	Mailer              *email.Mailer
	SMS                 types.SMS
	HTMLSanitize        *bluemonday.Policy
	Cookie              *securebytes.SecureBytes
	IsCookieSecure      bool
	TokenExpirationDays int
	TokenEncryptionKey  []byte
	Eip712Message       string
	Twitch              *TwitchConfig
	Twitter             *auth.TwitterConfig
	Google              *auth.GoogleConfig
	ClientToken         string
	WebhookToken        string
	GameserverHostUrl   string
	Commander           *ws.Commander
	Web3Params          *types.Web3Params
	botSecretKey        string

	// online user cache
	users chan func(userCacheList Transactor)

	// server clients
	// serverClients       chan func(serverClients ServerClientsList)
	// sendToServerClients chan *ServerClientMessage

	//tx stuff
	userCacheMap *Transactor

	walletOnlyConnect    bool
	storeItemExternalUrl string

	// supremacy client map
	ClientMap *sync.Map

	// captcha
	captcha *captcha

	JWTKey []byte
}

// NewAPI registers routes
func NewAPI(
	log *zerolog.Logger,
	mailer *email.Mailer,
	twilio types.SMS,
	addr string,
	HTMLSanitize *bluemonday.Policy,
	config *types.Config,
	externalUrl string,
	ucm *Transactor,
	isTestnetBlockchain bool,
	runBlockchainBridge bool,
	enablePurchaseSubscription bool,
	jwtKey []byte,
	environment string,
	ignoreRateLimitIPs []string,
	pxr *PassportExchangeRate,

) (*API, chi.Router) {

	api := &API{
		Web3Params:  config.Web3Params,
		ClientToken: config.AuthParams.GameserverToken,
		// webhook setup
		WebhookToken:      config.WebhookParams.GameserverWebhookToken,
		GameserverHostUrl: config.WebhookParams.GameserverHostUrl,

		TokenExpirationDays: config.TokenExpirationDays,
		TokenEncryptionKey:  []byte(config.EncryptTokensKey),
		Google: &auth.GoogleConfig{
			ClientID: config.AuthParams.GoogleClientID,
		},
		Twitch: &TwitchConfig{
			ClientID:     config.AuthParams.TwitchClientID,
			ClientSecret: config.AuthParams.TwitchClientSecret,
		},
		Twitter: &auth.TwitterConfig{
			APIKey:    config.AuthParams.TwitterAPIKey,
			APISecret: config.AuthParams.TwitterAPISecret,
		},
		Eip712Message: config.MetaMaskSignMessage,
		Cookie: securebytes.New(
			[]byte(config.CookieKey),
			securebytes.ASN1Serializer{}),
		IsCookieSecure:       config.CookieSecure,
		Log:                  log_helpers.NamedLogger(log, "api"),
		Addr:                 addr,
		Mailer:               mailer,
		SMS:                  twilio,
		HTMLSanitize:         HTMLSanitize,
		users:                make(chan func(userList Transactor)),
		userCacheMap:         ucm,
		walletOnlyConnect:    config.OnlyWalletConnect,
		storeItemExternalUrl: externalUrl,

		ClientMap:    &sync.Map{},
		JWTKey:       jwtKey,
		botSecretKey: config.BotSecret,

		captcha: &captcha{
			secret:    config.CaptchaSecret,
			siteKey:   config.CaptchaSiteKey,
			verifyUrl: "https://hcaptcha.com/siteverify",
		},
	}

	api.Commander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})

	cc := NewChainClients(log, api, config.Web3Params, isTestnetBlockchain, runBlockchainBridge, enablePurchaseSubscription)
	r := chi.NewRouter()
	r.Use(cors.New(
		cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(passlog.ChiLogger(zerolog.InfoLevel))

	var err error
	roadmapRoutes, err := RoadmapRoutes()
	if err != nil {
		log.Fatal().Msgf("failed to roadmap routes: %s", err.Error())
	}

	ws.Init(&ws.Config{
		Logger:        passlog.L,
		SkipRateLimit: os.Getenv("PASSPORT_ENVIRONMENT") == "staging" || os.Getenv("PASSPORT_ENVIRONMENT") == "development",
	})

	if runBlockchainBridge {
		_ = NewSupController(log, api, cc)
	}

	_ = NewAssetController(log, api)
	_ = NewCollectionController(log, api, isTestnetBlockchain)

	_ = NewCheckController(log, api)
	_ = NewUserActivityController(log, api)
	uc := NewUserController(log, api, &auth.GoogleConfig{
		ClientID: config.AuthParams.GoogleClientID,
	}, &auth.TwitchConfig{
		ClientID:     config.AuthParams.TwitchClientID,
		ClientSecret: config.AuthParams.TwitchClientSecret,
	}, &auth.DiscordConfig{
		ClientID:     config.AuthParams.DiscordClientID,
		ClientSecret: config.AuthParams.DiscordClientSecret,
	})
	_ = NewTransactionController(log, api)
	_ = NewFactionController(log, api)
	_ = NewRoleController(log, api)
	sc := NewSupremacyController(log, api)
	_ = NewGamebarController(log, api)
	_ = NewStoreController(log, api)
	d := DevRoutes(ucm)

	r.Mount("/api/admin", AdminRoutes(ucm))
	r.Mount("/api/roadmap", roadmapRoutes)
	r.Handle("/metrics", promhttp.Handler())
	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			if environment != "development" {
				r.Use(DatadogTracer.Middleware())
			}

			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
			r.Mount("/check", CheckRouter(log_helpers.NamedLogger(log, "check router")))
			r.Mount("/files", FileRouter(api))
			r.Mount("/nfts", api.NFTRoutes())
			r.Mount("/moderator", ModeratorRoutes())
			if environment == "development" {
				r.Mount("/dev", d.R)
			}

			//r.Get("/verify", WithError(api.Auth.VerifyAccountHandler))
			r.Get("/get-nonce", WithError(api.GetNonce))
			//r.Get("/auth/twitter", WithError(api.Auth.TwitterAuth))
			if os.Getenv("PASSPORT_ENVIRONMENT") != "staging" {
				r.Get("/withdraw/holding/{user_address}", WithError(api.HoldingSups))
				r.Get("/withdraw/check/{address}", WithError(api.GetMaxWithdrawAmount))
				r.Get("/withdraw/check", WithError(api.CheckCanWithdraw))
				r.Get("/deposit/check", WithError(api.CheckCanDeposit))
				r.Get("/withdraw/{address}/{nonce}/{amount}/{chain}", WithError(api.WithdrawSups))

				r.Get("/1155/{address}/{token_id}/{nonce}/{amount}", WithError(api.Withdraw1155))
			}
			r.Get("/1155/contracts", WithError(api.Get1155Contracts))

			r.Get("/asset/{hash}", WithError(api.AssetGet))
			r.Get("/asset/{collection_address}/{token_id}", WithError(api.AssetGetByCollectionAndTokenID))
			r.Get("/whitelist/check", WithError(api.WhitelistOnlyWalletCheck))

			r.Get("/collection/1155/all", WithError(api.Get1155Collections))
			r.Get("/collection/{collection_slug}", WithError(api.Get1155Collection))

			r.Route("/early", func(r chi.Router) {
				r.Get("/check", WithError(api.CheckUserEarlyContributor))
				r.Post("/sign", WithError(api.EarlyContributorSignMessage))
			})
			r.Route("/auth", func(r chi.Router) {
				r.Get("/check", WithError(api.AuthCheckHandler))
				r.Get("/logout", WithError(api.AuthLogoutHandler))
				r.Post("/token", WithError(api.TokenLoginHandler))
				r.Post("/external", api.ExternalLoginHandler)
				r.Post("/wallet", WithError(api.WalletLoginHandler))
				r.Post("/email", WithError(api.EmailLoginHandler))
				r.Post("/email_signup", WithError(api.EmailSignupVerifyHandler))
				r.Post("/signup", WithError(api.SignupHandler))
				r.Post("/forgot", WithError(api.ForgotPasswordHandler))
				r.Post("/reset", WithError(api.ResetPasswordHandler))
				r.Post("/change_password", WithError(WithUser(api, api.ChangePasswordHandler)))
				r.Post("/new_password", WithError(WithUser(api, api.NewPasswordHandler)))
				r.Post("/google", WithError(api.GoogleLoginHandler))
				r.Post("/facebook", WithError(api.FacebookLoginHandler))
				r.Post("/tfa", WithError(api.TFAVerifyHandler))
				r.Get("/twitter", WithError(api.TwitterAuth))

				r.Post("/verify_code", WithError(api.VerifyCodeHandler))

				r.Post("/bot_list", api.BotListHandler)
				r.Post("/bot_token", api.BotTokenLoginHandler)
			})
		})
		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.

		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Route("/ws", func(r chi.Router) {
			r.Use(ws.TrimPrefix("/api/ws"))
			r.Mount("/public", ws.NewServer(func(s *ws.Server) {
				s.WS("/exchange_rates", HubKeySUPSExchangeRates, WithPassportExchangeRate(pxr, ExchangeRatesHandler))
				s.WS("/sups_remaining", HubKeySUPSRemainingSubscribe, uc.TotalSupRemainingHandler)
			}))
			r.Mount("/store", ws.NewServer(func(s *ws.Server) {
			}))
			r.Mount("/user/{userId}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(true, true))
				s.WS("/*", HubKeyUserGet, api.MustSecure(uc.GetHandler))
				s.Mount("/commander", api.Commander)
				s.WS("/sups", HubKeyUserSupsSubscribe, api.MustSecure(api.UserSupsUpdatedSubscribeHandler))
				s.WS("/transactions", HubKeyUserTransactionsSubscribe, api.MustSecure(api.UserTransactionsSubscribeHandler))
			}))
		})
	})

	api.SupremacyController = sc

	api.State, err = db.StateGet(isTestnetBlockchain)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Fatal().Err(err).Msgf("failed to init state object")
	}
	return api, r
}

// RecordUserActivity adds a UserActivity to the db for the current user
func (api *API) RecordUserActivity(
	ctx context.Context,
	userID string,
	action string,
	objectType types.ObjectType,
	objectID *string,
	objectSlug *string,
	objectName *string,
	changes ...*types.UserActivityChangeData,
) {
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		api.Log.Err(err).Msgf("issue creating uuid from %s", userID)
	}

	oldData, newData, err := types.UserActivityGetDataChanges(changes)
	if err != nil {
		api.Log.Err(err).Msg("issue getting oldData and newData JSON")
	}

	err = db.UserActivityCreate(
		types.UserID(userUUID),
		action,
		objectType,
		objectID,
		objectSlug,
		objectName,
		oldData,
		newData,
	)
	if err != nil {
		api.Log.Err(err).Msg("issue saving user activity")
	}
}

type PassportExchangeRate struct {
	isEnabled           atomic.Bool
	isCurrentBlockAfter atomic.Bool
}

func (pxr *PassportExchangeRate) GetIsEnable() bool {
	return pxr.isEnabled.Load()
}

func (pxr *PassportExchangeRate) GetIsCurrentBlockAfter() bool {
	return pxr.isCurrentBlockAfter.Load()
}

func (pxr *PassportExchangeRate) SetIsEnabledTrue() {
	pxr.isEnabled.Store(true)
}

func (pxr *PassportExchangeRate) SetIsCurrentBlockAfterTrue() {
	pxr.isCurrentBlockAfter.Store(true)
}

func WithPassportExchangeRate(pxr *PassportExchangeRate, fn func(ctx context.Context, pxr *PassportExchangeRate, key string, payload []byte, reply ws.ReplyFunc) error) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		return fn(ctx, pxr, key, payload, reply)
	}
}

const HubKeySUPSExchangeRates = "SUPS:EXCHANGE"

func ExchangeRatesHandler(ctx context.Context, pxr *PassportExchangeRate, key string, payload []byte, reply ws.ReplyFunc) error {
	exchangeRates, err := payments.FetchExchangeRates(pxr.GetIsEnable() && pxr.GetIsCurrentBlockAfter())
	if err != nil {
		return terror.Error(err, "Unable to fetch exchange rates.")
	}
	reply(exchangeRates)
	return nil
}
