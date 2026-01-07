# Audit Service

A Go microservice for managing audit logs, providing both read and write operations for generalized audit events.

## Overview

The Audit Service answers: "who made this request, what did they ask for, and did they get that data" by providing APIs for logging and querying audit events. It provides a generalized, reusable audit logging solution that can be used across different services.

**Note:** Audit service is **optional**. Services can function normally with or without audit logging enabled. See [AUDIT_SERVICE.md](../exchange/AUDIT_SERVICE.md) for configuration details.

## Features

- **Read and write operations** for audit logs
- **Data exchange event logging** - Create and query data exchange events
- **Management event logging** - Create and query management events
- **Advanced filtering** by date, consumer, provider, and status
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

### Data Exchange Events
- `GET /api/data-exchange-events` - Retrieve data exchange event logs with filtering
- `POST /api/data-exchange-events` - Create a new data exchange event (used by Orchestration Engine)

### Management Events
- `GET /api/management-events` - Retrieve management event logs with filtering
- `POST /api/management-events` - Create a new management event (used by Portal Backend)

### System Endpoints
- `GET /health` - Service health check
- `GET /version` - Service version information

### Query Parameters

**Data Exchange Events** (`GET /api/data-exchange-events`) supports:
- `status` - Filter by status (success/failure)
- `startDate` - Start date filter (YYYY-MM-DD)
- `endDate` - End date filter (YYYY-MM-DD)
- `applicationId` - Filter by application ID
- `schemaId` - Filter by schema ID
- `consumerId` - Filter by consumer ID
- `providerId` - Filter by provider ID
- `limit` - Results per page (default: 50, max: 1000)
- `offset` - Pagination offset

**Management Events** (`GET /api/management-events`) supports:
- `eventType` - Filter by event type
- `status` - Filter by status
- `actorType` - Filter by actor type
- `actorId` - Filter by actor ID
- `actorRole` - Filter by actor role
- `targetResource` - Filter by target resource type
- `targetResourceId` - Filter by target resource ID
- `startDate` - Start date filter (YYYY-MM-DD)
- `endDate` - End date filter (YYYY-MM-DD)
- `limit` - Results per page (default: 50, max: 1000)
- `offset` - Pagination offset

### Response Format

**Data Exchange Events Response:**
```json
{
  "total": 100,
  "limit": 50,
  "offset": 0,
  "events": [
    {
      "id": "uuid",
      "timestamp": "2024-01-01T12:00:00Z",
      "status": "success",
      "applicationId": "app-123",
      "schemaId": "schema-456",
      "consumerId": "consumer-123",
      "providerId": "provider-456",
      "onBehalfOfOwnerId": "owner-789",
      "requestedData": {...},
      "additionalInfo": {...},
      "createdAt": "2024-01-01T12:00:00Z"
    }
  ]
}
```

**Management Events Response:**
```json
{
  "total": 50,
  "limit": 50,
  "offset": 0,
  "events": [
    {
      "id": "uuid",
      "eventType": "CREATE",
      "status": "success",
      "timestamp": "2024-01-01T12:00:00Z",
      "actorType": "USER",
      "actorId": "member-123",
      "actorRole": "MEMBER",
      "targetResource": "SCHEMAS",
      "targetResourceId": "schema-456",
      "metadata": null,
      "createdAt": "2024-01-01T12:00:00Z"
    }
  ]
}
```

**Note**: `eventType` values are: `CREATE`, `UPDATE`, `DELETE`. `targetResource` values include: `MEMBERS`, `SCHEMAS`, `SCHEMA-SUBMISSIONS`, `APPLICATIONS`, `APPLICATION-SUBMISSIONS`, `POLICY-METADATA`.

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

## Optional Deployment

The Audit Service is **optional** and can be deployed separately from the main OpenDIF services. Services (Orchestration Engine, Portal Backend) will function normally without audit logging enabled.

### Enabling/Disabling Audit Logging in Other Services

To enable or disable audit logging in services that use the audit-service (e.g., Orchestration Engine, Portal Backend), configure the following environment variables:

#### Enable Audit Logging

Set the audit service URL and optionally enable audit:

```bash
# Required: Set the audit service URL
AUDIT_CONNECTION_SERVICEURL=http://localhost:3001

# Optional: Explicitly enable audit (defaults to true if URL is set)
ENABLE_AUDIT=true
```

**Example for Orchestration Engine:**
```bash
# In exchange/orchestration-engine/.env or environment
export AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
export ENABLE_AUDIT=true
```

**Example for Portal Backend:**
```bash
# In portal-backend/.env or environment
export AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
export ENABLE_AUDIT=true
```

#### Disable Audit Logging

You can disable audit logging in two ways:

**Option 1: Set ENABLE_AUDIT=false**
```bash
# In your service's .env or environment
ENABLE_AUDIT=false
# AUDIT_CONNECTION_SERVICEURL can be set or unset (ignored when ENABLE_AUDIT=false)
```

**Option 2: Omit or Empty AUDIT_CONNECTION_SERVICEURL**
```bash
# Leave AUDIT_CONNECTION_SERVICEURL unset or empty
# Services will automatically disable audit logging if URL is not configured
# AUDIT_CONNECTION_SERVICEURL=
```

**Note:** When audit logging is disabled, services continue to function normally. Audit operations are asynchronous (fire-and-forget) and will not block requests even if the audit service is unavailable.

### Deployment Steps

#### Step 1: Deploy Audit Service (Optional)

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

#### Step 2: Configure Services to Use Audit Service

See the [Enabling/Disabling Audit Logging](#enablingdisabling-audit-logging-in-other-services) section above for detailed instructions on configuring other services to use the audit-service.

#### Step 3: Verify Audit Logging

1. Make a request to Orchestration Engine or Portal Backend
2. Check audit service logs: `GET http://localhost:3001/api/audit-logs`
3. Verify events are being logged with trace IDs

### Docker Compose Usage

The audit service has its own `docker-compose.yml` for standalone deployment:

```bash
# Deploy audit service separately
cd audit-service
docker compose up -d

# Or include in your main docker-compose.yml (optional)
# See audit-service/docker-compose.yml for reference
```

**Note:** The main `exchange/docker-compose.yml` does not include audit-service by default. Deploy it separately or add it manually if needed.

### Environment Variables Summary

**For Audit Service (this service):**
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

**For Services Using Audit Service (Orchestration Engine, Portal Backend, etc.):**
| Variable | Default | Description |
|----------|---------|-------------|
| `AUDIT_CONNECTION_SERVICEURL` | Empty | Audit service base URL (e.g., `http://localhost:3001`) |
| `ENABLE_AUDIT` | `true` (if URL set) | Enable/disable audit logging (`true`/`false`) |

**How it works:**
- If `AUDIT_CONNECTION_SERVICEURL` is empty or unset → Audit logging is **disabled**
- If `AUDIT_CONNECTION_SERVICEURL` is set and `ENABLE_AUDIT` is not `false` → Audit logging is **enabled**
- If `ENABLE_AUDIT=false` → Audit logging is **disabled** (regardless of URL)

### Graceful Degradation

- Services continue to function normally if audit service is unavailable
- No errors are thrown when audit service URL is not configured
- Audit operations are asynchronous (fire-and-forget) to avoid blocking requests
- Services can be started before audit service is ready

## Docker

### Using SQLite (Default)

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

### Using PostgreSQL

To use PostgreSQL instead of SQLite, set `DB_TYPE=postgres` and provide PostgreSQL connection details:

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

