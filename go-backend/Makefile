# Makefile for eSports Fantasy Backend

# Variables
BINARY_NAME=esports-fantasy
BUILD_DIR=build
CMD_DIR=cmd/server

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Show this help message
	@echo "eSports Fantasy Backend - Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: ## Run the application
	@echo "🚀 Starting $(BINARY_NAME)..."
	@go run $(CMD_DIR)/main.go

# Install dependencies
deps: ## Install/update dependencies
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

# Run tests
test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Format code
fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Run linter
lint: ## Run linter
	@echo "🔍 Running linter..."
	@golangci-lint run
	@echo "✅ Linting complete"

# Install tools
tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Tools installed"

# Generate swagger docs
swagger: ## Generate swagger documentation
	@echo "📚 Generating swagger documentation..."
	@swag init -g cmd/server/main.go -o api/docs
	@echo "✅ Swagger docs generated"

# Run in development mode
dev: ## Run in development mode with auto-reload
	@echo "🔄 Starting in development mode..."
	@air

# Database migrations
migrate-up: ## Run database migrations up
	@echo "⬆️  Running database migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ Migrations complete"

migrate-down: ## Run database migrations down
	@echo "⬇️  Reverting database migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down
	@echo "✅ Migrations reverted"

# Docker targets
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(BINARY_NAME) .
	@echo "✅ Docker image built"

docker-run: ## Run Docker container
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 --env-file .env $(BINARY_NAME)

# Development shortcuts
start: run ## Alias for run
dev-setup: deps tools ## Setup development environment
full-build: clean deps fmt test build ## Full build pipeline

.PHONY: help build run deps clean test test-coverage fmt lint tools swagger dev migrate-up migrate-down docker-build docker-run start dev-setup full-build