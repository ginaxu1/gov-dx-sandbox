# Policy Decision Point

Policy enforcement service for data access authorization decisions.

## Running

```bash
# Local development
go run main.go

# Docker
docker build -t pdp .
docker run -p 8082:8082 pdp

# With Docker Compose
docker compose --env-file .env.local up --build
```

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/decide` | POST | Authorization decision endpoint |
| `/health` | GET | Health check |
| `/debug` | GET | Debug information |

## Example Usage

```bash
# Policy Decision
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {"id": "test-app", "name": "Test App", "type": "mobile_app"},
    "request": {"resource": "person_data", "action": "read", "data_fields": ["person.fullName"]},
    "timestamp": "2025-09-09T16:30:00Z"
  }'
```

## Configuration

Environment variables:
- `PORT` - Service port (default: 8082)
- `ENVIRONMENT` - Environment (local/production)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)
