package api

import (
	"context"
	"net/http"
	"passport"
	"passport/db"
	"passport/email"
	"passport/log_helpers"

	"nhooyr.io/websocket"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/hub/v2/ext/auth"
	zerologger "github.com/ninja-software/hub/v2/ext/zerolog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// API server
type API struct {
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

	// server clients
	serverClients       chan func(serverClients ServerClientsList)
	sendToServerClients chan *ServerClientMessage

	// transaction channel
	transaction chan *Transaction
}

// NewAPI registers routes
func NewAPI(
	log *zerolog.Logger,
	cancelOnPanic context.CancelFunc,
	conn *pgxpool.Pool,
	googleClientID string,
	mailer *email.Mailer,
	addr string,
	twitchExtensionSecret []byte,
	twitchClientID string,
	twitchClientSecret string,
	HTMLSanitize *bluemonday.Policy,
	config *passport.Config,
) *API {
	msgBus, cleanUpFunc := messagebus.NewMessageBus(log_helpers.NamedLogger(log, "message bus"))
	api := &API{
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
			WelcomeMsg: &hub.WelcomeMsg{
				Key:     "WELCOME",
				Payload: nil,
			},
			AcceptOptions: &websocket.AcceptOptions{
				InsecureSkipVerify: true, // TODO: set this depending on environment
				OriginPatterns:     []string{config.PassportWebHostURL, config.GameserverHostURL},
			},
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
		// tx channel
		transaction: make(chan *Transaction),
	}

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(cors.New(cors.Options{
		AllowedOrigins: []string{config.PassportWebHostURL, config.GameserverHostURL},
	}).Handler)
	var err error
	api.Auth, err = auth.New(api.Hub, &auth.Config{
		CreateUserIfNotExist:     true,
		CreateAndGetOAuthUserVia: auth.IdTypeID,
		Google: &auth.GoogleConfig{
			ClientID: googleClientID,
		},
		TwitchExtensionSecret: twitchExtensionSecret,
		CookieSecure:          config.CookieSecure,
		UserController: &UserGetter{
			Log:    log_helpers.NamedLogger(log, "user getter"),
			Conn:   conn,
			Mailer: mailer,
		},
		Tokens:        api.Tokens,
		Eip712Message: config.MetaMaskSignMessage,
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
			r.Get("/verify", WithError(api.Auth.VerifyAccountHandler))
			r.Get("/get-nonce", WithError(api.Auth.GetNonce))
		})
		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Handle("/ws", api.Hub)
	})

	_ = NewAssetController(log, conn, api)
	_ = NewServerClientController(log, conn, api)
	_ = NewCheckController(log, conn, api)
	_ = NewUserActivityController(log, conn, api)
	_ = NewUserController(log, conn, api, &auth.GoogleConfig{
		ClientID: googleClientID,
	}, twitchClientID, twitchClientSecret)
	_ = NewAuthController(log, conn, api)
	_ = NewFactionController(log, conn, api)
	_ = NewOrganisationController(log, conn, api)
	_ = NewRoleController(log, conn, api)
	_ = NewProductController(log, conn, api)
	_ = NewSupremacyController(log, conn, api)

	//api.Hub.Events.AddEventHandler(hub.EventOnline, api.ClientOnline)
	api.Hub.Events.AddEventHandler(auth.EventLogin, api.ClientAuth)
	api.Hub.Events.AddEventHandler(auth.EventLogout, api.ClientLogout)
	api.Hub.Events.AddEventHandler(hub.EventOffline, api.ClientOffline)

	// Run the server client channel listener
	go api.HandleServerClients()

	// Run the transaction channel listener
	go api.HandleTransactions()

	return api
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
