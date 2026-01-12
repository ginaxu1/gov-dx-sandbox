# Audit Service

A generic Go microservice for centralized audit logging across distributed services, providing distributed tracing and comprehensive event tracking.

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](../LICENSE)

## Overview

The Audit Service tracks "who did what, when, and with what result" by providing APIs for logging and querying audit events across distributed services. It supports flexible database backends and distributed tracing with trace IDs.

**Key Features:**

- 📝 Create and retrieve audit logs via REST API
- 🔍 Filter by trace ID, event type, status, and more
- 🗄️ Multiple database backends (SQLite, PostgreSQL)
- 🚀 Zero configuration - works out of the box with in-memory database
- 📊 Distributed tracing support
- 🔌 CORS-enabled for cross-origin requests

## Quick Start

### Prerequisites

- Go 1.21 or higher
- (Optional) PostgreSQL for production deployments

### Installation & Running

```bash
# Clone the repository (if not already cloned)
git clone https://github.com/OpenDIF/opendif-core.git
cd opendif-core/audit-service

# Install dependencies
go mod tidy

# Run with in-memory database (no configuration needed)
go run .
```

Service starts on `http://localhost:3001`

### Test the API

```bash
# Health check
curl http://localhost:3001/health

# Create an audit log
curl -X POST http://localhost:3001/api/audit-logs \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2024-01-20T10:00:00Z",
    "status": "SUCCESS",
    "actorType": "SERVICE",
    "actorId": "test-service",
    "targetType": "RESOURCE",
    "eventType": "TEST_EVENT"
  }'

# Get audit logs
curl http://localhost:3001/api/audit-logs
```

## Configuration

### Database Options

The service supports three database modes:

| Mode                  | Configuration                    | Use Case                     |
| --------------------- | -------------------------------- | ---------------------------- |
| **In-Memory SQLite**  | No config needed                 | Development, testing         |
| **File-Based SQLite** | `DB_TYPE=sqlite` OR `DB_PATH` set | Single-server deployments    |
| **PostgreSQL**        | `DB_TYPE=postgres` + credentials | Production, high concurrency |

**Examples:**

```bash
# In-memory (default - no configuration)
go run .

# File-based SQLite (option 1: explicit DB_TYPE)
export DB_TYPE=sqlite
export DB_PATH=./data/audit.db
go run .

# File-based SQLite (option 2: DB_PATH alone)
export DB_PATH=./data/audit.db
go run .

# PostgreSQL
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_USERNAME=postgres
export DB_PASSWORD=your_password
export DB_NAME=audit_db
go run .
```

See [docs/DATABASE_CONFIGURATION.md](docs/DATABASE_CONFIGURATION.md) for complete database setup guide.

### Environment Variables

Copy `.env.example` to `.env` and configure:

| Variable               | Default                 | Description                                 |
| ---------------------- | ----------------------- | ------------------------------------------- |
| `PORT`                 | `3001`                  | Service port                                |
| `DB_TYPE`              | -                       | Database type: `sqlite` or `postgres`. If not set, uses in-memory SQLite |
| `DB_PATH`              | `./data/audit.db`       | SQLite database path (only used when `DB_TYPE=sqlite` or `DB_PATH` is explicitly set) |
| `LOG_LEVEL`            | `info`                  | Log level: `debug`, `info`, `warn`, `error` |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Allowed CORS origins                        |

For PostgreSQL configuration and advanced settings, see [.env.example](.env.example).

### Event Type Configuration

Event types are configurable via `config/enums.yaml`. This allows you to customize the audit service for your specific use case. The service comes with generic default event types, but you can add project-specific ones.

**Default Event Types:**
- `MANAGEMENT_EVENT` - Administrative operations
- `USER_MANAGEMENT` - User-related operations
- `DATA_FETCH` - Data retrieval operations

**Customizing Event Types:**

Edit `config/enums.yaml` to add your own event types:

```yaml
enums:
  eventTypes:
    - MANAGEMENT_EVENT
    - USER_MANAGEMENT
    - DATA_FETCH
    - YOUR_CUSTOM_EVENT_TYPE
    - ANOTHER_EVENT_TYPE
```

See [config/README.md](config/README.md) for detailed configuration options.

## API Endpoints

### Core Endpoints

| Method | Endpoint          | Description                              |
| ------ | ----------------- | ---------------------------------------- |
| POST   | `/api/audit-logs` | Create audit log entry                   |
| GET    | `/api/audit-logs` | Retrieve audit logs (filtered/paginated) |
| GET    | `/health`         | Health check                             |
| GET    | `/version`        | Version information                      |

### Quick API Examples

**Create Audit Log:**

```bash
curl -X POST http://localhost:3001/api/audit-logs \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-01-20T10:00:00Z",
    "eventType": "MANAGEMENT_EVENT",
    "eventAction": "READ",
    "status": "SUCCESS",
    "actorType": "SERVICE",
    "actorId": "my-service",
    "targetType": "SERVICE",
    "targetId": "target-service"
  }'
```

