package main

import (
	"database/sql"
	"errors"
	"net/url"
	"passport"
	"passport/api"
	"passport/db"
	"passport/email"
	"passport/log_helpers"
	"passport/seed"
	"time"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/common"

	_ "github.com/lib/pq" //postgres drivers for initialization

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

					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{envPrefix + "_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
					&cli.StringFlag{Name: "sentry_dsn_backend", Value: "", EnvVars: []string{envPrefix + "_SENTRY_DSN_BACKEND", "SENTRY_DSN_BACKEND"}, Usage: "Sends error to remote server. If set, it will send error."},
					&cli.StringFlag{Name: "sentry_server_name", Value: "dev-pc", EnvVars: []string{envPrefix + "_SENTRY_SERVER_NAME", "SENTRY_SERVER_NAME"}, Usage: "The machine name that this program is running on."},
					&cli.Float64Flag{Name: "sentry_sample_rate", Value: 1, EnvVars: []string{envPrefix + "_SENTRY_SAMPLE_RATE", "SENTRY_SAMPLE_RATE"}, Usage: "The percentage of trace sample to collect (0.0-1)"},

					&cli.StringFlag{Name: "passport_web_host_url", Value: "http://localhost:5003", EnvVars: []string{envPrefix + "_HOST_URL_FRONTEND"}, Usage: "The Public Site URL used for CORS and links (eg: in the mailer)"},
					&cli.StringFlag{Name: "gameserver_web_host_url", Value: "http://localhost:8084", EnvVars: []string{"GAMESERVER_HOST_URL"}, Usage: "The host for the gameserver, to allow it to connect"},

					&cli.StringFlag{Name: "api_addr", Value: ":8086", EnvVars: []string{envPrefix + "_API_ADDR", "API_ADDR"}, Usage: "host:port to run the API"},
					&cli.BoolFlag{Name: "cookie_secure", Value: true, EnvVars: []string{envPrefix + "_COOKIE_SECURE", "COOKIE_SECURE"}, Usage: "set cookie secure"},
					&cli.StringFlag{Name: "google_client_id", Value: "467953368642-8cobg822tej2i50ncfg4ge1pm4c5v033.apps.googleusercontent.com", EnvVars: []string{envPrefix + "_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"}, Usage: "Google Client ID for OAuth functionaility."},

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

					/****************************
					 *		Bridge details		*
					 ***************************/
					// ETH
					&cli.StringFlag{Name: "usdc_addr", Value: "0x8BB4eC208CDDE7761ac7f3346deBb9C931f80A33", Usage: "USDC contract address"},
					&cli.StringFlag{Name: "weth_addr", Value: "0x8cAEF228f1C322e34B04AD77f70b9f4bDdbd0fFD", Usage: "WETH contract address"},

					// BSC
					&cli.StringFlag{Name: "busd_addr", Value: "0xeAf33Ba4AcA3fE3110EAddD7D4cf0897121583D0", Usage: "BUSD contract address"},
					&cli.StringFlag{Name: "wbnb_addr", Value: "0xb2564d8Fd501868340eF0A1281B2aDA3E4506C7F", Usage: "WBNB contract address"},
					&cli.StringFlag{Name: "sup_addr", Value: "0xED4664f5F37307abf8703dD39Fd6e72F421e7DE2", Usage: "SUP contract address"},

					// wallet/contract addressed
					&cli.StringFlag{Name: "purchase_addr", Value: "0x5591eBC09A89A8B11D9644eC1455e294Fd3BAbB5", Usage: "Purchase wallet address"},

					&cli.StringFlag{Name: "withdraw_addr", Value: "0xEebebbd60e8800Db03e48E8052d9Ff7fb5cF81D2", Usage: "Withdraw contract address"},
					&cli.StringFlag{Name: "redemption_addr", Value: "0xBF7301CB28D765072D39CD21429f6E9Ec81ECe92", Usage: "Redemption contract address"},

					// chain id
					&cli.Int64Flag{Name: "bsc_chain_id", Value: 97, Usage: "BSC Chain ID"},
					&cli.Int64Flag{Name: "eth_chain_id", Value: 5, Usage: "ETH Chain ID"},

					// node address
					&cli.StringFlag{Name: "bsc_node_addr", Value: "wss://speedy-nodes-nyc.moralis.io/1375aa321ac8ac6cfba6aa9c/bsc/testnet/ws", Usage: "Binance WS node URL"},
					&cli.StringFlag{Name: "eth_node_addr", Value: "wss://speedy-nodes-nyc.moralis.io/1375aa321ac8ac6cfba6aa9c/eth/goerli/ws", Usage: "Ethereum WS node URL"},

					// exchange rates
					&cli.StringFlag{Name: "usdc_to_sups", Value: "8.333333333333333333", Usage: "Exchange rate for 1 USDC to SUPS"},
					&cli.StringFlag{Name: "busd_to_sups", Value: "8.333333333333333333", Usage: "Exchange rate for 1 BUSD to SUPS"},
					&cli.StringFlag{Name: "sup_price", Value: "0.12", Usage: "Exchange rate for 1 SUP to USD"},
					&cli.StringFlag{Name: "weth_to_sups", Value: "21000", Usage: "Exchange rate for 1 WETH to SUPS"},
					&cli.StringFlag{Name: "wbnb_to_sups", Value: "3000", Usage: "Exchange rate for 1 WBNB to SUPS"},
				},

				Usage: "run server",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					environment := c.String("environment")
					log := log_helpers.LoggerInitZero(environment)

					log.Info().Msg("zerolog initialised")
					nlog := log.Level(zerolog.DebugLevel)
					log = &nlog
					g := &run.Group{}
					// Listen for os.interrupt
					g.Add(run.SignalHandler(ctx, os.Interrupt))
					// start the server
					g.Add(func() error { return ServeFunc(c, ctx, log) }, func(err error) { cancel() })

					err := g.Run()

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
	poolConfig.MaxConns = 95

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

	UsdcAddr := ctxCLI.String("usdc_addr")
	BusdAddr := ctxCLI.String("busd_addr")
	WethAddr := ctxCLI.String("weth_addr")
	WbnbAddr := ctxCLI.String("wbnb_addr")
	SupAddr := ctxCLI.String("sup_addr")
	PurchaseAddr := ctxCLI.String("purchase_addr")
	WithdrawAddr := ctxCLI.String("withdraw_addr")
	RedemptionAddr := ctxCLI.String("redemption_addr")
	BscNodeAddr := ctxCLI.String("bsc_node_addr")
	EthNodeAddr := ctxCLI.String("eth_node_addr")
	BSCChainID := ctxCLI.Int64("bsc_chain_id")
	ETHChainID := ctxCLI.Int64("eth_chain_id")

	USDCToSUPS, err := decimal.NewFromString(ctxCLI.String("usdc_to_sups"))
	if err != nil {
		return err
	}
	BUSDToSUPS, err := decimal.NewFromString(ctxCLI.String("busd_to_sups"))
	if err != nil {
		return err
	}
	WETHToSUPS, err := decimal.NewFromString(ctxCLI.String("weth_to_sups"))
	if err != nil {
		return err
	}
	WBNBToSUPS, err := decimal.NewFromString(ctxCLI.String("wbnb_to_sups"))
	if err != nil {
		return err
	}
	SUPToUSD, err := decimal.NewFromString(ctxCLI.String("sup_price"))
	if err != nil {
		return err
	}

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
		BridgeParams: &passport.BridgeParams{
			UsdcAddr:       common.HexToAddress(UsdcAddr),
			BusdAddr:       common.HexToAddress(BusdAddr),
			WethAddr:       common.HexToAddress(WethAddr),
			WbnbAddr:       common.HexToAddress(WbnbAddr),
			SupAddr:        common.HexToAddress(SupAddr),
			PurchaseAddr:   common.HexToAddress(PurchaseAddr),
			WithdrawAddr:   common.HexToAddress(WithdrawAddr),
			RedemptionAddr: common.HexToAddress(RedemptionAddr),
			BscNodeAddr:    BscNodeAddr,
			EthNodeAddr:    EthNodeAddr,
			BSCChainID:     BSCChainID,
			ETHChainID:     ETHChainID,
			USDCToSUPS:     USDCToSUPS,
			BUSDToSUPS:     BUSDToSUPS,
			WETHToSUPS:     WETHToSUPS,
			WBNBToSUPS:     WBNBToSUPS,
			SUPToUSD:       SUPToUSD,
		},
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
	api := api.NewAPI(log, cancelOnPanic, pgxconn, txConn, googleClientID, mailer, apiAddr, twitchClientID, twitchClientSecret, HTMLSanitizePolicy, config, twitterAPIKey, twitterAPISecret, discordClientID, discordClientSecret, gameserverToken)
	return api.Run(ctx)
}
