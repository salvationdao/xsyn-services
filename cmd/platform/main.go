package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os/signal"
	"passport"
	"passport/api"
	"passport/comms"
	"passport/db"
	"passport/email"
	"passport/helpers"
	"passport/passdb"
	"passport/passlog"
	"passport/payments"
	"passport/rpcclient"
	"passport/seed"
	"passport/sms"
	"runtime"
	"strings"
	"time"

	_ "net/http/pprof"
	rpprof "runtime/pprof"

	"github.com/jackc/pgx/v4/stdlib"
	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/oklog/run"

	_ "github.com/lib/pq" //postgres drivers for initialization

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"

	"github.com/rs/zerolog"

	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Variable passed in at compile time using `-ldflags`
var (
	Version          string // -X main.Version=$(git describe --tags --abbrev=0)
	GitHash          string // -X main.GitHash=$(git rev-parse HEAD)
	GitBranch        string // -X main.GitBranch=$(git rev-parse --abbrev-ref HEAD)
	BuildDate        string // -X main.BuildDate=$(date -u +%Y%m%d%H%M%S)
	UnCommittedFiles string // -X main.UnCommittedFiles=$(git status --porcelain | wc -l)"
)

const SentryReleasePrefix = "ninja_syndicate-passport_api"
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
				// This is not using the built in version so ansible can more easily read the version
				Name: "version",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "full", Usage: "Prints full version and build info", Value: false},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("full") {
						fmt.Printf("Version=%s\n", Version)
						fmt.Printf("Commit=%s\n", GitHash)
						fmt.Printf("Branch=%s\n", GitBranch)
						fmt.Printf("BuildDate=%s\n", BuildDate)
						fmt.Printf("WorkingCopyState=%s uncommitted\n", UnCommittedFiles)
						return nil
					}
					fmt.Printf("%s-\n", Version)
					return nil
				},
			},
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

					&cli.BoolFlag{Name: "is_testnet_blockchain", Value: false, EnvVars: []string{envPrefix + "_IS_TESTNET_BLOCKCHAIN"}, Usage: "Update state according to testnet"},
					&cli.BoolFlag{Name: "run_blockchain_bridge", Value: true, EnvVars: []string{envPrefix + "_RUN_BLOCKCHAIN_BRIDGE"}, Usage: "Run the bridge to blockchain data"},

					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{envPrefix + "_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
					&cli.StringFlag{Name: "sentry_dsn_backend", Value: "", EnvVars: []string{envPrefix + "_SENTRY_DSN_BACKEND", "SENTRY_DSN_BACKEND"}, Usage: "Sends error to remote server. If set, it will send error."},
					&cli.StringFlag{Name: "sentry_server_name", Value: "dev-pc", EnvVars: []string{envPrefix + "_SENTRY_SERVER_NAME", "SENTRY_SERVER_NAME"}, Usage: "The machine name that this program is running on."},
					&cli.Float64Flag{Name: "sentry_sample_rate", Value: 1, EnvVars: []string{envPrefix + "_SENTRY_SAMPLE_RATE", "SENTRY_SAMPLE_RATE"}, Usage: "The percentage of trace sample to collect (0.0-1)"},
					&cli.StringFlag{Name: "log_level", Value: "TraceLevel", EnvVars: []string{envPrefix + "_LOG_LEVEL"}, Usage: "Set the log level for zerolog (Options: PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel"},

					&cli.StringFlag{Name: "passport_web_host_url", Value: "http://localhost:5003", EnvVars: []string{envPrefix + "_HOST_URL_FRONTEND"}, Usage: "The Public Site URL used for CORS and links (eg: in the mailer)"},
					&cli.StringFlag{Name: "gameserver_web_host_url", Value: "http://localhost:8084", EnvVars: []string{"GAMESERVER_HOST_URL"}, Usage: "The host for the gameserver, to allow it to connect"},

					&cli.StringFlag{Name: "api_addr", Value: ":8086", EnvVars: []string{envPrefix + "_API_ADDR", "API_ADDR"}, Usage: "host:port to run the API"},
					&cli.BoolFlag{Name: "cookie_secure", Value: true, EnvVars: []string{envPrefix + "_COOKIE_SECURE", "COOKIE_SECURE"}, Usage: "set cookie secure"},
					&cli.StringFlag{Name: "google_client_id", Value: "467953368642-8cobg822tej2i50ncfg4ge1pm4c5v033.apps.googleusercontent.com", EnvVars: []string{envPrefix + "_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"}, Usage: "Google Client ID for OAuth functionaility."},

					// SMS stuff
					&cli.StringFlag{Name: "twilio_sid", Value: "", EnvVars: []string{envPrefix + "_TWILIO_ACCOUNT_SID"}, Usage: "Twilio account sid"},
					&cli.StringFlag{Name: "twilio_api_key", Value: "", EnvVars: []string{envPrefix + "_TWILIO_API_KEY"}, Usage: "Twilio api key"},
					&cli.StringFlag{Name: "twilio_api_secret", Value: "", EnvVars: []string{envPrefix + "_TWILIO_API_SECRET"}, Usage: "Twilio api secret"},
					&cli.StringFlag{Name: "sms_from_number", Value: "", EnvVars: []string{envPrefix + "_SMS_FROM_NUMBER"}, Usage: "Number to send SMS from"},

					&cli.StringFlag{Name: "mail_domain", Value: "njs.dev", EnvVars: []string{envPrefix + "_MAIL_DOMAIN", "MAIL_DOMAIN"}, Usage: "Domain used for MailGun"},
					&cli.StringFlag{Name: "mail_apikey", Value: "", EnvVars: []string{envPrefix + "_MAIL_APIKEY", "MAIL_APIKEY"}, Usage: "MailGun API key"},
					&cli.StringFlag{Name: "mail_sender", Value: "Ninja Software <noreply@njs.dev>", EnvVars: []string{envPrefix + "_MAIL_SENDER", "MAIL_SENDER"}, Usage: "Default address emails are sent from"},

					&cli.BoolFlag{Name: "jwt_encrypt", Value: true, EnvVars: []string{envPrefix + "_JWT_ENCRYPT", "JWT_ENCRYPT"}, Usage: "set if to encrypt jwt tokens or not"},
					&cli.StringFlag{Name: "jwt_encrypt_key", Value: "ITF1vauAxvJlF0PLNY9btOO9ZzbUmc6X", EnvVars: []string{envPrefix + "_JWT_KEY", "JWT_KEY"}, Usage: "supports key sizes of 16, 24 or 32 bytes"},
					&cli.IntFlag{Name: "jwt_expiry_days", Value: 1, EnvVars: []string{envPrefix + "_JWT_EXPIRY_DAYS", "JWT_EXPIRY_DAYS"}, Usage: "expiry days for auth tokens"},
					&cli.StringFlag{Name: "metamask_sign_message", Value: "", EnvVars: []string{envPrefix + "_METAMASK_SIGN_MESSAGE", "METAMASK_SIGN_MESSAGE"}, Usage: "message to show in metamask key sign flow, needs to match frontend"},

					&cli.StringFlag{Name: "twitch_client_id", Value: "123", EnvVars: []string{envPrefix + "_TWITCH_CLIENT_ID", "TWITCH_CLIENT_ID"}, Usage: "Twitch client ID for verifying web account tokens sent with requests"},
					&cli.StringFlag{Name: "twitch_client_secret", Value: "123", EnvVars: []string{envPrefix + "_TWITCH_CLIENT_SECRET", "TWITCH_CLIENT_SECRET"}, Usage: "Twitch client secret for verifying web account tokens sent with requests"},
					&cli.StringFlag{Name: "twitter_api_key", Value: "123", EnvVars: []string{envPrefix + "_TWITTER_API_KEY", "TWITTER_API_KEY"}, Usage: "Twitter API key for requests used in the OAuth 1.0a flow"},
					&cli.StringFlag{Name: "twitter_api_secret", Value: "123", EnvVars: []string{envPrefix + "_TWITTER_API_SECRET", "TWITTER_API_SECRET"}, Usage: "Twitter API key for requests used in the OAuth 1.0a flow"},
					&cli.StringFlag{Name: "discord_client_id", Value: "123", EnvVars: []string{envPrefix + "_DISCORD_CLIENT_ID", "DISCORD_CLIENT_ID"}, Usage: "Discord client ID for verifying web account tokens sent with requests"},
					&cli.StringFlag{Name: "discord_client_secret", Value: "123", EnvVars: []string{envPrefix + "_DISCORD_CLIENT_SECRET", "DISCORD_CLIENT_SECRET"}, Usage: "Discord client secret for verifying web account tokens sent with requests"},

					&cli.StringFlag{Name: "gameserver_token", Value: "aG93cyBpdCBnb2luZyBtYWM=", EnvVars: []string{envPrefix + "_GAMESERVER_TOKEN"}, Usage: "Token to auth gameserver client"},
					&cli.BoolFlag{Name: "only_wallet", Value: true, EnvVars: []string{envPrefix + "_ONLY_WALLET"}, Usage: "Set passport to only accept wallet logins"},
					&cli.StringFlag{Name: "whitelist_check_endpoint", Value: "https://stories.supremacy.game", EnvVars: []string{envPrefix + "_WHITELIST_ENDPOINT"}, Usage: "Endpoint to check if user is whitelisted"},

					&cli.BoolFlag{Name: "pprof", Value: true, EnvVars: []string{envPrefix + "_PPROF"}, Usage: "record pprof at regular interval to help debug"},
					&cli.IntFlag{Name: "pprof_second", Value: 10, EnvVars: []string{envPrefix + "_PPROF_SECOND"}, Usage: "record pprof at x second interval"},
					&cli.IntFlag{Name: "pprof_port", Value: 6060, EnvVars: []string{envPrefix + "_PPROF_PORT"}, Usage: "pprof local listening port"},

					// setup for webhook
					&cli.StringFlag{Name: "gameserver_webhook_secret", Value: "e1BD3FF270804c6a9edJDzzDks87a8a4fde15c7=", EnvVars: []string{"GAMESERVER_WEBHOOK_SECRET"}, Usage: "Authorization key to passport webhook"},
					&cli.StringFlag{Name: "gameserver_host_url", Value: "http://localhost:8084", EnvVars: []string{"GAMESERVER_HOST_URL"}, Usage: "Authorization key to passport webhook"},
					&cli.StringFlag{Name: "jwt_key", Value: "9a5b8421bbe14e5a904cfd150a9951d3", EnvVars: []string{"STREAM_SITE_JWT_KEY"}, Usage: "JWT Key for signing token on stream site"},

					/****************************
					 *		Bridge details		*
					 ***************************/
					// ETH
					&cli.StringFlag{Name: "usdc_addr", Value: "0x8BB4eC208CDDE7761ac7f3346deBb9C931f80A33", EnvVars: []string{envPrefix + "_USDC_CONTRACT_ADDR"}, Usage: "USDC contract address"},

					// BSC
					&cli.StringFlag{Name: "busd_addr", Value: "0xeAf33Ba4AcA3fE3110EAddD7D4cf0897121583D0", EnvVars: []string{envPrefix + "_BUSD_CONTRACT_ADDR"}, Usage: "BUSD contract address"},
					&cli.StringFlag{Name: "sup_addr", Value: "0x5e8b6999B44E011F485028bf1AF0aF601F845304", EnvVars: []string{envPrefix + "_SUP_CONTRACT_ADDR"}, Usage: "SUP contract address"},

					// wallet/contract addresses
					&cli.StringFlag{Name: "operator_addr", Value: "0xc01c2f6DD7cCd2B9F8DB9aa1Da9933edaBc5079E", EnvVars: []string{envPrefix + "_OPERATOR_WALLET_ADDR"}, Usage: "Wallet address for administration"},
					&cli.StringFlag{Name: "signer_private_key", Value: "0x5f3b57101caf01c3d91e50809e70d84fcc404dd108aa8a9aa3e1a6c482267f48", EnvVars: []string{envPrefix + "_SIGNER_PRIVATE_KEY"}, Usage: "Private key for signing (usually operator)"},
					&cli.StringFlag{Name: "purchase_addr", Value: "0x52b38626D3167e5357FE7348624352B7062fE271", EnvVars: []string{envPrefix + "_PURCHASE_WALLET_ADDR"}, Usage: "Wallet address to receive payments and deposits"},

					&cli.StringFlag{Name: "withdraw_addr", Value: "0x9DAcEA338E4DDd856B152Ce553C7540DF920Bb15", EnvVars: []string{envPrefix + "_WITHDRAW_CONTRACT_ADDR"}, Usage: "Withdraw contract address"},

					// chain id
					&cli.Int64Flag{Name: "bsc_chain_id", Value: 97, EnvVars: []string{envPrefix + "_BSC_CHAIN_ID"}, Usage: "BSC Chain ID"},
					&cli.Int64Flag{Name: "eth_chain_id", Value: 5, EnvVars: []string{envPrefix + "_ETH_CHAIN_ID"}, Usage: "ETH Chain ID"},

					// node address
					&cli.StringFlag{Name: "bsc_node_addr", Value: "wss://speedy-nodes-nyc.moralis.io/6bc5ccfe2d00f7a5ae0ba00a/bsc/testnet/ws", EnvVars: []string{envPrefix + "_BSC_WS_NODE_URL"}, Usage: "Binance WS node URL"},
					&cli.StringFlag{Name: "eth_node_addr", Value: "wss://speedy-nodes-nyc.moralis.io/6bc5ccfe2d00f7a5ae0ba00a/eth/goerli/ws", EnvVars: []string{envPrefix + "_ETH_WS_NODE_URL"}, Usage: "Ethereum WS node URL"},
					//router address for exchange rates
					&cli.StringFlag{Name: "bsc_router_addr", Value: "0x10ED43C718714eb63d5aA57B78B54704E256024E", EnvVars: []string{envPrefix + "_BSC_ROUTER_ADDR"}, Usage: "BSC Router address"},
					&cli.BoolFlag{Name: "enable_purchase_subscription", Value: false, EnvVars: []string{envPrefix + "_ENABLE_PURCHASE_SUBSCRIPTION"}, Usage: "Poll payments and price"},
					&cli.BoolFlag{Name: "avant_testnet", Value: false, EnvVars: []string{envPrefix + "_AVANT_TESTNET"}, Usage: "Use testnet for Avant data scraper"},
					&cli.BoolFlag{Name: "skip_update_users_mixed_case", Value: false, EnvVars: []string{envPrefix + "_SKIP_UPDATE_USERS_MIXED_CASE"}, Usage: "Set to true after users have been all updated as mixed case"},

					//moralis key- set in env vars
					//moralis key- set in env vars
					//moralis key- set in env vars
					&cli.IntFlag{Name: "database_max_pool_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_POOL_CONNS"}, Usage: "Database max pool conns"},
					&cli.IntFlag{Name: "database_max_idle_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_IDLE_CONNS"}, Usage: "Database max idle conns"},
					&cli.IntFlag{Name: "database_max_open_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_OPEN_CONNS"}, Usage: "Database max open conns"},
					&cli.StringFlag{Name: "moralis_key", Value: "91Xp2ke5eOVMavAsqdOoiXN4lg0n0AieW5kTJoupdyQBhL2k9XvMQtFPSA4opX2s", EnvVars: []string{envPrefix + "_MORALIS_KEY"}, Usage: "Key to connect to moralis API"},
				},

				Usage: "run server",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					environment := c.String("environment")
					level := c.String("log_level")
					log := log_helpers.LoggerInitZero(environment, level)
					if environment == "production" || environment == "staging" {
						logPtr := zerolog.New(os.Stdout)
						log = &logPtr
					}
					passlog.New(environment, level)
					log.Info().Msg("zerolog initialised")
					tracer.Start(
						tracer.WithEnv(environment),
						tracer.WithService(envPrefix),
						tracer.WithServiceVersion(Version),
					)
					defer tracer.Stop()

					if c.Bool("pprof") {
						pint := c.Int("pprof_second")
						pport := c.Int("pprof_port")
						// dumping pprof at period bases
						pprofMonitor(pint, pport)
					}

					g := &run.Group{}
					// Listen for os.interrupt
					g.Add(run.SignalHandler(ctx, os.Interrupt))
					// start the server
					g.Add(func() error { return ServeFunc(c, log) }, func(err error) { cancel() })

					err := g.Run()
					if errors.Is(err, run.SignalError{Signal: os.Interrupt}) {
						err = terror.Warn(err)
					}
					log_helpers.TerrorEcho(ctx, err, log)
					return nil
				},
			},
			{
				Name: "sync",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "database_user", Value: "passport", EnvVars: []string{"PASSPORT_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{"PASSPORT_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{"PASSPORT_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5432", EnvVars: []string{"PASSPORT_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "passport", EnvVars: []string{"PASSPORT_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{"PASSPORT_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},
					&cli.StringFlag{Name: "gameserver_web_host_url", Value: "http://localhost:8084", EnvVars: []string{"GAMESERVER_HOST_URL"}, Usage: "The host for the gameserver, to allow it to connect"},
					&cli.IntFlag{Name: "database_max_pool_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_POOL_CONNS"}, Usage: "Database max pool conns"},
					&cli.IntFlag{Name: "database_max_idle_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_IDLE_CONNS"}, Usage: "Database max idle conns"},
					&cli.IntFlag{Name: "database_max_open_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_OPEN_CONNS"}, Usage: "Database max open conns"},
				},
				Usage: "sync items over from supremacy-gameserver",
				Action: func(c *cli.Context) error {
					return SuperMigrate(c)
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
					&cli.StringFlag{Name: "passport_web_host_url", Value: "http://localhost:8086", EnvVars: []string{"PASSPORT_WEB_HOST_URL"}, Usage: "The API Url where files are hosted"},
					&cli.BoolFlag{Name: "database_prod", Value: false, EnvVars: []string{"PASSPORT_DB_PROD", "DB_PROD"}, Usage: "seed the database (prod)"},
					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{"PASSPORT_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
					&cli.BoolFlag{Name: "seed", EnvVars: []string{"PASSPORT_DB_SEED", "DB_SEED"}, Usage: "seed the database"},

					&cli.IntFlag{Name: "database_max_pool_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_POOL_CONNS"}, Usage: "Database max pool conns"},
					&cli.IntFlag{Name: "database_max_idle_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_IDLE_CONNS"}, Usage: "Database max idle conns"},
					&cli.IntFlag{Name: "database_max_open_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_OPEN_CONNS"}, Usage: "Database max open conns"},
				},
				Usage: "seed the database",
				Action: func(c *cli.Context) error {

					databaseMaxPoolConns := c.Int("database_max_pool_conns")
					databaseMaxIdleConns := c.Int("database_max_idle_conns")
					databaseMaxOpenConns := c.Int("database_max_open_conns")
					databaseUser := c.String("database_user")
					databasePass := c.String("database_pass")
					databaseTxUser := c.String("database_tx_user")
					databaseTxPass := c.String("database_tx_pass")
					databaseHost := c.String("database_host")
					databasePort := c.String("database_port")
					databaseName := c.String("database_name")
					databaseAppName := c.String("database_application_name")
					databaseProd := c.Bool("database_prod")
					passportWebHostUrl := c.String("passport_web_host_url")

					pgxconn, err := pgxconnect(
						databaseUser,
						databasePass,
						databaseHost,
						databasePort,
						databaseName,
						databaseAppName,
						Version,
						databaseMaxPoolConns,
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
						databaseMaxIdleConns,
						databaseMaxOpenConns,
					)
					if err != nil {
						return terror.Panic(err)
					}

					seeder := seed.NewSeeder(pgxconn, txConn, passportWebHostUrl, passportWebHostUrl)
					return seeder.Run(databaseProd)
				},
			},
			{
				Name:  "shortcodes",
				Usage: "print shortcodes",
				Action: func(c *cli.Context) error {
					for i := 0; i < 9; i++ {
						_, _ = helpers.GenerateMetadataHashID("9cdf55aa-217b-4821-aa77-bc8555195f23", i, true)
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		terror.Echo(err)
		os.Exit(1) // so ci knows it no good
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
	maxPool int,
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
	poolConfig.MaxConns = int32(maxPool)

	ctx := context.Background()
	conn, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, terror.Panic(err, "could not initialise database")
	}

	return conn, nil
}

func sqlConnect(
	databaseTxUser string,
	databaseTxPass string,
	databaseHost string,
	databasePort string,
	databaseName string,
	maxIdle int,
	maxOpen int,
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
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	conn := stdlib.OpenDB(*cfg)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(maxIdle)
	conn.SetMaxOpenConns(maxOpen)
	return conn, nil

}

func txConnect(
	databaseTxUser string,
	databaseTxPass string,
	databaseHost string,
	databasePort string,
	databaseName string,
	maxIdle int,
	maxOpen int,
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

	conn, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, terror.Error(err)
	}
	conn.SetMaxIdleConns(maxIdle)
	conn.SetMaxOpenConns(maxOpen)
	return conn, nil
}

