package api

import (
	"context"
	"database/sql"
	"fmt"
	"passport"
	"passport/db"
	"passport/email"
	"sync"

	"github.com/ninja-software/log_helpers"

	SentryTracer "github.com/ninja-syndicate/hub/ext/sentry"

	"github.com/shopspring/decimal"

	"nhooyr.io/websocket"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-syndicate/hub/ext/auth"
	zerologger "github.com/ninja-syndicate/hub/ext/zerolog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// API server
type API struct {
	SupremacyController *SupremacyControllerWS
	State               *passport.State
	SupUSD              decimal.Decimal
	Log                 *zerolog.Logger
	Addr                string
	Mailer              *email.Mailer
	HTMLSanitize        *bluemonday.Policy
	Hub                 *hub.Hub
	Conn                *pgxpool.Pool
	Tokens              *Tokens
	*auth.Auth
	*messagebus.MessageBus
	ClientToken       string
	WebhookToken      string
	GameserverHostUrl string
	BridgeParams      *passport.BridgeParams

	// online user cache
	users chan func(userCacheList UserCacheMap)

	// server clients
	// serverClients       chan func(serverClients ServerClientsList)
	// sendToServerClients chan *ServerClientMessage

	//tx stuff
	transactionCache *TransactionCache
	userCacheMap     *UserCacheMap

	walletOnlyConnect    bool
	storeItemExternalUrl string

	// supremacy client map
	ClientMap *sync.Map
}

// NewAPI registers routes
func NewAPI(
	log *zerolog.Logger,
	conn *pgxpool.Pool,
	txConn *sql.DB,
	mailer *email.Mailer,
	addr string,
	HTMLSanitize *bluemonday.Policy,
	config *passport.Config,
	externalUrl string,
	tc *TransactionCache,
	ucm *UserCacheMap,
	isTestnetBlockchain bool,
	runBlockchainBridge bool,
	msgBus *messagebus.MessageBus,
	enablePurchaseSubscription bool,
) (*API, chi.Router) {

	api := &API{
		SupUSD:       decimal.New(12, -2),
		BridgeParams: config.BridgeParams,
		ClientToken:  config.AuthParams.GameserverToken,
		// webhook setup
		WebhookToken:      config.WebhookParams.GameserverWebhookToken,
		GameserverHostUrl: config.WebhookParams.GameserverHostUrl,

		Tokens: &Tokens{
			Conn:                conn,
			Mailer:              mailer,
			tokenExpirationDays: config.TokenExpirationDays,
			encryptToken:        config.EncryptTokens,
			encryptTokenKey:     config.EncryptTokensKey,
		},
		MessageBus: msgBus,
		Hub: hub.New(&hub.Config{
			ClientOfflineFn: func(client *hub.Client) {
				msgBus.UnsubAll(client)
			},
			Log:    zerologger.New(*log_helpers.NamedLogger(log, "hub library")),
			Tracer: SentryTracer.New(),
			WelcomeMsg: &hub.WelcomeMsg{
				Key:     "WELCOME",
				Payload: nil,
			},
			AcceptOptions: &websocket.AcceptOptions{
				InsecureSkipVerify: config.InsecureSkipVerifyCheck,
				OriginPatterns:     []string{"*"},
			},
			WebsocketReadLimit: 104857600,
		}),
		Log:          log_helpers.NamedLogger(log, "api"),
		Conn:         conn,
		Addr:         addr,
		Mailer:       mailer,
		HTMLSanitize: HTMLSanitize,
		// server clients
		// serverClients:       make(chan func(serverClients ServerClientsList)),
		// sendToServerClients: make(chan *ServerClientMessage),
		//382
		// user cache map
		users: make(chan func(userList UserCacheMap)),

		// object to hold transaction stuff
		transactionCache: tc,
		userCacheMap:     ucm,

		walletOnlyConnect:    config.OnlyWalletConnect,
		storeItemExternalUrl: externalUrl,

		ClientMap: &sync.Map{},
	}

	cc := NewChainClients(log, api, config.BridgeParams, isTestnetBlockchain, runBlockchainBridge, enablePurchaseSubscription)
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	}).Handler)

	var err error
	api.Auth, err = auth.New(api.Hub, &auth.Config{
		CreateUserIfNotExist:     true,
		CreateAndGetOAuthUserVia: auth.IdTypeID,
		Google: &auth.GoogleConfig{
			ClientID: config.AuthParams.GoogleClientID,
		},
		Twitch: &auth.TwitchConfig{
			ClientID:     config.AuthParams.TwitchClientID,
			ClientSecret: config.AuthParams.TwitchClientSecret,
		},
		Twitter: &auth.TwitterConfig{
			APIKey:    config.AuthParams.TwitterAPIKey,
			APISecret: config.AuthParams.TwitterAPISecret,
		},
		Discord: &auth.DiscordConfig{
			ClientID:     config.AuthParams.DiscordClientID,
			ClientSecret: config.AuthParams.DiscordClientSecret,
		},
		CookieSecure: config.CookieSecure,
		UserController: &UserGetter{
			Log:    log_helpers.NamedLogger(log, "user getter"),
			Conn:   conn,
			Mailer: mailer,
		},
		Tokens:                 api.Tokens,
		Eip712Message:          config.MetaMaskSignMessage,
		OnlyWalletConnect:      config.OnlyWalletConnect,
		WhitelistCheckEndpoint: fmt.Sprintf("%s/api/whitelist", config.WhitelistEndpoint),
	})
	if err != nil {
		log.Fatal().Msgf("failed to init hub auther: %s", err.Error())
	}

	r.Handle("/metrics", promhttp.Handler())
	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
			r.Mount("/check", CheckRouter(log_helpers.NamedLogger(log, "check router"), conn))
			r.Mount("/files", FileRouter(conn, api))
			r.Mount("/nfts", api.NFTRoutes())
			r.Mount("/admin", AdminRoutes(ucm))

			r.Get("/verify", WithError(api.Auth.VerifyAccountHandler))
			r.Get("/get-nonce", WithError(api.Auth.GetNonce))
			r.Get("/withdraw/{address}/{nonce}/{amount}", WithError(api.WithdrawSups))
			r.Get("/withdraw-tx-hash/{refundID}/{txHash}", WithError(api.UpdatePendingRefund))

			r.Get("/asset/{hash}", WithError(api.AssetGet))
			r.Get("/asset/{collection_address}/{token_id}", WithError(api.AssetGetByCollectionAndTokenID))
			r.Get("/auth/twitter", WithError(api.Auth.TwitterAuth))
			r.Get("/whitelist/check", WithError(api.WhitelistOnlyWalletCheck))
			r.Get("/faction-data", WithError(api.FactionGetData))
			r.Route("/early", func(r chi.Router) {
				r.Get("/check", WithError(api.CheckUserEarlyContributor))
				r.Post("/sign", WithError(api.EarlyContributorSignMessage))
			})
		})
		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Handle("/ws", api.Hub)
	})

	if runBlockchainBridge {
		_ = NewSupController(log, conn, api, cc)
	}
	_ = NewAssetController(log, conn, api)
	_ = NewCollectionController(log, conn, api, isTestnetBlockchain)
	_ = NewServerClientController(log, conn, api)
	_ = NewCheckController(log, conn, api)
	_ = NewUserActivityController(log, conn, api)
	_ = NewUserController(log, conn, api, &auth.GoogleConfig{
		ClientID: config.AuthParams.GoogleClientID,
	}, &auth.TwitchConfig{
		ClientID:     config.AuthParams.TwitchClientID,
		ClientSecret: config.AuthParams.TwitchClientSecret,
	}, &auth.DiscordConfig{
		ClientID:     config.AuthParams.DiscordClientID,
		ClientSecret: config.AuthParams.DiscordClientSecret,
	})
	_ = NewTransactionController(log, conn, api)
	_ = NewFactionController(log, conn, api)
	_ = NewOrganisationController(log, conn, api)
	_ = NewRoleController(log, conn, api)
	_ = NewProductController(log, conn, api)
	sc := NewSupremacyController(log, conn, api)
	_ = NewGamebarController(log, conn, api)
	_ = NewStoreController(log, conn, api)

	api.SupremacyController = sc

	//api.Hub.Events.AddEventHandler(hub.EventOnline, api.ClientOnline)
	api.Hub.Events.AddEventHandler(auth.EventLogin, api.ClientAuth, func(err error) {})
	api.Hub.Events.AddEventHandler(auth.EventLogout, api.ClientLogout, func(err error) {})
	api.Hub.Events.AddEventHandler(hub.EventOffline, api.ClientOffline, func(err error) {})

	api.State, err = db.StateGet(context.Background(), isTestnetBlockchain, api.Conn)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to init state object")
	}
	return api, r
}

// RecordUserActivity adds a UserActivity to the db for the current user
func (api *API) RecordUserActivity(
	ctx context.Context,
	userID string,
	action string,
	objectType passport.ObjectType,
	objectID *string,
	objectSlug *string,
	objectName *string,
	changes ...*passport.UserActivityChangeData,
) {
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		api.Log.Err(err).Msgf("issue creating uuid from %s", userID)
	}

	oldData, newData, err := passport.UserActivityGetDataChanges(changes)
	if err != nil {
		api.Log.Err(err).Msg("issue getting oldData and newData JSON")
	}

	err = db.UserActivityCreate(
		ctx,
		api.Conn,
		passport.UserID(userUUID),
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
