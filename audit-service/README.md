# Audit Service

A Go microservice for managing audit logs, providing both read and write operations for generalized audit events.

## Overview

The Audit Service answers: "who made this request, what did they ask for, and did they get that data" by providing APIs for logging and querying audit events. It provides a generalized, reusable audit logging solution that can be used across different services.

**Note:** Audit service is **optional**. Services can function normally with or without audit logging enabled. See [AUDIT_SERVICE.md](../exchange/AUDIT_SERVICE.md) for configuration details.

## Features

- **Read and write operations** for generalized audit logs
- **Distributed tracing support** - Track requests across services using trace IDs
- **Flexible event classification** - Support for custom event types and actions
- **Advanced filtering** by trace ID, event type, and status
- **Pagination support** for large datasets
- **CORS enabled** for cross-origin requests

## Quick Start

### Prerequisites

- Go 1.21+
- No database setup required (uses SQLite by default)

### Run the Service

```bash
# Install dependencies
go mod tidy

# Run the service (uses SQLite database at ./data/audit.db by default)
go run main.go

# Or build and run
go build -o audit-service
./audit-service
```

The service runs on port 3001 by default and automatically creates a SQLite database at `./data/audit.db` if it doesn't exist.

## Configuration

### Database Configuration

The audit-service supports two database options:

#### SQLite (Default)

SQLite works out of the box with no external setup. Supports both file-based and in-memory databases.

**Environment Variables:**
- `DB_TYPE` - Set to `sqlite` (default, can be omitted)
- `DB_PATH` - Path to SQLite database file (default: `./data/audit.db`)
  - Use `:memory:` for in-memory database (data lost on restart)

**Examples:**
```bash
# File-based SQLite (default)
go run main.go

# In-memory SQLite (for testing)
DB_PATH=:memory: go run main.go

# Custom file path
DB_PATH=/var/lib/audit/audit.db go run main.go
```

#### PostgreSQL (External Database)

Use PostgreSQL for production deployments requiring high concurrency.

**Environment Variables:**
- `DB_TYPE` - Set to `postgres`
- `DB_HOST` - Database host (default: `localhost`)
- `DB_PORT` - Database port (default: `5432`)
- `DB_USERNAME` - Database username (default: `postgres`)
- `DB_PASSWORD` - Database password (required)
- `DB_NAME` - Database name (default: `audit_db`)
- `DB_SSLMODE` - SSL mode (default: `disable`)

**Example:**
```bash
DB_TYPE=postgres \
DB_HOST=localhost \
DB_PORT=5432 \
DB_USERNAME=postgres \
DB_PASSWORD=your_password \
DB_NAME=audit_db \
go run main.go
```

**Switching:**
Simply change `DB_TYPE` environment variable:
- `DB_TYPE=sqlite` (or omit) → SQLite
- `DB_TYPE=postgres` → PostgreSQL

### Service Configuration

**Environment Variables:**
- `PORT` - Service port (default: `3001`)
- `ENVIRONMENT` - Environment mode (default: `production`)
- `AUDIT_ENUMS_CONFIG` - Path to enum configuration YAML file (default: `config/enums.yaml`)

## API Endpoints

### Audit Logs
- `GET /api/audit-logs` - Retrieve audit logs with filtering and pagination
- `POST /api/audit-logs` - Create a new audit log entry

### System Endpoints
- `GET /health` - Service health check
- `GET /version` - Service version information

### Query Parameters

**Get Audit Logs** (`GET /api/audit-logs`) supports:
- `traceId` - Filter by trace ID (UUID format)
- `eventType` - Filter by event type (e.g., `POLICY_CHECK`, `MANAGEMENT_EVENT`)
- `limit` - Maximum number of logs to return (default: 100, max: 1000)
- `offset` - Number of logs to skip for pagination (default: 0)

### Request Format

**Create Audit Log** (`POST /api/audit-logs`):

**Required Fields:**
- `timestamp` - ISO 8601 timestamp (e.g., `2024-01-20T10:00:00Z`)
- `status` - Event status: `SUCCESS` or `FAILURE`
- `actorType` - Actor type: `SERVICE`, `ADMIN`, `MEMBER`, or `SYSTEM`
- `actorId` - Actor identifier (email, UUID, or service name)
- `targetType` - Target type: `SERVICE` or `RESOURCE`

**Optional Fields:**
- `traceId` - UUID for distributed tracing (nullable for standalone events)
- `eventType` - User-defined event type (e.g., `POLICY_CHECK`, `MANAGEMENT_EVENT`)
- `eventAction` - Event action: `CREATE`, `READ`, `UPDATE`, `DELETE`
- `targetId` - Target identifier (resource ID or service name)
- `requestMetadata` - JSON object with request payload (without PII/sensitive data)
- `responseMetadata` - JSON object with response or error details
- `additionalMetadata` - JSON object with additional context-specific data

