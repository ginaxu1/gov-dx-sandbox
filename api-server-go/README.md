# API Server Go

API Server for government data exchange portal management with PostgreSQL integration.

## Features

- Consumer and Provider Management
- Schema and Grants Management  
- Admin Functions and Metrics
- PostgreSQL with connection pooling

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
make test-all
```

## Database

PostgreSQL with automatic schema initialization. Tables:
- consumers, consumer_apps
- provider_submissions, provider_profiles, provider_schemas  
- consumer_grants, provider_metadata