package db_test

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
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

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"

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

	err = resource.Expire(60) // Tell docker to hard kill the container in 60 seconds
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

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
	b.StopTimer()
	ctx := context.Background()
	transactionChannel := make(chan *Transaction)

	r, err := http.Get(fmt.Sprintf("https://randomuser.me/api/?results=%d&inc=name,email&nat=au,us,gb&noinfo", 50))
	if err != nil {
		b.Fatal()
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
		b.Fatal()
	}
	if len(result.Results) == 0 {
		b.Fatal()
	}

	// Loop over results
	var users []*passport.User
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
			b.Fatal()
		}

		q := `UPDATE users
    			set sups = 1000000000000000000000000
    			where id = $1 `

		// add 100mil sups
		_, err = conn.Exec(ctx, q, u.ID)
		if err != nil {
			b.Fatal()
		}
		users = append(users, u)
	}

	singleConn, err := conn.Acquire(ctx)
	if err != nil {
		b.Fatal()
	}

	go func() {
		// spin up 100 go routines that each send 100 transactions
		for outI := 0; outI < 100; outI++ {
			go func(users []*passport.User, outI int) {
				for inI := 0; inI < 80; inI++ {
					rand.Seed(time.Now().Unix())
					randUser1 := users[rand.Intn(len(users))]
					randUser2 := users[rand.Intn(len(users))]
					for randUser1.ID == randUser2.ID {
						randUser2 = users[rand.Intn(len(users))]
					}
					max := big.NewInt(9223372036854775807)

					random := max.Rand(rand.New(rand.NewSource(time.Now().UnixNano())), max)
					transactionChannel <- &Transaction{
						From:                 randUser1.ID,
						To:                   randUser2.ID,
						Amount:               passport.BigInt{Int: *random},
						TransactionReference: fmt.Sprintf("%d:%d", outI, inI),
					}
				}
			}(users, outI)
		}
	}()

	b.StartTimer()
	b.Log("starting tx")
	count := 5000
	for {
		transaction := <-transactionChannel
		// before we begin the db tx for the transaction, log that we are attempting a transaction (log success defaults to false, we set it to true if it succeeds)
		logID := uuid.UUID{}
		q := `INSERT INTO xsyn_transaction_log (from_id, to_id,  amount, transaction_reference) VALUES($1, $2, $3, $4) RETURNING id`

		err := pgxscan.Get(ctx, conn, &logID, q, transaction.From, transaction.To, transaction.Amount.Int.String(), transaction.TransactionReference) // we can use any pgx connection for this, so we just use the pool
		if err != nil {
			b.Log(err.Error())
			b.Fatal()
		}

		err = func() error {
			// begin the db tx for the transaction
			tx, err := singleConn.Begin(ctx)
			if err != nil {
				return terror.Error(err)
			}
			defer func(tx pgx.Tx, ctx context.Context) {
				err := tx.Rollback(ctx)
				if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
					b.Log(err.Error())
					b.Fatal()
				}
			}(tx, ctx)

			fromUser, err := db.UserGet(ctx, tx, transaction.From)
			if err != nil {
				return terror.Error(err)
			}

			q = `UPDATE users SET sups = sups - $2 WHERE id = $1`

			_, err = tx.Exec(ctx, q, fromUser.ID, transaction.Amount.Int.String())
			if err != nil {
				return terror.Error(err)
			}

			q = `UPDATE users SET sups = sups + $2 WHERE id = $1`

			_, err = tx.Exec(ctx, q, transaction.To, transaction.Amount.Int.String())
			if err != nil {
				return terror.Error(err)
			}

			err = tx.Commit(ctx)
			if err != nil {
				return terror.Error(err)
			}
			return nil
		}()

		if err != nil {
			b.Log(err.Error())
			b.Fatal()
		}
		// update the transaction log to be successful
		q = `UPDATE xsyn_transaction_log SET status = 'success'	WHERE id = $1`

		_, err = conn.Exec(ctx, q, logID)
		if err != nil {
			b.Log(err.Error())
			b.Fatal()
		}
		count--
		if count == 0 {
			singleConn.Release()
			return
		}
	}
}
