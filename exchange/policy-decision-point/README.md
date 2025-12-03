# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Overview

The PDP provides attribute-based access control (ABAC) with field-level permissions. It uses Open Policy Agent (OPA) v1 with Rego v1 policies and stores policy metadata in PostgreSQL for real-time evaluation.

**Technology**: Go + Open Policy Agent (OPA) v1 + Rego v1 + PostgreSQL  
**Port**: 8082

## Features

- **Real-time Policy Evaluation** - Policies loaded from database on startup
- **Field-level Access Control** - Granular permissions for individual data fields
- **Consent Management** - Automatic consent requirement calculation
- **Allow List Management** - Dynamic application authorization for restricted fields
- **OPA v1 Integration** - Modern Open Policy Agent with Rego v1 syntax
- **Database-driven** - Policy metadata stored in PostgreSQL

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+

### Run the Service

```bash
# Install dependencies
go mod download

# Run locally
go run main.go

# Or build and run
go build -o policy-decision-point
./policy-decision-point
```

The service runs on port 8082 by default.

## Configuration

### Environment Variables

```bash
# Database Configuration (Choreo)
CHOREO_DB_PDP_HOSTNAME=your-db-host
CHOREO_DB_PDP_PORT=your-db-port
CHOREO_DB_PDP_USERNAME=your-db-username
CHOREO_DB_PDP_PASSWORD=your-db-password
CHOREO_DB_PDP_DATABASENAME=your-db-name

# Or use standard DB variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=pdp
DB_SSLMODE=disable
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/policy/decide` | POST | Authorization decision |
| `/api/v1/policy/metadata` | POST | Create policy metadata for fields |
| `/api/v1/policy/update-allowlist` | POST | Update allow list for applications |
| `/health` | GET | Health check |

### Authorization Request

**Endpoint:** `POST /api/v1/policy/decide`

**Request:**
```json
{
  "consumer_id": "passport-app",
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

### Policy Metadata Management

**Create Policy Metadata:** `POST /api/v1/policy/metadata`

```json
{
  "schema_id": "schema-123",
  "sdl": "type Person { fullName: String }"
}
```

**Update Allow List:** `POST /api/v1/policy/update-allowlist`

```json
{
  "application_id": "passport-app",
  "records": [
    {
      "field_name": "person.fullName",
      "schema_id": "schema-123"
    }
  ],
  "grant_duration": "ONE_MONTH"
}
```

## Access Control Logic

### Field Types

1. **Public Fields** (`access_control_type: "public"`)
   - Any app can access
   - Consent required only if `consent_required: true`

2. **Restricted Fields** (`access_control_type: "restricted"`)
   - Only apps in `allow_list` can access
   - Consent required if `consent_required: true`

### Decision Logic

- **Allow**: All requested fields are authorized for the app
- **Deny**: Any requested field is not authorized for the app
- **Consent Required**: Any requested field has `consent_required: true`

### Consent Logic

Consent requirement is calculated as: `!is_owner && access_control_type != "public"`

- **Owner Fields** (`is_owner: true`): No consent required
- **Public Fields**: No consent required for non-owners
- **Restricted Fields**: Consent required for non-owners

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Test authorization scenarios
curl -X POST http://localhost:8082/api/v1/policy/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "required_fields": ["person.fullName"]
  }'
```

## Architecture

### Database Schema

**`policy_metadata` Table:**
- `id` (UUID) - Primary key
- `field_name` (TEXT) - Data field name
- `display_name` (TEXT) - Human-readable name
- `access_control_type` (ENUM) - public/restricted
- `is_owner` (BOOLEAN) - Field ownership flag
- `allow_list` (JSONB) - Authorized applications with expiration
- `created_at`, `updated_at` (TIMESTAMP)

### Policy Evaluation Flow

```
Request → Load Policy Metadata → OPA Evaluation → Consent Check → Decision
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
