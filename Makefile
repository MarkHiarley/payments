SHELL := /bin/bash

MIGRATIONS_DIR := internal/database/migrations
MIGRATE_IMAGE  ?= migrate/migrate:v4.18.3
COVERAGE_FILE  := coverage.out
# Must match the golangci-lint version in .github/workflows/ci.yml; bump both together.
GOLANGCI_LINT_VERSION ?= v2.1.6

# Reuse the same values as docker compose. Missing .env is fine: defaults below match .env.example.
-include .env
export

POSTGRES_HOST          ?= localhost
POSTGRES_USER          ?= payments
POSTGRES_PASSWORD      ?= payments
POSTGRES_DB            ?= payments
POSTGRES_PORT_EXTERNAL ?= 5432

# Host-side URL: migrations run against the port published by the db container.
MIGRATE_DB_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT_EXTERNAL)/$(POSTGRES_DB)?sslmode=disable

.DEFAULT_GOAL := help
.PHONY: help up down logs migrate migrate-down test test-race cover lint build fmt run e2e

help: ## List available targets
	@grep -hE '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.env:
	@cp -n .env.example .env
	@echo "Created .env from .env.example — edit it if you need different credentials."

up: .env ## Create .env if missing, then start the whole stack in the background
	docker compose up -d

down: ## Stop and remove containers (the pgdata volume is kept)
	docker compose down

logs: ## Follow logs from every service
	docker compose logs -f

migrate: ## Apply all pending migrations (host network, uses the published db port)
	docker run --rm --network host \
		-v "$(CURDIR)/$(MIGRATIONS_DIR):/migrations:ro" \
		$(MIGRATE_IMAGE) \
		-path=/migrations -database "$(MIGRATE_DB_URL)" up

migrate-down: ## Roll back the last applied migration
	docker run --rm --network host \
		-v "$(CURDIR)/$(MIGRATIONS_DIR):/migrations:ro" \
		$(MIGRATE_IMAGE) \
		-path=/migrations -database "$(MIGRATE_DB_URL)" down 1

test: ## Run the test suite
	go test -count=1 ./...

test-race: ## Run the test suite with the race detector
	go test -race -count=1 ./...

cover: ## Run the test suite with coverage and print the total
	go test -count=1 -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -func=$(COVERAGE_FILE) | tail -n 1

lint: ## Check formatting and run go vet + golangci-lint
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt required on:"; echo "$$unformatted"; exit 1; \
	fi
	go vet ./...
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --config .golangci.yml

build: ## Build the api binary into bin/api
	@mkdir -p bin
	go build -o bin/api ./cmd/api

fmt: ## Format every Go file with gofmt
	gofmt -w .

run: ## Run the api locally, reading .env
	go run ./cmd/api

e2e: ## Run end-to-end tests (fails while no e2e suite exists)
	@if [ ! -d tests/e2e ]; then \
		echo "e2e: tests/e2e/ does not exist yet — add the end-to-end suite there first."; \
		exit 1; \
	fi
	go test -count=1 -tags e2e ./tests/e2e/...
