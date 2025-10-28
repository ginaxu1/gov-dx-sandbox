# Integration Tests for OpenDIF Platform

Comprehensive integration tests for the OpenDIF data exchange platform using Go and testcontainers.

## Overview

This test suite validates the interactions between services and their infrastructure (databases, Redis, OPA) to ensure the system works correctly end-to-end.

## Features

- **Isolated Test Environment**: Uses Docker Compose or testcontainers for complete service isolation
- **Real Components**: Tests actual databases, Redis, and OPA - no mocking
- **Automated**: Single command test execution
- **Stateless**: Each test run starts from a clean, known state

## Test Scenarios

### 1. Authentication & Authorization (`auth_test.go`)

Tests authentication and authorization flows:
- OPA health and connectivity
- Database authentication queries
- Policy metadata checks
- JWT validation (when applicable)
- Authorization policy enforcement

### 2. Audit Logging (`audit_stream_test.go`)

Tests asynchronous audit logging:
- Audit log creation in database
- Redis stream message processing
- Audit log filtering and querying
- Audit middleware integration

### 3. Data Exchange Flow (`data_exchange_flow_test.go`)

Tests core data exchange workflows:
- Provider metadata queries
- Consumer grants verification
- Provider schemas validation
- Consent workflow testing
- Data retrieval verification

### 4. GraphQL API (`graphql_api_test.go`)

Tests GraphQL federation:
- Unified schemas table
- Schema introspection
- Version tracking
- Federation capabilities

## Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for Docker Compose mode)
- PostgreSQL 15+ (managed via containers)
- Redis (managed via containers)
- OPA (optional, managed via containers)

## Quick Start

### Option 1: Using Docker Compose (Recommended)

1. **Start services**:
   ```bash
   make run-compose
   ```

2. **Run tests with Docker Compose**:
   ```bash
   make test-compose
   ```

3. **Stop services**:
   ```bash
   make down-compose
   ```

### Option 2: Using Testcontainers

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Run all tests**:
   ```bash
   make test
   ```

3. **Run specific tests**:
   ```bash
   go test -v -run TestAuthentication ./...
   ```

### Option 3: Using Existing Services

1. **Ensure services are running**:
   - PostgreSQL on `localhost:5432`
   - Redis on `localhost:6379`
   - OPA on `localhost:8181` (optional)

2. **Set environment variables**:
   ```bash
   export DATABASE_URL="postgres://user:pass@localhost:5432/opendif_test?sslmode=disable"
   export REDIS_URL="localhost:6379"
   ```

3. **Run tests**:
   ```bash
   make test
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://test_user:test_password@localhost:5432/opendif_test` |
| `REDIS_URL` | Redis connection string | `localhost:6379` |
| `OPA_URL` | OPA base URL | `http://localhost:8181` |
| `USE_COMPOSE` | Use Docker Compose mode | `false` |

### Test Configuration

The tests use two modes:

1. **Testcontainers Mode** (default): Each test spins up its own containers
2. **Docker Compose Mode**: All tests share a single Docker Compose environment

## Running Tests

### Run All Tests

```bash
make test
```

### Run Specific Test Suite

```bash
# Authentication tests
go test -v -run TestAuthentication ./...

# Audit logging tests
go test -v -run TestAuditStream ./...

# Data exchange tests
go test -v -run TestDataExchange ./...

# GraphQL tests
go test -v -run TestGraphQL ./...
```

### Run in Short Mode

```bash
make test-short
```

This skips integration tests and only runs unit tests.

### Run with Coverage

```bash
make coverage
```

This generates a coverage report in `coverage.html`.

## Directory Structure

```
integration-tests/go/
├── docker-compose.yml          # Docker Compose configuration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── Makefile                    # Build automation
├── README.md                   # This file
│
├── main_test.go                # Test main (setup/teardown)
├── utils.go                    # Test utilities
│
├── auth_test.go                # Authentication & authorization tests
├── audit_stream_test.go        # Audit logging tests
├── data_exchange_flow_test.go  # Data exchange flow tests
├── graphql_api_test.go         # GraphQL API tests
│
└── testdata/                   # Test fixtures
    ├── init-db.sql             # Database initialization
    ├── seed-data.sql           # Seed data
    └── policies/               # OPA policies
        ├── allow_all.rego
        └── deny_passport.rego
```

## Test Data

The tests use a predefined set of test data stored in `testdata/seed-data.sql`:

- **Entities**: `entity-1`, `entity-2`, `entity-3`
- **Consumers**: `passport-app`, `test-consumer`, `unauthorized-app`
- **Providers**: `provider-drp`, `provider-rgd`
- **Schemas**: Sample GraphQL schemas
- **Policy Metadata**: Field-level access control policies

## Troubleshooting

### Docker Issues

**Problem**: Docker containers fail to start

**Solution**:
```bash
# Check Docker status
docker ps

# Clean up old containers
docker-compose down -v
make clean-all

# Restart Docker
sudo systemctl restart docker
```

### Database Connection Issues

**Problem**: Cannot connect to database

**Solution**:
```bash
# Check database logs
make logs

# Verify database is running
docker-compose ps

# Restart database
docker-compose restart postgres
```

### Test Timeout Issues

**Problem**: Tests timeout before completing

**Solution**:
```bash
# Increase timeout
export TEST_TIMEOUT=20m
make test

# Or run specific tests
go test -v -timeout 30m ./...
```

## Continuous Integration

These tests are designed to run in CI/CD pipelines. Example GitHub Actions workflow:

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.24'
      - name: Run integration tests
        run: |
          cd integration-tests/go
          make test
```

## Best Practices

1. **Keep Tests Independent**: Each test should be able to run in isolation
2. **Clean State**: Always start with a clean database state
3. **Idempotent**: Running tests multiple times should produce the same results
4. **Fast**: Integration tests should complete in reasonable time (<10 minutes)
5. **Document**: Add comments explaining complex test scenarios

## Contributing

When adding new tests:

1. Follow the existing test structure
2. Use descriptive test names
3. Add comments for complex scenarios
4. Update this README with new test descriptions
5. Ensure tests pass in both Testcontainers and Docker Compose modes

## License

This project is part of the OpenDIF platform.

