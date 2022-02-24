package main

import (
	"database/sql"
	"errors"
	"net/url"
	"passport"
	"passport/api"
	"passport/db"
	"passport/email"
	"passport/payments"
	"passport/seed"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/log_helpers"
	"github.com/oklog/run"
	"github.com/shopspring/decimal"

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
					&cli.BoolFlag{Name: "run_blockchain_bridge", Value: false, EnvVars: []string{envPrefix + "_RUN_BLOCKCHAIN_BRIDGE"}, Usage: "Run the bridge to blockchain data"},

					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{envPrefix + "_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
					&cli.StringFlag{Name: "sentry_dsn_backend", Value: "", EnvVars: []string{envPrefix + "_SENTRY_DSN_BACKEND", "SENTRY_DSN_BACKEND"}, Usage: "Sends error to remote server. If set, it will send error."},
					&cli.StringFlag{Name: "sentry_server_name", Value: "dev-pc", EnvVars: []string{envPrefix + "_SENTRY_SERVER_NAME", "SENTRY_SERVER_NAME"}, Usage: "The machine name that this program is running on."},
					&cli.Float64Flag{Name: "sentry_sample_rate", Value: 1, EnvVars: []string{envPrefix + "_SENTRY_SAMPLE_RATE", "SENTRY_SAMPLE_RATE"}, Usage: "The percentage of trace sample to collect (0.0-1)"},
					&cli.StringFlag{Name: "log_level", Value: "DebugLevel", EnvVars: []string{envPrefix + "_LOG_LEVEL"}, Usage: "Set the log level for zerolog (Options: PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel"},

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
					&cli.BoolFlag{Name: "only_wallet", Value: true, EnvVars: []string{envPrefix + "_ONLY_WALLET"}, Usage: "Set passport to only accept wallet logins"},
					&cli.StringFlag{Name: "whitelist_check_endpoint", Value: "https://stories.supremacy.game", EnvVars: []string{envPrefix + "_WHITELIST_ENDPOINT"}, Usage: "Endpoint to check if user is whitelisted"},

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

					&cli.StringFlag{Name: "eth_nft_addr", Value: "0xC1ce98F52E771Bd82938c4Cb6CCaA40Dc2B3258D", EnvVars: []string{envPrefix + "_NFT_CONTRACT_ADDR"}, Usage: "NFT contract address for minting"},
					&cli.StringFlag{Name: "eth_nft_staking_addr", Value: "0xceED4Db9234e7374fe3132a2610c31275712685C", EnvVars: []string{envPrefix + "_NFT_STAKING_CONTRACT_ADDR"}, Usage: "NFT staking contract address for locking"},

					// chain id
					&cli.Int64Flag{Name: "bsc_chain_id", Value: 56, EnvVars: []string{envPrefix + "_BSC_CHAIN_ID"}, Usage: "BSC Chain ID"},
					&cli.Int64Flag{Name: "eth_chain_id", Value: 1, EnvVars: []string{envPrefix + "_ETH_CHAIN_ID"}, Usage: "ETH Chain ID"},

					// node address
					&cli.StringFlag{Name: "bsc_node_addr", Value: "wss://thrumming-misty-bush.bsc.quiknode.pro/f08252fc1f9a373108c84dad3c9ab46868781b24/", EnvVars: []string{envPrefix + "_BSC_WS_NODE_URL"}, Usage: "Binance WS node URL"},
					&cli.StringFlag{Name: "eth_node_addr", Value: "wss://sparkling-polished-glade.quiknode.pro/a68ec6502e56dd3292f33c276c81cc6360877e58/", EnvVars: []string{envPrefix + "_ETH_WS_NODE_URL"}, Usage: "Ethereum WS node URL"},
					//router address for exchange rates
					&cli.StringFlag{Name: "bsc_router_addr", Value: "0x10ED43C718714eb63d5aA57B78B54704E256024E", EnvVars: []string{envPrefix + "_BSC_ROUTER_ADDR"}, Usage: "BSC Router address"},
					&cli.BoolFlag{Name: "enable_purchase_subscription", Value: false, EnvVars: []string{envPrefix + "_ENABLE_PURCHASE_SUBSCRIPTION"}, Usage: "Scrape payments every 20 seconds"},
					//moralis key- set in env vars
					&cli.StringFlag{Name: "moralis_key", Value: "oijl9YX0BIopm9fRitAYhMWuJOrqr7CE1xl5FIO9XncEdOx5CvxkwMOKm2bv4s0p", EnvVars: []string{envPrefix + "_MORALIS_KEY"}, Usage: "Key to connect to moralis API"},
				},

				Usage: "run server",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					environment := c.String("environment")
					level := c.String("log_level")
					log := log_helpers.LoggerInitZero(environment, level)
					log.Info().Msg("zerolog initialised")

					enablePurchaseSubscription := c.Bool("enable_purchase_subscription")
					if enablePurchaseSubscription {
						err := SyncFunc(c, log)
						if err != nil {
							log.Error().Err(err).Msg("sync")
						}

						go func() {
							t := time.NewTicker(20 * time.Second)
							for range t.C {
								err := SyncFunc(c, log)
								if err != nil {
									log.Error().Err(err).Msg("sync")
								}
							}
						}()
					}

					g := &run.Group{}
					// Listen for os.interrupt
					g.Add(run.SignalHandler(ctx, os.Interrupt))
					// start the server
					g.Add(func() error { return ServeFunc(c, ctx, log) }, func(err error) { cancel() })

					err := g.Run()
					if errors.Is(err, run.SignalError{Signal: os.Interrupt}) {
						err = terror.Warn(err)
					}
					log_helpers.TerrorEcho(ctx, err, log)
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
					&cli.StringFlag{Name: "passport_web_host_url", Value: "http://localhost:8086", EnvVars: []string{"PASSPORT_WEB_HOST_URL"}, Usage: "The API Url where files are hosted"},
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
					passportWebHostUrl := c.String("passport_web_host_url")

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

					seeder := seed.NewSeeder(pgxconn, txConn, passportWebHostUrl, passportWebHostUrl)
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
func SyncFunc(ctxCLI *cli.Context, log *zerolog.Logger) error {
	databaseUser := ctxCLI.String("database_user")
	databasePass := ctxCLI.String("database_pass")
	databaseHost := ctxCLI.String("database_host")
	databasePort := ctxCLI.String("database_port")
	databaseName := ctxCLI.String("database_name")
	databaseAppName := ctxCLI.String("database_application_name")

	databaseTxUser := ctxCLI.String("database_tx_user")
	databaseTxPass := ctxCLI.String("database_tx_pass")

	conn, err := pgxconnect(
		databaseUser,
		databasePass,
		databaseHost,
		databasePort,
		databaseName,
		databaseAppName,
		Version,
	)
	if err != nil {
		return err
	}

	txConn, err := txConnect(
		databaseTxUser,
		databaseTxPass,
		databaseHost,
		databasePort,
		databaseName,
	)
	if err != nil {
		return err
	}

	records1, err := payments.BNB(3)
	if err != nil {
		return err
	}

	z := decimal.Zero
	totalSupsSold := decimal.Zero
	for _, r := range records1 {
		sups, err := decimal.NewFromString(r.Sups)
		if err != nil {
			return err
		}
		totalSupsSold = totalSupsSold.Add(sups)
		d, err := decimal.NewFromString(r.Value)
		if err != nil {
			log.Error().Err(err).Msg("parse decimal from string")
		}
		z = z.Add(d)
	}
	log.Info().Str("sym", "BNB").Str("sups", totalSupsSold.StringFixed(4)).Str("total", z.StringFixed(4)).Msg("total inputs")

	records2, err := payments.BUSD(3)
	if err != nil {
		return err
	}

	z = decimal.Zero
	totalSupsSold = decimal.Zero
	for _, r := range records2 {
		sups, err := decimal.NewFromString(r.Sups)
		if err != nil {
			return err
		}
		totalSupsSold = totalSupsSold.Add(sups)
		d, err := decimal.NewFromString(r.Value)
		if err != nil {
			log.Error().Err(err).Msg("parse decimal from string")
		}
		if d.GreaterThan(decimal.NewFromInt(500000)) {
			fmt.Println("BIG!", d.String(), r.TxHash)
		}
		z = z.Add(d)
	}
	log.Info().Str("sym", "BUSD").Str("sups", totalSupsSold.StringFixed(4)).Str("total", z.StringFixed(4)).Str("total", z.StringFixed(4)).Msg("total inputs")

	records3, err := payments.ETH(3)
	if err != nil {
		return err
	}
	totalSupsSold = decimal.Zero
	z = decimal.Zero
	for _, r := range records3 {
		sups, err := decimal.NewFromString(r.Sups)
		if err != nil {
			return err
		}
		totalSupsSold = totalSupsSold.Add(sups)
		d, err := decimal.NewFromString(r.Value)
		if err != nil {
			log.Error().Err(err).Msg("parse decimal from string")
		}
		z = z.Add(d)
	}
	log.Info().Str("sym", "ETH").Str("sups", totalSupsSold.StringFixed(4)).Str("total", z.StringFixed(4)).Str("total", z.StringFixed(4)).Msg("total inputs")
	records4, err := payments.USDC(3)
	if err != nil {
		return err
	}
	totalSupsSold = decimal.Zero
	z = decimal.Zero
	for _, r := range records4 {
		sups, err := decimal.NewFromString(r.Sups)
		if err != nil {
			return err
		}
		totalSupsSold = totalSupsSold.Add(sups)
		d, err := decimal.NewFromString(r.Value)
		if err != nil {
			log.Error().Err(err).Msg("parse decimal from string")
		}
		z = z.Add(d)
	}
	log.Info().Str("sym", "USDC").Str("sups", totalSupsSold.StringFixed(4)).Str("total", z.StringFixed(4)).Str("total", z.StringFixed(4)).Msg("total inputs")

	records1 = append(records1, records2...)
	records1 = append(records1, records3...)
	records1 = append(records1, records4...)
	log.Info().Int("records", len(records1)).Msg("Syncing payments...")
	successful := 0
	skipped := 0
	for _, r := range records1 {
		ctx := context.Background()
		err = payments.StoreRecord(ctx, conn, txConn, r)
		if err != nil && strings.Contains(err.Error(), "duplicate key") {
			skipped++
			continue
		}
		if err != nil && !strings.Contains(err.Error(), "duplicate key") {
			fmt.Println(r.Value, r.JSON.Value)
			log.Error().Str("sym", r.Symbol).Str("txid", r.TxHash).Err(err).Msg("store record")
			continue
		}
		successful++
	}
	log.Info().Int("skipped", skipped).Int("successful", successful).Int("failed", len(records1)-successful).Msg("Synced payments.")
	return nil
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

	MoralisKey := ctxCLI.String("moralis_key")
	UsdcAddr := ctxCLI.String("usdc_addr")
	BusdAddr := ctxCLI.String("busd_addr")
	SupAddr := ctxCLI.String("sup_addr")
	PurchaseAddr := ctxCLI.String("purchase_addr")
	WithdrawAddr := ctxCLI.String("withdraw_addr")
	EthNftAddr := ctxCLI.String("eth_nft_addr")
	EthNftStakingAddr := ctxCLI.String("eth_nft_staking_addr")
	OperatorAddr := ctxCLI.String("operator_addr")
	SignerPrivateKey := ctxCLI.String("signer_private_key")
	BscNodeAddr := ctxCLI.String("bsc_node_addr")
	EthNodeAddr := ctxCLI.String("eth_node_addr")
	BSCChainID := ctxCLI.Int64("bsc_chain_id")
	ETHChainID := ctxCLI.Int64("eth_chain_id")
	BSCRouterAddr := ctxCLI.String("bsc_router_addr")

	isTestnetBlockchain := ctxCLI.Bool("is_testnet_blockchain")
	runBlockchainBridge := ctxCLI.Bool("run_blockchain_bridge")
	fmt.Println("RUN", runBlockchainBridge)
	mailDomain := ctxCLI.String("mail_domain")
	mailAPIKey := ctxCLI.String("mail_apikey")
	mailSender := ctxCLI.String("mail_sender")
	externalURL := ctxCLI.String("passport_web_host_url")
	insecuritySkipVerify := false
	if environment == "development" || environment == "testing" {
		insecuritySkipVerify = true
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
			MoralisKey:   MoralisKey,
			OperatorAddr: common.HexToAddress(OperatorAddr),
			UsdcAddr:     common.HexToAddress(UsdcAddr),
			BusdAddr:     common.HexToAddress(BusdAddr),
			SupAddr:      common.HexToAddress(SupAddr),
			PurchaseAddr: common.HexToAddress(PurchaseAddr),
			WithdrawAddr: common.HexToAddress(WithdrawAddr),

			EthNftAddr:        common.HexToAddress(EthNftAddr),
			EthNftStakingAddr: common.HexToAddress(EthNftStakingAddr),
			SignerPrivateKey:  SignerPrivateKey,
			BscNodeAddr:       BscNodeAddr,
			EthNodeAddr:       EthNodeAddr,
			BSCChainID:        BSCChainID,
			ETHChainID:        ETHChainID,
			BSCRouterAddr:     common.HexToAddress(BSCRouterAddr),
		},
		OnlyWalletConnect:       ctxCLI.Bool("only_wallet"),
		WhitelistEndpoint:       ctxCLI.String("whitelist_check_endpoint"),
		InsecureSkipVerifyCheck: insecuritySkipVerify,
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

	tc := api.NewTransactionCache(txConn, log)

	// initialise user cache map
	ucm := api.NewUserCacheMap(pgxconn)
	err = ucm.Initialise()
	if err != nil {
		return terror.Error(err)
	}

	// API Server
	ctx, cancelOnPanic := context.WithCancel(ctx)
	api := api.NewAPI(log, cancelOnPanic, pgxconn, txConn, googleClientID, mailer, apiAddr, twitchClientID, twitchClientSecret, HTMLSanitizePolicy, config, twitterAPIKey, twitterAPISecret, discordClientID, discordClientSecret, gameserverToken, externalURL, tc, ucm, isTestnetBlockchain, runBlockchainBridge)

	return api.Run(ctx)
}
