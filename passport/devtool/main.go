package main

import (
	"fmt"
	"net/url"
	"os"
	"time"
	"xsyn-services/boiler"

	"github.com/bxcodec/faker/v3"
	"github.com/ninja-software/terror/v2"
	"github.com/urfave/cli/v2"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/volatiletech/null/v8"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"log"
)

const envPrefix = "PASSPORT"

func main() {
	if os.Getenv("PASSPORT_ENVIRONMENT") == "production" {
		log.Fatal("Only works in dev and staging environment")

	}

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
					&cli.StringFlag{Name: "database_user", Value: "passport", EnvVars: []string{envPrefix + "_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{envPrefix + "_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{envPrefix + "_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5432", EnvVars: []string{envPrefix + "_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "passport", EnvVars: []string{envPrefix + "_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{envPrefix + "_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},
				},

				Usage: "run server",
				Action: func(c *cli.Context) error {
					params := url.Values{}
					params.Add("sslmode", "disable")

					connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
						"passport",
						"dev",
						"localhost",
						"5432",
						"passport",
						params.Encode(),
					)
					cfg, err := pgx.ParseConfig(connString)
					if err != nil {
						log.Fatal(err)
					}

					conn := stdlib.OpenDB(*cfg)
					if err != nil {
						log.Fatal(err)
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

	params := url.Values{}
	params.Add("sslmode", "disable")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		"passport",
		"dev",
		"localhost",
		"5432",
		"passport",
		params.Encode(),
	)
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		log.Fatal(err)
	}
	conn := stdlib.OpenDB(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	if fillSups != nil && *fillSups {
		result, err := conn.Exec(
			`
			UPDATE users u set sups = 1000000000000000000000000
			WHERE EXISTS (SELECT 1 FROM roles r WHERE r.id = u.role_id AND r.name = 'Bot');
			`,
		)
		if err != nil {
			log.Fatal(err)
		}

		number, err := result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(number, "users' sups is updated")
	}

	// generate bot user
	if botGenNum != nil && *botGenNum > 0 {
		fmt.Println(*botGenNum)

		// get faction
		factions, err := boiler.Factions().All(conn)
		if err != nil {
			log.Fatal(err)
		}

		// get bot role
		botRole, err := boiler.Roles(boiler.RoleWhere.Name.EQ("Bot")).One(conn)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := conn.Begin()
		if err != nil {
			log.Fatal(err)
		}

		defer tx.Rollback()

		for _, faction := range factions {
			i := 0
			for i < *botGenNum {
				i++

				user := boiler.User{
					Username:  faker.Name(),
					FactionID: null.StringFrom(faction.ID),
					RoleID:    null.StringFrom(botRole.ID),
					Verified:  true,
					Sups:      decimal.New(1000000, 18), // 1M sups
				}

				err = user.Insert(tx, boil.Infer())
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("generated", *botGenNum, "bots for each faction")
	}
}
