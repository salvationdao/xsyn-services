package client_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"xsyn-services/passport/passlog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var conn *pgxpool.Pool

const initSQL = `
CREATE USER passport WITH ENCRYPTED PASSWORD 'dev';
GRANT ALL PRIVILEGES ON DATABASE passport TO passport;
CREATE USER passport_tx WITH ENCRYPTED PASSWORD 'dev-tx';
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
`

func TestMain(m *testing.M) {
	passlog.New("testing", "TraceLevel")
	passlog.L.Info().Msg("Spinning up docker container for postgres...")

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	user := "dev"
	password := "dev"
	dbName := "passport"

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13-alpine",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(60) // Tell docker to hard kill the container in 60 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		ctx := context.Background()
		connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			"localhost",
			resource.GetPort("5432/tcp"),
			dbName,
		)

		pgxPoolConfig, err := pgxpool.ParseConfig(connString)
		if err != nil {
			passlog.L.Err(err).Msg("parse config")
			return terror.Error(err, "")
		}

		pgxPoolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace

		conn, err = pgxpool.ConnectConfig(ctx, pgxPoolConfig)
		if err != nil {
			passlog.L.Warn().Err(err).Msg("connect to db")
			return terror.Error(err, "")
		}
		_, err = conn.Exec(ctx, initSQL)
		if err != nil {
			passlog.L.Err(err).Msg("setup roles")
			return terror.Error(err, "")
		}

		passlog.L.Info().Msg("running migrations")

		mig, err := migrate.New("file://../../migrations", connString)
		if err != nil {
			log.Fatal(err)
		}
		if err := mig.Up(); err != nil {
			log.Fatal(err)
		}

		passlog.L.Info().Msg("db is ready")

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	passlog.L.Info().Msg("running tests")
	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	// TODO: Figure out how to spinup api server
	// api, routes := api.NewAPI(log,
	// 	pgxconn,
	// 	txConn,
	// 	mailer,
	// 	twilio,
	// 	apiAddr,
	// 	HTMLSanitizePolicy,
	// 	config,
	// 	externalURL,
	// 	ucm,
	// 	isTestnetBlockchain,
	// 	runBlockchainBridge,
	// 	msgBus,
	// 	enablePurchaseSubscription,
	// 	jwtKeyByteArray,
	// )

	os.Exit(code)
}

func TestAuth(t *testing.T) {
	t.Run("user can login", func(t *testing.T) {})
	t.Run("user can logout", func(t *testing.T) {})
}

func TestProfile(t *testing.T) {
	t.Run("user can view profile", func(t *testing.T) {})
	t.Run("user can change username", func(t *testing.T) {})
}

func TestPurchases(t *testing.T) {
	t.Run("user can list collections", func(t *testing.T) {})
	t.Run("user can list store items for a collection", func(t *testing.T) {})
	t.Run("user can purchase remaining lootbox items", func(t *testing.T) {})
	t.Run("user can purchase remaining megas items", func(t *testing.T) {})
	t.Run("user can list their purchased items", func(t *testing.T) {})
	t.Run("user can view a purchased item", func(t *testing.T) {})
}

func TestChainOperations(t *testing.T) {
	t.Run("user can withdraw SUPS", func(t *testing.T) {})
	t.Run("user can deposit SUPS", func(t *testing.T) {})
	t.Run("user can purchase SUPS", func(t *testing.T) {})
}

func TestTransactions(t *testing.T) {
	t.Run("user can list their own transactions", func(t *testing.T) {})
	t.Run("user can filter transactions by group and subgroup", func(t *testing.T) {})
}
