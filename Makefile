APP_NAME := my-go-app
CMD_PATH := ./cmd/api
BIN_DIR := bin
BIN := $(BIN_DIR)/api

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)

.PHONY: all build run test vet fmt lint tidy clean docker docker-run help

all: build

help: ## Show targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-14s %s\n", $$1, $$2}'

build: ## Build binary to bin/api
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o $(BIN) $(CMD_PATH)

run: ## Run the API locally (loads .env if present)
	@set -a; [ -f .env ] && . ./.env; set +a; go run $(CMD_PATH)

test: ## Run tests with race detector
	go test -race -count=1 ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format source
	gofmt -w .

lint: ## Run golangci-lint (requires install)
	golangci-lint run ./...

tidy: ## Tidy go.mod
	go mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html

docker: ## Build Docker image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(APP_NAME):$(VERSION) .

docker-run: ## Run container on :8080
	docker run --rm -p 8080:8080 $(APP_NAME):$(VERSION)
