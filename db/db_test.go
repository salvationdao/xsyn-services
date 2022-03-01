package db_test

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"passport"
	"passport/db"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/ory/dockertest/v3/docker"

	"github.com/ninja-software/terror/v2"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/ory/dockertest/v3"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

var conn *pgxpool.Pool
var txConn *sql.DB

//go:embed migrations
var migrations embed.FS

func TestMain(m *testing.M) {
	fmt.Println("Spinning up docker container for postgres...")

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	user := "test"
	password := "dev"
	dbName := "test"

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

	err = resource.Expire(300) // Tell docker to hard kill the container in 60 seconds
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		ctx := context.Background()

		txConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			"localhost",
			resource.GetPort("5432/tcp"),
			dbName,
		)
		txConn, err = sql.Open("postgres", txConnString)

		connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			"localhost",
			resource.GetPort("5432/tcp"),
			dbName,
		)

		pgxPoolConfig, err := pgxpool.ParseConfig(connString)
		if err != nil {
			return terror.Error(err, "")
		}

		pgxPoolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace

		conn, err = pgxpool.ConnectConfig(ctx, pgxPoolConfig)
		if err != nil {
			return terror.Error(err, "")
		}

		fmt.Println("Running Migration...")

		source, err := httpfs.New(http.FS(migrations), "migrations")
		if err != nil {
			log.Fatal(err)
		}

		mig, err := migrate.NewWithSourceInstance("embed", source, connString)
		if err != nil {
			log.Fatal(err)
		}
		if err := mig.Up(); err != nil {
			log.Fatal(err)
		}
		source.Close()

		fmt.Println("Postgres Ready.")

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	fmt.Println("Running tests...")

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

type Transaction struct {
	From                 passport.UserID `json:"from"`
	To                   passport.UserID `json:"to"`
	Amount               passport.BigInt `json:"amount"`
	TransactionReference string          `json:"transactionReference"`
}

func BenchmarkTransactions(b *testing.B) {
	ctx := context.Background()
	r, err := http.Get(fmt.Sprintf("https://randomuser.me/api/?results=%d&inc=name,email&nat=au,us,gb&noinfo", 50))
	if err != nil {
		log.Fatal()
	}

	var result struct {
		Results []struct {
			Name struct {
				First string
				Last  string
			}
			Email string
		}
	}

	// Decode json
	err = json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		log.Fatal()
	}
	r.Body.Close()
	if len(result.Results) == 0 {
		log.Fatal()
	}

	// insert user that can go negative (offchain user)
	offChainUser := &passport.User{ID: passport.OnChainUserID, RoleID: passport.UserRoleOffChain, Username: passport.OnChainUsername, Verified: true}
	err = db.InsertSystemUser(ctx, conn, offChainUser)
	if err != nil {
		log.Fatal()
	}

	var userList []*passport.User
	// Loop over results
	for _, result := range result.Results {

		// Create user
		u := &passport.User{
			FirstName: result.Name.First,
			LastName:  result.Name.Last,
			Email:     passport.NewString(result.Email),
		}

		u.Username = fmt.Sprintf("%s%s", u.FirstName, u.LastName)

		// Insert
		err = db.UserCreate(ctx, conn, u)
		if err != nil {
			log.Fatal()
		}

		userList = append(userList, u)
	}

	transactionChannel := make(chan *passport.NewTransaction)
	amountToTransfer := big.NewInt(7223372036854775807)
	go func() {
		// spin up 100 go routines that each send 100 transactions
		for outI := 0; outI < 1000; outI++ {
			go func(outI int) {
				for inI := 0; inI < 80; inI++ {
					rand.Seed(time.Now().Unix())
					randUser1 := userList[rand.Intn(len(userList))]

					transactionChannel <- &passport.NewTransaction{
						From:                 offChainUser.ID,
						To:                   randUser1.ID,
						Amount:               *amountToTransfer,
						NotSafe: true,
						TransactionReference: passport.TransactionReference(fmt.Sprintf("%d:%d", outI, inI)),
					}
				}
			}(outI)
		}
	}()
}
