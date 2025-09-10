# Exchange Services

Microservices-based data exchange platform with policy enforcement and consent management.

## Important: Build State Management

**Before running any commands, ensure you're in the correct build state:**

```bash
# Local development (default)
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/restore-local-build.sh

# Production/Choreo deployment
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/prepare-docker-build.sh
```

**Current State Check:**
- **Local Development**: go.mod files use `../shared/` paths
- **Production/Choreo**: go.mod files use `/app/shared/` paths

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)

### Start Services
```bash
# Start services
docker-compose up -d

# Run tests
cd integration-tests && ./run-all-tests.sh
```

## Services

# Run individual test suites
./test-pdp.sh                    # PDP policy tests
./test-complete-flow.sh          # End-to-end flow tests
./test-consent-flow.sh           # Basic consent flow tests
./test-complete-consent-flow.sh  # Full consent flow tests
```

## What's in Exchange

### Services
- **Policy Decision Point (PDP)** - Port 8082: Authorization service using OPA and Rego policies
- **Consent Engine (CE)** - Port 8081: Consent management and workflow coordination

### Key Features
- **ABAC Authorization**: Attribute-based access control with field-level permissions
- **Consent Management**: Complete consent workflow for data owners
- **GraphQL Schema Support**: Convert GraphQL SDL to provider metadata
- **Multi-Environment**: Local development and production deployment support
- **Docker Ready**: Containerized services with Docker Compose orchestration

## Architecture

The Data Exchange Platform implements a comprehensive consent-based data sharing system with policy enforcement and consent management. The platform consists of three main services coordinated by the Orchestration Engine.

### Service Architecture

- **Orchestration Engine (OE)**: Coordinates data access requests between PDP and Consent Engine
- **Policy Decision Point (PDP) - Port 8082**: ABAC authorization using Open Policy Agent (OPA)
- **Consent Engine (CE) - Port 8081**: Manages data owner consent workflow

### Directory Structure

```
exchange/
├── policy-decision-point/    # Policy service (Port 8082)
├── consent-engine/           # Consent service (Port 8081)
├── shared/                   # Shared utilities and packages
├── docker-compose.yml        # Multi-environment orchestration
├── .env.local               # Local environment config
├── .env.production          # Production environment config
├── scripts/                 # Essential development scripts
└── Makefile                 # Convenience commands
```

## Data Flow

### 1. Data Consumer Request

The Data Consumer sends a GetData request to the Orchestration Engine:

```json
{
  "dataConsumer": {
    "id": "passport-app",
    "type": "application"
  },
  "dataOwner": {
    "id": "user-nuwan-fernando-456"
  },
  "request": {
    "type": "GraphQL",
    "query_fields": [
      "fullName",
      "nic",
      "birthDate",
      "permanentAddress"
    ]
  }
}
```

### 2. Policy Decision Point (PDP) Evaluation

The Orchestration Engine forwards the request to the Policy Decision Point (PDP) for authorization:

- **PDP Service**: Embedded OPA engine evaluates Rego policies
- **Data Sources**: 
  - `/policies/main.rego` - Authorization rules
  - `provider-metadata.json` - Field-level metadata and consent requirements

### 3. Authorization Decision Scenarios

#### Access Permitted + Consent Required
```json
{
  "decision": {
    "allow": true,
    "deny_reason": null,
    "consent_required": true,
    "consent_required_fields": [
      "person.permanentAddress",
      "person.birthDate"
    ],
    "data_owner": "user-nuwan-fernando-456",
    "expiry_time": "30d",
    "conditions": {}
  }
}
```

#### Access Denied/Insufficient Permissions
```json
{
  "decision": {
    "allow": false,
    "deny_reason": "Consumer not authorized for requested fields",
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
  }
}
```

#### Access Permitted + No Consent Required
```json
{
  "decision": {
    "allow": true,
    "deny_reason": null,
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
  }
}
```

### 4. Orchestration Engine Response

Based on the PDP decision, the Orchestration Engine:

- **If `allow: false`**: Immediately rejects the request and sends an error back to the Data Consumer
- **If `allow: true` and consent required**: Makes a second API call to the Consent Management Engine (CE)
- **If `allow: true` and no consent needed**: Proceeds to fetch data from Data Providers

## API Reference

| Service | Port | Endpoints | Documentation |
|---------|------|-----------|---------------|
| **Policy Decision Point** | 8082 | `/decide`, `/health`, `/debug` | [PDP README](policy-decision-point/README.md) |
| **Consent Engine** | 8081 | `/consent`, `/consent/{id}`, `/consent/portal`, `/data-owner/{owner}`, `/consumer/{consumer}`, `/admin/expiry-check`, `/health` | [CE README](consent-engine/README.md) |

### Quick API Examples

```bash
# Policy Decision
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{"consumer": {"id": "passport-app"}, "request": {"resource": "person_data", "action": "read", "data_fields": ["person.fullName"]}}'

