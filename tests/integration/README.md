# Integration Tests

End-to-end integration tests for the Data Exchange Platform covering GraphQL workflows, consent management, and policy decisions.

---

## Quick Start

### Option 1: Automated Script (Recommended)

Run the automated test script that mimics the CI/CD workflow:

```bash
cd tests/integration
./run-local-tests.sh
```

This script:
- Checks Docker and Go are available
- Installs dependencies
- Starts all services
- Waits for services to be healthy
- Runs tests with race detection
- Cleans up automatically

**Skip go.mod check for local development:**
```bash
SKIP_GO_MOD_CHECK=1 ./run-local-tests.sh
```

### Directory Structure

```
tests/integration/
├── README.md                    # This file
├── docker-compose.test.yml     # Docker Compose configuration for tests
├── go.mod                      # Go module definition
├── consent/                    # Consent Engine integration tests
│   └── consent_test.go
├── audit/                      # Audit Service integration tests
│   └── audit_test.go
├── graphql_flow_test.go        # GraphQL flow integration test
├── services_integration_test.go # Service health checks
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

### Option 2: Manual Steps

```bash
cd tests/integration
docker compose -f docker-compose.test.yml up -d

# Wait for services to be ready
go test -v ./...

docker compose -f docker-compose.test.yml down -v
```

**Prerequisites:** Docker and Docker Compose installed.

### Run Specific Test Suites
```bash
# Consent Engine tests
go test -v ./consent/...

# Audit Service tests
go test -v ./audit/...

```

### With Database Verification
```bash
TEST_VERIFY_DB=true go test -v ./...
```

---

## Test Scenarios

### GraphQL Flow Tests

**`TestGraphQLFlow_SuccessPath`** - Complete success path:
- Creates policy metadata in PDP
- Adds application to allowlist
- Creates consent record
- Executes GraphQL query
- Verifies PDP and Consent Engine integration

**`TestGraphQLFlow_MissingPolicyMetadata`** - Tests behavior when field lacks policy metadata (expects authorization error)

**`TestGraphQLFlow_UnauthorizedApp`** - Tests behavior when app has no consent (expects consent error)

**`TestGraphQLFlow_ServiceTimeout`** - Tests resilience when PDP is unavailable

**`TestGraphQLFlow_InvalidQuery`** - Tests malformed GraphQL query handling

**`TestGraphQLFlow_MissingToken`** - Tests authentication failure (missing JWT)

### Consent Management Workflow

Tests consent management scenarios including different data ownership patterns.

#### Scenario A: Data Owner is NOT the Provider
- **Setup**: Provider (DRP) requests data owned by RGD
- **Expected**: Consent required, SMS OTP sent to data owner

#### Scenario B: Data Owner IS the Provider
- **Setup**: Provider (DRP) requests data owned by DRP
- **Expected**: No consent required, direct access

### Policy Decision Point Tests

Tests authorization decisions and consent requirements.

**Test Cases**:
- Public field access
- Restricted field access (authorized)
- Restricted field access (consent required)
- Unauthorized access

### Audit Service Tests

**`TestAudit_CreateDataExchangeEvent`** - Tests creating a data exchange audit event

**`TestAudit_CreateDataExchangeEventFailure`** - Tests creating a failure event

**`TestAudit_GetDataExchangeEvents`** - Tests retrieving data exchange events

**`TestAudit_FilterByConsumer`** - Tests filtering events by consumer ID

**`TestAudit_FilterByStatus`** - Tests filtering events by status

**`TestAudit_FilterByDateRange`** - Tests filtering events by date range

**`TestAudit_Pagination`** - Tests pagination of event results

**`TestAudit_InvalidRequest`** - Tests handling of invalid requests

**`TestAudit_DatabaseVerification`** - Tests database state verification

**`TestAudit_CreateManagementEvent`** - Tests creating management events

**`TestAudit_GetManagementEvents`** - Tests retrieving management events

### Consent Engine Tests

**`TestConsent_CreateAndRetrieve`** - Tests basic consent creation and retrieval

**`TestConsent_InvalidRequest`** - Tests edge cases for invalid consent requests

**`TestConsent_GetByConsumer`** - Tests retrieving consents by consumer ID

**`TestConsent_StatusUpdate`** - Tests consent status updates

**`TestConsent_ExpiryCheck`** - Tests consent expiry handling

**`TestConsent_DatabaseVerification`** - Tests database state verification

### Service Health Tests

**`TestPortalBackend_Health`** - Verifies Portal Backend health endpoint

---

## Configuration

### Environment Variables

**Database Credentials:**
- `POSTGRES_PASSWORD` - Database password (default: `test-password-change-in-production` in `docker-compose.test.yml`; **override for security**)
- `POSTGRES_USER` - Database username (default: `postgres`)
- `POSTGRES_DB` - Database name (default: `postgres`)

> **Note:** The default `POSTGRES_PASSWORD` is for convenience in local testing only. Always set a strong password for production or CI environments.

**Service URLs (Optional):**
- `ORCHESTRATION_ENGINE_URL` - Default: `http://127.0.0.1:4000/public/graphql`
- `PDP_URL` - Default: `http://127.0.0.1:8082/api/v1/policy`
- `CONSENT_ENGINE_URL` - Default: `http://127.0.0.1:8081/consents`
- `PORTAL_BACKEND_URL` - Default: `http://127.0.0.1:3000`

