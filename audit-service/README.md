# Audit Service

## Overview

The Audit Service provides reliable, asynchronous audit logging for the OpenDIF ecosystem. It captures audit events from various services and stores them persistently in PostgreSQL. The system uses Redis Streams as an asynchronous buffer to ensure no audit logs are lost, even if the audit service is temporarily unavailable.

## Architecture

```
┌─────────────────┐       ┌───────────────────┐
│ Frontend UI/App │──────▶│  Audit Service    │
└─────────────────┘       │  (API Server)     │
                          │  (GET /api/logs)  │
                          └─────────▲─────────┘
                                    │
                                    │ Reads from
                                    │
┌─────────────────┐       ┌─────────▼─────────┐
│ api-server-go   │───┐   │   PostgreSQL      │
└─────────────────┘   │   │   Database        │
                      │   └─────────▲─────────┘
                      │             │
┌─────────────────┐   │ (XADD)      │ Writes to
│ orchestration-  │───┼──▶┌─────────┴─────────┐
│ engine          │   │   │  Redis Streams    │
└─────────────────┘   │   │  (audit-events)   │
                      │   └─────────┬─────────┘
┌─────────────────┐   │             │
│ ...any other    │───┘             │ (XREADGROUP)
│ producer...     │                 │
└─────────────────┘       ┌─────────▼─────────┐
                          │  Audit Service    │
                          │  (Consumer)       │
                          └───────────────────┘
```

## Features

- **Reliable Message Delivery**: Uses Redis Streams with consumer groups for guaranteed message processing
- **Fault Tolerance**: Messages are persisted in Redis and can be reprocessed after service restarts
- **Retry Logic**: Built-in retry mechanism with configurable max retries and Dead Letter Queue (DLQ)
- **Graceful Degradation**: Continues to serve API requests even when Redis is unavailable
- **Environment Flexibility**: Configurable Redis connection via environment variables

## Quick Start

### Prerequisites

1. **Redis Server**:
   ```bash
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

2. **PostgreSQL Database**:
   ```bash
   # Set up your PostgreSQL database
   export CHOREO_DB_AUDIT_HOSTNAME=localhost
   export CHOREO_DB_AUDIT_PORT=5432
   export CHOREO_DB_AUDIT_USERNAME=postgres
   export CHOREO_DB_AUDIT_PASSWORD=password
   export CHOREO_DB_AUDIT_DATABASENAME=gov_dx_sandbox
   ```

### Running the Service

```bash
# Set environment variables
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=""  # Optional

# Start the audit service
cd audit-service
go run main.go
```

Expected output:
```
Audit service started. Waiting for events...
Consumer group audit-processors ensured for stream audit-events
Redis Stream consumer started.
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/audit-logs` | GET | Retrieve audit logs with filtering |

### Health Check

```bash
curl http://localhost:8081/health
```

### Retrieve Audit Logs

```bash
# Get all audit logs
curl http://localhost:8081/audit-logs

# Filter by consumer ID
curl "http://localhost:8081/audit-logs?consumer_id=test-app"

# Filter by date range
curl "http://localhost:8081/audit-logs?start_date=2024-01-01&end_date=2024-01-31"
```

## How Audit Logging Works

### 1. Request Flow

When a request is made to services with audit middleware:

1. **Request Reception**: Service receives the request
2. **Authentication**: JWT token is validated to extract consumer information
3. **Processing**: Request is processed normally
4. **Response Generation**: Response is generated
5. **Audit Logging**: Audit info is captured and sent to Redis Streams

### 2. Audit Data Captured

For each request, the following audit information is captured:

```json
{
  "event_id": "uuid-generated-event-id",
  "consumer_id": "application-id-from-jwt",
  "provider_id": "schema-id-from-active-schema",
  "requested_data": "{\"query\": \"...\", \"variables\": {...}}",
  "response_data": "{\"data\": {...}, \"errors\": [...]}",
  "transaction_status": "success|failure",
  "user_agent": "client-user-agent",
  "ip_address": "client-ip-address",
  "timestamp": 1698000000
}
```

### 3. Redis Streams Flow

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Server    │    │   Redis         │    │   Audit Service │
│                 │    │   Streams       │    │                 │
│ 1. Publish      │───▶│ 2. Store in     │───▶│ 3. Consume      │
│    Audit Event  │    │    Stream       │    │    Messages     │
│                 │    │                 │    │                 │
│                 │    │ 4. Acknowledge  │◀───│ 5. Process &    │
│                 │    │    (XACK)       │    │    Save to DB   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Testing the Complete Flow

### Step 1: Start the Audit Service

```bash
cd audit-service
export REDIS_ADDR=localhost:6379
go run main.go
```

### Step 2: Start the Orchestration Engine (with audit middleware)

```bash
cd exchange/orchestration-engine-go
export REDIS_ADDR=localhost:6379
go run main.go
```

### Step 3: Make a Test Request

```bash
curl -X POST http://localhost:4000/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "query": "query { person(nic: \"123456789V\") { fullName address } }"
  }'
```

### Step 4: Verify Audit Logs

1. **Check Redis Stream**:
   ```bash
   redis-cli xlen audit-events
   redis-cli xrange audit-events - +
   ```

2. **Check Database**:
   ```sql
   SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT 10;
   ```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_ADDR` | `localhost:6379` | Redis server address |
| `REDIS_PASSWORD` | `""` | Redis password (optional) |
| `CHOREO_DB_AUDIT_HOSTNAME` | `localhost` | PostgreSQL hostname |
| `CHOREO_DB_AUDIT_PORT` | `5432` | PostgreSQL port |
| `CHOREO_DB_AUDIT_USERNAME` | `postgres` | PostgreSQL username |
| `CHOREO_DB_AUDIT_PASSWORD` | `password` | PostgreSQL password |
| `CHOREO_DB_AUDIT_DATABASENAME` | `gov_dx_sandbox` | PostgreSQL database name |

### Redis Streams Configuration

- **Stream Name**: `audit-events`
- **Consumer Group**: `audit-processors`
- **Dead Letter Queue**: `audit-events_dlq`
- **Max Retries**: 5
- **Block Timeout**: 5 seconds
- **Pending Timeout**: 1 minute

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

### Redis Streams Monitoring

```bash
# Check stream length
redis-cli xlen audit-events

# Check consumer groups
redis-cli xinfo groups audit-events

# Check pending messages
redis-cli xpending audit-events audit-processors
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
