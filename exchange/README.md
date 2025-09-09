# Exchange Services

Data exchange platform implementing consent management and policy-based authorization for government data sharing.

## Services

- **Policy Decision Point (PDP)**: Port 8082 - ABAC authorization using Open Policy Agent
- **Consent Engine (CE)**: Port 8081 - Consent record management and portal

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.24+ (for local development)

### Start Services
```bash
# Start all services
docker-compose up -d

# Check health
curl http://localhost:8082/health  # PDP
curl http://localhost:8081/health  # CE

# View logs
docker-compose logs -f
```

### Run Tests
```bash
# Integration tests
cd integration-tests && ./run-all-tests.sh

# Unit tests
go test ./...
```

### Development
```bash
# Run locally
cd policy-decision-point && go run main.go
cd consent-engine && go run main.go
```

## Configuration

Unified configuration system with environment-based settings:

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | `local` | Environment (local/production) |
| `PORT` | 8081/8082 | Service port (auto-detected) |
| `LOG_LEVEL` | `debug`/`warn` | Log level |
| `LOG_FORMAT` | `text`/`json` | Log format |

## API Endpoints

### Policy Decision Point (Port 8082)
- `POST /decide` - Authorization decisions
- `GET /health` - Health check

### Consent Engine (Port 8081)
- `POST /consent` - Create consent record
- `GET /consent/{id}` - Get consent status
- `PUT /consent/{id}` - Update consent
- `GET /consent-portal/{id}` - Consent portal
- `GET /health` - Health check

## Deployment

### Docker
```bash
docker-compose up -d --build
```

### WSO2 Choreo
Services configured with component definitions in `.choreo/` directories.

## Project Structure

```
exchange/
├── config/                    # Configuration management
├── consent-engine/           # Consent service
├── policy-decision-point/    # Policy service
├── utils/                    # Shared utilities
├── integration-tests/        # Test suite
└── orchestration-engine-go/  # Orchestration service
```