**Setting Variables:**
```bash
# Export before running
export POSTGRES_PASSWORD=your-password
go test ./...

# Or inline
POSTGRES_PASSWORD=your-password go test ./...
```

**Security Notes:**
- Never commit `.env` files or credentials
- Test credentials should differ from production
- `docker-compose.test.yml` uses environment variable substitution

---

## Running Tests

### Run All Tests
```bash
docker compose -f docker-compose.test.yml up -d
go test -v ./...
docker compose -f docker-compose.test.yml down -v
```

### Run Specific Test
```bash
go test -v -run TestGraphQLFlow_SuccessPath
go test -v -run TestGraphQLFlow_MissingPolicyMetadata
```

### Run with Verbose Output
```bash
go test -v ./...
```

---

## Test Services

The `docker-compose.test.yml` starts:

- **PostgreSQL Databases**:
  - `pdp-db` (5433) - Database for Policy Decision Point
  - `ce-db` (5434) - Database for Consent Engine
  - `audit-db` (5435) - Database for Audit Service
- **Policy Decision Point** (8082) - Policy evaluation service
- **Consent Engine** (8081) - Consent management service
- **Audit Service** (3001) - Audit logging service
- **Orchestration Engine** (4000) - GraphQL orchestration service

All services run on `test-network` Docker network.

---

## Test Data

Tests use unique identifiers (timestamp-based) for isolation:
- Schema IDs: `test-schema-123` (matches `schema.graphql`)
- App IDs: `test-consumer-app-{timestamp}`
- Test constants: `testNIC`, `testEmail`, `testOwnerID`

Tests automatically clean up created resources using `t.Cleanup()`.

---

## Troubleshooting

**Services not starting:**
```bash
# Check logs
docker compose -f docker-compose.test.yml logs

# Verify services are healthy
docker compose -f docker-compose.test.yml ps
```

**Database connection errors:**
- Verify databases are healthy: `docker compose -f docker-compose.test.yml ps pdp-db ce-db audit-db`
- Check database logs: `docker compose -f docker-compose.test.yml logs pdp-db`

**Port conflicts:**
```bash
# Check what's using ports
lsof -i :4000  # Orchestration Engine
lsof -i :8081  # Consent Engine
lsof -i :8082  # Policy Decision Point
lsof -i :5433  # PDP Database
lsof -i :5434  # Consent Engine Database
lsof -i :5435  # Audit Service Database
```

**Test failures:**
- Ensure all services are healthy before running tests
- Check service logs for errors
- Verify GraphQL schema matches test expectations (`schema.graphql`)

---

## Test Coverage

Tests cover:
- Complete GraphQL request/response flow
- Policy metadata and allowlist management
- Consent creation, retrieval, status updates, and expiry
- Audit event creation, retrieval, filtering, and pagination
- Management event tracking
- Authorization failures (missing metadata, unauthorized app)
- Service resilience (timeout scenarios)
- Invalid query handling
- Authentication (missing tokens)
- Service health checks
- Database state verification

---

## Architecture

```
Test Runner (go test)
    ↓
Docker Compose Services
    ├── PostgreSQL Databases (pdp-db, ce-db, audit-db)
    ├── Policy Decision Point (8082)
    ├── Consent Engine (8081)
    ├── Audit Service (3001)
    └── Orchestration Engine (4000)
    ↓
Test Utilities (testutils/)
    ├── HTTP client helpers
    └── Database helpers
```

---

## Contributing

When adding new tests:

1. **Use unique IDs** - Prevents test data conflicts
2. **Add cleanup** - Use `t.Cleanup()` for resource cleanup
3. **Use helpers** - Leverage `testutils` functions
4. **Document** - Add godoc comments to test functions
5. **Isolate** - Each test should be independent
6. **Include error handling** - Check for service availability
7. **Test edge cases** - Include both success and failure scenarios

---

## Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)

---

## License

This project is part of the OpenDIF platform.