func SyncPayments(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) error {

	records1, err := payments.BNB()
	if err != nil {
		return fmt.Errorf("get bnb payments: %w", err)
	}

	log.Info().Int("records", len(records1)).Str("sym", "BNB").Msg("fetch purchases")

	records2, err := payments.BUSD()
	if err != nil {
		return fmt.Errorf("get busd payments: %w", err)
	}

	log.Info().Int("records", len(records2)).Str("sym", "BUSD").Msg("fetch purchases")

	records3, err := payments.ETH()
	if err != nil {
		return fmt.Errorf("get eth payments: %w", err)
	}
	log.Info().Int("records", len(records3)).Str("sym", "ETH").Msg("fetch purchases")
	records4, err := payments.USDC()
	if err != nil {
		return fmt.Errorf("get usdc payments: %w", err)
	}
	log.Info().Int("records", len(records4)).Str("sym", "USDC").Msg("fetch purchases")

	records := []*payments.PurchaseRecord{}
	records = append(records, records1...)
	records = append(records, records2...)
	records = append(records, records3...)
	records = append(records, records4...)

	log.Info().Int("records", len(records1)).Msg("Syncing payments...")
	successful := 0
	skipped := 0
	failed := 0
	for _, r := range records {
		ctx := context.Background()

		exists, err := db.TransactionExists(ctx, conn, r.TxHash)
		if err != nil {
			log.Error().Str("sym", r.Symbol).Str("txid", r.TxHash).Err(err).Msg("check record exists")
			failed++
			continue
		}
		if exists {
			skipped++
			continue
		}

		user, err := payments.CreateOrGetUser(ctx, conn, common.HexToAddress(r.FromAddress))
		if err != nil {
			failed++
			log.Error().Str("sym", r.Symbol).Str("txid", r.TxHash).Err(err).Msg("create new user for payment insertion")
			continue
		}

		input, _, err := payments.ProcessValues(r.Sups, r.ValueInt, r.ValueDecimals)
		if err != nil {
			return err
		}

		if input.Equal(decimal.Zero) {
			log.Warn().Str("sym", r.Symbol).Str("txid", r.TxHash).Msg("zero value payment")
			skipped++
			continue
		}

		err = payments.StoreRecord(ctx, passport.XsynSaleUserID, user.ID, ucm, r)
		if err != nil && strings.Contains(err.Error(), "duplicate key") {
			skipped++
			continue
		}
		if err != nil && !strings.Contains(err.Error(), "duplicate key") {
			failed++
			log.Error().Str("sym", r.Symbol).Str("txid", r.TxHash).Err(err).Msg("duplicate key when inserting payment record")
			continue
		}

		successful++

	}
	log.Info().Int("skipped", skipped).Int("successful", successful).Int("failed", failed).Msg("synced payments")

	return nil

}
func SyncDeposits(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) error {
	depositRecords, err := payments.GetDeposits(isTestnet)
	if err != nil {
		return fmt.Errorf("get deposits: %w", err)
	}
	_, _, err = payments.ProcessDeposits(depositRecords, ucm)
	if err != nil {
		return fmt.Errorf("process deposits: %w", err)
	}

	return nil

}
func SyncWithdraw(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet, enableWithdrawRollback bool) error {
	// Update with TX hash first
	withdrawRecords, err := payments.GetWithdraws(isTestnet)
	if err != nil {
		return fmt.Errorf("get withdraws: %w", err)
	}
	success, skipped := payments.UpdateSuccessfulWithdrawsWithTxHash(withdrawRecords)
	passlog.L.Info().Int("success", success).Int("skipped", skipped).Msg("add tx hashes to pending refunds")

	refundsSuccess, refundsSkipped, err := payments.ReverseFailedWithdraws(ucm, enableWithdrawRollback)
	if err != nil {
		return fmt.Errorf("process withdraws: %w", err)
	}
	passlog.L.Info().Int("success", refundsSuccess).Int("skipped", refundsSkipped).Msg("refunds processed")

	return nil

}
func SyncNFTs(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) error {
	nftOwnerStatuses, err := payments.GetNFTOwnerRecords(isTestnet)
	if err != nil {
		return fmt.Errorf("get nft owners: %w", err)
	}

	ownerupdated, ownerskipped, err := payments.UpdateOwners(nftOwnerStatuses, isTestnet)
	if err != nil {
		return fmt.Errorf("update nft owners: %w", err)
	}
	passlog.L.Info().Int("updated", ownerupdated).Int("skipped", ownerskipped).Msg("synced nft ownerships")
	return nil
}

