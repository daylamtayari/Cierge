APP_NAME := cierge

VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

BIN_DIR := bin

# Dev

.PHONY: dev
dev: ## Run server in dev mode with hot reload
	@which air > /dev/null  2>&1 || (echo "Error: 'air' is not installed and required for hot reload."; exit 1)
	air -c .air.toml

.PHONY: run
run: ## Run server
	go run ./cmd/server

# Utilities and cleanup

.PHONY: tidy
tidy: ## Tidy all modules
	go work sync
	go mod tidy
	cd cmd/cli && go mod tidy
	cd lambdas/reservation && go mod tidy
	cd pkg/errcol && go mod tidy
	cd pkg/resy && go mod tidy
	cd pkg/opentable && go mod tidy

.PHONY: clean
clean: ## Clean bin directory
	rm -rf $(BIN_DIR)

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