# Consent Management  
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{"consumer_id": "passport-app", "data_owner": "user-123", "data_fields": ["person.permanentAddress"], "purpose": "passport application", "expiry_days": 30}'
```

> **For detailed API documentation and examples, see individual service READMEs.**

## Building for Local Development

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for local development)

### Quick Commands
```bash
# Start services
make start              # Local (default)
make start-prod         # Production

# Management
make stop              # Stop all services
make status            # Check service status
make logs              # View logs
make test              # Run API tests
make clean             # Clean up containers
```

### Environment Configuration
```bash
# Local Development
docker compose --env-file .env.local up --build

# Production Testing
docker compose --env-file .env.production up --build

# Stop services
docker compose down
```

### Testing

#### Unit Tests
```bash
# Policy Decision Point
cd policy-decision-point && go test -v

# Consent Engine  
cd consent-engine && go test -v
```

#### Integration Tests
```bash
# All integration tests
cd integration-tests && ./run-all-tests.sh

# Individual test suites
./test-pdp.sh                    # PDP policy tests
./test-consent-flow.sh           # Basic consent flow tests
./test-complete-flow.sh          # End-to-end flow tests
./test-complete-consent-flow.sh  # Full consent flow tests
```

#### API Tests
```bash
# Basic API tests
./scripts/test.sh

# Service management
./scripts/manage.sh status       # Check service health
./scripts/manage.sh logs         # View service logs
```

> **For detailed testing information, see [Integration Tests README](integration-tests/README.md)**

## Building for Production (Choreo)

### Step 1: Prepare for Choreo
```bash
./scripts/prepare-docker-build.sh
```

### Step 2: Commit and Push
```bash
git add .
git commit -m "Prepare for Choreo deployment"
git push origin main
```

### Step 3: Deploy to Choreo

**Choreo Console Configuration:**
- **Component Directory (build context)**: `.` (repository root)
- **Dockerfile Path**: `exchange/consent-engine/Dockerfile` or `exchange/policy-decision-point/Dockerfile`

**Why Root Build Context:**
- Allows access to `shared/` directory from repository root
- Uses single `Dockerfile` for both local and Choreo deployment
- Shared utilities copied via `COPY shared/ /app/shared/`
- No code duplication required

### Step 4: Restore Local Development
```bash
./scripts/restore-local-build.sh
```

## Configuration

### Environment Variables

**Local (`.env.local`):**
```bash
ENVIRONMENT=local
BUILD_VERSION=dev
LOG_LEVEL=debug
LOG_FORMAT=text
```

**Production (`.env.production`):**
```bash
ENVIRONMENT=production
BUILD_VERSION=latest
LOG_LEVEL=warn
LOG_FORMAT=json
```

### Environment Files

| File | Purpose |
|------|---------|
| `.env.local` | Local development (debug logging) |
| `.env.production` | Production (warn logging, JSON format) |

## Scripts

Essential management scripts in `scripts/` directory:

| Script | Purpose |
|--------|---------|
| `common.sh` | Common configuration and functions |
| `manage.sh` | Consolidated service management |
| `test.sh` | API testing with help functionality |
| `restore-local-build.sh` | Restore to local development state |
| `prepare-docker-build.sh` | Prepare for production/Choreo deployment |

### Quick Commands
```bash
# Service management
./scripts/manage.sh start-local    # Start local environment
./scripts/manage.sh start-prod     # Start production environment
./scripts/manage.sh stop           # Stop all services
./scripts/manage.sh status         # Check service status and health
./scripts/manage.sh logs [service] # View logs (all or specific service)

# Testing
./scripts/test.sh                  # Run API tests
./scripts/test.sh help             # Show test help
```

> **For detailed script documentation, see [Scripts README](scripts/README.md)**

## Troubleshooting

**Issue: "shared directory not found" in Choreo build**
- **Solution**: Ensure build context is set to repository root (`.`) in Choreo console

**Issue: Local tests fail with import errors**
- **Solution**: Run `./scripts/restore-local-build.sh` to restore local development setup

**Issue: Docker build fails with go.mod errors**
- **Solution**: Run `./scripts/prepare-docker-build.sh` before building for Choreo

**Issue: Choreo build fails with COPY errors**
- **Solution**: Verify Component Directory is set to `.` (repository root) in Choreo console

## Development

### Adding New Services
1. Create service directory with `Dockerfile`
2. Add to `docker-compose.yml`
3. Add health checks
4. Update environment files if needed

### Environment Switching Best Practices
1. Use `docker compose --env-file .env.local up --build` for local development
2. Use `docker compose --env-file .env.production up --build` for production testing
3. Always stop services before switching environments: `docker compose down`
4. Use `make clean` to remove old containers when switching
5. Check status with `make status` after switching