func SyncFunc(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet, enableWithdrawRollback bool) error {
	go func(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) {
		if db.GetBoolWithDefault(db.KeyEnableSyncPayments, false) {
			err := SyncPayments(ucm, conn, log, isTestnet)
			if err != nil {
				passlog.L.Err(err).Msg("failed to sync payments")
			}
		}
	}(ucm, conn, log, isTestnet)
	go func(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) {
		if db.GetBoolWithDefault(db.KeyEnableSyncDeposits, false) {
			err := SyncDeposits(ucm, conn, log, isTestnet)
			if err != nil {
				passlog.L.Err(err).Msg("failed to sync deposits")
			}
		}
	}(ucm, conn, log, isTestnet)
	go func(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) {
		if db.GetBoolWithDefault(db.KeyEnableSyncNFTOwners, false) {
			err := SyncNFTs(ucm, conn, log, isTestnet)
			if err != nil {
				passlog.L.Err(err).Msg("failed to sync nf ts")
			}
		}
	}(ucm, conn, log, isTestnet)
	go func(ucm *api.Transactor, conn *pgxpool.Pool, log *zerolog.Logger, isTestnet bool) {
		if db.GetBoolWithDefault(db.KeyEnableSyncWithdraw, false) {
			err := SyncWithdraw(ucm, conn, log, isTestnet, enableWithdrawRollback)
			if err != nil {
				passlog.L.Err(err).Msg("failed to sync withdraw")
			}
		}
	}(ucm, conn, log, isTestnet)
	return nil
}
func ServeFunc(ctxCLI *cli.Context, log *zerolog.Logger) error {
	databaseMaxPoolConns := ctxCLI.Int("database_max_pool_conns")
	databaseMaxIdleConns := ctxCLI.Int("database_max_idle_conns")
	databaseMaxOpenConns := ctxCLI.Int("database_max_open_conns")
	environment := ctxCLI.String("environment")
	sentryDSNBackend := ctxCLI.String("sentry_dsn_backend")
	sentryServerName := ctxCLI.String("sentry_server_name")
	sentryTraceRate := ctxCLI.Float64("sentry_sample_rate")
	skipUpdateUsersMixedCase := ctxCLI.Bool("skip_update_users_mixed_case")
	sentryRelease := fmt.Sprintf("%s@%s", SentryReleasePrefix, Version)
	err := log_helpers.SentryInit(sentryDSNBackend, sentryServerName, sentryRelease, environment, sentryTraceRate, log)
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

	MoralisKey := ctxCLI.String("moralis_key")
	UsdcAddr := ctxCLI.String("usdc_addr")
	BusdAddr := ctxCLI.String("busd_addr")
	SupAddr := ctxCLI.String("sup_addr")
	PurchaseAddr := ctxCLI.String("purchase_addr")
	WithdrawAddr := ctxCLI.String("withdraw_addr")
	OperatorAddr := ctxCLI.String("operator_addr")
	SignerPrivateKey := ctxCLI.String("signer_private_key")
	BscNodeAddr := ctxCLI.String("bsc_node_addr")
	EthNodeAddr := ctxCLI.String("eth_node_addr")
	BSCChainID := ctxCLI.Int64("bsc_chain_id")
	ETHChainID := ctxCLI.Int64("eth_chain_id")
	BSCRouterAddr := ctxCLI.String("bsc_router_addr")

	enablePurchaseSubscription := ctxCLI.Bool("enable_purchase_subscription")

	isTestnetBlockchain := ctxCLI.Bool("is_testnet_blockchain")
	runBlockchainBridge := ctxCLI.Bool("run_blockchain_bridge")

	jwtKey := ctxCLI.String("jwt_key")
	mailDomain := ctxCLI.String("mail_domain")
	mailAPIKey := ctxCLI.String("mail_apikey")
	mailSender := ctxCLI.String("mail_sender")
	twilioSid := ctxCLI.String("twilio_sid")
	twilioApiKey := ctxCLI.String("twilio_api_key")
	twilioApiSecrete := ctxCLI.String("twilio_api_secret")
	smsFromNumber := ctxCLI.String("sms_from_number")
	externalURL := ctxCLI.String("passport_web_host_url")
	insecuritySkipVerify := false
	if environment == "development" || environment == "testing" {
		insecuritySkipVerify = true
	}

	gameserverWebhookToken := ctxCLI.String("gameserver_webhook_secret")
	if gameserverWebhookToken == "" {
		return terror.Panic(fmt.Errorf("missing passort webhook token"))
	}

	gameserverHostUrl := ctxCLI.String("gameserver_host_url")
	if gameserverHostUrl == "" {
		return terror.Panic(fmt.Errorf("missing passort webhook token"))
	}

	gameserverToken := ctxCLI.String("gameserver_token")
	if gameserverToken == "" {
		return terror.Panic(fmt.Errorf("missing gameserver auth token"))
	}

	googleClientID := ctxCLI.String("google_client_id")
	if googleClientID == "" {
		return terror.Panic(fmt.Errorf("missing google client id"))
	}

	twitchClientID := ctxCLI.String("twitch_client_id")
	if twitchClientID == "" {
		return terror.Panic(fmt.Errorf("no twitch client id"))
	}

	twitchClientSecret := ctxCLI.String("twitch_client_secret")
	if twitchClientSecret == "" {
		return terror.Panic(fmt.Errorf("no twitch client secret"))
	}

	twitterAPIKey := ctxCLI.String("twitter_api_key")
	if twitterAPIKey == "" {
		return terror.Panic(fmt.Errorf("no twitter api key"))
	}

	twitterAPISecret := ctxCLI.String("twitter_api_secret")
	if twitterAPISecret == "" {
		return terror.Panic(fmt.Errorf("no twitter api secret"))
	}

	discordClientID := ctxCLI.String("discord_client_id")
	if discordClientID == "" {
		return terror.Panic(fmt.Errorf("no discord client id"))
	}

	discordClientSecret := ctxCLI.String("discord_client_secret")
	if discordClientSecret == "" {
		return terror.Panic(fmt.Errorf("no discord client secret"))
	}

	config := &passport.Config{
		CookieSecure:        ctxCLI.Bool("cookie_secure"),
		PassportWebHostURL:  ctxCLI.String("passport_web_host_url"),
		GameserverHostURL:   ctxCLI.String("gameserver_web_host_url"),
		EncryptTokens:       ctxCLI.Bool("jwt_encrypt"),
		EncryptTokensKey:    ctxCLI.String("jwt_encrypt_key"),
		TokenExpirationDays: ctxCLI.Int("jwt_expiry_days"),
		MetaMaskSignMessage: ctxCLI.String("metamask_sign_message"),
		BridgeParams: &passport.BridgeParams{
			MoralisKey:       MoralisKey,
			OperatorAddr:     common.HexToAddress(OperatorAddr),
			UsdcAddr:         common.HexToAddress(UsdcAddr),
			BusdAddr:         common.HexToAddress(BusdAddr),
			SupAddr:          common.HexToAddress(SupAddr),
			PurchaseAddr:     common.HexToAddress(PurchaseAddr),
			WithdrawAddr:     common.HexToAddress(WithdrawAddr),
			SignerPrivateKey: SignerPrivateKey,
			BscNodeAddr:      BscNodeAddr,
			EthNodeAddr:      EthNodeAddr,
			BSCChainID:       BSCChainID,
			ETHChainID:       ETHChainID,
			BSCRouterAddr:    common.HexToAddress(BSCRouterAddr),
		},
		OnlyWalletConnect:       ctxCLI.Bool("only_wallet"),
		WhitelistEndpoint:       ctxCLI.String("whitelist_check_endpoint"),
		InsecureSkipVerifyCheck: insecuritySkipVerify,
		AuthParams: &passport.AuthParams{
			GameserverToken:     gameserverToken,
			GoogleClientID:      googleClientID,
			TwitchClientID:      twitchClientID,
			TwitchClientSecret:  twitchClientSecret,
			TwitterAPIKey:       twitterAPIKey,
			TwitterAPISecret:    twitterAPISecret,
			DiscordClientID:     discordClientID,
			DiscordClientSecret: discordClientSecret,
		},
		WebhookParams: &passport.WebhookParams{
			GameserverWebhookToken: gameserverWebhookToken,
			GameserverHostUrl:      gameserverHostUrl,
		},
	}

	txConn, err := txConnect(
		databaseTxUser,
		databaseTxPass,
		databaseHost,
		databasePort,
		databaseName,
		databaseMaxIdleConns,
		databaseMaxOpenConns,
	)
	if err != nil {
		return terror.Panic(err)
	}

	pgxconn, err := pgxconnect(
		databaseUser,
		databasePass,
		databaseHost,
		databasePort,
		databaseName,
		databaseAppName,
		Version,
		databaseMaxPoolConns,
	)
	if err != nil {
		return terror.Panic(err)
	}

	sqlConnect, err := sqlConnect(
		databaseUser,
		databasePass,
		databaseHost,
		databasePort,
		databaseName,
		databaseMaxIdleConns,
		databaseMaxOpenConns,
	)
	if err != nil {
		return terror.Panic(err)
	}
	err = passdb.New(pgxconn, sqlConnect)
	if err != nil {
		return terror.Panic(err)
	}
	count := 0
	err = db.IsSchemaDirty(context.Background(), pgxconn, &count)
	if err != nil {
		return terror.Error(api.ErrCheckDBQuery)
	}
	if count > 0 {
		return terror.Error(api.ErrCheckDBDirty)
	}

	// Mailer
	mailer, err := email.NewMailer(mailDomain, mailAPIKey, mailSender, config, log)
	if err != nil {
		return terror.Panic(err, "Mailer init failed")
	}

	// SMS
	twilio, err := sms.NewTwilio(twilioSid, twilioApiKey, twilioApiSecrete, smsFromNumber, environment)
	if err != nil {
		return terror.Panic(err, "SMS init failed")
	}

	// HTML Sanitizer
	HTMLSanitizePolicy := bluemonday.UGCPolicy()
	HTMLSanitizePolicy.AllowAttrs("class").OnElements("img", "table", "tr", "td", "p")

	msgBus := messagebus.NewMessageBus(log_helpers.NamedLogger(log, "message bus"))

	// initialise user cache map
	ucm, err := api.NewTX(pgxconn, msgBus)
	if err != nil {
		return terror.Error(err)
	}

	jwtKeyByteArray, err := base64.StdEncoding.DecodeString(jwtKey)
	if err != nil {
		return terror.Error(err, "Failed to convert string to byte array")
	}

	// API Server
	api, routes := api.NewAPI(log,
		pgxconn,
		txConn,
		mailer,
		twilio,
		apiAddr,
		HTMLSanitizePolicy,
		config,
		externalURL,
		ucm,
		isTestnetBlockchain,
		runBlockchainBridge,
		msgBus,
		enablePurchaseSubscription,
		jwtKeyByteArray,
	)

	passlog.L.Info().Msg("start rpc server")
	s := comms.NewServer(ucm, msgBus, api.SupremacyController.Txs, log, pgxconn, api.ClientMap, twilio)
	err = comms.StartServer(s)
	if err != nil {
		return terror.Error(err)
	}

	apiServer := &http.Server{
		Addr:    api.Addr,
		Handler: routes,
	}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		api.Log.Info().Msg("Stopping API")
		err := apiServer.Shutdown(context.Background())
		if err != nil {
			api.Log.Warn().Err(err).Msg("")
		}
		os.Exit(1)
	}()

	go func() {
		gameserverAddr := ctxCLI.String("gameserver_web_host_url")
		gameserverURL, err := url.Parse(gameserverAddr)
		if err != nil {
			passlog.L.Err(err).Msg("parse gameserver addr")
			return
		}
		hostname := gameserverURL.Hostname()
		rpcAddrs := []string{
			fmt.Sprintf("%s:11001", hostname),
			fmt.Sprintf("%s:11002", hostname),
			fmt.Sprintf("%s:11003", hostname),
			fmt.Sprintf("%s:11004", hostname),
			fmt.Sprintf("%s:11005", hostname),
			fmt.Sprintf("%s:11006", hostname),
			fmt.Sprintf("%s:11007", hostname),
			fmt.Sprintf("%s:11008", hostname),
			fmt.Sprintf("%s:11009", hostname),
			fmt.Sprintf("%s:11010", hostname),
			fmt.Sprintf("%s:11011", hostname),
			fmt.Sprintf("%s:11012", hostname),
			fmt.Sprintf("%s:11013", hostname),
			fmt.Sprintf("%s:11014", hostname),
			fmt.Sprintf("%s:11015", hostname),
			fmt.Sprintf("%s:11016", hostname),
			fmt.Sprintf("%s:11017", hostname),
			fmt.Sprintf("%s:11018", hostname),
			fmt.Sprintf("%s:11019", hostname),
			fmt.Sprintf("%s:11020", hostname),
			fmt.Sprintf("%s:11021", hostname),
			fmt.Sprintf("%s:11022", hostname),
			fmt.Sprintf("%s:11023", hostname),
			fmt.Sprintf("%s:11024", hostname),
			fmt.Sprintf("%s:11025", hostname),
			fmt.Sprintf("%s:11026", hostname),
			fmt.Sprintf("%s:11027", hostname),
			fmt.Sprintf("%s:11028", hostname),
			fmt.Sprintf("%s:11029", hostname),
			fmt.Sprintf("%s:11030", hostname),
			fmt.Sprintf("%s:11031", hostname),
			fmt.Sprintf("%s:11032", hostname),
			fmt.Sprintf("%s:11033", hostname),
			fmt.Sprintf("%s:11034", hostname),
			fmt.Sprintf("%s:11035", hostname),
		}
		rpcClient := &rpcclient.XrpcClient{
			Addrs: rpcAddrs,
		}
		rpcclient.SetGlobalClient(rpcClient)
	}()

	if enablePurchaseSubscription {
		l := passlog.L.With().Str("svc", "avant_scraper").Logger()
		db.PutInt(db.KeyLatestWithdrawBlock, 0)
		db.PutInt(db.KeyLatestDepositBlock, 0)
		db.PutInt(db.KeyLatestETHBlock, 0)
		db.PutInt(db.KeyLatestBNBBlock, 0)
		db.PutInt(db.KeyLatestBUSDBlock, 0)
		db.PutInt(db.KeyLatestUSDCBlock, 0)

		enableWithdrawRollback := db.GetBoolWithDefault(db.KeyEnableWithdrawRollback, false)
		if !enableWithdrawRollback {
			l.Debug().Bool("enable_withdraw_rollback", enableWithdrawRollback).Msg("withdraw rollback is disabled")
		} else {
			l.Debug().Bool("enable_withdraw_rollback", enableWithdrawRollback).Msg("withdraw rollback is enabled")
		}
		avantTestnet := ctxCLI.Bool("avant_testnet")
		err := SyncFunc(ucm, pgxconn, log, avantTestnet, enableWithdrawRollback)
		if err != nil {
			log.Error().Err(err).Msg("sync")
		}

		go func() {
			t := time.NewTicker(20 * time.Second)
			for range t.C {
				enableWithdrawRollback := db.GetBoolWithDefault(db.KeyEnableWithdrawRollback, false)
				if !enableWithdrawRollback {
					l.Debug().Bool("enable_withdraw_rollback", enableWithdrawRollback).Msg("withdraw rollback is disabled")
				} else {
					l.Debug().Bool("enable_withdraw_rollback", enableWithdrawRollback).Msg("withdraw rollback is enabled")
				}
				err := SyncFunc(ucm, pgxconn, log, avantTestnet, enableWithdrawRollback)
				if err != nil {
					log.Error().Err(err).Msg("sync")
				}
			}
		}()
	}

	if !skipUpdateUsersMixedCase {
		go func() {
			passlog.L.Info().Msg("updating all users to mixed case")
			err = db.UserMixedCaseUpdateAll()
			if err != nil {
				passlog.L.Error().Err(err).Msg("updating all users to mixed case failed")
			}
		}()
	}

	api.Log.Info().Msg("Starting API")
	return apiServer.ListenAndServe()
}

