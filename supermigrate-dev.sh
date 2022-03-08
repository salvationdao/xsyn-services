#!/bin/bash

./bin/migrate -database "postgres://passport:dev@10.25.26.16:5432/passport-db?sslmode=disable" -path db/migrations up
./bin/migrate -database "postgres://gameserver:dev@10.25.26.16:5432/gameserver-db?sslmode=disable" -path ../supremacy-gameserver/server/db/migrations up


PASSPORT_DATABASE_USER=passport \
PASSPORT_DATABASE_PASS=dev \
PASSPORT_DATABASE_HOST=10.25.26.16 \
PASSPORT_DATABASE_PORT=5432 \
PASSPORT_DATABASE_NAME=passport-db \
PASSPORT_DATABASE_APPLICATION_NAME="API Server" \
go run cmd/platform/main.go s

GAMESERVER_DATABASE_USER=gameserver \
GAMESERVER_DATABASE_PASS=dev \
GAMESERVER_DATABASE_HOST=10.25.26.16 \
GAMESERVER_DATABASE_PORT=5432 \
GAMESERVER_DATABASE_NAME=gameserver-db \
go run cmd/gameserver/main.go s