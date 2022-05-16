package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"log"
)

func main() {
	if os.Getenv("PASSPORT_ENVIRONMENT") == "staging" || os.Getenv("PASSPORT_ENVIRONMENT") == "production" {
		log.Fatal("Only works in dev environment")
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
	result, err := conn.Exec(`
		UPDATE users set sups = 1000000000000000000000000;
	`)
	if err != nil {
		log.Fatal(err)
	}

	number, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(number, "users' sups is updated")
}
