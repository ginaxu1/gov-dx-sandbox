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

### Admin Portal
- `GET /audit/events` - Returns all logs with filtering (no authentication required)

### Provider Portal
- `GET /audit/provider/events` - Returns logs where provider ID matches authenticated JWT

### Consumer Portal
- `GET /audit/consumer/events` - Returns logs where consumer ID matches authenticated JWT

### Query Parameters

All endpoints support:
- `consumer_id` - Filter by consumer ID
- `provider_id` - Filter by provider ID
- `transaction_status` - Filter by status (SUCCESS/FAILURE)
- `start_date` - Start date filter (YYYY-MM-DD)
- `end_date` - End date filter (YYYY-MM-DD)
- `limit` - Results per page (default: 50, max: 1000)
- `offset` - Pagination offset

### Response Format

```json
{
  "events": [
    {
      "event_id": "uuid",
      "timestamp": "2024-01-01T12:00:00Z",
      "consumer_id": "consumer-123",
      "provider_id": "provider-456",
      "transaction_status": "SUCCESS",
      "citizen_hash": "hashed-citizen-id"
    }
  ],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```

### Authentication

- **Admin Portal**: No authentication required (internal use)
- **Provider Portal**: JWT token with `provider_id` claim
- **Consumer Portal**: JWT token with `consumer_id` claim

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

- **Sensitive data excluded**: `requested_data` and `response_data` are intentionally omitted
- **Role-based access**: Each endpoint returns only data relevant to the authenticated user's role
- **JWT validation**: Provider and consumer endpoints validate JWT tokens
