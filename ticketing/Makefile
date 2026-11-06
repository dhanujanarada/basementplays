SHELL := /bin/sh
GO ?= go

.PHONY: help dev migrate seed test test-race test-integration test-integration-race lint fmt tidy db-create db-drop psql docker-build

help:
	@printf '%s\n' \
		'dev              run the API using ticketing/.env' \
		'migrate          apply pending SQL migrations' \
		'seed             create the placeholder sample-50 seat plan' \
		'test             run unit tests' \
		'test-race        run unit tests with the race detector' \
		'test-integration run Postgres integration tests (needs TEST_DATABASE_URL)' \
		'test-integration-race run Postgres integration tests under -race' \
		'lint             go vet + gofmt check' \
		'fmt              gofmt -w .' \
		'tidy             go mod tidy' \
		'db-create        create the ticketing_dev database if missing' \
		'db-drop          drop the ticketing_dev database' \
		'psql             open psql on ticketing_dev' \
		'docker-build     build the container image with podman'

dev:
	@test -f .env || { echo "error: ticketing/.env is missing; cp .env.example .env"; exit 1; }
	@set -a; . ./.env; set +a; $(GO) run ./cmd/ticketing serve

migrate:
	@set -a; [ -f .env ] && . ./.env; set +a; $(GO) run ./cmd/ticketing migrate

seed:
	@set -a; [ -f .env ] && . ./.env; set +a; $(GO) run ./cmd/seed

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-integration:
	@set -a; [ -f .env ] && . ./.env; set +a; \
		TEST_DATABASE_URL="$${TEST_DATABASE_URL:-$${DATABASE_URL}}" \
		$(GO) test -tags=integration ./... -v

test-integration-race:
	@set -a; [ -f .env ] && . ./.env; set +a; \
		TEST_DATABASE_URL="$${TEST_DATABASE_URL:-$${DATABASE_URL}}" \
		$(GO) test -race -tags=integration ./... -v

lint:
	$(GO) vet ./...
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "gofmt required for:"; echo "$$files"; exit 1; fi

fmt:
	gofmt -w .

tidy:
	$(GO) mod tidy

db-create:
	@psql -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = 'ticketing_dev'" | grep -q 1 \
		|| createdb ticketing_dev
	@echo "ticketing_dev ready"

db-drop:
	dropdb --if-exists ticketing_dev

psql:
	psql ticketing_dev

docker-build:
	podman build -t basementplays/ticketing:dev .
