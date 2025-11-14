# Integration Tests

Comprehensive integration tests for the OpenDIF data exchange platform, with both shell-based and Go-based test suites.

## Overview

The integration tests validate the complete data exchange workflow from data request to consent management, ensuring all components work together correctly.

## Two Test Suites

### 1. **Go-based Integration Tests** (New)

A comprehensive integration test suite written in Go that focuses on validating the interactions between services and their backing infrastructure (databases, Redis, OPA).

**Location**: `integration-tests/go/`

**Features**:
- Database and infrastructure validation
- Authentication & authorization testing
- Audit logging verification
- Data exchange flow validation
- GraphQL API testing

### 2. **Shell-based Integration Tests** (Existing)

Quick smoke tests and manual validation scripts for end-to-end workflows.

**Location**: `integration-tests/*.sh`

**Features**:
- End-to-end workflow testing
- Service health checks
- Consent flow validation
- Policy Decision Point testing

## Directory Structure

```
integration-tests/
├── README.md                    # This file
│
├── go/                          # Go-based integration tests
│   ├── main_test.go             # Test runner and setup/teardown
│   ├── utils.go                 # Test utilities (DB, Redis, OPA helpers)
│   ├── auth_test.go             # Authentication & authorization tests
│   ├── audit_stream_test.go     # Audit logging tests
│   ├── data_exchange_flow_test.go # Data exchange flow tests
│   ├── graphql_api_test.go      # GraphQL API tests
│   ├── docker-compose.yml       # Docker Compose configuration
│   ├── go.mod                   # Go module definition
│   ├── Makefile                 # Build automation
│   ├── README.md                # Detailed Go tests documentation
│   └── testdata/                # Test fixtures
│       ├── init-db.sql          # Database schema initialization
│       ├── seed-data.sql        # Seed data for tests
│       └── policies/            # OPA test policies
│
└── *.sh                         # Shell-based integration tests
    ├── run-all-tests.sh         # Run all integration tests
    ├── test-complete-flow.sh    # Complete end-to-end workflow test
    ├── test-consent-flow.sh     # Consent management workflow test
    └── test-pdp.sh              # Policy Decision Point test
```

## Quick Start

### Go-based Tests

```bash
cd integration-tests/go

# Install dependencies
make deps

# Run tests (requires services running)
make test

# Or use Docker Compose
make run-compose
USE_COMPOSE=true make test
```

### Shell-based Tests

```bash
cd integration-tests

# Run all shell tests
./run-all-tests.sh

# Run specific test
./test-consent-flow.sh
```

## Prerequisites

Before running integration tests, ensure all services are running:

1. **Consent Engine** (Port 8081) 
2. **Policy Decision Point** (Port 8082)
3. **Orchestration Engine** (Port 4000)
4. **API Server** (Port 3000)
5. **Audit Service** (Port 3001)

### Starting Services

```bash
# Start all services using Docker Compose
cd /Users/tmp/gov-dx-sandbox
make start-exchange

# Or start individual services:
# Terminal 1 - Consent Engine
cd exchange/consent-engine
go run main.go

# Terminal 2 - Policy Decision Point
cd exchange/policy-decision-point
go run main.go

# Terminal 3 - Orchestration Engine
cd exchange/orchestration-engine-go
go run main.go
```

## Go-based Test Coverage

### Scenario 1: Authentication & Authorization
- OPA connectivity tests
- Database authentication queries  
- Policy metadata validation
- JWT validation (placeholder)
- Allow-list enforcement

### Scenario 2: Audit Logging
- Audit log creation in database
- Redis stream message processing
- Audit log filtering
- Audit middleware integration

### Scenario 3: Data Exchange Flow
- Provider metadata queries
- Consumer grants verification
- Provider schemas validation
- Consent workflow testing

### Scenario 4: GraphQL API
- Schema availability checks
- Schema introspection
- Federation capabilities
- Version tracking

## Shell-based Test Scenarios