**Example Request:**
```json
{
  "traceId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-20T10:00:00Z",
  "eventType": "POLICY_CHECK",
  "eventAction": "READ",
  "status": "SUCCESS",
  "actorType": "SERVICE",
  "actorId": "orchestration-engine",
  "targetType": "SERVICE",
  "targetId": "policy-decision-point",
  "requestMetadata": {
    "schemaId": "schema-123",
    "requestedFields": ["name", "address"]
  },
  "responseMetadata": {
    "decision": "ALLOWED"
  }
}
```

### Response Format

**Get Audit Logs Response:**
```json
{
  "logs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "timestamp": "2024-01-20T10:00:00Z",
      "traceId": "550e8400-e29b-41d4-a716-446655440000",
      "eventType": "POLICY_CHECK",
      "eventAction": "READ",
      "status": "SUCCESS",
      "actorType": "SERVICE",
      "actorId": "orchestration-engine",
      "targetType": "SERVICE",
      "targetId": "policy-decision-point",
      "requestMetadata": {
        "schemaId": "schema-123"
      },
      "responseMetadata": {
        "decision": "ALLOWED"
      },
      "additionalMetadata": null,
      "createdAt": "2024-01-20T10:00:00Z"
    }
  ],
  "total": 100,
  "limit": 100,
  "offset": 0
}
```

**Create Audit Log Response:**
Returns the created audit log entry in the same format as above.

For complete API documentation, see [openapi.yaml](./openapi.yaml).

## Testing

### Local Test Setup

Tests use SQLite by default and create temporary database files. No additional setup is required.

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover
```

**Note**: Tests use in-memory SQLite databases and clean up automatically after execution.

## Health Check

- `GET /health` - Returns service health status

## Security

- **CORS enabled**: Cross-origin requests are supported via CORS middleware
- **Internal service**: Service is intended for internal use within the OpenDIF ecosystem
- **No authentication required**: Authentication is handled by upstream services (Orchestration Engine, Portal Backend)
- **Database security**: Uses configurable SSL modes for database connections

## Deployment

The Audit Service is **optional** and can be deployed separately from the main OpenDIF services. Services (Orchestration Engine, Portal Backend) will function normally without audit logging enabled.

### Deployment Steps

**Using Docker Compose:**
```bash
cd audit-service
docker compose up -d
```

**Using Docker:**
```bash
# Build image
docker build -t audit-service .

# Run container
docker run -d \
  -p 3001:3001 \
  -e DB_PATH=/data/audit.db \
  -v audit-data:/data \
  --name audit-service \
  audit-service
```

**Using Standalone Binary:**
```bash
cd audit-service
go build -o audit-service
./audit-service
```

### Verify Deployment

1. Check service health: `GET http://localhost:3001/health`
2. Check service version: `GET http://localhost:3001/version`
3. Query audit logs: `GET http://localhost:3001/api/audit-logs`

### Environment Variables Summary

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_TYPE` | `sqlite` | Database type: `sqlite` or `postgres` |
| `DB_PATH` | `./data/audit.db` | SQLite database file path (when `DB_TYPE=sqlite`) |
| `DB_HOST` | `localhost` | PostgreSQL host (when `DB_TYPE=postgres`) |
| `DB_PORT` | `5432` | PostgreSQL port (when `DB_TYPE=postgres`) |
| `DB_USERNAME` | `postgres` | PostgreSQL username (when `DB_TYPE=postgres`) |
| `DB_PASSWORD` | - | PostgreSQL password (required when `DB_TYPE=postgres`) |
| `DB_NAME` | `audit_db` | PostgreSQL database name (when `DB_TYPE=postgres`) |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode (when `DB_TYPE=postgres`) |
| `DB_MAX_OPEN_CONNS` | `1` (SQLite) / `25` (PostgreSQL) | Maximum open database connections |
| `DB_MAX_IDLE_CONNS` | `1` (SQLite) / `5` (PostgreSQL) | Maximum idle database connections |
| `DB_CONN_MAX_LIFETIME` | `1h` | Connection max lifetime |
| `DB_CONN_MAX_IDLE_TIME` | `15m` | Connection max idle time |
| `PORT` | `3001` | Service port |
| `ENVIRONMENT` | `production` | Environment mode |
| `AUDIT_ENUMS_CONFIG` | `config/enums.yaml` | Enum configuration file path |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Comma-separated list of allowed CORS origins |

**Note:** For information on enabling/disabling audit logging in other services (Orchestration Engine, Portal Backend), refer to the deployment documentation of those respective services.

### Docker Examples

**Using SQLite (Default):**
```bash
# Build image
docker build -t audit-service .

# Run container with SQLite (default)
docker run -d \
  -p 3001:3001 \
  -e DB_PATH=/data/audit.db \
  -v audit-data:/data \
  --name audit-service \
  audit-service
```

**Using PostgreSQL:**
```bash
# Run container with PostgreSQL
docker run -d \
  -p 3001:3001 \
  -e DB_TYPE=postgres \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USERNAME=user \
  -e DB_PASSWORD=password \
  -e DB_NAME=audit_db \
  -e DB_SSLMODE=disable \
  --name audit-service \
  audit-service
```

**Note:** The audit service has its own `docker-compose.yml` for standalone deployment. The main `exchange/docker-compose.yml` does not include audit-service by default. Deploy it separately or add it manually if needed.

