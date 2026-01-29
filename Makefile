.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down setup

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building..."
	@go build -o bin/server ./cmd/server

run: ## Run the application locally
	@echo "Running server..."
	@go run ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t url-shortener:latest .

docker-up: ## Start Docker Compose stack
	@echo "Starting Docker Compose stack..."
	@docker compose up --build -d

docker-down: ## Stop Docker Compose stack
	@echo "Stopping Docker Compose stack..."
	@docker compose down

docker-logs: ## View Docker Compose logs
	@docker compose logs -f app

setup: ## Setup local development environment
	@echo "Setting up development environment..."
	@cp -n .env.example .env || true
	@echo "Created .env file (if it didn't exist)"
	@go mod download
	@echo "Dependencies downloaded"

migrate-up: ## Run database migrations (requires running PostgreSQL)
	@echo "Running migrations..."
	@psql $(DB_DSN) < migrations/001_create_urls_table.up.sql

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@psql $(DB_DSN) < migrations/001_create_urls_table.down.sql

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

api-test: docker-up ## Run API tests (requires running service)
	@echo "Waiting for service to be ready..."
	@sleep 5
	@bash test-api.sh

.DEFAULT_GOAL := help
