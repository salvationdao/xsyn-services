package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"net/url"
	"passport"
	"passport/api"
	"passport/db"
	"passport/email"
	"passport/log_helpers"
	"passport/seed"
	"time"

	_ "github.com/lib/pq" //postgres drivers for initialization

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"

	"github.com/rs/zerolog"

	"context"
	"fmt"
	"os"

	"github.com/oklog/run"
	"github.com/urfave/cli/v2"
)

// Version build Version
const Version = "v0.1.0"

// passed in using
//   ```sh
//   go build -ldflags "-X main.SentryVersion=" main.go
//   ```
var SentryVersion string

const envPrefix = "PASSPORT"

func main() {
	app := &cli.App{
		Compiled: time.Now(),
		Usage:    "Run the passport server or database administration commands",
		Authors: []*cli.Author{
			{
				Name:  "Ninja Software",
				Email: "hello@ninjasoftware.com.au",
			},
		},
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "database_tx_user", Value: "passport_tx", EnvVars: []string{"PASSPORT_DATABASE_TX_USER", "DATABASE_TX_USER"}, Usage: "The database transaction user"},
					&cli.StringFlag{Name: "database_tx_pass", Value: "dev-tx", EnvVars: []string{"PASSPORT_DATABASE_TX_PASS", "DATABASE_TX_PASS"}, Usage: "The database transaction pass"},

					&cli.StringFlag{Name: "database_user", Value: "passport", EnvVars: []string{envPrefix + "_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{envPrefix + "_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{envPrefix + "_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5432", EnvVars: []string{envPrefix + "_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "passport", EnvVars: []string{envPrefix + "_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{envPrefix + "_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},

					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{envPrefix + "_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
					&cli.StringFlag{Name: "sentry_dsn_backend", Value: "", EnvVars: []string{envPrefix + "_SENTRY_DSN_BACKEND", "SENTRY_DSN_BACKEND"}, Usage: "Sends error to remote server. If set, it will send error."},
					&cli.StringFlag{Name: "sentry_server_name", Value: "dev-pc", EnvVars: []string{envPrefix + "_SENTRY_SERVER_NAME", "SENTRY_SERVER_NAME"}, Usage: "The machine name that this program is running on."},
					&cli.Float64Flag{Name: "sentry_sample_rate", Value: 1, EnvVars: []string{envPrefix + "_SENTRY_SAMPLE_RATE", "SENTRY_SAMPLE_RATE"}, Usage: "The percentage of trace sample to collect (0.0-1)"},

					&cli.StringFlag{Name: "passport_web_host_url", Value: "http://localhost:5003", EnvVars: []string{envPrefix + "_HOST_URL_FRONTEND"}, Usage: "The Public Site URL used for CORS and links (eg: in the mailer)"},
					&cli.StringFlag{Name: "gameserver_web_host_url", Value: "http://localhost:8084", EnvVars: []string{"GAMESERVER_HOST_URL"}, Usage: "The host for the gameserver, to allow it to connect"},

					&cli.StringFlag{Name: "api_addr", Value: ":8086", EnvVars: []string{envPrefix + "_API_ADDR", "API_ADDR"}, Usage: "host:port to run the API"},
					&cli.BoolFlag{Name: "cookie_secure", Value: true, EnvVars: []string{envPrefix + "_COOKIE_SECURE", "COOKIE_SECURE"}, Usage: "set cookie secure"},
					&cli.StringFlag{Name: "google_client_id", Value: "593683501366-gk7ab1nnskc1tft14bk8ebsja1bce24v.apps.googleusercontent.com", EnvVars: []string{envPrefix + "_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"}, Usage: "Google Client ID for OAuth functionaility."},

					&cli.StringFlag{Name: "mail_domain", Value: "njs.dev", EnvVars: []string{envPrefix + "_MAIL_DOMAIN", "MAIL_DOMAIN"}, Usage: "Domain used for MailGun"},
					&cli.StringFlag{Name: "mail_apikey", Value: "", EnvVars: []string{envPrefix + "_MAIL_APIKEY", "MAIL_APIKEY"}, Usage: "MailGun API key"},
					&cli.StringFlag{Name: "mail_sender", Value: "Ninja Software <noreply@njs.dev>", EnvVars: []string{envPrefix + "_MAIL_SENDER", "MAIL_SENDER"}, Usage: "Default address emails are sent from"},

					&cli.BoolFlag{Name: "jwt_encrypt", Value: true, EnvVars: []string{envPrefix + "_JWT_ENCRYPT", "JWT_ENCRYPT"}, Usage: "set if to encrypt jwt tokens or not"},
					&cli.StringFlag{Name: "jwt_encrypt_key", Value: "ITF1vauAxvJlF0PLNY9btOO9ZzbUmc6X", EnvVars: []string{envPrefix + "_JWT_KEY", "JWT_KEY"}, Usage: "supports key sizes of 16, 24 or 32 bytes"},
					&cli.IntFlag{Name: "jwt_expiry_days", Value: 1, EnvVars: []string{envPrefix + "_JWT_EXPIRY_DAYS", "JWT_EXPIRY_DAYS"}, Usage: "expiry days for auth tokens"},
					&cli.StringFlag{Name: "metamask_sign_message", Value: "", EnvVars: []string{envPrefix + "_METAMASK_SIGN_MESSAGE", "METAMASK_SIGN_MESSAGE"}, Usage: "message to show in metamask key sign flow, needs to match frontend"},

					&cli.StringFlag{Name: "twitch_extension_secret", Value: "", EnvVars: []string{envPrefix + "_TWITCH_EXTENSION_SECRET", "TWITCH_EXTENSION_SECRET"}, Usage: "Twitch extension secret for verifying shadow account tokens sent with requests"},
					&cli.StringFlag{Name: "twitch_client_id", Value: "", EnvVars: []string{envPrefix + "_TWITCH_CLIENT_ID", "TWITCH_CLIENT_ID"}, Usage: "Twitch client ID for verifying web account tokens sent with requests"},
					&cli.StringFlag{Name: "twitch_client_secret", Value: "", EnvVars: []string{envPrefix + "_TWITCH_CLIENT_SECRET", "TWITCH_CLIENT_SECRET"}, Usage: "Twitch client secret for verifying web account tokens sent with requests"},
				},

				Usage: "run server",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					environment := c.String("environment")
					log := log_helpers.LoggerInitZero(environment)
					log.Info().Msg("zerolog initialised")
					log.Level(zerolog.DebugLevel)
					g := &run.Group{}
					// Listen for os.interrupt
					g.Add(run.SignalHandler(ctx, os.Interrupt))
					// start the server
					g.Add(func() error { return ServeFunc(c, ctx, log) }, func(err error) { cancel() })

					err := g.Run()
					if errors.Is(err, run.SignalError{Signal: os.Interrupt}) {
						err = terror.Warn(err)
					}
					log_helpers.TerrorEcho(sentry.CurrentHub(), err, log)
					return nil
				},
			},
			{
				Name: "db",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "database_tx_user", Value: "passport_tx", EnvVars: []string{"PASSPORT_DATABASE_TX_USER", "DATABASE_TX_USER"}, Usage: "The database transaction user"},
					&cli.StringFlag{Name: "database_tx_pass", Value: "dev-tx", EnvVars: []string{"PASSPORT_DATABASE_TX_PASS", "DATABASE_TX_PASS"}, Usage: "The database transaction pass"},

					&cli.StringFlag{Name: "database_user", Value: "passport", EnvVars: []string{"PASSPORT_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{"PASSPORT_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{"PASSPORT_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5432", EnvVars: []string{"PASSPORT_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "passport", EnvVars: []string{"PASSPORT_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{"PASSPORT_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},
					&cli.BoolFlag{Name: "database_prod", Value: false, EnvVars: []string{"PASSPORT_DB_PROD", "DB_PROD"}, Usage: "seed the database (prod)"},
					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{"PASSPORT_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},

					&cli.BoolFlag{Name: "seed", EnvVars: []string{"PASSPORT_DB_SEED", "DB_SEED"}, Usage: "seed the database"},
				},
				Usage: "seed the database",
				Action: func(c *cli.Context) error {
					databaseUser := c.String("database_user")
					databasePass := c.String("database_pass")
					databaseTxUser := c.String("database_tx_user")
					databaseTxPass := c.String("database_tx_pass")
					databaseHost := c.String("database_host")
					databasePort := c.String("database_port")
					databaseName := c.String("database_name")
					databaseAppName := c.String("database_application_name")
					databaseProd := c.Bool("database_prod")

					pgxconn, err := pgxconnect(
						databaseUser,
						databasePass,
						databaseHost,
						databasePort,
						databaseName,
						databaseAppName,
						Version,
					)
					if err != nil {
						return terror.Error(err)
					}

					txConn, err := txConnect(
						databaseTxUser,
						databaseTxPass,
						databaseHost,
						databasePort,
						databaseName,
					)
					if err != nil {
						return terror.Panic(err)
					}

					seeder := seed.NewSeeder(pgxconn, txConn)
					return seeder.Run(databaseProd)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		terror.Echo(err)
	}
}

func pgxconnect(
	DatabaseUser string,
	DatabasePass string,
	DatabaseHost string,
	DatabasePort string,
	DatabaseName string,
	DatabaseApplicationName string,
	APIVersion string,
) (*pgxpool.Pool, error) {
	params := url.Values{}
	params.Add("sslmode", "disable")
	if DatabaseApplicationName != "" {
		params.Add("application_name", fmt.Sprintf("%s %s", DatabaseApplicationName, APIVersion))
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		DatabaseUser,
		DatabasePass,
		DatabaseHost,
		DatabasePort,
		DatabaseName,
		params.Encode(),
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, terror.Panic(err, "could not initialise database")
	}
	poolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace
	poolConfig.MaxConns = 100

	ctx := context.Background()
	conn, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, terror.Panic(err, "could not initialise database")
	}

	return conn, nil
}

func txConnect(
	databaseTxUser string,
	databaseTxPass string,
	databaseHost string,
	databasePort string,
	databaseName string,
) (*sql.DB, error) {
	params := url.Values{}
	params.Add("sslmode", "disable")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		databaseTxUser,
		databaseTxPass,
		databaseHost,
		databasePort,
		databaseName,
		params.Encode(),
	)

	fmt.Println(connString)

	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, terror.Error(err)
	}

	return conn, nil
}

func ServeFunc(ctxCLI *cli.Context, ctx context.Context, log *zerolog.Logger) error {
	environment := ctxCLI.String("environment")
	sentryDSNBackend := ctxCLI.String("sentry_dsn_backend")
	sentryServerName := ctxCLI.String("sentry_server_name")
	sentryTraceRate := ctxCLI.Float64("sentry_sample_rate")
	err := log_helpers.SentryInit(sentryDSNBackend, sentryServerName, SentryVersion, environment, sentryTraceRate, log)
	switch errors.Unwrap(err) {
	case log_helpers.ErrSentryInitEnvironment:
		return terror.Error(err, fmt.Sprintf("got environment %s", environment))
	case log_helpers.ErrSentryInitDSN, log_helpers.ErrSentryInitVersion:
		if terror.GetLevel(err) == terror.ErrLevelPanic {
			// if the level is panic then in a prod environment
			// so keep panicing
			return terror.Panic(err)
		}
	default:
		if err != nil {
			return terror.Error(err)
		}
	}

	apiAddr := ctxCLI.String("api_addr")
	databaseUser := ctxCLI.String("database_user")
	databasePass := ctxCLI.String("database_pass")
	databaseTxUser := ctxCLI.String("database_tx_user")
	databaseTxPass := ctxCLI.String("database_tx_pass")
	databaseHost := ctxCLI.String("database_host")
	databasePort := ctxCLI.String("database_port")
	databaseName := ctxCLI.String("database_name")
	databaseAppName := ctxCLI.String("database_application_name")

	mailDomain := ctxCLI.String("mail_domain")
	mailAPIKey := ctxCLI.String("mail_apikey")
	mailSender := ctxCLI.String("mail_sender")

	config := &passport.Config{
		CookieSecure:        ctxCLI.Bool("cookie_secure"),
		PassportWebHostURL:  ctxCLI.String("passport_web_host_url"),
		GameserverHostURL:   ctxCLI.String("gameserver_web_host_url"),
		EncryptTokens:       ctxCLI.Bool("jwt_encrypt"),
		EncryptTokensKey:    ctxCLI.String("jwt_encrypt_key"),
		TokenExpirationDays: ctxCLI.Int("jwt_expiry_days"),
		MetaMaskSignMessage: ctxCLI.String("metamask_sign_message"),
	}

	pgxconn, err := pgxconnect(
		databaseUser,
		databasePass,
		databaseHost,
		databasePort,
		databaseName,
		databaseAppName,
		Version,
	)
	if err != nil {
		return terror.Panic(err)
	}

	txConn, err := txConnect(
		databaseTxUser,
		databaseTxPass,
		databaseHost,
		databasePort,
		databaseName,
	)
	if err != nil {
		return terror.Panic(err)
	}

	count := 0
	err = db.IsSchemaDirty(ctx, pgxconn, &count)
	if err != nil {
		return terror.Error(api.ErrCheckDBQuery)
	}
	if count > 0 {
		return terror.Error(api.ErrCheckDBDirty)
	}

	twitchExtensionSecretBase64 := ctxCLI.String("twitch_extension_secret")
	if twitchExtensionSecretBase64 == "" {
		return terror.Panic(fmt.Errorf("missing twitch extension secret"))
	}

	twitchExtensionSecret, err := base64.StdEncoding.DecodeString(twitchExtensionSecretBase64)
	if err != nil {
		return terror.Panic(err, "Failed to decode twitch extension secret")
	}

	twitchClientID := ctxCLI.String("twitch_client_id")
	if twitchClientID == "" {
		return terror.Panic(fmt.Errorf("no twitch client id"))
	}

	twitchClientSecret := ctxCLI.String("twitch_client_secret")
	if twitchClientSecret == "" {
		return terror.Panic(fmt.Errorf("no twitch client secret"))
	}

	// Mailer
	mailer, err := email.NewMailer(mailDomain, mailAPIKey, mailSender, config, log)
	if err != nil {
		return terror.Panic(err, "Mailer init failed")
	}

	// HTML Sanitizer
	HTMLSanitizePolicy := bluemonday.UGCPolicy()
	HTMLSanitizePolicy.AllowAttrs("class").OnElements("img", "table", "tr", "td", "p")

	// API Server
	ctx, cancelOnPanic := context.WithCancel(ctx)
	api := api.NewAPI(log, cancelOnPanic, pgxconn, txConn, ctxCLI.String("google_client_id"), mailer, apiAddr, twitchExtensionSecret, twitchClientID, twitchClientSecret, HTMLSanitizePolicy, config)
	return api.Run(ctx)
}
