//go:build tools && windows
// +build tools,windows

package server

//go:generate go build -o ../bin/migrate.exe -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
//go:generate go build -o ../bin/arelo.exe github.com/makiuchi-d/arelo
//go:generate go build -o ../bin/xcaddy github.com/caddyserver/xcaddy/cmd/xcaddy
//go:generate go build -o ../bin/sqlboiler github.com/volatiletech/sqlboiler/v4
//go:generate go build -o ../bin/sqlboiler-psql github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql

import (
	_ "github.com/caddyserver/caddy/v2"
	_ "github.com/caddyserver/xcaddy/cmd/xcaddy"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/makiuchi-d/arelo"
	_ "github.com/volatiletech/sqlboiler/v4"
	_ "github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql"
)
