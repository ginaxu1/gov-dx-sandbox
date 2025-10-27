# Audit Service

## Overview

The Audit Service is a **hybrid service** that performs two separate jobs in the OpenDIF ecosystem:

1. **Consumer (Write Path)**: Asynchronously processes audit messages from Redis Streams and saves them to PostgreSQL
   - Triggered by: `streamConsumer.Start(ctx)` in main.go
   - Runs as: Background goroutine  
   - Reads from: Redis Stream `audit-events`
   - Writes to: PostgreSQL `audit_logs` table

2. **API Server (Read Path)**: Provides REST API endpoints for frontend portals to query and retrieve audit logs
   - Triggered by: `httpServer.ListenAndServe()` in main.go
   - Runs as: HTTP server on port 3001
   - Reads from: PostgreSQL `audit_logs` table
   - Serves: GET /api/logs and POST /api/logs endpoints

This dual-role architecture allows the service to reliably consume high-volume audit events while simultaneously serving read requests to administrative and entity portals.

## Architecture

```
┌─────────────────┐                    ┌─────────────────┐
│ Other Services  │                    │ Frontend        │
│ (Producers)     │                    │ Portals         │
└────────┬────────┘                    └────────┬─────────┘
         │                                     │
         │ Publish to Stream                   │ GET /api/logs
         │                                     │
         ▼                                     ▼
┌─────────────────────────────────────────────────────────┐
│                    Audit Service                       │
│                                                        │
│  ┌────────────────┐          ┌──────────────────┐    │
│  │   Consumer     │          │   API Server     │    │
│  │  (Write Path)  │          │   (Read Path)    │    │
│  │                │          │                  │    │
│  │ Reads from     │          │ Serves REST API │    │
│  │ Redis Streams  │          │ Port 3001       │    │
│  │                │          │                 │    │
│  └───────┬────────┘          └────────┬─────────┘    │
│          │                             │             │
│          │ Saves to Database           │ Reads from DB│
└───────────┼─────────────────────────────┼──────────────┘
            │                             │
            │  ┌───────────────────────┐  │
            └─▶│   PostgreSQL          │◀─┘
               │   audit_logs         │
               └───────────────────────┘
```

## The Two Jobs Explained

### 1. Consumer (Write Path)
- **Purpose**: Background process that consumes audit events from Redis Streams
- **Trigger**: Started by `streamConsumer.Start(ctx)` in main.go
- **Flow**: Reads messages from Redis Stream `audit-events` → Processes them → Saves to PostgreSQL
- **Features**: Automatic retry, dead letter queue, fault tolerance

### 2. API Server (Read Path)
- **Purpose**: Serves HTTP requests from frontend portals
- **Trigger**: Started by `httpServer.ListenAndServe()` in main.go
- **Flow**: Handles GET/POST requests from frontends → Queries PostgreSQL → Returns results
- **Features**: Filtering, pagination, CORS support

## Components

1. **Consumer Package** (`consumer/`): Handles Redis Streams consumption and message processing
2. **API Server** (`main.go`): HTTP server providing REST endpoints
3. **Simple CORS Middleware**: Adds CORS headers for cross-origin requests
4. **Handlers** (`handlers/`): Request handlers for different operations
5. **Services** (`services/`): Business logic for audit operations
6. **Database** (`database.go`): PostgreSQL connection management

## Features

- **REST API**: Simple HTTP endpoints for creating and querying audit logs
- **Filtering**: Filter by status, date range, consumer ID, provider ID
- **Pagination**: Built-in support for limit and offset
- **CORS Support**: Automatic CORS headers for cross-origin requests
- **Database Views**: Optimized queries using PostgreSQL views
- **Graceful Shutdown**: Handles shutdown signals gracefully
- **Health Checks**: Health and version endpoints for monitoring

## Quick Start

### Prerequisites

**PostgreSQL Database**:
```bash
# Set up your PostgreSQL database
export CHOREO_DB_AUDIT_HOSTNAME=localhost
export CHOREO_DB_AUDIT_PORT=5432
export CHOREO_DB_AUDIT_USERNAME=postgres
export CHOREO_DB_AUDIT_PASSWORD=password
export CHOREO_DB_AUDIT_DATABASENAME=gov_dx_sandbox
export DB_SSLMODE=disable  # For local development
```

### Running the Service

```bash
# Build the service
cd audit-service
go build -o audit-service .

# Or run directly
go run main.go

# Or use the Makefile
make build
make run
```

Expected output:
```
INFO Audit Service starting environment=development port=3001
INFO Connecting to database host=localhost port=5432 database=gov_dx_sandbox
INFO Successfully connected to PostgreSQL database
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/version` | GET | Version information |
| `/api/logs` | GET | Retrieve audit logs with filtering |
| `/api/logs` | POST | Create new audit log entry |

### Health Check

```bash
curl http://localhost:3001/health
```

Response:
```json
{
  "service": "audit-service",
  "status": "healthy"
}
```

### Retrieve Audit Logs

```bash
# Get all audit logs
curl http://localhost:3001/api/logs

# Filter by status
curl "http://localhost:3001/api/logs?status=success"

# Filter by date range
curl "http://localhost:3001/api/logs?startDate=2024-01-01&endDate=2024-01-31"

# With pagination
curl "http://localhost:3001/api/logs?limit=10&offset=0"
```

### Create Audit Log

```bash
curl -X POST http://localhost:3001/api/logs \
  -H "Content-Type: application/json" \
  -d '{
    "status": "success",
    "requestedData": "query { personInfo(nic: \"199512345678\") { fullName } }",
    "applicationId": "app-123",
    "schemaId": "schema-456"
  }'
```

## How It Works

### Creating Audit Logs

You can create audit logs via the REST API:

