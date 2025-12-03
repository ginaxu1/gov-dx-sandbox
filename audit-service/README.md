# Audit Service

A Go microservice providing read-only access to audit logs with role-based endpoints for different portals.

## Overview

The Audit Service answers: "who made this request, what did they ask for, and did they get that data" by exposing audit data through secure, role-based API endpoints.

## Features

- **Read-only access** to audit logs
- **Role-based endpoints** for Admin, Provider, and Consumer portals
- **Advanced filtering** by date, consumer, provider, and status
- **Pagination support** for large datasets
- **JWT authentication** for provider and consumer endpoints

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
- `CHOREO_DB_AUDIT_HOSTNAME` - Database hostname
- `CHOREO_DB_AUDIT_PORT` - Database port
- `CHOREO_DB_AUDIT_USERNAME` - Database username
- `CHOREO_DB_AUDIT_PASSWORD` - Database password
- `CHOREO_DB_AUDIT_DATABASENAME` - Database name

**Fallback Environment Variables (Local Development):**
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USERNAME` - Database username (default: user)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: gov_dx_sandbox)
- `DB_SSLMODE` - SSL mode (default: disable)
- `PORT` - Service port (default: 3001)

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
      "requestedData": {...},
      "additionalInfo": {...}
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
      "eventType": "SCHEMA_CREATED",
      "status": "success",
      "actorType": "MEMBER",
      "actorId": "member-123",
      "targetResource": "SCHEMA",
      "targetResourceId": "schema-456",
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ]
}
```

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
- **Read-only by design**: Endpoints are designed for audit log retrieval and creation
- **No authentication required**: Service is intended for internal use within the OpenDIF ecosystem
