# Exchange Services

Microservices-based data exchange platform with policy enforcement and consent management.

## Quick Start

### Prerequisites
- Docker & Docker Compose
- curl, jq (for testing)

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

The `scripts/` directory contains 3 essential management scripts:
- **`common.sh`** - Common configuration and functions
- **`manage.sh`** - Consolidated management script for all operations
- **`test.sh`** - API testing with help functionality

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

**Choreo Configuration:**
- Each service has its own `.choreo/component.yaml` file
- Production environment variables are configured in Choreo Console
- Services are deployed independently and can be scaled separately

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