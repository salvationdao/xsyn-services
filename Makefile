PACKAGE=passport

# Names and Versions
DOCKER_CONTAINER=$(PACKAGE)-db
MIGRATE_VERSION=v4.12.2

# Paths
BIN = $(CURDIR)/bin
SERVER = $(CURDIR)

# DB Settings
LOCAL_DEV_DB_USER=$(PACKAGE)
LOCAL_DEV_DB_PASS=dev
LOCAL_DEV_DB_HOST=localhost
LOCAL_DEV_DB_PORT=5432
LOCAL_DEV_DB_DATABASE=$(PACKAGE)
DB_CONNECTION_STRING="postgres://$(LOCAL_DEV_DB_USER):$(LOCAL_DEV_DB_PASS)@$(LOCAL_DEV_DB_HOST):$(LOCAL_DEV_DB_PORT)/$(LOCAL_DEV_DB_DATABASE)?sslmode=disable"

# Make Commands
.PHONY: build
build:
	go build -o platform cmd/platform/main.go

.PHONY: tools
tools: go-mod-tidy
	@mkdir -p $(BIN)
	go get -u golang.org/x/tools/cmd/goimports
	go generate -tags tools ./tools/...
	# cd $(BIN) && ./xcaddy build
	# sudo setcap cap_net_bind_service=+ep ./bin/caddy

.PHONY: docker-start
docker-start:
	docker start $(DOCKER_CONTAINER) || docker run -d -p $(LOCAL_DEV_DB_PORT):5432 --name $(DOCKER_CONTAINER) -e POSTGRES_USER=$(PACKAGE) -e POSTGRES_PASSWORD=dev -e POSTGRES_DB=$(PACKAGE) postgres:13-alpine

.PHONY: docker-stop
docker-stop:
	docker stop $(DOCKER_CONTAINER)

.PHONY: docker-remove
docker-remove:
	docker rm $(DOCKER_CONTAINER)

.PHONY: docker-setup
docker-setup:
	docker exec -it $(DOCKER_CONTAINER) psql -U $(PACKAGE) -c 'CREATE EXTENSION IF NOT EXISTS pg_trgm; CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'


.PHONY: db-setup
db-setup:
	psql -U postgres -f init.sql


.PHONY: db-version
db-version:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations version

.PHONY: db-drop
db-drop:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations drop -f

.PHONY: db-migrate
db-migrate:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations up

.PHONY: db-migrate-down
db-migrate-down:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations down

.PHONY: db-migrate-down-one
db-migrate-down-one:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations down 1

.PHONY: db-migrate-up-one
db-migrate-up-one:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations up 1


.PHONY: db-seed
db-seed:
	go run cmd/platform/main.go db --seed

.PHONY: db-reset
db-reset: db-drop db-migrate db-seed

.PHONY: go-mod-download
go-mod-download:
	go mod download

.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy

.PHONY: init
init: db-setup deps tools go-mod-tidy db-reset

.PHONY: init-docker
init-docker: docker-start deps tools go-mod-tidy docker-setup db-reset

.PHONY: deps
deps: go-mod-download

.PHONY: serve
serve:
	${BIN}/air -c .air.toml

.PHONY: serve-arelo
serve-arelo:
	${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*' -- go run cmd/platform/main.go serve

.PHONY: lb
lb:
	cd $(BIN) && ./caddy run

.PHONY: wt
wt:
	wt --window 0 --tabColor #4747E2 --title "Boilerplate - Server" -p "PowerShell" -d ./server powershell -NoExit "${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*' -- go run cmd/platform/main.go serve" ; split-pane --tabColor #4747E2 --title "Boilerplate - Load Balancer" -p "PowerShell" -d ./ powershell -NoExit make lb ; split-pane -H -s 0.8 --tabColor #4747E2 --title "Boilerplate - Admin Frontend" --suppressApplicationTitle -p "PowerShell" -d ./web powershell -NoExit "$$env:BROWSER='none' \; npm run admin-start" ; split-pane -H -s 0.5 --tabColor #4747E2 --title "Boilerplate - Public Frontend" --suppressApplicationTitle -p "PowerShell" -d ./web powershell -NoExit "$$env:BROWSER='none' \; npm run public-start"

