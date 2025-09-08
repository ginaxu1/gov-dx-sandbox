# Exchange Services

## Overview

The exchange has three services: Orchestration Engine (coordinates requests), Policy Decision Point (authorization), and Consent Engine (consent workflow).

### How it works:

The **Orchestration Engine (OE)** coordinates between the PDP and CE to get a final answer on data access. The PDP and CE each run as self-contained services in their own Docker containers and are coordinated by the OE:

1. **App** sends data request to **Orchestration Engine**
2. **Orchestration Engine** forwards request to **PDP** for authorization decision
3. **PDP** evaluates request and determines if consent is required
4. If consent required, **Orchestration Engine** coordinates with **Consent Engine**
5. **Consent Engine** manages consent workflow and returns consent status
6. **Orchestration Engine** makes final data access decision based on PDP and CE responses
7. **Orchestration Engine** returns final decision to **App**

## Services

### Orchestration Engine (OE)
- **Purpose**: Coordinates data access requests between PDP and Consent Engine
- **Technology**: Go + HTTP client coordination
- **Function**: Manages complete consent flow, makes final data access decisions

### Policy Decision Point (PDP) - Port 8080
- **Purpose**: ABAC authorization using Open Policy Agent (OPA)
- **Technology**: Go + Rego policies
- **Function**: Evaluates requests and determines consent requirements
- **Documentation**: [policy-decision-point/README.md](policy-decision-point/README.md)

### Consent Engine (CE) - Port 8081
- **Purpose**: Manages data owner consent workflow
- **Technology**: Go + In-memory storage
- **Function**: Creates, manages, and tracks consent records
- **Documentation**: [consent-engine/README.md](consent-engine/README.md)

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)

### Start Services
```bash
# Start both services
docker-compose up -d

# Check service status
docker ps
```

### Run Tests
```bash
# Run all tests
cd integration-tests
./run-all-tests.sh

# Run individual test suites
./test-pdp.sh                    # PDP policy tests
./test-complete-flow.sh          # End-to-end flow tests
./test-consent-flow.sh           # Basic consent flow tests to verify services are working
./test-complete-consent-flow.sh  # Full consent flow tests for integration testing before releases
```

## Consent Flow

The platform implements the complete consent flow coordinated by the Orchestration Engine:

1. **AppUser** initiates login request to **App**
2. **App** requests data from **Orchestration Engine**
3. **Orchestration Engine** forwards request to **PDP** for authorization
4. **PDP** evaluates request and determines if consent is required
5. If consent required, **Orchestration Engine** coordinates with **Consent Engine**
6. **Consent Engine** manages consent workflow and user interaction
7. **Data Owner** provides consent via consent portal
8. **Orchestration Engine** makes final data access decision
9. **Orchestration Engine** returns final decision to **App**

## API Endpoints

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

## Configuration

### Consumer Grants
Defines which consumers can access which data fields:
```json
{
  "passport-app": {
    "approved_fields": [
      "person.fullName",
      "person.nic", 
      "person.photo"
    ]
  }
}
```

### Provider Metadata
Defines field-level consent requirements and ownership:
```json
{
  "fields": {
    "person.fullName": { "consent_required": false, "owner": "drp" },
    "person.permanentAddress": {
      "consent_required": true,
      "owner": "drp",
      "expiry_time": "30d"
    }
  }
}
```

## Development

### Local Development
```bash
# Run PDP locally
cd policy-decision-point
go run .

# Run Consent Engine locally
cd consent-engine
go run .
```

### Container Development
```bash
# Rebuild and restart services
docker-compose down
docker-compose build
docker-compose up -d

# View logs
docker logs pdp_app
docker logs ce_app
```

## Testing

```bash
# Run all tests
cd integration-tests && ./run-all-tests.sh

# Run individual test suites
./test-pdp.sh                    # PDP authorization tests
./test-consent-flow.sh           # Basic consent flow tests
./test-complete-consent-flow.sh  # Full integration tests
```

See [tests/README.md](tests/README.md) for detailed test documentation.

## Recent Updates

### Policy Decision Point
- Fixed non-consent-required field handling
- Improved consent field analysis
- Enhanced error handling
- Updated Docker containers

### Consent Engine
- Fixed consent endpoint routing
- Improved error handling
- Enhanced response consistency
- Updated Docker containers

### Test Infrastructure
- Centralized test scripts in `/tests` directory
- Added comprehensive test coverage
- Created automated test runner
- Improved test documentation

## Troubleshooting

### Services Not Responding
```bash
# Check container status
docker ps

# Check logs
docker logs pdp_app
docker logs ce_app

# Restart services
docker-compose restart
```

### Test Failures
1. Ensure both services are running and healthy
2. Check that ports 8080 and 8081 are available
3. Verify latest code changes are deployed
4. Check container logs for errors

### Port Conflicts
```bash
# Check port usage
lsof -i :8080
lsof -i :8081

# Kill processes using ports
lsof -ti:8080 | xargs kill -9
lsof -ti:8081 | xargs kill -9
```

## Security Considerations

- Default deny policy ensures secure-by-default behavior
- All requests are logged with decision outcomes
- Input validation prevents malformed requests
- Consent requirements are enforced at the field level
- Consumer authorization is verified against grants data
- Resource access is validated against metadata
- Non-root user execution in containers

## Contributing

When contributing to the Exchange Services:

1. Follow the existing code structure and patterns
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure all tests pass before submitting
5. Use the centralized test runner for validation
