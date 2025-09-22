# Makefile for consent-engine

.PHONY: help test test-local test-postgres test-all build run clean setup-test-db

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run tests with in-memory engine (default)"
	@echo "  test-local    - Run tests with local PostgreSQL"
	@echo "  test-postgres - Run tests with PostgreSQL (requires TEST_USE_POSTGRES=true)"
	@echo "  test-all      - Run all test variants"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application locally"
	@echo "  clean         - Clean up test databases"
	@echo "  setup-test-db - Start test PostgreSQL container"

# Test with in-memory engine (default)
test:
	@echo "Running tests with in-memory engine..."
	go test ./... -count=1 -v

# Test with local PostgreSQL
test-local: setup-test-db
	@echo "Running tests with local PostgreSQL..."
	@export TEST_USE_POSTGRES=true && \
	export TEST_DB_HOST=localhost && \
	export TEST_DB_PORT=5433 && \
	export TEST_DB_USERNAME=test_user && \
	export TEST_DB_PASSWORD=test_password && \
	export TEST_DB_DATABASE=consent_engine_test && \
	export TEST_DB_SSLMODE=disable && \
	go test ./... -count=1 -v

# Test with PostgreSQL (requires environment variables)
test-postgres:
	@echo "Running tests with PostgreSQL..."
	@if [ "$$TEST_USE_POSTGRES" != "true" ]; then \
		echo "Error: TEST_USE_POSTGRES must be set to 'true'"; \
		exit 1; \
	fi
	go test ./... -count=1 -v

# Run all test variants
test-all: test test-postgres

# Build the application
build:
	@echo "Building consent-engine..."
	go build -o consent-engine .

# Run the application locally
run: setup-test-db
	@echo "Starting consent-engine..."
	@export CHOREO_OPENDIF_DB_HOSTNAME=localhost && \
	export CHOREO_OPENDIF_DB_PORT=5433 && \
	export CHOREO_OPENDIF_DB_USERNAME=test_user && \
	export CHOREO_OPENDIF_DB_PASSWORD=test_password && \
	export CHOREO_OPENDIF_DB_DATABASENAME=consent_engine_test && \
	export DB_SSLMODE=disable && \
	go run .

# Clean up test databases
clean:
	@echo "Cleaning up test databases..."
	docker-compose -f docker-compose.test.yml down -v
	docker volume prune -f

# Start test PostgreSQL container
setup-test-db:
	@echo "Starting test PostgreSQL container..."
	docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout 30s bash -c 'until docker-compose -f docker-compose.test.yml exec postgres-test pg_isready -U test_user -d consent_engine_test; do sleep 1; done'
	@echo "PostgreSQL is ready!"
