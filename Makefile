PACKAGE=passport

# Names and Versions
DOCKER_CONTAINER=$(PACKAGE)-db
MIGRATE_VERSION=v4.12.2

# Paths
BIN = $(CURDIR)/bin
SERVER = $(CURDIR)

# DB Settings
LOCAL_DEV_DB_USER?=$(PACKAGE)
LOCAL_DEV_DB_PASS?=dev
LOCAL_DEV_DB_TX_USER?=$(PACKAGE)_tx
LOCAL_DEV_DB_TX_PASS?=dev-tx
LOCAL_DEV_DB_HOST?=localhost
LOCAL_DEV_DB_PORT?=5432
LOCAL_DEV_DB_DATABASE?=$(PACKAGE)
DB_CONNECTION_STRING="postgres://$(LOCAL_DEV_DB_USER):$(LOCAL_DEV_DB_PASS)@$(LOCAL_DEV_DB_HOST):$(LOCAL_DEV_DB_PORT)/$(LOCAL_DEV_DB_DATABASE)?sslmode=disable"

GITVERSION=`git describe --tags --abbrev=0`
GITVERSION_CHECK := $(shell git describe --tags --abbrev=0  || echo fail)
GITBRANCH=`git rev-parse --abbrev-ref HEAD`
GITHASH=`git rev-parse HEAD`
BUILDDATE=`date -u +%Y%m%d%H%M%S`
GITSTATE=`git status --porcelain | wc -l`
REPO_ROOT=`git rev-parse --show-toplevel`

# Make Commands
.PHONY: setup-git
setup-git:
	ln -s ${REPO_ROOT}/.pre-commit ${REPO_ROOT}/.git/hooks/pre-commit

.PHONY: clean
clean:
	rm -rf deploy

.PHONY: deploy-package
deploy-prep: clean tools build
	mkdir -p deploy
	cp $(BIN)/migrate deploy/.
	cp -r ./init deploy/.
	cp -r ./configs deploy/.
	cp -r ./asset deploy/.
	cp -r $(CURDIR)/migrations deploy/.

define BUILD_ERROR_GIT_VER

	Can not contiue, there are no git tags set, 
	Tag the current version, I.E.
	git tag 0.0.0-dev 

endef

.PHONY: build
build:
ifeq ($(GITVERSION_CHECK), fail)
	$(error $(BUILD_ERROR_GIT_VER))
endif
	cd $(SERVER) && go build \
		-ldflags " -X main.Version=${GITVERSION} -X main.GitHash=${GITHASH} -X main.GitBranch=${GITBRANCH} -X main.BuildDate=${BUILDDATE} -X main.UnCommittedFiles=$(GITSTATE)" \
		-o deploy/passport-api \
		cmd/platform/main.go