func SuperMigrate(c *cli.Context) error {
	databaseMaxPoolConns := c.Int("database_max_pool_conns")
	databaseMaxIdleConns := c.Int("database_max_idle_conns")
	databaseMaxOpenConns := c.Int("database_max_open_conns")
	databaseUser := c.String("database_user")
	databasePass := c.String("database_pass")
	databaseHost := c.String("database_host")
	databasePort := c.String("database_port")
	databaseName := c.String("database_name")
	databaseAppName := c.String("database_application_name")
	gameserverAddr := c.String("gameserver_web_host_url")
	passlog.New("development", "InfoLevel")
	pgxconn, err := pgxconnect(
		databaseUser,
		databasePass,
		databaseHost,
		databasePort,
		databaseName,
		databaseAppName,
		Version,
		databaseMaxPoolConns,
	)
	if err != nil {
		return terror.Panic(err)
	}

	sqlConnect, err := sqlConnect(
		databaseUser,
		databasePass,
		databaseHost,
		databasePort,
		databaseName,
		databaseMaxIdleConns,
		databaseMaxOpenConns,
	)
	if err != nil {
		return terror.Panic(err)
	}
	err = passdb.New(pgxconn, sqlConnect)
	if err != nil {
		return terror.Panic(err)
	}
	gameserverURL, err := url.Parse(gameserverAddr)
	if err != nil {
		return terror.Panic(err)
	}
	hostname := gameserverURL.Hostname()
	rpcAddrs := []string{
		fmt.Sprintf("%s:11001", hostname),
		fmt.Sprintf("%s:11002", hostname),
		fmt.Sprintf("%s:11003", hostname),
		fmt.Sprintf("%s:11004", hostname),
		fmt.Sprintf("%s:11005", hostname),
		fmt.Sprintf("%s:11006", hostname),
		fmt.Sprintf("%s:11007", hostname),
		fmt.Sprintf("%s:11008", hostname),
		fmt.Sprintf("%s:11009", hostname),
		fmt.Sprintf("%s:11010", hostname),
		fmt.Sprintf("%s:11011", hostname),
		fmt.Sprintf("%s:11012", hostname),
		fmt.Sprintf("%s:11013", hostname),
		fmt.Sprintf("%s:11014", hostname),
		fmt.Sprintf("%s:11015", hostname),
		fmt.Sprintf("%s:11016", hostname),
		fmt.Sprintf("%s:11017", hostname),
		fmt.Sprintf("%s:11018", hostname),
		fmt.Sprintf("%s:11019", hostname),
		fmt.Sprintf("%s:11020", hostname),
		fmt.Sprintf("%s:11021", hostname),
		fmt.Sprintf("%s:11022", hostname),
		fmt.Sprintf("%s:11023", hostname),
		fmt.Sprintf("%s:11024", hostname),
		fmt.Sprintf("%s:11025", hostname),
		fmt.Sprintf("%s:11026", hostname),
		fmt.Sprintf("%s:11027", hostname),
		fmt.Sprintf("%s:11028", hostname),
		fmt.Sprintf("%s:11029", hostname),
		fmt.Sprintf("%s:11030", hostname),
		fmt.Sprintf("%s:11031", hostname),
		fmt.Sprintf("%s:11032", hostname),
		fmt.Sprintf("%s:11033", hostname),
		fmt.Sprintf("%s:11034", hostname),
		fmt.Sprintf("%s:11035", hostname),
	}
	rpcClient := &rpcclient.XrpcClient{
		Addrs: rpcAddrs,
	}
	rpcclient.SetGlobalClient(rpcClient)
	err = db.SyncStoreItems()
	if err != nil {
		return terror.Panic(err)
	}
	err = db.SyncPurchasedItems()
	if err != nil {
		return terror.Panic(err)
	}
	return nil
}

