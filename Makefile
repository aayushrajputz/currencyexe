# Exchange Rate Service Makefile

.PHONY: build run test clean docker-build docker-run help

# Variables
BINARY_NAME=exchange-rate-service
DOCKER_IMAGE=exchange-rate-service
PORT=8080

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/server

run: build ## Build and run the application
	@echo "Starting $(BINARY_NAME) on port $(PORT)..."
	@./$(BINARY_NAME)

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -cover ./...

clean: ## Clean build artifacts
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME).exe

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container on port $(PORT)..."
	@docker run -p $(PORT):$(PORT) $(DOCKER_IMAGE)

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run



dev: ## Run in development mode with auto-reload (requires air)
	@echo "Starting development server..."
	@air

health-check: ## Test the health endpoint
	@echo "Testing health endpoint..."
	@curl -s http://localhost:$(PORT)/health | jq .
