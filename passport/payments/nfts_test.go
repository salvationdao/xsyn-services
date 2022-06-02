package payments_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"testing"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var connPool *pgxpool.Pool
var conn *sql.DB

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

		connPool, err = pgxpool.ConnectConfig(ctx, pgxPoolConfig)
		if err != nil {
			passlog.L.Warn().Err(err).Msg("connect to db")
			return terror.Error(err, "")
		}

		// set up normal connPool
		params := url.Values{}
		params.Add("sslmode", "disable")

		cfg, err := pgx.ParseConfig(connString)
		if err != nil {
			return terror.Error(err, "")

		}
		conn = stdlib.OpenDB(*cfg)
		if err != nil {
			return terror.Error(err, "")

		}
		conn.SetMaxIdleConns(100)
		conn.SetMaxOpenConns(500)

		_, err = connPool.Exec(ctx, initSQL)
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

	os.Exit(code)
}

func TestAuth(t *testing.T) {
	const (
		MINTABLE       = "MINTABLE"
		STAKABLE       = "STAKABLE"
		UNSTAKABLE     = "UNSTAKABLE"
		UNSTAKABLE_OLD = "UNSTAKABLE_OLD"
	)
	onChainStatusSlice := []string{
		MINTABLE,
		STAKABLE,
		UNSTAKABLE,
		UNSTAKABLE_OLD,
	}
	var collection *boiler.Collection
	var users []*boiler.User
	var userAssets []*boiler.UserAsset

	t.Run("Insert test collection", func(t *testing.T) {
		collection = &boiler.Collection{
			Name:               "test collection",
			Slug:               "test-collection",
			MintContract:       null.StringFrom("test-mint"),
			StakeContract:      null.StringFrom("test-stake"),
			StakingContractOld: null.StringFrom("test-stake old"),
			IsVisible:          null.BoolFrom(true),
			ContractType:       null.StringFrom("ERC-721"),
		}

		err := collection.Insert(conn, boil.Infer())
		if err != nil {
			t.Fatal("failed to insert col")
		}
	})
	t.Run("Insert Users", func(t *testing.T) {
		amount := 100
		for i := 0; i < amount; i++ {
			userDetails := fmt.Sprintf("%d%d%d%d", i, i, i, i)
			usr := &boiler.User{
				Username:      userDetails,
				PublicAddress: null.StringFrom(userDetails),
			}
			err := usr.Insert(conn, boil.Infer())
			if err != nil {
				t.Fatal("failed to insert users")
			}
			users = append(users, usr)
		}

		userCount, err := boiler.Users().Count(conn)
		if err != nil {
			t.Fatal("failed to get user count")
		}
		if userCount != int64(amount) {
			t.Fatalf("user count wrong, expected: %d, got: %d",int64(amount), userCount)
		}
	})
	t.Run("Insert UserAssets", func(t *testing.T) {
		amount := 100
		for i := 0; i < amount; i++ {
			rand.Seed(time.Now().UnixNano())
			userAssetDeet := fmt.Sprintf("userasset-%d%d%d%d", i, i, i, i)

			usrAss := &boiler.UserAsset{
				CollectionID:  collection.ID,
				TokenID:       int64(i),
				Hash:          userAssetDeet,
				OwnerID:       users[rand.Intn(len(users))].ID,
				Name:          userAssetDeet,
				OnChainStatus: onChainStatusSlice[rand.Intn(len(onChainStatusSlice))],
			}
			err := usrAss.Insert(conn, boil.Infer())
			if err != nil {
				t.Fatal("failed to insert user assets")
			}

			userAssets = append(userAssets, usrAss)
		}

		userAssCount, err := boiler.UserAssets().Count(conn)
		if err != nil {
			t.Fatal("failed to get user count")
		}
		if userAssCount != int64(amount) {
			t.Fatalf("user asset count wrong, expected: %d, got: %d",int64(amount), userAssCount)
		}
	})
  	var nftStatus map[int]*payments.NFTOwnerStatus

	t.Run("UpdateOwners", func(t *testing.T) {
		// seed items
		// TODO: here
		_,_, err := payments.UpdateOwners(nftStatus, collection)
		if err != nil {
			t.Fatal("failed to update owners")
		}


	})
}