.PHONY: tools-darwin
tools-darwin: go-mod-tidy
	@mkdir -p $(BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.43.0 go get -u golang.org/x/tools/cmd/goimports
	go generate -tags tools ./tools/...

.PHONY: tools
tools: go-mod-tidy
	@mkdir -p $(BIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.43.0 go get -u golang.org/x/tools/cmd/goimports
	go generate -tags tools ./tools/...

.PHONY: tools-windows
tools-windows: go-mod-tidy
	-mkdir bin
	cd $(BIN) && powershell Invoke-WebRequest -Uri "https://github.com/golangci/golangci-lint/releases/download/v1.43.0/golangci-lint-1.43.0-windows-amd64.zip" -OutFile "./golangci-lint.zip" -UseBasicParsing
	cd $(BIN) && powershell Expand-Archive './golangci-lint.zip' -DestinationPath './'\
	 && powershell rm './golangci-lint.zip'\
	 && powershell mv ./golangci-lint-1.43.0-windows-amd64/golangci-lint.exe ./golangci-lint.exe\
	 && powershell rm -r ./golangci-lint-1.43.0-windows-amd64
	go generate -tags tools ./tools/...


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
	docker exec -it $(DOCKER_CONTAINER)\
 	psql -U $(LOCAL_DEV_DB_USER) -c\
 	"CREATE USER $(LOCAL_DEV_DB_TX_USER) WITH ENCRYPTED PASSWORD '$(LOCAL_DEV_DB_TX_PASS)';\
	CREATE EXTENSION IF NOT EXISTS pg_trgm;\
	CREATE EXTENSION IF NOT EXISTS pgcrypto;\
	CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"


.PHONY: db-setup
db-setup:
	psql -h $(LOCAL_DEV_DB_HOST) -p $(LOCAL_DEV_DB_PORT) -U postgres -f init.sql


.PHONY: db-version
db-version:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations version

.PHONY: db-drop
db-drop:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations drop -f

.PHONY: db-migrate
db-migrate:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations up

.PHONY: db-migrate-down
db-migrate-down:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations down

.PHONY: db-migrate-down-one
db-migrate-down-one:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations down 1

.PHONY: db-migrate-up-one
db-migrate-up-one:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations up 1

# make sure `make tools` is done
.PHONY: db-boiler
db-boiler:
	$(BIN)/sqlboiler $(BIN)/sqlboiler-psql --wipe --config sqlboiler.toml

.PHONY: db-seed
db-seed:
	go run seed/main.go db

# targeting file 20220705053059_player_syndicate_table.up
.PHONY: db-migrate-before-syndicate-table
db-migrate-before-syndicate-table:
	$(BIN)/migrate -database $(DB_CONNECTION_STRING) -path $(CURDIR)/migrations up 55

.PHONY: db-reset
db-reset: db-drop db-migrate-before-syndicate-table db-seed db-migrate db-boiler

.PHONY: db-reset-windows
db-reset-windows: db-drop db-migrate-before-syndicate-table db-seed db-migrate

.PHONY: go-mod-download
go-mod-download:
	go mod download

.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy -compat=1.18

.PHONY: init
init: db-setup deps tools go-mod-tidy db-reset

.PHONY: init-docker
init-docker: docker-start tools go-mod-tidy docker-setup db-reset

.PHONY: deps
deps: go-mod-download

.PHONY: serve
serve:
	${BIN}/air -c ./passport/.air.toml

# TODO: add linter with arelo
.PHONY: serve-arelo
serve-arelo:
	${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*'  -- go run passport/main.go serve

.PHONY: test
test:
	echo "first" && echo "second"

.PHONY: sync
sync:
	go run passport/main.go sync

.PHONY: lb
lb:
	cd $(BIN) && ./caddy run

.PHONY: wt
wt:
	wt --window 0 --tabColor #4747E2 --title "Boilerplate - Server" -p "PowerShell" -d ./server powershell -NoExit "${BIN}/arelo -p '**/*.go' -i '**/.*' -i '**/*_test.go' -i 'tools/*' -- go run cmd/platform/main.go serve" ; split-pane --tabColor #4747E2 --title "Boilerplate - Load Balancer" -p "PowerShell" -d ./ powershell -NoExit make lb ; split-pane -H -s 0.8 --tabColor #4747E2 --title "Boilerplate - Admin Frontend" --suppressApplicationTitle -p "PowerShell" -d ./web powershell -NoExit "$$env:BROWSER='none' \; npm run admin-start" ; split-pane -H -s 0.5 --tabColor #4747E2 --title "Boilerplate - Public Frontend" --suppressApplicationTitle -p "PowerShell" -d ./web powershell -NoExit "$$env:BROWSER='none' \; npm run public-start"

docker-db-dump:
	mkdir -p ./tmp
	docker exec -it ${DOCKER_CONTAINER} /usr/local/bin/pg_dump -U ${LOCAL_DEV_DB_USER} > tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONY: docker-db-restore
docker-db-restore:
	ifeq ("$(wildcard tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql)","")
		$(error tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql is missing restore will fail)
	endif
		docker exec -it ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid <> pg_backend_pid() AND datname = '${LOCAL_DEV_DB_DATABASE}'"
		docker exec -it ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "DROP DATABASE $(LOCAL_DEV_DB_DATABASE)"
		docker exec -i  ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d postgres < init.sql
		docker exec -i  ${DOCKER_CONTAINER} /usr/local/bin/psql -U ${LOCAL_DEV_DB_USER} -d $(LOCAL_DEV_DB_DATABASE) < tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONY: db-dump
db-dump:
	mkdir -p ./tmp
	pg_dump -U ${LOCAL_DEV_DB_USER} > tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

.PHONE: db-restore
db-restore:
	ifeq ("$(wildcard tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql)","")
		$(error tmp/$(LOCAL_DEV_DB_DATABASE)_dump.sql is missing restore will fail)
	endif
		psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid <> pg_backend_pid() AND datname = '${LOCAL_DEV_DB_DATABASE}'"
		psql -U ${LOCAL_DEV_DB_USER} -d postgres -c "DROP DATABASE $(LOCAL_DEV_DB_DATABASE)"
		psql -U ${LOCAL_DEV_DB_USER} -d postgres < init.sql
		psql -U ${LOCAL_DEV_DB_USER} -d $(LOCAL_DEV_DB_DATABASE) < tmp/${LOCAL_DEV_DB_DATABASE}_dump.sql

dev_tool_fill_bot_sups:
	go run ./passport/devtool/*.go -fill_bot_sups

dev_tool_gen_bot_100:
	go run ./passport/devtool/*.go -bot_gen_number=100

.PHONE: dev-give-mech
dev-give-mech:
	curl -i -H "X-Authorization: NinjaDojo_!" -k https://api.xsyndev.io/api/dev/give-mechs/${public_address}

.PHONE: dev-give-mechs
dev-give-mechs:
	make dev-give-mech public_address=0x04c2C035F908C73FBff4FAf3818824170C938640
