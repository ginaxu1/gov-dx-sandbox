# Consent Engine (CE)

Service that manages data owner consent workflows for data access requests with user JWT authentication support for public endpoints and internal access for service-to-service communication.

## Overview

The Consent Engine (CE) manages the lifecycle of data access consents. It coordinates workflows between data consumers, data owners (citizens), and the Orchestration Engine. It supports JWT authentication for user-facing interactions and provides internal APIs for service-to-service communication.

**Technology**: Go + In-memory storage (Test) / PostgreSQL (Prod)
**Port**: 8081

## Features

- **Consent Workflow Management** - Create, approve, reject, revoke, and expire consents
- **Dual Authentication Model** - JWT for users, internal access for services
- **Data Owner Verification** - Ensures only the data owner can manage their consents
- **Integration Ready** - Works with Orchestration Engine and Policy Decision Point
- **Audit Ready** - Tracks consent status changes

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+ (for persistence)

### Run the Service

```bash
# Install dependencies
go mod download

# Run locally (uses defaults or .env.local)
go run main.go

# Or build and run
go build -o consent-engine
./consent-engine
```

The service runs on port 8081 by default.

## Configuration

### Environment Variables

**Required (User JWT - Asgardeo):**
- `ASGARDEO_JWKS_URL` - JWKS endpoint URL
- `ASGARDEO_ISSUER` - JWT issuer URL
- `ASGARDEO_AUDIENCE` - JWT audience
- `ASGARDEO_ORG_NAME` - Organization name

**Database (Choreo/Local):**
- `CHOREO_OPENDIF_DB_HOSTNAME` - Database host
- `CHOREO_OPENDIF_DB_PORT` - Database port
- `CHOREO_OPENDIF_DB_USERNAME` - Database username
- `CHOREO_OPENDIF_DB_PASSWORD` - Database password
- `CHOREO_OPENDIF_DB_DATABASENAME` - Database name

**Service Configuration:**
- `PORT` - Service port (default: 8081)
- `CONSENT_PORTAL_URL` - URL for the Consent Portal
- `ORCHESTRATION_ENGINE_URL` - URL for the Orchestration Engine
- `ENVIRONMENT` - `production` or `development`

## API Endpoints

| Endpoint | Method | Description | Auth |
|----------|--------|-------------|------|
| `/consents` | POST | Create new consent | Internal |
| `/consents/{id}` | GET | Get consent information | User JWT |
| `/consents/{id}` | PUT | Update consent status | User JWT |
| `/consents/{id}` | PATCH | Partially update consent | Internal |
| `/consents/{id}` | DELETE | Revoke consent | User JWT |
| `/data-info/{id}` | GET | Get data owner information | None |
| `/health` | GET | Health check | None |

### Create Consent (Internal)

**Endpoint:** `POST /consents`

```bash
curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "passport-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "user@example.com",
        "fields": ["person.photo"]
      }
    ],
    "purpose": "passport_application",
    "session_id": "session_123"
  }'
```

### Get Consent (User)

**Endpoint:** `GET /consents/{id}`

```bash
curl -X GET http://localhost:8081/consents/consent_123 \
  -H "Authorization: Bearer <jwt_token>"
```

## Authentication

- **User-facing Endpoints** (`GET/PUT/DELETE /consents/{id}`): Require a valid JWT token where the `email` claim matches the consent owner.
- **Internal Endpoints** (`POST /consents`): Require network-level access (no auth header).

## Testing

```bash
# Run unit tests
go test ./...

# Run integration tests (requires DB)
make test-local
```

## Docker

```bash
# Build image
docker build -t consent-engine .

# Run container
docker run -p 8081:8081 \
  -e CHOREO_OPENDIF_DB_HOSTNAME=host.docker.internal \
  --env-file .env.local \
  consent-engine
```