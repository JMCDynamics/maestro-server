# Variables
include .env
export $(shell sed 's/=.*//' .env)

AIR_CMD=air
GO_CMD=go
SRC_DIR=./cmd/api
MAIN_FILE=main.go
BUILD_DIR=./build
MIGRATE_CMD=migrate
DATABASE_URL=postgres://${DATABASE_NAME}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_NAME}?sslmode=disable
MIGRATION_DIR=./migrations

.PHONY: run
run:
	$(AIR_CMD) -c .air.toml

.PHONY: build
build:
	$(GO_CMD) build -o $(BUILD_DIR)/app $(SRC_DIR)/$(MAIN_FILE)

.PHONY: install
install:
	$(GO_CMD) mod tidy

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: install-air
install-air:
	go install github.com/cosmtrek/air@latest

.PHONY: migrate-up
migrate-up:
	$(MIGRATE_CMD) -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	$(MIGRATE_CMD) -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: migrate-reset
migrate-reset:
	$(MIGRATE_CMD) -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down
	$(MIGRATE_CMD) -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up

.PHONY: help
help:
	@echo "Available make commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*## ' Makefile | awk 'BEGIN {FS=":.*## "}{printf "  %-20s %s\n
