package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"passport/email"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/log_helpers"

	SentryTracer "github.com/ninja-syndicate/hub/ext/sentry"

	"github.com/shopspring/decimal"

	"nhooyr.io/websocket"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
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
	Routes              chi.Router
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
) *API {

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
		Routes:       chi.NewRouter(),
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

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(middleware.Logger)
	api.Routes.Use(cors.New(cors.Options{
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

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
			r.Mount("/check", CheckRouter(log_helpers.NamedLogger(log, "check router"), conn))
			r.Mount("/files", FileRouter(conn, api))
			r.Get("/verify", api.WithError(api.Auth.VerifyAccountHandler))
			r.Get("/get-nonce", api.WithError(api.Auth.GetNonce))
			r.Get("/withdraw/{address}/{nonce}/{amount}", api.WithError(api.WithdrawSups))
			r.Get("/mint-nft/{address}/{nonce}/{collectionSlug}/{externalTokenID}", api.WithError(api.MintAsset))
			r.Get("/asset/{hash}", api.WithError(api.AssetGet))
			r.Get("/asset/{collection_address}/{token_id}", api.WithError(api.AssetGetByCollectionAndTokenID))
			r.Get("/auth/twitter", api.WithError(api.Auth.TwitterAuth))

			r.Get("/whitelist/check", api.WithError(api.WhitelistOnlyWalletCheck))
			r.Get("/faction-data", api.WithError(api.FactionGetData))
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
	return api
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

// AssetGet grabs asset's metadata via token id
func (api *API) AssetGet(w http.ResponseWriter, r *http.Request) (int, error) {
	// Get token id
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid asset hash"), "Invalid Asset Hash.")
	}

	// Get asset via token id
	asset, err := db.AssetGet(context.Background(), api.Conn, hash)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset")
	}
	if asset == nil {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("asset is nil"), "Asset doesn't exist")
	}

	// openseas object
	//asset

	// Encode result
	err = json.NewEncoder(w).Encode(asset)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to encode JSON")
	}

	return http.StatusOK, nil
}

// AssetGet grabs asset's metadata via token id
func (api *API) AssetGetByCollectionAndTokenID(w http.ResponseWriter, r *http.Request) (int, error) {
	collectionAddress := chi.URLParam(r, "collection_address")
	if collectionAddress == "" {
		return http.StatusBadRequest, terror.Warn(errors.New("collection_address not provided in URL"), "metadata")
	}
	tokenIDStr := chi.URLParam(r, "token_id")
	if tokenIDStr == "" {
		return http.StatusBadRequest, terror.Warn(errors.New("token_id not provided in URL"), "metadata")
	}
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		return http.StatusBadRequest, terror.Warn(err, "get asset from db")
	}
	asset, err := db.AssetGetFromMintContractAndID(context.Background(), api.Conn, string(common.HexToAddress(collectionAddress).Hex()), uint64(tokenID))
	if err != nil {
		return http.StatusBadRequest, terror.Warn(err, "get asset from db")
	}
	// Get asset via token id
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset")
	}

	// Encode result
	err = json.NewEncoder(w).Encode(asset)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to encode JSON")
	}

	return http.StatusOK, nil
}
