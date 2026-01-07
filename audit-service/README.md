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
- PostgreSQL 13+

### Run the Service

```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go

# Or build and run
go build -o audit-service
./audit-service
```

The service runs on port 3001 by default.

## Configuration

### Environment Variables

**Choreo Environment Variables (Primary):**
- `CHOREO_OPENDIF_DATABASE_HOSTNAME` or `CHOREO_OPENDIF_DB_HOSTNAME` - Database hostname
- `CHOREO_OPENDIF_DATABASE_PORT` or `CHOREO_OPENDIF_DB_PORT` - Database port
- `CHOREO_OPENDIF_DATABASE_USERNAME` or `CHOREO_OPENDIF_DB_USERNAME` - Database username
- `CHOREO_OPENDIF_DATABASE_PASSWORD` or `CHOREO_OPENDIF_DB_PASSWORD` - Database password
- `CHOREO_OPENDIF_DATABASE_DATABASENAME` or `CHOREO_OPENDIF_DB_DATABASENAME` - Database name

**Fallback Environment Variables (Local Development):**
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USERNAME` - Database username (default: user)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: gov_dx_sandbox)
- `DB_SSLMODE` - SSL mode (default: require for Choreo, disable for local)
- `PORT` - Service port (default: 3001)
- `ENVIRONMENT` - Environment (development/production, default: production)

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

Create `.env.local` with PostgreSQL credentials:

```bash
TEST_DB_USERNAME=postgres
TEST_DB_PASSWORD=your_password
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_DATABASE=audit_service_test
TEST_DB_SSLMODE=disable
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover
```

**Note**: Tests automatically load credentials from `.env.local` and skip gracefully if PostgreSQL is unavailable.

## Health Check

- `GET /health` - Returns service health status

## Security

- **CORS enabled**: Cross-origin requests are supported via CORS middleware
- **Internal service**: Service is intended for internal use within the OpenDIF ecosystem
- **No authentication required**: Authentication is handled by upstream services (Orchestration Engine, Portal Backend)
- **Database security**: Uses SSL connections in production (Choreo) with configurable SSL modes

## Optional Deployment

The Audit Service is **optional** and can be deployed separately from the main OpenDIF services. Services (Orchestration Engine, Portal Backend) will function normally without audit logging enabled.

### Enabling/Disabling Audit Logging

#### Option 1: Disable via Environment Variable
Set `ENABLE_AUDIT=false` in your service configuration:
```bash
# In orchestration-engine/.env or portal-backend/.env
ENABLE_AUDIT=false
```

#### Option 2: Omit Audit Service URL
Leave `CHOREO_AUDIT_CONNECTION_SERVICEURL` unset or empty:
```bash
# Services will automatically disable audit logging if URL is not configured
# CHOREO_AUDIT_CONNECTION_SERVICEURL=
```

#### Option 3: Enable Audit Logging
Set the audit service URL:
```bash
# In orchestration-engine/.env or portal-backend/.env
CHOREO_AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
ENABLE_AUDIT=true  # Optional, defaults to true if URL is provided
```

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

**For Orchestration Engine:**
```bash
# In exchange/orchestration-engine/.env
CHOREO_AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
ENABLE_AUDIT=true
```

**For Portal Backend:**
```bash
# In portal-backend/.env
CHOREO_AUDIT_CONNECTION_SERVICEURL=http://localhost:3001
ENABLE_AUDIT=true
```

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

### Environment Variables for Enabling/Disabling

| Service | Environment Variable | Default | Description |
|---------|---------------------|---------|------------|
| Orchestration Engine | `CHOREO_AUDIT_CONNECTION_SERVICEURL` | Empty | Audit service base URL |
| Orchestration Engine | `ENABLE_AUDIT` | `true` (if URL set) | Enable/disable audit logging |
| Portal Backend | `CHOREO_AUDIT_CONNECTION_SERVICEURL` | Empty | Audit service base URL |
| Portal Backend | `ENABLE_AUDIT` | `true` (if URL set) | Enable/disable audit logging |

### Graceful Degradation

- Services continue to function normally if audit service is unavailable
- No errors are thrown when audit service URL is not configured
- Audit operations are asynchronous (fire-and-forget) to avoid blocking requests
- Services can be started before audit service is ready

## Docker

```bash
# Build image
docker build -t audit-service .

# Run container
docker run -p 3001:3001 \
  -e DB_HOST=host.docker.internal \
  --env-file .env \
  audit-service
```

