package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"xsyn-services/boiler"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/shopspring/decimal"

	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"log"
)

func main() {
	if os.Getenv("PASSPORT_ENVIRONMENT") == "staging" || os.Getenv("PASSPORT_ENVIRONMENT") == "production" {
		log.Fatal("Only works in dev environment")
	}
	fillSups := flag.Bool("fill_sups", false, "trigger db to filled 1M sup for all users")
	botGenNum := flag.Int("bot_gen_number", 0, "generate x amount of bot users on each faction")

	flag.Parse()

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
			UPDATE users set sups = 1000000000000000000000000;
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
					Username:  uuid.Must(uuid.NewV4()).String(),
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
