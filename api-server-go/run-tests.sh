#!/bin/bash

# Test database configuration
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5434
export TEST_DB_USER=test_user
export TEST_DB_PASSWORD=test_password
export TEST_DB_NAME=api_server_test
export TEST_DB_SSLMODE=disable

echo "Setting up test database..."

# Create test database if it doesn't exist
PGPASSWORD=$TEST_DB_PASSWORD createdb -h $TEST_DB_HOST -p $TEST_DB_PORT -U $TEST_DB_USER $TEST_DB_NAME 2>/dev/null || echo "Database may already exist"

echo "Running tests with PostgreSQL test database..."
echo "Host: $TEST_DB_HOST:$TEST_DB_PORT"
echo "Database: $TEST_DB_NAME"
echo "User: $TEST_DB_USER"

# Run the tests
go test ./tests/... -v

echo "Tests completed."