# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Overview

- **Technology**: Go + Open Policy Agent (OPA) v1 + Rego v1 policies + PostgreSQL
- **Port**: 8082
- **Purpose**: Attribute-based access control (ABAC) with field-level permissions
- **Database**: Policy metadata stored in PostgreSQL with real-time policy evaluation

## Quick Start

```bash
# Run locally
cd policy-decision-point && go run main.go

# Run tests
go test -v

# Docker
docker build -t pdp . && docker run -p 8082:8082 pdp
```

## Environment Variables

The service requires the following environment variables for database connection:

```bash
export CHOREO_DB_PDP_HOSTNAME=your-db-host
export CHOREO_DB_PDP_PORT=your-db-port
export CHOREO_DB_PDP_USERNAME=your-db-username
export CHOREO_DB_PDP_PASSWORD=your-db-password
export CHOREO_DB_PDP_DATABASENAME=your-db-name
```

## Database Schema

The service uses a PostgreSQL database with the following key table:

### `policy_metadata` Table
- `id` (UUID) - Primary key
- `field_name` (TEXT) - Name of the data field
- `display_name` (TEXT) - Human-readable name
- `description` (TEXT) - Field description
- `source` (ENUM) - Source system (primary/fallback)
- `is_owner` (BOOLEAN) - Whether the field owner is the data owner
- `owner` (TEXT) - Owner identifier (default: "CITIZEN")
- `access_control_type` (ENUM) - Access control type (public/restricted)
- `allow_list` (JSONB) - List of authorized applications
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/decide` | POST | Authorization decision |
| `/policy-metadata` | POST | Create policy metadata for fields |
| `/allow-list` | POST | Update allow list for applications |
| `/health` | GET | Health check |

## Authorization Request

**Endpoint:** `POST /decide`

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

**cURL Example:**
```bash
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "passport-app",
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
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_001",
    "required_fields": ["person.fullName"]
  }'

# Consent required field
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "passport-app",
    "app_id": "passport-app",
    "request_id": "req_002",
    "required_fields": ["person.nic", "person.photo"]
  }'

# Restricted field (unauthorized app)
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "unknown-app",
    "app_id": "unknown-app",
    "request_id": "req_003",
    "required_fields": ["person.birthDate"]
  }'

# Authorized restricted field
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "driver-app",
    "app_id": "driver-app",
    "request_id": "req_004",
    "required_fields": ["person.birthDate"]
  }'
```

## Data Files

- `policies/main.rego` - OPA v1 authorization policies using Rego v1 syntax
- `database.go` - Database service for policy metadata management
- `policy-evaluator.go` - Core policy evaluation logic using OPA v1

## Features

- **Real-time Policy Evaluation**: Policies are loaded from database on startup
- **Field-level Access Control**: Granular permissions for individual data fields
- **Consent Management**: Automatic consent requirement calculation based on field ownership
- **Allow List Management**: Dynamic application authorization for restricted fields
- **OPA v1 Integration**: Modern Open Policy Agent with Rego v1 syntax
- **Database-driven**: Policy metadata stored in PostgreSQL for real-time updates

## Policy Metadata Management

### Create Policy Metadata

**Endpoint:** `POST /policy-metadata`

**Request:**
```json
{
  "field_name": "person.fullName",
  "display_name": "Full Name",
  "description": "Complete name of the person",
  "source": "primary",
  "is_owner": true,
  "access_control_type": "public",
  "allow_list": []
}
```

**Response:**
```json
{
  "success": true,
  "message": "Created policy metadata for field person.fullName",
  "id": "82907bce-38c4-44b5-9392-7f7fd70c8c67"
}
```

### Update Allow List

**Endpoint:** `POST /allow-list`

**Request:**
```json
{
  "field_name": "person.fullName",
  "application_id": "passport-app",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Updated allow list for field person.fullName with application passport-app"
}
```

## Field Configuration

Fields are configured in the database via the policy metadata endpoints. The consent logic is:

- **Consent Required**: `!is_owner && access_control_type != "public"`
- **Owner Fields**: No consent required regardless of access control type
- **Public Fields**: No consent required for non-owners
- **Restricted Fields**: Consent required for non-owners

### Consent Logic Implementation

The consent requirement is automatically calculated by the policy engine based on:

1. **Field Ownership** (`is_owner` field):
   - `true`: Field owner is the data owner → No consent required
   - `false`: Field owner is not the data owner → Consent may be required

2. **Access Control Type** (`access_control_type` field):
   - `"public"`: Public access → No consent required for non-owners
   - `"restricted"`: Restricted access → Consent required for non-owners

3. **Application Authorization** (`allow_list` field):
   - Applications must be in the allow list to access restricted fields
   - Allow list entries include expiration timestamps

### Example Field Configurations

```json
{
  "fields": {
    "person.fullName": {
      "owner": "CITIZEN",
      "provider": "primary",
      "is_owner": true,
      "access_control_type": "public",
      "allow_list": [
        {
          "application_id": "passport-app",
          "expires_at": "2024-12-31T23:59:59Z"
        }
      ]
    },
    "person.birthDate": {
      "owner": "CITIZEN",
      "provider": "primary", 
      "is_owner": false,
      "access_control_type": "restricted",
      "allow_list": [
        {
          "application_id": "passport-app",
          "expires_at": "2024-12-31T23:59:59Z"
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