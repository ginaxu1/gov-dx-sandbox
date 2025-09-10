# Exchange Services

Microservices-based data exchange platform with policy enforcement and consent management.

## Important: Build State Management

**Before running any commands, ensure you're in the correct build state:**

```bash
# Restore to LOCAL DEVELOPMENT state (default)
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/restore-local-build.sh

# Prepare for PRODUCTION/CHOREO deployment
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/prepare-docker-build.sh
```

**Current State Check:**
- **Local Development**: go.mod files use `../shared/` paths
- **Production/Choreo**: go.mod files use `/app/shared/` paths

## Quick Start

### Prerequisites
- Docker & Docker Compose
- curl, jq (for testing)

**Note**: Make sure you're in local development state before running commands (see Build State Management above)

### Environment Configuration

```bash
# Local Development
docker compose --env-file .env.local up --build

# Production Testing
docker compose --env-file .env.production up --build

# Stop services
docker compose down
```

### Environment Files

| File | Purpose |
|------|---------|
| `.env.local` | Local development (debug logging) |
| `.env.production` | Production (warn logging, JSON format) |

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

# Direct Docker Compose
make local              # docker compose --env-file .env.local up --build
make prod               # docker compose --env-file .env.production up --build
```

## API Endpoints

| Service | Port | Endpoints |
|---------|------|-----------|
| **Policy Decision Point** | 8082 | `/decide`, `/health`, `/debug` |
| **Consent Engine** | 8081 | `/consent`, `/consent/{id}`, `/consent/portal`, `/data-owner/{owner}`, `/consumer/{consumer}`, `/admin/expiry-check`, `/health` |

### Example API Calls

```bash
# Policy Decision
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {"id": "test-app", "name": "Test App", "type": "mobile_app"},
    "request": {"resource": "person_data", "action": "read", "data_fields": ["person.fullName"]},
    "timestamp": "2025-09-09T16:30:00Z"
  }'

# Consent Management
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "test-app",
    "data_owner": "test-owner", 
    "data_fields": ["person.fullName"],
    "purpose": "testing",
    "expiry_days": 30
  }'

# Get Consent by ID
curl -X GET http://localhost:8081/consent/{consent-id}

# Get Consents by Data Owner
curl -X GET http://localhost:8081/data-owner/{owner-id}

# Get Consents by Consumer
curl -X GET http://localhost:8081/consumer/{consumer-id}
```

## Architecture

```
exchange/
├── policy-decision-point/    # Policy service (Port 8082)
├── consent-engine/           # Consent service (Port 8081)
├── docker-compose.yml        # Multi-environment orchestration
├── .env.local               # Local environment config
├── .env.production          # Production environment config
├── scripts/                 # Essential development scripts
│   ├── common.sh            # Common configuration
│   ├── manage.sh            # Consolidated management script
│   └── test.sh              # API testing
└── Makefile                 # Convenience commands
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

## Scripts

The `scripts/` directory contains essential management scripts:
- **`common.sh`** - Common configuration and functions
- **`manage.sh`** - Consolidated management script for all operations
- **`test.sh`** - API testing with help functionality
- **`restore-local-build.sh`** - Restore to local development state
- **`prepare-docker-build.sh`** - Prepare for production/Choreo deployment

### Build State Management Scripts

```bash
# Restore to LOCAL DEVELOPMENT (default state)
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/restore-local-build.sh

# Prepare for PRODUCTION/CHOREO deployment
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/prepare-docker-build.sh
```

### Usage
```bash
# Service management
./scripts/manage.sh start-local    # Start local environment
./scripts/manage.sh start-prod     # Start production environment
./scripts/manage.sh stop           # Stop all services
./scripts/manage.sh status         # Check service status and health
./scripts/manage.sh logs [service] # View logs (all or specific service)
./scripts/manage.sh help           # Show available commands

# Testing
./scripts/test.sh                  # Run API tests
./scripts/test.sh help             # Show test help
```

## Testing

```bash
# Unit tests
cd policy-decision-point && go test -v ./...
cd consent-engine && go test -v ./...

# Integration tests
cd integration-tests && ./run-all-tests.sh

# API tests
./scripts/test.sh
```