### 1. Complete Data Exchange Flow (`test-complete-flow.sh`)

Tests the complete end-to-end workflow from data request to data retrieval.

**Test Steps**:
1. Create provider profile
2. Submit and approve schema
3. Create consumer application
4. Request data access
5. Handle consent workflow (if required)
6. Retrieve authorized data

### 2. Consent Management Workflow (`test-consent-flow.sh`)

Tests consent management scenarios including different data ownership patterns.

#### Scenario A: Data Owner is NOT the Provider
- **Setup**: Provider (DRP) requests data owned by RGD
- **Expected**: Consent required, SMS OTP sent to data owner

#### Scenario B: Data Owner IS the Provider
- **Setup**: Provider (DRP) requests data owned by DRP
- **Expected**: No consent required, direct access

### 3. Policy Decision Point Tests (`test-pdp.sh`)

Tests authorization decisions and consent requirements.

**Test Cases**:
- Public field access
- Restricted field access (authorized)
- Restricted field access (consent required)
- Unauthorized access

## Running Tests

### Run All Shell Tests
```bash
cd integration-tests
./run-all-tests.sh
```

### Run All Go Tests
```bash
cd integration-tests/go
make test
```

### Run Specific Tests
```bash
# Shell tests
./integration-tests/test-consent-flow.sh
./integration-tests/test-pdp.sh

# Go tests
cd integration-tests/go
go test -v -run TestAuthentication
go test -v -run TestAuditStream
```

## Configuration

### Go Test Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `USE_COMPOSE` | Use Docker Compose mode | `false` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://test_user:test_password@localhost:5432/opendif_test` |
| `REDIS_URL` | Redis connection string | `localhost:6379` |

## Test Data

The Go tests use a predefined set of test data stored in `integration-tests/go/testdata/seed-data.sql`:

- **Entities**: `entity-1`, `entity-2`, `entity-3`
- **Consumers**: `passport-app`, `test-consumer`, `unauthorized-app`
- **Providers**: `provider-drp`, `provider-rgd`
- **Schemas**: Sample GraphQL schemas
- **Policy Metadata**: Field-level access control policies

## Expected Test Results

### Successful Test Run
```
✅ All services are running
✅ Complete data exchange flow passed
✅ Consent workflow (data owner ≠ provider) passed
✅ Consent workflow (data owner = provider) passed
✅ Policy Decision Point authorization tests passed
✅ Go integration tests passed
✅ All integration tests completed successfully
```

## Troubleshooting

### Service Not Running
```
❌ Error: Service not responding on port 8081
```
**Solution**: Start the service: `make start-exchange`

### Database Connection Issues
```
❌ Error: Cannot connect to database
```
**Solution**: 
```bash
cd integration-tests/go
docker-compose up -d postgres
```

### Test Timeout Issues
```bash
# Increase timeout for Go tests
export TEST_TIMEOUT=20m
cd integration-tests/go
make test
```

## CI/CD Integration

Both test suites can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
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
      
      - name: Run shell tests
        run: |
          cd integration-tests
          ./run-all-tests.sh
      
      - name: Run Go tests
        run: |
          cd integration-tests/go
          make deps
          make test
```

## Contributing

When adding new tests:

1. **For Shell Tests**: Follow naming convention: `test-{feature}-{scenario}.sh`
2. **For Go Tests**: Follow Go testing best practices and naming conventions
3. **Include error handling**: Check for service availability
4. **Add documentation**: Update this README
5. **Test edge cases**: Include both success and failure scenarios
6. **Clean up**: Ensure tests don't leave test data in the system

## Best Practices

1. **Keep Tests Independent**: Each test should be able to run in isolation
2. **Clean State**: Always start with a clean database state
3. **Idempotent**: Running tests multiple times should produce the same results
4. **Fast**: Integration tests should complete in reasonable time
5. **Document**: Add comments explaining complex test scenarios

## License

This project is part of the OpenDIF platform.
