.PHONY: test test-short clean deps setup run-compose down-compose help

# Default target
.DEFAULT_GOAL := help

# Variables
TEST_TIMEOUT := 10m
GO := go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	$(GO) mod download
	$(GO) mod tidy

setup: ## Set up the test environment
	@echo "Setting up integration test environment..."
	@mkdir -p testdata/policies
	@echo "Setup complete!"

test: deps ## Run all integration tests
	@echo "Running integration tests..."
	$(GO) test -v -timeout $(TEST_TIMEOUT) ./...

test-short: deps ## Run tests in short mode
	@echo "Running integration tests (short mode)..."
	$(GO) test -v -short -timeout $(TEST_TIMEOUT) ./...

test-compose: ## Run tests using Docker Compose
	@echo "Running integration tests with Docker Compose..."
	USE_COMPOSE=true $(GO) test -v -timeout $(TEST_TIMEOUT) ./...

test-unit: deps ## Run unit tests only
	@echo "Running unit tests..."
	$(GO) test -v -run TestUnit -timeout $(TEST_TIMEOUT) ./...

test-integration: deps ## Run integration tests only
	@echo "Running integration tests..."
	$(GO) test -v -run TestIntegration -timeout $(TEST_TIMEOUT) ./...

clean: ## Clean up test artifacts
	@echo "Cleaning up..."
	rm -rf coverage.out
	rm -rf *.out
	find . -name "*.test" -delete

run-compose: ## Start Docker Compose services
	@echo "Starting Docker Compose services..."
	docker compose up -d

down-compose: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	docker compose down -v

logs: ## Show Docker Compose logs
	docker compose logs -f

clean-all: clean down-compose ## Clean everything including containers
	@echo "Cleaning everything..."
	docker compose down -v --rmi local || true

coverage: deps ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GO) test -v -coverprofile=coverage.out -coverpkg=./... ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run || echo "golangci-lint not installed, skipping..."

vet: deps ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...

check: fmt vet lint ## Run all checks

.DEFAULT_GOAL := help

