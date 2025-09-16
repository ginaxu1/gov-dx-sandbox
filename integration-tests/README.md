# Integration Tests

Comprehensive integration tests for the Data Exchange Platform covering end-to-end workflows, consent management, and policy decision scenarios.

## Overview

The integration tests validate the complete data exchange workflow from data request to consent management, ensuring all components work together correctly.

## Test Structure

```
integration-tests/
├── README.md                    # This file
├── run-all-tests.sh            # Run all integration tests
├── test-complete-flow.sh       # Complete end-to-end workflow test
├── test-consent-flow.sh        # Consent management workflow test
└── test-pdp.sh                 # Policy Decision Point test
```

## Prerequisites

Before running integration tests, ensure all services are running:

1. **Consent Engine** (Port 8081) 
2. **Policy Decision Point** (Port 8082)
3. **Orchestration Engine** (Port 4000)

### Starting Services

```bash
# Start all services using Docker Compose
cd /Users/tmp/gov-dx-sandbox
make start-exchange

# Or start individual services:
# Terminal 1 - Consent Engine
cd /Users/tmp/gov-dx-sandbox/exchange/consent-engine
go run main.go

# Terminal 2 - Policy Decision Point
cd /Users/tmp/gov-dx-sandbox/exchange/policy-decision-point
go run main.go

# Terminal 3 - Orchestration Engine
cd /Users/tmp/gov-dx-sandbox/exchange/orchestration-engine-go
go run main.go
```

## Test Scenarios

### 1. Complete Data Exchange Flow (`test-complete-flow.sh`)

Tests the complete end-to-end workflow from data request to data retrieval.

**Test Steps:**
1. Create provider profile
2. Submit and approve schema
3. Create consumer application
4. Request data access
5. Handle consent workflow (if required)
6. Retrieve authorized data

**Expected Outcome:**
- All services respond correctly
- Data access is properly authorized
- Consent workflow functions as expected

### 2. Consent Management Workflow (`test-consent-flow.sh`)

Tests consent management scenarios including different data ownership patterns.

**Test Scenarios:**

#### Scenario A: Data Owner is NOT the Provider
- **Setup**: Provider (DRP) requests data owned by RGD
- **Fields**: `person.permanentAddress`, `person.photo`
- **Expected**: Consent required, SMS OTP sent to data owner
- **Workflow**:
  1. PDP determines consent required
  2. Consent Engine creates consent record
  3. Data owner receives SMS OTP
  4. Data owner approves/denies consent
  5. Data access proceeds based on consent decision

#### Scenario B: Data Owner IS the Provider
- **Setup**: Provider (DRP) requests data owned by DRP
- **Fields**: `person.fullName`, `person.nic`
- **Expected**: No consent required, direct access
- **Workflow**:
  1. PDP determines no consent required
  2. Data access proceeds immediately
  3. No consent record created

#### Scenario C: Mixed Ownership
- **Setup**: Provider requests data from multiple owners
- **Fields**: `person.fullName` (DRP), `person.birthDate` (RGD)
- **Expected**: Consent required only for RGD fields
- **Workflow**:
  1. PDP determines consent required for RGD fields
  2. Consent Engine creates consent record for RGD data
  3. DRP data accessed immediately
  4. RGD data accessed after consent approval

### 3. Policy Decision Point Tests (`test-pdp.sh`)

Tests authorization decisions and consent requirements.

**Test Cases:**

#### Public Field Access
```bash
# Request: person.fullName (public field)
# Expected: ALLOWED, no consent required
```

#### Restricted Field Access (Authorized)
```bash
# Request: person.birthDate (restricted, app in allow_list)
# Expected: ALLOWED, no consent required
```

#### Restricted Field Access (Consent Required)
```bash
# Request: person.permanentAddress (restricted, cross-provider)
# Expected: ALLOWED, consent required
```

#### Unauthorized Access
```bash
# Request: person.nic (restricted, app not in allow_list)
# Expected: DENIED
```

## Running Tests

### Run All Tests
```bash
cd /Users/tmp/gov-dx-sandbox/integration-tests
./run-all-tests.sh
```

### Run Specific Test
```bash
# Complete flow test
./test-complete-flow.sh

# Consent workflow test
./test-consent-flow.sh

# Policy Decision Point test
./test-pdp.sh
```

## Test Data

### Provider Metadata
The tests use the following provider metadata structure:

