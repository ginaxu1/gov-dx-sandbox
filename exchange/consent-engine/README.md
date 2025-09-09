# Consent Engine

Consent management service for data access authorization.

## Running

```bash
# Local development
go run main.go

# Docker
docker build -t ce .
docker run -p 8081:8081 ce

# With Docker Compose
docker compose --env-file .env.local up --build
```

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/consent` | POST | Create consent record |
| `/consent/{id}` | GET, PUT, DELETE | Get, update, or revoke consent |
| `/consent/portal` | GET, POST | Portal info and processing |
| `/data-owner/{owner}` | GET | Get consents by data owner |
| `/consumer/{consumer}` | GET | Get consents by consumer |
| `/admin/expiry-check` | POST | Check consent expiry |
| `/health` | GET | Health check |

## Example Usage

```bash
# Create Consent
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "test-app",
    "data_owner": "test-owner", 
    "data_fields": ["person.fullName"],
    "purpose": "testing",
    "expiry_days": 30
  }'

# Get Consent by ID
curl -X GET http://localhost:8081/consent/{consent-id}

# Get Consents by Data Owner
curl -X GET http://localhost:8081/data-owner/{owner-id}
```

## Configuration

Environment variables:
- `PORT` - Service port (default: 8081)
- `ENVIRONMENT` - Environment (local/production)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)