**Get Audit Logs:**

```bash
# All logs
curl http://localhost:3001/api/audit-logs

# Filter by trace ID
curl http://localhost:3001/api/audit-logs?traceId=550e8400-e29b-41d4-a716-446655440000

# Filter by event type
curl http://localhost:3001/api/audit-logs?eventType=MANAGEMENT_EVENT&status=SUCCESS
```

## Development

### Project Structure

```
audit-service/
├── config/          # Configuration management
├── database/        # Database connection layer
├── middleware/      # HTTP middleware (CORS)
├── v1/              # API Version 1
│   ├── database/    # Repository interface & implementation
│   ├── handlers/    # HTTP handlers
│   ├── models/      # Domain models & DTOs
│   ├── services/    # Business logic
│   └── testutil/    # Test utilities
├── docs/            # Documentation
└── main.go          # Entry point
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with verbose output
go test ./... -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Tests use in-memory SQLite and require no external dependencies.

### Building

```bash
# Build binary
go build -o audit-service

# Run binary
./audit-service

# Build with version information
go build -ldflags="-X main.Version=1.0.0 -X main.GitCommit=$(git rev-parse HEAD)" -o audit-service
```

## Integration

The Audit Service is designed to be integrated into any microservices architecture. It provides a simple REST API that can be called from any service.

### Integration Pattern

1. **Configure your service** to point to the audit service URL
2. **Send audit events** via HTTP POST to `/api/audit-logs`
3. **Query audit logs** via HTTP GET from `/api/audit-logs`

### Client Libraries

You can integrate the audit service using:

- **HTTP Client**: Direct HTTP calls to the REST API
- **Shared Client Package**: Use the `shared/audit` package (if available in your project)
- **Custom Wrapper**: Create your own client library

### Example Integration

```go
// Example: Log an audit event from your service
auditRequest := map[string]interface{}{
    "traceId":    traceID,
    "timestamp":  time.Now().UTC().Format(time.RFC3339),
    "eventType":  "YOUR_EVENT_TYPE",
    "status":     "SUCCESS",
    "actorType":  "SERVICE",
    "actorId":    "your-service-name",
    "targetType": "RESOURCE",
    "targetId":   "resource-id",
}

// POST to http://audit-service:3001/api/audit-logs
```

### Graceful Degradation

- Services continue to function normally if audit service is unavailable
- No errors are thrown when audit service URL is not configured
- Audit operations should be asynchronous (fire-and-forget) to avoid blocking requests
- Services can be started before audit service is ready

## Deployment

### Docker

```bash
# Build image
docker build -t audit-service .

# Run with file-based SQLite
docker run -d \
  -p 3001:3001 \
  -v audit-data:/data \
  -e DB_TYPE=sqlite \
  -e DB_PATH=/data/audit.db \
  audit-service

# Run with PostgreSQL
docker run -d \
  -p 3001:3001 \
  -e DB_TYPE=postgres \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=your_password \
  -e DB_NAME=audit_db \
  audit-service
```

### Docker Compose

The audit service includes a `docker-compose.yml` for standalone deployment:

```bash
# Deploy audit service
cd audit-service
docker compose up -d

# View logs
docker compose logs -f

# Stop service
docker compose down
```

### Production Considerations

1. **Database**: Use PostgreSQL for production deployments
2. **Logging**: Set `LOG_LEVEL=info` or `LOG_LEVEL=warn` in production
3. **CORS**: Configure `CORS_ALLOWED_ORIGINS` appropriately
4. **Monitoring**: Monitor service health via `/health` endpoint
5. **Backup**: Implement database backup strategy
6. **High Availability**: Consider deploying multiple instances behind a load balancer
7. **Security**: Implement authentication/authorization if exposing the service publicly

## Troubleshooting

### Common Issues

**Service won't start:**

- Check port 3001 is available: `lsof -i :3001`
- Verify database configuration
- Check logs for error messages

**Database locked error (SQLite):**

- Ensure `DB_MAX_OPEN_CONNS=1` (default)
- Switch to PostgreSQL for high concurrency

**Connection timeout (PostgreSQL):**

- Verify database is running and accessible
- Check credentials and SSL settings
- Verify network connectivity

**Event type validation errors:**

- Check that your event types are defined in `config/enums.yaml`
- Verify the enum configuration file is being loaded correctly
- Check service logs for validation error details

See [docs/DATABASE_CONFIGURATION.md](docs/DATABASE_CONFIGURATION.md) for detailed troubleshooting.

## Contributing

We welcome contributions! Please see:

- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Architecture overview
- [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md) - Code of conduct

## License

This project is licensed under the Apache License 2.0 - see [LICENSE](../LICENSE) for details.

## Support

For issues, questions, or contributions, please use the project's issue tracker and discussion forums.
