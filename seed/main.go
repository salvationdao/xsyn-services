package main

import (
	"database/sql"
	"embed"
	"net/url"
	"time"

	"github.com/jackc/pgx/v4/stdlib"

	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"

	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

//go:embed *.sql
var SeedSQL embed.FS

// Variable passed in at compile time using `-ldflags`
var (
	Version          string // -X main.Version=$(git describe --tags --abbrev=0)
	GitHash          string // -X main.GitHash=$(git rev-parse HEAD)
	GitBranch        string // -X main.GitBranch=$(git rev-parse --abbrev-ref HEAD)
	BuildDate        string // -X main.BuildDate=$(date -u +%Y%m%d%H%M%S)
	UnCommittedFiles string // -X main.UnCommittedFiles=$(git status --porcelain | wc -l)"
)

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
				// This is not using the built-in version so ansible can more easily read the version
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
				Name: "db",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "database_user", Value: "passport", EnvVars: []string{"PASSPORT_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{"PASSPORT_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{"PASSPORT_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5432", EnvVars: []string{"PASSPORT_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "passport", EnvVars: []string{"PASSPORT_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{"PASSPORT_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},
					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{"PASSPORT_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
				},
				Usage: "seed the database",
				Action: func(c *cli.Context) error {
					dbConn, err := sqlConnect(
						c.String("database_user"),
						c.String("database_pass"),
						c.String("database_host"),
						c.String("database_port"),
						c.String("database_name"),
						50,
						50,
					)
					if err != nil {
						return err
					}

					return Seed(dbConn, SeedSQL)
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

func sqlConnect(
	databaseTxUser string,
	databaseTxPass string,
	databaseHost string,
	databasePort string,
	databaseName string,
	maxIdle int,
	maxOpen int,
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
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	conn := stdlib.OpenDB(*cfg)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(maxIdle)
	conn.SetMaxOpenConns(maxOpen)
	return conn, nil

}

// Seed seeds database
func Seed(conn *sql.DB, seedSQL embed.FS) error {
	seedFiles := []string{
		"01_roles.sql",
		"02_blobs.sql",
		"03_factions.sql",
		"04_collections.sql",
		"05_users.sql",
		"06_store_items.sql",
		"07_purchased_items.sql",
		"08_api_keys.sql",
	}

	for _, file := range seedFiles {
		fileData, err := seedSQL.ReadFile(file)
		if err != nil {
			return err
		}
		_, err = conn.Exec(string(fileData))
		if err != nil {
			return err
		}
	}

	return nil
}
