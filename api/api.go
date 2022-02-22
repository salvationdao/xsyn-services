package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/db"
	"passport/email"
	"strconv"
	"time"

	"github.com/ninja-software/log_helpers"

	SentryTracer "github.com/ninja-syndicate/hub/ext/sentry"

	"github.com/shopspring/decimal"

	"nhooyr.io/websocket"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
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
	State        *passport.State
	SupUSD       decimal.Decimal
	Log          *zerolog.Logger
	Routes       chi.Router
	Addr         string
	Mailer       *email.Mailer
	HTMLSanitize *bluemonday.Policy
	Hub          *hub.Hub
	Conn         *pgxpool.Pool
	Tokens       *Tokens
	*auth.Auth
	*messagebus.MessageBus
	ClientToken  string
	BridgeParams *passport.BridgeParams

	// online user cache
	users chan func(userCacheList UserCacheMap)

	// server clients
	serverClients       chan func(serverClients ServerClientsList)
	sendToServerClients chan *ServerClientMessage

	//tx stuff
	transaction      chan *passport.NewTransaction
	heldTransactions chan func(heldTxList map[passport.TransactionReference]*passport.NewTransaction)

	// treasury ticker map
	treasuryTickerMap map[ServerClientName]*tickle.Tickle

	// Supremacy Sups Pool
	supremacySupsPool chan func(*SupremacySupPool)

	// War Machine Queue Contract
	factionWarMachineContractMap map[passport.FactionID]chan func(*WarMachineContract)
	fastAssetRepairCenter        chan func(RepairQueue)
	standardAssetRepairCenter    chan func(RepairQueue)

	// Queue Reward
	TxConn *sql.DB

	walletOnlyConnect    bool
	storeItemExternalUrl string
}

