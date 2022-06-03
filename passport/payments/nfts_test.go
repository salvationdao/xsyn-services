package payments_test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/ninja-software/terror/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"testing"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
)

var connPool *pgxpool.Pool

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
		conn := stdlib.OpenDB(*cfg)
		if err != nil {
			return terror.Error(err, "")

		}
		conn.SetMaxIdleConns(100)
		conn.SetMaxOpenConns(500)
		err = passdb.New(conn)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

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
	var collection *boiler.Collection
	var users []*boiler.User
	var userAssets []*boiler.UserAsset

	t.Run("Insert user roles", func(t *testing.T) {
		role := &boiler.Role{
			ID:          "cca82653-c071-4171-92da-05b0808542e7",
			Name:        "Member",
			Tier:        3,
			Reserved:    true,
			Permissions: []string{},
		}

		err := role.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			t.Fatalf("failed to insert role: %s", err.Error())
		}
	})
	t.Run("Insert test collection", func(t *testing.T) {
		collection = &boiler.Collection{
			Name:               "test collection",
			Slug:               "test-collection",
			MintContract:       null.StringFrom(common.HexToAddress(strconv.FormatUint(rand.Uint64(), 16)).Hex()),
			StakeContract:      null.StringFrom(common.HexToAddress(strconv.FormatUint(rand.Uint64(), 16)).Hex()),
			StakingContractOld: null.StringFrom(common.HexToAddress(strconv.FormatUint(rand.Uint64(), 16)).Hex()),
			IsVisible:          null.BoolFrom(true),
			ContractType:       null.StringFrom("ERC-721"),
		}

		err := collection.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			t.Fatalf("failed to insert col: %s", err.Error())
		}
	})
	t.Run("Insert Users", func(t *testing.T) {
		amount := 100
		for i := 0; i < amount; i++ {
			userDetails := fmt.Sprintf("%d%d%d%d%d%d%d%d", i, i, i, i, i, i, i, i)
			usr := &boiler.User{
				Username:      userDetails,
				PublicAddress: null.StringFrom(common.HexToAddress(strconv.FormatUint(rand.Uint64(), 16)).Hex()),
			}
			err := usr.Insert(passdb.StdConn, boil.Infer())
			if err != nil {
				t.Fatalf("failed to insert users: %s", err.Error())
			}
			users = append(users, usr)
		}

		userCount, err := boiler.Users().Count(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to get user count: %s", err.Error())
		}
		if userCount != int64(amount) {
			t.Fatalf("user count wrong, expected: %d, got: %d", int64(amount), userCount)
		}
	})
	t.Run("Insert UserAssets", func(t *testing.T) {
		amount := 100
		for i := 0; i < amount; i++ {
			userAssetDeet := fmt.Sprintf("userasset-%d%d%d%d", i, i, i, i)
			chainStatus := db.MINTABLE
			if i >= 25 {
				chainStatus = db.STAKABLE
			}
			if i >= 50 {
				chainStatus = db.UNSTAKABLE
			}
			if i >= 75 {
				chainStatus = db.UNSTAKABLEOLD
			}

			usrAss := &boiler.UserAsset{
				CollectionID:  collection.ID,
				TokenID:       int64(i),
				Hash:          userAssetDeet,
				OwnerID:       users[i].ID,
				Name:          userAssetDeet,
				OnChainStatus: string(chainStatus),
			}
			err := usrAss.Insert(passdb.StdConn, boil.Infer())
			if err != nil {
				t.Fatalf("failed to insert user assets: %s", err.Error())
			}

			userAssets = append(userAssets, usrAss)
		}

		userAssCount, err := boiler.UserAssets().Count(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to get user count: %s", err.Error())
		}
		if userAssCount != int64(amount) {
			t.Fatalf("user asset count wrong, expected: %d, got: %d", int64(amount), userAssCount)
		}

		userAssCountMintable, err := boiler.UserAssets(
			boiler.UserAssetWhere.OnChainStatus.EQ(string(db.MINTABLE)),
		).Count(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to get user mintable count: %s", err.Error())
		}
		if userAssCountMintable != int64(25) {
			t.Fatalf("user asset count wrong, expected: %d, got: %d", int64(25), userAssCountMintable)
		}

		userAssCountStakable, err := boiler.UserAssets(
			boiler.UserAssetWhere.OnChainStatus.EQ(string(db.STAKABLE)),
		).Count(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to get user stakable count: %s", err.Error())
		}
		if userAssCountStakable != int64(25) {
			t.Fatalf("user asset count wrong, expected: %d, got: %d", int64(25), userAssCountStakable)
		}

		userAssCountUnstakable, err := boiler.UserAssets(
			boiler.UserAssetWhere.OnChainStatus.EQ(string(db.UNSTAKABLE)),
		).Count(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to get user unstakable count: %s", err.Error())
		}
		if userAssCountUnstakable != int64(25) {
			t.Fatalf("user asset count wrong, expected: %d, got: %d", int64(25), userAssCountUnstakable)
		}

		userAssCountUnstakableOld, err := boiler.UserAssets(
			boiler.UserAssetWhere.OnChainStatus.EQ(string(db.UNSTAKABLEOLD)),
		).Count(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to get user unstakable old count: %s", err.Error())
		}
		if userAssCountUnstakableOld != int64(25) {
			t.Fatalf("user asset count wrong, expected: %d, got: %d", int64(25), userAssCountUnstakableOld)
		}
	})
	t.Run("UpdateOwners mintable to stakable", func(t *testing.T) {
		mintableToStakable := map[int]*payments.NFTOwnerStatus{}

		firstAssetOwner := common.HexToAddress(users[0].PublicAddress.String)
		firstAssetOwnerID := users[0].ID
		secondAssetOwner := common.HexToAddress(users[1].PublicAddress.String)
		secondAssetOwnerID := users[1].ID

		mintableToStakable[0] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         firstAssetOwner,
			OnChainStatus: db.STAKABLE,
		}
		mintableToStakable[1] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         secondAssetOwner,
			OnChainStatus: db.STAKABLE,
		}

		success, fail, err := payments.UpdateOwners(mintableToStakable, collection)
		if err != nil {
			t.Fatalf("failed to update owners: %s", err.Error())
		}

		if fail > 0 {
			t.Fatalf("mintableToStakable UpdateOwners failed wrong, expected: %d, got: %d", 0, fail)
		}

		if success != 2 {
			t.Fatalf("mintableToStakable UpdateOwners success wrong, expected: %d, got: %d", 2, success)
		}

		// get items and check status
		err = userAssets[0].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 1: %s", err.Error())
		}
		err = userAssets[1].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 2: %s", err.Error())
		}

		// check owner id is right
		if userAssets[0].OwnerID != firstAssetOwnerID {
			t.Fatalf("asset 1 owner changed, expected: %s, got: %s", firstAssetOwnerID, userAssets[0].OwnerID)
		}
		if userAssets[1].OwnerID != secondAssetOwnerID {
			t.Fatalf("asset 2 owner changed, expected: %s, got: %s", secondAssetOwnerID, userAssets[1].OwnerID)
		}
		// check on chain status is right
		if userAssets[0].OnChainStatus != string(db.STAKABLE) {
			t.Fatalf("asset 1 wrong on chain status, expected: %s, got: %s", string(db.STAKABLE), userAssets[0].OnChainStatus)
		}
		if userAssets[1].OnChainStatus != string(db.STAKABLE) {
			t.Fatalf("asset 2 wrong on chain status, expected: %s, got: %s", string(db.STAKABLE), userAssets[1].OnChainStatus)
		}

	})
	t.Run("UpdateOwners stakable to unstakable", func(t *testing.T) {
		stakableToUnstakable := map[int]*payments.NFTOwnerStatus{}

		firstAssetIndex := 25
		secondAssetIndex := 26
		firstAssetOwner := common.HexToAddress(users[firstAssetIndex].PublicAddress.String)
		firstAssetOwnerID := users[firstAssetIndex].ID
		secondAssetOwner := common.HexToAddress(users[secondAssetIndex].PublicAddress.String)
		secondAssetOwnerID := users[secondAssetIndex].ID

		stakableToUnstakable[firstAssetIndex] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         firstAssetOwner,
			OnChainStatus: db.UNSTAKABLE,
		}
		stakableToUnstakable[secondAssetIndex] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         secondAssetOwner,
			OnChainStatus: db.UNSTAKABLE,
		}

		success, fail, err := payments.UpdateOwners(stakableToUnstakable, collection)
		if err != nil {
			t.Fatalf("failed to update owners: %s", err.Error())
		}

		if fail > 0 {
			t.Fatalf("stakableToUnstakable UpdateOwners failed wrong, expected: %d, got: %d", 0, fail)
		}

		if success != 2 {
			t.Fatalf("stakableToUnstakable UpdateOwners success wrong, expected: %d, got: %d", 2, success)
		}

		// get items and check status
		err = userAssets[firstAssetIndex].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 1: %s", err.Error())
		}
		err = userAssets[secondAssetIndex].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 2: %s", err.Error())
		}

		// check owner id is right
		if userAssets[firstAssetIndex].OwnerID != firstAssetOwnerID {
			t.Fatalf("asset 1 owner changed, expected: %s, got: %s", firstAssetOwnerID, userAssets[firstAssetIndex].OwnerID)
		}
		if userAssets[secondAssetIndex].OwnerID != secondAssetOwnerID {
			t.Fatalf("asset 2 owner changed, expected: %s, got: %s", secondAssetOwnerID, userAssets[secondAssetIndex].OwnerID)
		}
		// check on chain status is right
		if userAssets[firstAssetIndex].OnChainStatus != string(db.UNSTAKABLE) {
			t.Fatalf("asset 1 wrong on chain status, expected: %s, got: %s", string(db.UNSTAKABLE), userAssets[firstAssetIndex].OnChainStatus)
		}
		if userAssets[secondAssetIndex].OnChainStatus != string(db.UNSTAKABLE) {
			t.Fatalf("asset 2 wrong on chain status, expected: %s, got: %s", string(db.UNSTAKABLE), userAssets[secondAssetIndex].OnChainStatus)
		}

	})
	t.Run("UpdateOwners unstakable to stakable", func(t *testing.T) {
		unstakableToStakable := map[int]*payments.NFTOwnerStatus{}

		firstAssetIndex := 50
		secondAssetIndex := 51
		firstAssetOwner := common.HexToAddress(users[firstAssetIndex].PublicAddress.String)
		firstAssetOwnerID := users[firstAssetIndex].ID
		secondAssetOwner := common.HexToAddress(users[secondAssetIndex].PublicAddress.String)
		secondAssetOwnerID := users[secondAssetIndex].ID

		unstakableToStakable[firstAssetIndex] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         firstAssetOwner,
			OnChainStatus: db.STAKABLE,
		}
		unstakableToStakable[secondAssetIndex] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         secondAssetOwner,
			OnChainStatus: db.STAKABLE,
		}

		success, fail, err := payments.UpdateOwners(unstakableToStakable, collection)
		if err != nil {
			t.Fatalf("failed to update owners: %s", err.Error())
		}

		if fail > 0 {
			t.Fatalf("unstakableToStakable UpdateOwners failed wrong, expected: %d, got: %d", 0, fail)
		}

		if success != 2 {
			t.Fatalf("unstakableToStakable UpdateOwners success wrong, expected: %d, got: %d", 2, success)
		}

		// get items and check status
		err = userAssets[firstAssetIndex].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 1: %s", err.Error())
		}
		err = userAssets[secondAssetIndex].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 2: %s", err.Error())
		}

		// check owner id is right
		if userAssets[firstAssetIndex].OwnerID != firstAssetOwnerID {
			t.Fatalf("asset 1 owner changed, expected: %s, got: %s", firstAssetOwnerID, userAssets[firstAssetIndex].OwnerID)
		}
		if userAssets[secondAssetIndex].OwnerID != secondAssetOwnerID {
			t.Fatalf("asset 2 owner changed, expected: %s, got: %s", secondAssetOwnerID, userAssets[secondAssetIndex].OwnerID)
		}
		// check on chain status is right
		if userAssets[firstAssetIndex].OnChainStatus != string(db.STAKABLE) {
			t.Fatalf("asset 1 wrong on chain status, expected: %s, got: %s", string(db.STAKABLE), userAssets[firstAssetIndex].OnChainStatus)
		}
		if userAssets[secondAssetIndex].OnChainStatus != string(db.STAKABLE) {
			t.Fatalf("asset 2 wrong on chain status, expected: %s, got: %s", string(db.STAKABLE), userAssets[secondAssetIndex].OnChainStatus)
		}
	})
	t.Run("UpdateOwners on chain transfer", func(t *testing.T) {
		onChainTransfer := map[int]*payments.NFTOwnerStatus{}

		assetIndex := 27
		assetOldOwnerID := userAssets[assetIndex].OwnerID
		assetNewOwnerID := users[80].ID
		assetNewOwnerPublicAddress := users[80].PublicAddress

		onChainTransfer[assetIndex] = &payments.NFTOwnerStatus{
			Collection:    common.HexToAddress(collection.MintContract.String),
			Owner:         common.HexToAddress(assetNewOwnerPublicAddress.String),
			OnChainStatus: db.STAKABLE,
		}

		success, fail, err := payments.UpdateOwners(onChainTransfer, collection)
		if err != nil {
			t.Fatalf("failed to update owners: %s", err.Error())
		}

		if fail > 0 {
			t.Fatalf("unstakableToStakable UpdateOwners failed wrong, expected: %d, got: %d", 0, fail)
		}

		if success != 1 {
			t.Fatalf("unstakableToStakable UpdateOwners success wrong, expected: %d, got: %d", 1, success)
		}

		// get items and check status
		err = userAssets[assetIndex].Reload(passdb.StdConn)
		if err != nil {
			t.Fatalf("failed to reload asset 1: %s", err.Error())
		}

		// check owner id is right
		if userAssets[assetIndex].OwnerID != assetNewOwnerID {
			t.Fatalf("asset 1 owner changed, expected: %s, got: %s, old: %s", assetNewOwnerID, userAssets[assetIndex].OwnerID, assetOldOwnerID)
		}

		// check on chain status is right
		if userAssets[assetIndex].OnChainStatus != string(db.STAKABLE) {
			t.Fatalf("asset 1 wrong on chain status, expected: %s, got: %s", string(db.STAKABLE), userAssets[assetIndex].OnChainStatus)
		}
	})
	t.Run("OwnerRecordToOwnerStatus", func(t *testing.T) {
		from1 :=  common.HexToAddress(strconv.FormatUint(rand.Uint64(), 16)).Hex()
		from2 :=  common.HexToAddress(strconv.FormatUint(rand.Uint64(), 16)).Hex()

		records := []*payments.NFTOwnerRecord{
			{
				FromAddress:from1,
				ToAddress:   common.HexToAddress(collection.StakeContract.String).Hex(),
				TokenID:     0,
			},
			{
				FromAddress: from2,
				ToAddress:   common.HexToAddress(collection.StakingContractOld.String).Hex(),
				TokenID:     1,
			},
		}

		results := payments.OwnerRecordToOwnerStatus(records, collection)
		// check owner hasnt changed
		if results[0].Owner.Hex() != from1 {
			t.Fatalf("owner changed, expected: %s, got %s", from1, results[0].Owner.Hex())
		}
		if results[1].Owner.Hex() != from2 {
			t.Fatalf("owner changed, expected: %s, got %s", from2, results[1].Owner.Hex())
		}

		if results[0].OnChainStatus != db.UNSTAKABLE {
			t.Fatalf("wrong on chain status returned for token: %d, expected: %s, got %s", 0, db.UNSTAKABLE, results[0].OnChainStatus)
		}
		if results[1].OnChainStatus != db.UNSTAKABLEOLD {
			t.Fatalf("wrong on chain status returnedfor token: %d, expected: %s, got %s", 1, db.UNSTAKABLEOLD, results[1].OnChainStatus)
		}
	})
}
