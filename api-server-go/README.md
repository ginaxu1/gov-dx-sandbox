# API Server Go

API Server for government data exchange portal management with PostgreSQL database integration.


## Quick Start

1. **Start PostgreSQL**:
```bash
make setup-test-db
```

2. **Run the application**:
```bash
make run
```

3. **Test the API**:
```bash
curl http://localhost:3000/health
```

## Docker Deployment

```bash
cd ../exchange
docker-compose up postgres api-server-go
```

## API Documentation

OpenAPI spec available at `/openapi.yaml` when running.

## Testing

```bash
# Run all tests (including PostgreSQL integration tests)
make test-all

# Run tests with local PostgreSQL
make test-local

# Run tests with Docker
make test-docker
```

## Database

**PostgreSQL Database** with automatic schema initialization and connection pooling.

### Database Tables
- `consumers` - Consumer organization information
- `consumer_apps` - Consumer application submissions
- `provider_submissions` - Provider registration submissions
- `provider_profiles` - Approved provider profiles
- `provider_schemas` - Provider data schemas and SDL definitions
- `consumer_grants` - Consumer access grants and permissions
- `provider_metadata` - Field-level metadata and access controls

### Database Configuration

The application uses environment variables for database configuration:

```bash
# Required database connection settings
CHOREO_OPENDIF_DB_HOSTNAME=localhost
CHOREO_OPENDIF_DB_PORT=5432
CHOREO_OPENDIF_DB_USERNAME=postgres
CHOREO_OPENDIF_DB_PASSWORD=password
CHOREO_OPENDIF_DB_DATABASENAME=api_server

# Optional database optimization settings
DB_SSLMODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m
DB_QUERY_TIMEOUT=30s
DB_CONNECT_TIMEOUT=10s
DB_RETRY_ATTEMPTS=3
DB_RETRY_DELAY=1s
DB_TRANSACTION_TIMEOUT=60s
DB_ENABLE_MONITORING=true
DB_HEALTH_CHECK_INTERVAL=30s
```

### Health Monitoring

The application provides a comprehensive health monitoring endpoint:

- `/health` - Complete health check with database status, connection pool metrics, and utilization monitoring