```json
{
  "fields": {
    "person.fullName": {
      "owner": "citizen",
      "provider": "drp",
      "consent_required": false,
      "access_control_type": "public",
      "allow_list": []
    },
    "person.birthDate": {
      "owner": "rgd",
      "provider": "drp",
      "consent_required": false,
      "access_control_type": "restricted",
      "allow_list": [
        {
          "consumerId": "passport-app",
          "expires_at": 1757560679,
          "grant_duration": "30d"
        }
      ]
    },
    "person.permanentAddress": {
      "owner": "rgd",
      "provider": "drp",
      "consent_required": true,
      "access_control_type": "restricted",
      "allow_list": [
        {
          "consumerId": "passport-app",
          "expires_at": 1757560679,
          "grant_duration": "30d"
        }
      ]
    },
    "person.nic": {
      "owner": "citizen",
      "provider": "drp",
      "consent_required": true,
      "access_control_type": "restricted",
      "allow_list": []
    }
  }
}
```

### Allow List Construction

The `allow_list` is constructed through the following process:

1. **Schema Submission**: Provider submits GraphQL SDL schema
2. **Admin Approval**: Admin approves schema, triggering metadata generation
3. **Consumer Authorization**: Consumers are added to allow_list through:
   - Direct admin action
   - Consent approval workflow
   - Pre-approved MOUs (Memorandum of Understanding)
   - API integration

4. **Allow List Entry**: Each entry contains:
   - `consumerId`: Authorized application ID
   - `expires_at`: Epoch timestamp for expiry
   - `grant_duration`: Human-readable duration

### Test Data Setup

Before running tests, ensure the following data is set up:

1. **Provider Profile**: DRP provider profile exists
2. **Approved Schema**: Schema is approved and metadata generated
3. **Consumer Authorization**: `passport-app` is in allow_list for restricted fields
4. **Test Consumers**: Various test consumers with different authorization levels

### Test Applications
- **passport-app**: Authorized consumer for restricted fields
- **unauthorized-app**: Consumer not in allow_list
- **test-app**: General test consumer

## Expected Test Results

### Successful Test Run
```
✅ All services are running
✅ Complete data exchange flow passed
✅ Consent workflow (data owner ≠ provider) passed
✅ Consent workflow (data owner = provider) passed
✅ Mixed ownership consent workflow passed
✅ Policy Decision Point authorization tests passed
✅ All integration tests completed successfully
```

### Common Issues and Solutions

#### Service Not Running
```
❌ Error: Orchestration Engine not responding on port 4000
```
**Solution**: Start the services: `make start-exchange`

#### Consent Engine Error
```
❌ Error: Consent Engine not responding on port 8081
```
**Solution**: Start the Consent Engine: `cd consent-engine && go run main.go`

#### Policy Decision Point Error
```
❌ Error: Policy Decision Point not responding on port 8082
```
**Solution**: Start the PDP: `cd policy-decision-point && go run main.go`

#### Test Data Issues
```
❌ Error: Provider profile not found
```
**Solution**: Ensure test data is properly set up in provider-metadata.json

## Test Coverage

The integration tests cover:

1. **Service Health Checks**: All services are running and responsive
2. **API Endpoints**: All major endpoints are functional
3. **Data Flow**: Complete request-to-response workflow
4. **Authorization**: Proper access control enforcement
5. **Consent Management**: Consent creation, approval, and tracking
6. **Error Handling**: Proper error responses and status codes
7. **Edge Cases**: Invalid requests, missing data, unauthorized access
8. **Cross-Service Communication**: Services communicate correctly

## Debugging

### Enable Verbose Logging
```bash
# Set log level to debug
export LOG_LEVEL=debug
./run-all-tests.sh
```

### Check Service Logs
```bash
# Check API Server logs
tail -f /tmp/api-server.log

# Check Consent Engine logs
tail -f /tmp/consent-engine.log

# Check PDP logs
tail -f /tmp/pdp.log
```

### Manual Testing
```bash
# Test Orchestration Engine health
curl http://localhost:4000/health

# Test Consent Engine health
curl http://localhost:8081/health

# Test PDP health
curl http://localhost:8082/health
```

## Contributing

When adding new integration tests:

1. **Follow naming convention**: `test-{feature}-{scenario}.sh`
2. **Include error handling**: Check for service availability
3. **Add documentation**: Update this README with new test scenarios
4. **Test edge cases**: Include both success and failure scenarios
5. **Clean up**: Ensure tests don't leave test data in the system

## Troubleshooting

### Port Conflicts
If you get port conflicts, check what's running:
```bash
lsof -i :4000  # Orchestration Engine
lsof -i :8081  # Consent Engine
lsof -i :8082  # Policy Decision Point
```

### Permission Issues
Make sure test scripts are executable:
```bash
chmod +x *.sh
```

### Test Data Cleanup
If tests fail and leave test data, clean up:
```bash
# Reset API Server data (if using in-memory storage)
curl -X DELETE http://localhost:8080/admin/reset

# Reset Consent Engine data
curl -X DELETE http://localhost:8081/admin/reset
```