## Production Deployment

### WSO2 Choreo

**Important**: For production deployment, WSO2 Choreo manages the deployment process. You don't need to use Docker Compose directly in production.

**How it works:**
1. Choreo's CI/CD pipeline builds the Docker image from your service's Dockerfile
2. Choreo injects production environment variables at runtime (configured in Choreo Console)
3. Each service is deployed independently as a Choreo component

**Local Production Testing:**
```bash
# Test production configuration locally
docker compose --env-file .env.production up --build
```

### Choreo Deployment Workflow

**Step 1: Prepare for Choreo**
```bash
./scripts/prepare-docker-build.sh
```

This script:
- Updates go.mod files to use `/app/shared/` paths for Docker builds
- Prepares repository for Choreo deployment

**Step 2: Commit and Push**
```bash
git add .
git commit -m "Prepare for Choreo deployment"
git push origin main
```

**Step 3: Deploy to Choreo**

**Choreo Console Configuration:**
- **Component Directory (build context)**: `.` (repository root)
- **Dockerfile Path**: `exchange/consent-engine/Dockerfile`

**Why Root Build Context:**
- Allows access to `shared/` directory from repository root
- Uses single `Dockerfile` for both local and Choreo deployment
- Shared utilities copied via `COPY shared/ /app/shared/`
- No code duplication required

**Component.yaml Configuration:**
```yaml
build:
  type: docker
  context: .                    # Repository root
  dockerfile: exchange/consent-engine/Dockerfile
  args:
    - name: BUILD_VERSION
      value: latest
    - name: BUILD_TIME
      value: ""
    - name: GIT_COMMIT
      value: ""
```

**Step 4: Restore Local Development**
```bash
./scripts/restore-local-build.sh
```

This script:
- Restores go.mod files to use `../shared/` paths for local development
- Prepares repository for local development

### Dockerfile Usage

**Single Dockerfile for Both Environments**
- **Dockerfile**: Uses repository root build context for both local and Choreo
- **Build Context**: Repository root (`.`)
- **Shared Access**: `COPY shared/ /app/shared/` from repository root
- **Go Modules**: `/app/shared/` paths for Docker builds, `../shared/` for local development

**Dockerfile Structure:**
```dockerfile
# Multi-stage Dockerfile for Consent Engine
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy go mod files and source code
COPY consent-engine/go.mod consent-engine/go.sum ./
COPY consent-engine/ ./

# Copy shared packages from repository root
COPY shared/ /app/shared/

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service_binary .
```

**Go Module Configuration:**
```go
// go.mod with Docker paths (set by prepare-docker-build.sh)
replace github.com/gov-dx-sandbox/exchange/shared/config => /app/shared/config
replace github.com/gov-dx-sandbox/exchange/shared/constants => /app/shared/constants
replace github.com/gov-dx-sandbox/exchange/shared/utils => /app/shared/utils
replace github.com/gov-dx-sandbox/exchange/shared/handlers => /app/shared/handlers
```

### Key Benefits

1. **Single Dockerfile**: One Dockerfile works for both local and Choreo deployment
2. **No Code Duplication**: Shared utilities remain in single location (`exchange/shared/`)
3. **DRY Principles**: Maintains single source of truth for shared code
4. **Simplified Workflow**: Only go.mod paths change between environments
5. **Choreo Compatibility**: Works with Choreo's component-based architecture

### Troubleshooting

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

### Configuration Management
- **Local Development**: Use `.env.local` with Docker Compose
- **Production**: WSO2 Choreo manages environment variables at runtime

### Environment Switching Best Practices
1. Use `docker compose --env-file .env.local up --build` for local development
2. Use `docker compose --env-file .env.production up --build` for production testing
3. Always stop services before switching environments: `docker compose down`
4. Use `make clean` to remove old containers when switching
5. Check status with `make status` after switching

### Direct Docker Compose Usage
```bash
# Local development
docker compose --env-file .env.local up --build

# Production testing
docker compose --env-file .env.production up --build

# Stop services
docker compose down

# View logs
docker compose logs -f

# Clean up
docker compose down -v --remove-orphans
```