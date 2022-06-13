package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"xsyn-services/boiler"

	"github.com/bxcodec/faker/v3"
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

	fillSups := flag.Bool("fill_bot_sups", false, "trigger db to filled 1M sup for bot users")
	botGenNum := flag.Int("bot_gen_number", 0, "generate x amount of bot users on each faction")

	flag.Parse()

	params := url.Values{}
	params.Add("sslmode", "disable")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		"passport",
		"LDs!Eb3z6LDC9axTrtEZWynme",
		"localhost",
		"5454",
		"passport-db",
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