// NewAPI registers routes
func NewAPI(
	log *zerolog.Logger,
	cancelOnPanic context.CancelFunc,
	conn *pgxpool.Pool,
	txConn *sql.DB,
	googleClientID string,
	mailer *email.Mailer,
	addr string,
	twitchClientID string,
	twitchClientSecret string,
	HTMLSanitize *bluemonday.Policy,
	config *passport.Config,
	twitterAPIKey string,
	twitterAPISecret string,
	discordClientID string,
	discordClientSecret string,
	clientToken string,
	externalUrl string,
	isTestnetBlockchain bool,
	runBlockchainBridge bool,
) *API {
	msgBus, cleanUpFunc := messagebus.NewMessageBus(log_helpers.NamedLogger(log, "message bus"))
	api := &API{
		SupUSD:       decimal.New(12, -2),
		BridgeParams: config.BridgeParams,
		ClientToken:  clientToken,
		Tokens: &Tokens{
			Conn:                conn,
			Mailer:              mailer,
			tokenExpirationDays: config.TokenExpirationDays,
			encryptToken:        config.EncryptTokens,
			encryptTokenKey:     config.EncryptTokensKey,
		},
		MessageBus: msgBus,
		Hub: hub.New(&hub.Config{
			ClientOfflineFn: cleanUpFunc,
			Log:             zerologger.New(*log_helpers.NamedLogger(log, "hub library")),
			Tracer:          SentryTracer.New(),
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
		serverClients:       make(chan func(serverClients ServerClientsList)),
		sendToServerClients: make(chan *ServerClientMessage),
		//382
		// user cache map
		users: make(chan func(userList UserCacheMap)),

		// object to hold transaction stuff
		TxConn:           txConn,
		transaction:      make(chan *passport.NewTransaction),
		heldTransactions: make(chan func(heldTxList map[passport.TransactionReference]*passport.NewTransaction)),

		// treasury ticker map
		treasuryTickerMap: make(map[ServerClientName]*tickle.Tickle),
		supremacySupsPool: make(chan func(*SupremacySupPool)),

		// faction war machine contract
		factionWarMachineContractMap: make(map[passport.FactionID]chan func(*WarMachineContract)),

		walletOnlyConnect:    config.OnlyWalletConnect,
		storeItemExternalUrl: externalUrl,
	}

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	}).Handler)

	var err error
	api.Auth, err = auth.New(api.Hub, &auth.Config{
		CreateUserIfNotExist:     true,
		CreateAndGetOAuthUserVia: auth.IdTypeID,
		Google: &auth.GoogleConfig{
			ClientID: googleClientID,
		},
		Twitch: &auth.TwitchConfig{
			ClientID:     twitchClientID,
			ClientSecret: twitchClientSecret,
		},
		Twitter: &auth.TwitterConfig{
			APIKey:    twitterAPIKey,
			APISecret: twitterAPISecret,
		},
		Discord: &auth.DiscordConfig{
			ClientID:     discordClientID,
			ClientSecret: discordClientSecret,
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

	cc := &ChainClients{}
	if runBlockchainBridge {
		// Runs the listeners for all the chain bridges
		cc = RunChainListeners(log, api, config.BridgeParams, isTestnetBlockchain)

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
			r.Get("/mint-nft/{address}/{nonce}/{tokenID}", api.WithError(api.MintAsset))
			r.Get("/asset/{token_id}", api.WithError(api.AssetGet))
			r.Get("/auth/twitter", api.WithError(api.Auth.TwitterAuth))
			r.Get("/dummy-sale", api.WithError(api.Dummysale))
			if runBlockchainBridge {
				r.Get("/check-eth-tx/{tx_id}", api.WithError(cc.CheckEthTx))
				r.Get("/check-bsc-tx/{tx_id}", api.WithError(cc.CheckBscTx))
			}
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
	_ = NewCollectionController(log, conn, api)
	_ = NewServerClientController(log, conn, api)
	_ = NewCheckController(log, conn, api)
	_ = NewUserActivityController(log, conn, api)
	_ = NewUserController(log, conn, api, &auth.GoogleConfig{
		ClientID: googleClientID,
	}, &auth.TwitchConfig{
		ClientID:     twitchClientID,
		ClientSecret: twitchClientSecret,
	}, &auth.DiscordConfig{
		ClientID:     discordClientID,
		ClientSecret: discordClientSecret,
	})

	_ = NewFactionController(log, conn, api)
	_ = NewOrganisationController(log, conn, api)
	_ = NewRoleController(log, conn, api)
	_ = NewProductController(log, conn, api)
	_ = NewSupremacyController(log, conn, api)
	_ = NewGamebarController(log, conn, api)
	_ = NewStoreController(log, conn, api)

	//api.Hub.Events.AddEventHandler(hub.EventOnline, api.ClientOnline)
	api.Hub.Events.AddEventHandler(auth.EventLogin, api.ClientAuth)
	api.Hub.Events.AddEventHandler(auth.EventLogout, api.ClientLogout)
	api.Hub.Events.AddEventHandler(hub.EventOffline, api.ClientOffline)

	ctx := context.TODO()
	api.State, err = db.StateGet(ctx, isTestnetBlockchain, api.Conn)
	if err != nil {
		log.Fatal().Msgf("failed to init state object")
	}

	// Run the server client channel listener
	go api.HandleServerClients()

	// Run the transaction channel listeners
	go api.HandleTransactions()
	go api.HandleHeldTransactions()

	// Run the listener for the db user update event
	go api.DBListenForUserUpdateEvent()

	// Run the listener for the user cache
	go api.HandleUserCache()

	// Initialise treasury fund ticker
	go api.InitialiseTreasuryFundTicker()

	// Initial supremacy sup pool
	go api.StartSupremacySupPool()

	// Initial faction war machine contract
	api.factionWarMachineContractMap[passport.RedMountainFactionID] = make(chan func(*WarMachineContract))
	go api.InitialiseFactionWarMachineContract(passport.RedMountainFactionID)
	api.factionWarMachineContractMap[passport.BostonCyberneticsFactionID] = make(chan func(*WarMachineContract))
	go api.InitialiseFactionWarMachineContract(passport.BostonCyberneticsFactionID)
	api.factionWarMachineContractMap[passport.ZaibatsuFactionID] = make(chan func(*WarMachineContract))
	go api.InitialiseFactionWarMachineContract(passport.ZaibatsuFactionID)

	// Initialise repair center
	go api.InitialAssetRepairCenter()

	return api
}

//test function for remaining supply
func (api *API) Dummysale(w http.ResponseWriter, r *http.Request) (int, error) {
	// get amount from get url
	ctx := context.Background()
	amount := r.URL.Query().Get("amount")

	bigIntAmount := big.Int{}
	bigIntAmount.SetString(amount, 10)

	tx := &passport.NewTransaction{
		From:                 passport.XsynSaleUserID,
		To:                   passport.SupremacyGameUserID,
		TransactionReference: "test sale",
		Amount:               bigIntAmount,
	}

	select {
	case api.transaction <- tx:

	case <-time.After(10 * time.Second):
		api.Log.Err(errors.New("timeout on channel send exceeded"))
		panic("transaction send")
	}

	sups, err := db.UserBalance(ctx, api.Conn, passport.XsynSaleUserID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeySUPSRemainingSubscribe), sups.String())

	return http.StatusAccepted, nil
}

// Run the API service
func (api *API) Run(ctx context.Context) error {
	api.Log.Info().Msg("Starting API")

	server := &http.Server{
		Addr:    api.Addr,
		Handler: api.Routes,
	}

	go func() {
		<-ctx.Done()
		api.Log.Info().Msg("Stopping API")
		err := server.Shutdown(ctx)
		if err != nil {
			api.Log.Warn().Err(err).Msg("")
		}
	}()

	return server.ListenAndServe()
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
	tokenID := chi.URLParam(r, "token_id")
	if tokenID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid token id"), "Invalid Token ID")
	}

	// Convert token id from string to uint64
	_tokenID, err := strconv.ParseUint(string(tokenID), 10, 64)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed converting string token id to uint64")
	}

	// Get asset via token id
	asset, err := db.AssetGet(r.Context(), api.Conn, _tokenID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset")
	}
	if asset == nil {
		return http.StatusBadRequest, terror.Warn(err, "Asset doesn't exist")
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
