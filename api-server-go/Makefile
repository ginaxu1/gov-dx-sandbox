# API Server Go Makefile
.PHONY: help test test-local test-postgres test-all build run clean setup-test-db

help:
	@echo "Available targets:"
	@echo "  help         - Show this help message"
	@echo "  test         - Run all tests"
	@echo "  test-local   - Run tests with local PostgreSQL"
	@echo "  test-postgres - Run tests with PostgreSQL (requires TEST_USE_POSTGRES=true)"
	@echo "  test-all     - Run all test suites"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  clean        - Clean build artifacts"
	@echo "  setup-test-db - Start test PostgreSQL container"

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Test with local PostgreSQL
test-local:
	@echo "Running tests with local PostgreSQL..."
	@export TEST_USE_POSTGRES=true && \
	export TEST_DB_HOSTNAME=localhost && \
	export TEST_DB_PORT=5434 && \
	export TEST_DB_USERNAME=test_user && \
	export TEST_DB_PASSWORD=test_password && \
	export TEST_DB_DATABASENAME=api_server_test && \
	go test -v ./...

# Test with PostgreSQL (requires environment variables)
test-postgres:
	@echo "Running tests with PostgreSQL..."
	@if [ "$$TEST_USE_POSTGRES" != "true" ]; then \
		echo "Error: TEST_USE_POSTGRES must be set to 'true'"; \
		exit 1; \
	fi
	@go test -v ./...

# Run all test suites
test-all: test test-postgres

# Build the application
build:
	@echo "Building API Server Go..."
	@go build -o api-server-go .

# Run the application locally
run:
	@echo "Running API Server Go locally..."
	@go run .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f api-server-go
	@go clean

# Start test PostgreSQL container
setup-test-db:
	@echo "Starting test PostgreSQL container..."
	@docker-compose -f docker-compose.test.yml up -d postgres-test
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout 30s bash -c 'until docker-compose -f docker-compose.test.yml exec postgres-test pg_isready -U test_user -d api_server_test; do sleep 1; done'
	@echo "PostgreSQL is ready!"

# Stop test containers
stop-test-db:
	@echo "Stopping test containers..."
	@docker-compose -f docker-compose.test.yml down

# Run tests with Docker
test-docker:
	@echo "Running tests with Docker..."
	@docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	@docker-compose -f docker-compose.test.yml down
