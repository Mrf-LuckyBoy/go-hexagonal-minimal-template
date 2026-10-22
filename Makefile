APP_NAME      ?= app

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-14s %s\n", $$1, $$2}'

.PHONY: run
run: ## Run the app locally
	go run ./cmd

.PHONY: build
build: ## Build a static binary into bin/
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) ./cmd

.PHONY: test
test: ## Run tests with the race detector and coverage
	go test -race -cover ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	golangci-lint fmt ./...

.PHONY: tidy
tidy: ## Tidy go.mod / go.sum
	go mod tidy

.PHONY: docker-build
docker-build: ## Build the Docker image
	docker build -t $(APP_NAME):latest .

.PHONY: clean
clean: ## Remove build output
	rm -rf bin/