// pprofMonitor monitor to help debug some invisible issues
func pprofMonitor(intervalSecond, listenPort int) {
	if intervalSecond < 10 {
		intervalSecond = 10
	}
	if listenPort <= 0 || listenPort >= 65535 {
		listenPort = 6060
	}

	// auto record at interval
	err := os.Mkdir("/tmp/passport-pprof", 0755)
	if err != nil {
		log.Println("ERROR pprof mkdir fail", err)
	}

	go func() {
		lists := []string{
			"allocs",
			"block",
			"goroutine",
			"heap",
			"mutex",
			"threadcreate",
			"goroutine",
		}
		for {
			log.Printf("total goroutines %d\n", runtime.NumGoroutine())

			for _, list := range lists {
				t := time.Now().Format("2006-01-02T15:04:05")
				fName := fmt.Sprintf("/tmp/passport-pprof/%s-%s.dump", t, list)

				f, err := os.Create(fName)
				if err != nil {
					log.Println("ERROR failed to create pprof file", err)
					continue
				}

				err = rpprof.Lookup(list).WriteTo(f, 1)
				if err != nil {
					log.Println("ERROR failed to write pprof file", err)
					continue
				}

				err = f.Close()
				if err != nil {
					log.Println("ERROR failed to close pprof file", err)
					continue
				}
			}

			time.Sleep(time.Duration(intervalSecond) * time.Second)
		}
	}()
	// pprof for quick web check
	go func() {
		log.Println(
			http.ListenAndServe(
				fmt.Sprintf("localhost:%d", listenPort),
				nil,
			),
		)
	}()
}