```bash
POST /api/logs
Content-Type: application/json

{
  "status": "success",
  "requestedData": "query { ... }",
  "applicationId": "app-123",
  "schemaId": "schema-456"
}
```

### Querying Audit Logs

The service provides flexible querying with:
- **Status filtering**: `?status=success` or `?status=failure`
- **Date range**: `?startDate=2024-01-01&endDate=2024-12-31`
- **Consumer filtering**: `?consumerId=consumer-123`
- **Provider filtering**: `?providerId=provider-456`
- **Pagination**: `?limit=50&offset=0`

### Database Schema

The service uses the `audit_logs` table with this schema:

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(10) NOT NULL CHECK (status IN ('success', 'failure')),
    requested_data TEXT NOT NULL,
    application_id VARCHAR(255) NOT NULL,
    schema_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_application_id ON audit_logs(application_id);
CREATE INDEX idx_audit_logs_schema_id ON audit_logs(schema_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
```

### Redis Streams Consumer (Automatic)

When Redis is configured via `REDIS_ADDR`, the service automatically enables the Consumer (Write Path):
- **Location**: `audit-service/consumer/`
- **Stream name**: `audit-events`
- **Consumer group**: `audit-processors`
- **Features**: 
  - Automatic retry with configurable max attempts
  - Dead letter queue (DLQ) for failed messages
  - Message claiming for stuck messages
  - Fault tolerance and graceful degradation

**Operation**: The consumer runs as a background goroutine, automatically processing messages from Redis Streams and saving them to PostgreSQL. The API server continues to operate independently.

## Testing

### Unit Tests

```bash
cd audit-service
make test
```

### Integration Tests

```bash
# From the integration-tests directory
cd integration-tests
bash test-audit-service.sh
```

### Manual Testing

1. **Start the service**:
   ```bash
   cd audit-service
   go run main.go
   ```

2. **Test health endpoint**:
   ```bash
   curl http://localhost:3001/health
   ```

3. **Create a log entry**:
   ```bash
   curl -X POST http://localhost:3001/api/logs \
     -H "Content-Type: application/json" \
     -d '{
       "status": "success",
       "requestedData": "test query",
       "applicationId": "test-app",
       "schemaId": "test-schema"
     }'
   ```

4. **Query logs**:
   ```bash
   curl http://localhost:3001/api/logs
   ```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3001` | Service port |
| `ENVIRONMENT` | `production` | Environment (development/production) |
| `CHOREO_DB_AUDIT_HOSTNAME` | `localhost` | PostgreSQL hostname |
| `CHOREO_DB_AUDIT_PORT` | `5432` | PostgreSQL port |
| `CHOREO_DB_AUDIT_USERNAME` | `postgres` | PostgreSQL username |
| `CHOREO_DB_AUDIT_PASSWORD` | `password` | PostgreSQL password |
| `CHOREO_DB_AUDIT_DATABASENAME` | `gov_dx_sandbox` | PostgreSQL database name |
| `DB_SSLMODE` | `require` | SSL mode for database connection |
| `REDIS_ADDR` | `""` | Redis server address (e.g., `localhost:6379`) - enables Consumer |
| `REDIS_PASSWORD` | `""` | Redis password (optional) |
| `AUDIT_STREAM_NAME` | `audit-events` | Redis stream name for audit events |
| `AUDIT_GROUP_NAME` | `audit-processors` | Consumer group name |
| `AUDIT_MAX_RETRY` | `5` | Maximum retry attempts for failed messages |
| `AUDIT_BLOCK_TIMEOUT` | `5s` | Timeout for blocking stream reads |
| `AUDIT_PENDING_TIMEOUT` | `1m` | Timeout before claiming pending messages |

### Optional Database Pool Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_MAX_OPEN_CONNS` | `25` | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle connections |
| `DB_CONN_MAX_LIFETIME` | `1h` | Connection max lifetime |
| `DB_CONN_MAX_IDLE_TIME` | `30m` | Connection max idle time |
| `DB_QUERY_TIMEOUT` | `30s` | Query timeout |
| `DB_CONNECT_TIMEOUT` | `10s` | Connection timeout |
| `DB_RETRY_ATTEMPTS` | `10` | Connection retry attempts |
| `DB_RETRY_DELAY` | `2s` | Delay between retries |

## Development

### Building

```bash
go build .
```

### Running Tests

```bash
go test ./tests/... -v
```

### Docker

```bash
# Build image
docker build -t audit-service .

# Run container
docker run -p 8081:8081 \
  -e REDIS_ADDR=redis:6379 \
  -e CHOREO_DB_AUDIT_HOSTNAME=postgres \
  audit-service
```

## Monitoring

### Service Monitoring

```bash
# Check service health
curl http://localhost:3001/health

# Get version information
curl http://localhost:3001/version
```

### Database Monitoring

```sql
-- Check recent audit logs
SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;

-- Check logs by consumer
SELECT consumer_id, COUNT(*) FROM audit_logs GROUP BY consumer_id;

-- Check success/failure rates
SELECT transaction_status, COUNT(*) FROM audit_logs GROUP BY transaction_status;
```

## Troubleshooting

### Common Issues

1. **Redis Connection Failed**
   - Check Redis is running: `redis-cli ping`
   - Verify `REDIS_ADDR` environment variable
   - Check Redis authentication

2. **Database Connection Failed**
   - Verify PostgreSQL is running
   - Check database credentials
   - Ensure database exists

3. **No Audit Logs Appearing**
   - Check Redis Streams for messages: `redis-cli xlen audit-events`
   - Verify consumer group exists: `redis-cli xinfo groups audit-events`
   - Check service logs for errors

### Logs

The service logs important events:
- Redis connection status
- Message processing success/failure
- Database operations
- Error conditions

## License

This project is part of the OpenDIF ecosystem.
