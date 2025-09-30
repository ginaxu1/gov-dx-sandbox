# Audit Service

A simple Go microservice that provides read-only access to audit logs for different portals.

## Overview

The Audit Service implements the goal: "who made this request, what did they ask for, and did they get that data" by exposing audit data through secure API endpoints tailored to each role's access level.

## Features

- **Read-only access** to the `audit_logs` table
- **Role-based endpoints** for different portals
- **Filtering capabilities** by date, consumer, provider, status
- **Pagination support** for large datasets
- **JWT authentication** for provider and consumer endpoints

## API Endpoints

### For NDX Admin Portal
- `GET /audit/events` - Returns all logs with filtering capabilities

### For Data Provider Portal  
- `GET /audit/provider/events` - Returns logs only where the provider's ID matches the authenticated JWT

### For Data Consumer Portal
- `GET /audit/consumer/events` - Returns logs only where the consumer ID matches the authenticated JWT

## Query Parameters

All endpoints support the following query parameters:

- `consumer_id` - Filter by consumer ID
- `provider_id` - Filter by provider ID  
- `transaction_status` - Filter by status (SUCCESS/FAILURE)
- `start_date` - Filter by start date (YYYY-MM-DD format)
- `end_date` - Filter by end date (YYYY-MM-DD format)
- `limit` - Number of results per page (default: 50, max: 1000)
- `offset` - Number of results to skip for pagination

## Authentication

- **Admin Portal**: No authentication required (internal use)
- **Provider Portal**: Requires JWT token with `provider_id` claim
- **Consumer Portal**: Requires JWT token with `consumer_id` claim

## Response Format

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

## Security Notes

- **Sensitive data is excluded**: `requested_data` and `response_data` are intentionally omitted from API responses for security
- **Role-based access**: Each endpoint only returns data relevant to the authenticated user's role
- **JWT validation**: Provider and consumer endpoints validate JWT tokens to ensure proper authorization

## Environment Variables

### Choreo Environment Variables (Primary)
- `CHOREO_DB_AUDIT_HOSTNAME` - Database hostname
- `CHOREO_DB_AUDIT_PORT` - Database port
- `CHOREO_DB_AUDIT_USERNAME` - Database username
- `CHOREO_DB_AUDIT_PASSWORD` - Database password
- `CHOREO_DB_AUDIT_DATABASENAME` - Database name

### Fallback Environment Variables (Local Development)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database username (default: user)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: gov_dx_sandbox)
- `DB_SSLMODE` - SSL mode (default: disable)
- `PORT` - Service port (default: 3001)

## Running the Service

```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go

# Or build and run
go build -o audit-service
./audit-service
```

## Health Check

- `GET /health` - Returns service health status
