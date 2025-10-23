# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Overview

- **Technology**: Go + Open Policy Agent (OPA) + Rego policies
- **Port**: 8082
- **Purpose**: Attribute-based access control (ABAC) with field-level permissions

## Quick Start

```bash
# Run locally
cd policy-decision-point && go run main.go

# Run tests
go test -v

# Docker
docker build -t pdp . && docker run -p 8082:8082 pdp
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/decide` | POST | Authorization decision |
| `/health` | GET | Health check |

## Authorization Request

**Endpoint:** `POST /decide`

**Request:**
```json
{
  "application_id": "passport-app",
  "app_id": "passport-app", 
  "request_id": "req_123",
  "required_fields": ["person.fullName", "person.photo"]
}
```

**Response:**
```json
{
  "allow": true,
  "consent_required": true,
  "consent_required_fields": ["person.photo"]
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_123",
    "required_fields": ["person.fullName", "person.photo"]
  }'
```

## Access Control Logic

### Field Types

1. **Public Fields** (`access_control_type: "public"`)
   - Any app can access
   - Consent only required if `consent_required: true`

2. **Restricted Fields** (`access_control_type: "restricted"`)
   - Only apps in `allow_list` can access
   - Consent required if `consent_required: true`

### Decision Logic

- **Allow**: All requested fields are authorized for the app
- **Deny**: Any requested field is not authorized for the app
- **Consent Required**: Any requested field has `consent_required: true`

## Testing

### Test Different Scenarios

```bash
# Public field (no consent required)
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_001",
    "required_fields": ["person.fullName"]
  }'

# Consent required field
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_002",
    "required_fields": ["person.nic", "person.photo"]
  }'

# Restricted field (unauthorized app)
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "unknown-app",
    "app_id": "unknown-app",
    "request_id": "req_003",
    "required_fields": ["person.birthDate"]
  }'

# Authorized restricted field
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "driver-app",
    "app_id": "driver-app",
    "request_id": "req_004",
    "required_fields": ["person.birthDate"]
  }'
```

## Data Files

- `policies/main.rego` - OPA authorization policies

## Field Configuration

Fields are configured in the database and can be updated via the `/metadata/update` endpoint:

```json
{
  "fields": {
    "person.fullName": {
      "owner": "citizen",
      "provider": "drp",
      "consent_required": false,
      "access_control_type": "public",
      "allow_list": []
    },
    "person.birthDate": {
      "owner": "rgd",
      "provider": "drp", 
      "consent_required": false,
      "access_control_type": "restricted",
      "allow_list": [
        {
          "application_id": "driver-app",
          "expires_at": 1757560679,
          "grant_duration": "30d"
        }
      ]
    }
  }
}
```

## Health Check

```bash
curl http://localhost:8082/health
```

**Response:**
```json
{
  "service": "policy-decision-point",
  "status": "healthy"
}
```