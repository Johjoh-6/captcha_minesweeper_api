include .env
export

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: run quality control checks
.PHONY: audit
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## upgradeable: list direct dependencies that have upgrades available
.PHONY: upgradeable
upgradeable:
	go run github.com/oligot/go-mod-upgrade@latest

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## db/connect: connect to database
.PHONY: db/connect
db/connect:
	psql "$(DB_DSN)"

## sqlc: generate database code
.PHONY: sqlc
sqlc:
	sqlc generate

## dev/setup: run full local setup
.PHONY: dev/setup
dev/setup: migrations/up sqlc run

## tidy: tidy modfiles and format .go files
.PHONY: tidy
tidy:
	go mod tidy -v
	go fix ./...
	go fmt ./...

## build: build the cmd/api application
.PHONY: build
build:
	@echo 'Building cmd/api...'
		go build -ldflags='-s' -o=./bin/api ./cmd/api
		GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

## run: run the cmd/api application
.PHONY: run
run: build
	./bin/api

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" --build.bin "./bin/api" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"


# ==================================================================================== #
# DATABASE MIGRATIONS
# ==================================================================================== #

DB_DSN ?= postgres://captcha:captcha@localhost:5432/captcha?sslmode=disable

## migrations/new name=$1: create a new database migration
.PHONY: migrations/new
migrations/new:
	goose -dir ./sql/migrations create $(name) sql

## migrations/up: apply all up migrations
.PHONY: migrations/up
migrations/up:
	goose -dir ./sql/migrations postgres "$(DB_DSN)" up

## migrations/down: rollback the last migration
.PHONY: migrations/down
migrations/down:
	goose -dir ./sql/migrations postgres "$(DB_DSN)" down

## migrations/status: show migration status
.PHONY: migrations/status
migrations/status:
	goose -dir ./sql/migrations postgres "$(DB_DSN)" status

## migrations/reset: rollback all migrations
.PHONY: migrations/reset
migrations/reset:
	goose -dir ./sql/migrations postgres "$(DB_DSN)" reset
