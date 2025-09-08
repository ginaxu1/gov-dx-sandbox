# Exchange Services Test Suite

Comprehensive test scripts for the Exchange Services: Policy Decision Point (PDP) and Consent Engine (CE).

## Test Scripts

### Core Tests

- **`test-pdp.sh`** - PDP policy logic and authorization tests
- **`test-consent-flow.sh`** - Basic consent flow integration tests
- **`test-complete-flow.sh`** - End-to-end flow simulation tests
- **`test-complete-consent-flow.sh`** - Full consent flow with Consent Engine integration

### Test Runner

- **`run-all-tests.sh`** - Executes all test suites in sequence

## Prerequisites

Before running the tests, ensure the following services are running:

```bash
# Start both services
cd /path/to/exchange
docker-compose up -d

# Or start individually
docker-compose up -d pdp_app    # Policy Decision Point on port 8080
docker-compose up -d ce_app     # Consent Engine on port 8081
```

## Running Tests

### Run All Tests
```bash
cd tests
./run-all-tests.sh
```

### Run Individual Tests
```bash
cd tests

# Test PDP policy logic
./test-pdp.sh

# Test consent flow
./test-consent-flow.sh

# Test complete flow
./test-complete-flow.sh

# Test complete consent flow
./test-complete-consent-flow.sh
```

## Test Coverage

### Policy Decision Point (PDP) Tests
- ✅ No consent required scenarios
- ✅ Consent required scenarios  
- ✅ Unauthorized access scenarios
- ✅ Invalid consumer scenarios
- ✅ Field authorization logic
- ✅ Consumer authorization logic
- ✅ Action authorization logic

### Consent Engine (CE) Tests
- ✅ Consent record creation
- ✅ Consent status updates
- ✅ Consent record retrieval
- ✅ Consent portal functionality
- ✅ Data owner consent management
- ✅ Consumer consent management

### Integration Tests
- ✅ PDP ↔ CE communication
- ✅ Complete consent flow (AppUser → App → DataCustodian → PDP → CE)
- ✅ End-to-end authorization flow
- ✅ Consent portal interactions
- ✅ Data access after consent

## Test Scenarios

### 1. No Consent Required
- App requests data that doesn't require consent
- Expected: Direct data access without consent flow

### 2. Consent Required
- App requests data that requires consent
- Expected: Consent flow triggered, consent required

### 3. Unauthorized Access
- App requests data it's not authorized to access
- Expected: Access denied without consent flow

### 4. Invalid Consumer
- Unknown consumer requests data
- Expected: Access denied, consumer not found

### 5. Complete Consent Flow
- Full flow: AppUser → App → DataCustodian → PDP → ConsentEngine
- Expected: Complete consent flow with data access after consent

## Service Endpoints

### Policy Decision Point (Port 8080)
- `POST /decide` - Authorization decision endpoint
- `GET /debug` - Debug information endpoint

### Consent Engine (Port 8081)
- `POST /consent` - Create consent record
- `GET /consent/{id}` - Get consent record
- `PUT /consent/{id}` - Update consent record
- `DELETE /consent/{id}` - Revoke consent record
- `GET /consent-portal/` - Consent portal information
- `GET /data-owner/{owner}` - Get consents by data owner
- `GET /consumer/{consumer}` - Get consents by consumer

## Troubleshooting

### Services Not Responding
```bash
# Check if containers are running
docker ps

# Check container logs
docker logs pdp_app
docker logs ce_app

# Restart services
docker-compose restart
```

### Test Failures
1. Ensure both services are running and healthy
2. Check that the correct ports (8080, 8081) are available
3. Verify that the latest code changes are deployed in containers
4. Check container logs for any errors

### Rebuilding Services
```bash
# Rebuild and restart services
docker-compose down
docker-compose build
docker-compose up -d
```

## Test Data

The tests use the following test data:

### Consumer Grants
- `passport-app`: Authorized for `person.fullName`, `person.nic`, `person.photo`

### Provider Metadata
- `person.fullName`: No consent required
- `person.nic`: No consent required  
- `person.photo`: No consent required
- `person.permanentAddress`: Consent required (30d expiry)
- `person.birthDate`: Consent required (30d expiry)
- `person.ssn`: Consent required (30d expiry)

## Contributing

When adding new tests:
1. Follow the existing naming convention (`test-*.sh`)
2. Include proper error handling and colored output
3. Update this README with new test descriptions
4. Ensure tests are executable (`chmod +x`)
5. Test from the `tests/` directory to ensure relative paths work correctly
