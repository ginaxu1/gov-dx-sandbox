# Integration Tests

Comprehensive integration tests for the OpenDIF data exchange platform.

## Overview

The integration tests validate the complete data exchange workflow from data request to consent management, ensuring all components work together correctly.

## Directory Structure

```
tests/integration/
├── README.md                    # This file
├── docker-compose.test.yml     # Docker Compose configuration for tests
├── go.mod                      # Go module definition
├── consent/                    # Consent Engine integration tests
│   └── consent_test.go
├── audit/                      # Audit Service integration tests
│   └── audit_test.go
├── pdp/                        # Policy Decision Point tests
│   └── pdp_test.go
├── graphql_flow_test.go        # GraphQL flow integration test
└── testutils/                  # Test utilities
    ├── db.go                   # Database utilities
    └── http.go                 # HTTP utilities
```

## Prerequisites

Before running integration tests, ensure all services are running:

1. **Consent Engine** (Port 8081) 
2. **Policy Decision Point** (Port 8082)
3. **Orchestration Engine** (Port 4000)
4. **Audit Service** (Port 3001)

### Starting Services

```bash
# Start all services using Docker Compose
cd /Users/tmp/gov-dx-sandbox
make start-exchange
```

## Running Tests

### Run All Tests
```bash
cd tests/integration
docker compose -f docker-compose.test.yml up -d

# Wait for services to be ready
go test -v ./...

docker compose -f docker-compose.test.yml down -v
```

### Run Specific Test Suites
```bash
# Consent Engine tests
go test -v ./consent/...

# Audit Service tests
go test -v ./audit/...

# Policy Decision Point tests
go test -v ./pdp/...
```

### With Database Verification
```bash
TEST_VERIFY_DB=true go test -v ./...
```

## Test Scenarios

### 1. Complete Data Exchange Flow

Tests the complete end-to-end workflow from data request to data retrieval.

**Test Steps**:
1. Create provider profile
2. Submit and approve schema
3. Create consumer application
4. Request data access
5. Handle consent workflow (if required)
6. Retrieve authorized data

### 2. Consent Management Workflow

Tests consent management scenarios including different data ownership patterns.

#### Scenario A: Data Owner is NOT the Provider
- **Setup**: Provider (DRP) requests data owned by RGD
- **Expected**: Consent required, SMS OTP sent to data owner

#### Scenario B: Data Owner IS the Provider
- **Setup**: Provider (DRP) requests data owned by DRP
- **Expected**: No consent required, direct access

### 3. Policy Decision Point Tests

Tests authorization decisions and consent requirements.

**Test Cases**:
- Public field access
- Restricted field access (authorized)
- Restricted field access (consent required)
- Unauthorized access

## Contributing

When adding new tests:

1. **Include error handling**: Check for service availability
2. **Add documentation**: Update this README
3. **Test edge cases**: Include both success and failure scenarios
4. **Clean up**: Ensure tests don't leave test data in the system

## License

This project is part of the OpenDIF platform.
