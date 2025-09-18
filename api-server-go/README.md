# API Server (Go)

A REST API server for government data exchange portal management. Built with Go and runs on port 3000.

## Overview

The API server provides endpoints for managing:
- Consumer applications and authentication
- Provider submissions and profiles
- Provider schemas (with SDL support)
- Admin functions and allow list management
- Consent Management Workflow integration

## Architecture

### Single Asgardeo Client Pattern

This API server implements the **Single Asgardeo Client Pattern**:

- **One Asgardeo OAuth2 application** for the entire platform
- **Each consumer gets unique API credentials** from your system
- **All consumers use the same Asgardeo client** for token exchange
- **No per-consumer Asgardeo applications** needed

```
Consumer A (API Key: A123) ──┐
Consumer B (API Key: B456) ──┼──► API Server ──► Single Asgardeo Client ──► Asgardeo
Consumer C (API Key: C789) ──┘
```

## Quick Start

### Running the Server

```bash
# Start the API server
go run main.go

# Server starts on http://localhost:3000
```

### Health Check

```bash
curl http://localhost:3000/health
```

## API Endpoints

### Health & Debug
- `GET /health` - Check server health
- `GET /debug` - Debug information

### Consumer Management
- `GET /consumers` - List all consumers
- `POST /consumers` - Create new consumer
- `GET /consumers/{consumerId}` - Get specific consumer
- `PUT /consumers/{consumerId}` - Update consumer
- `DELETE /consumers/{consumerId}` - Delete consumer

### Consumer Applications
- `GET /consumer-applications` - List all consumer applications (admin view)
- `GET /consumer-applications/{consumerId}` - Get applications for specific consumer
- `POST /consumer-applications/{consumerId}` - Create application for specific consumer
- `GET /consumer-applications/{submissionId}` - Get specific application by submission ID
- `PUT /consumer-applications/{submissionId}` - Update application (admin approval)

### Provider Submissions
- `GET /provider-submissions` - List all provider submissions (admin view)
- `POST /provider-submissions` - Create new provider submission
- `GET /provider-submissions/{submissionId}` - Get specific provider submission
- `PUT /provider-submissions/{submissionId}` - Update provider submission (admin approval/rejection)

### Provider Management
- `GET /providers` - List all providers
- `GET /providers/{providerId}` - Get specific provider

### Provider Schema Management
- `GET /providers/{providerId}/schemas` - List approved schemas (status=approved, schemaId not null)
- `GET /providers/{providerId}/schema-submissions` - List provider's schema submissions (all statuses)
- `POST /providers/{providerId}/schema-submissions` - Create new schema submission (status: draft) or modify existing
- `GET /providers/{providerId}/schema-submissions/{schemaId}` - Get specific schema submission
- `PUT /providers/{providerId}/schema-submissions/{schemaId}` - Update schema submission (submit for review or admin approval/rejection)

### Authentication
- `POST /auth/exchange` - Exchange API credentials for Asgardeo access token
- `POST /auth/validate` - Validate Asgardeo access token

### Admin
- `GET /admin/metrics` - Get system metrics
- `GET /admin/recent-activity` - Get recent system activity
- `GET /admin/statistics` - Get detailed statistics

### Allow List Management
- `GET /admin/fields/{fieldName}/allow-list` - List consumers in allow_list for a field
- `POST /admin/fields/{fieldName}/allow-list` - Add consumer to allow_list for a field
- `PUT /admin/fields/{fieldName}/allow-list/{consumerId}` - Update consumer in allow_list
- `DELETE /admin/fields/{fieldName}/allow-list/{consumerId}` - Remove consumer from allow_list

## Detailed API Documentation

### Health Check

#### `GET /health`
**Description:** Check server health

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Consumer Management

#### `POST /consumers`
**Description:** Create a new consumer

**Payload:**
```json
{
  "consumerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string"
}
```

**Response:**
```json
{
  "consumerId": "string",
  "consumerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string",
  "createdAt": "datetime"
}
```

#### `GET /consumers`
**Description:** List all consumers

**Response:**
```json
{
  "count": 0,
  "items": [
    {
      "consumerId": "string",
      "consumerName": "string",
      "contactEmail": "string",
      "phoneNumber": "string",
      "createdAt": "datetime"
    }
  ]
}
```

#### `GET /consumers/{consumerId}`
**Description:** Get specific consumer

#### `PUT /consumers/{consumerId}`
**Description:** Update consumer

#### `DELETE /consumers/{consumerId}`
**Description:** Delete consumer

### Consumer Applications

#### `GET /consumer-applications`
**Description:** List all consumer applications (admin view)

**Response:**
```json
{
  "count": 0,
  "items": [
    {
      "submissionId": "string",
      "consumerId": "string",
      "status": "pending|approved|denied",
      "required_fields": {
        "fieldName": "boolean"
      },
      "createdAt": "datetime",
      "credentials": {
        "apiKey": "string",
        "apiSecret": "string"
      }
    }
  ]
}
```

#### `GET /consumer-applications/{consumerId}`
**Description:** Get applications for specific consumer

**Response:**
```json
{
  "count": 0,
  "items": [
    {
      "submissionId": "string",
      "consumerId": "string",
      "status": "pending|approved|denied",
      "required_fields": {
        "fieldName": "boolean"
      },
      "createdAt": "datetime"
    }
  ]
}
```

#### `POST /consumer-applications/{consumerId}`
**Description:** Create application for specific consumer

**Payload:**
```json
{
  "required_fields": {
    "fieldName": "boolean"
  }
}
```

**Response:**
```json
{
  "submissionId": "string",
  "consumerId": "string",
  "status": "pending",
  "required_fields": {
    "fieldName": "boolean"
  },
  "createdAt": "datetime"
}
```

#### `GET /consumer-applications/{submissionId}`
**Description:** Get specific application by submission ID

**Response:**
```json
{
  "submissionId": "string",
  "consumerId": "string",
  "status": "pending|approved|denied",
  "required_fields": {
    "fieldName": "boolean"
  },
  "createdAt": "datetime",
  "credentials": {
    "apiKey": "string",
    "apiSecret": "string"
  }
}
```

#### `PUT /consumer-applications/{submissionId}`
**Description:** Update application (admin approval)

**Payload:**
```json
{
  "status": "approved|denied",
  "required_fields": {
    "fieldName": "boolean"
  }
}
```

**Response:**
```json
{
  "submissionId": "string",
  "consumerId": "string",
  "status": "approved|denied",
  "required_fields": {
    "fieldName": "boolean"
  },
  "createdAt": "datetime",
  "credentials": {
    "apiKey": "string",
    "apiSecret": "string"
  },
  "providerId": "string"
}
```

### Provider Submissions

#### `GET /provider-submissions`
**Description:** List all provider submissions (admin view)

**Response:**
```json
{
  "count": 0,
  "items": [
    {
      "submissionId": "string",
      "providerName": "string",
      "contactEmail": "string",
      "phoneNumber": "string",
      "providerType": "government|board|business",
      "status": "pending|approved|rejected",
      "createdAt": "datetime"
    }
  ]
}
```

#### `POST /provider-submissions`
**Description:** Create new provider submission

**Payload:**
```json
{
  "providerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string",
  "providerType": "government|board|business"
}
```

**Response:**
```json
{
  "submissionId": "string",
  "providerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string",
  "providerType": "government|board|business",
  "status": "pending",
  "createdAt": "datetime"
}
```

#### `GET /provider-submissions/{submissionId}`
**Description:** Get specific provider submission

**Response:**
```json
{
  "submissionId": "string",
  "providerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string",
  "providerType": "government|board|business",
  "status": "pending|approved|rejected",
  "createdAt": "datetime"
}
```

#### `PUT /provider-submissions/{submissionId}`
**Description:** Update provider submission (admin approval/rejection)

**Payload:**
```json
{
  "status": "approved|rejected"
}
```

**Response:**
```json
{
  "submissionId": "string",
  "providerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string",
  "providerType": "government|board|business",
  "status": "approved|rejected",
  "createdAt": "datetime"
}
```

### Provider Management

#### `GET /providers`
**Description:** List all providers

**Response:**
```json
{
  "count": 2,
  "items": [
    {
      "providerId": "string",
      "providerName": "string",
      "contactEmail": "string",
      "phoneNumber": "string",
      "providerType": "government|board|business",
      "approvedAt": "datetime"
    }
  ]
}
```

#### `GET /providers/{providerId}`
**Description:** Get specific provider

**Response:**
```json
{
  "providerId": "string",
  "providerName": "string",
  "contactEmail": "string",
  "phoneNumber": "string",
  "providerType": "government|board|business",
  "approvedAt": "datetime"
}
```

### Provider Schema Management

#### `GET /providers/{providerId}/schemas`
**Description:** List approved schemas (status=approved, schemaId not null)

**Response:**
```json
{
  "count": 0,
  "items": [
    {
      "submissionId": "string",
      "providerId": "string",
      "schemaId": "string",
      "status": "approved",
      "sdl": "string",
      "createdAt": "datetime"
    }
  ]
}
```

#### `GET /providers/{providerId}/schema-submissions`
**Description:** List provider's schema submissions (all statuses)

#### `POST /providers/{providerId}/schema-submissions`
**Description:** Create new schema submission (status: draft) or modify existing

**Payload (New Schema):**
```json
{
  "sdl": "string"
}
```

**Payload (Modify Existing):**
```json
{
  "sdl": "string",
  "schema_id": "string"
}
```

**Response:**
```json
{
  "submissionId": "string",
  "providerId": "string",
  "status": "draft",
  "sdl": "string",
  "createdAt": "datetime"
}
```

#### `GET /providers/{providerId}/schema-submissions/{schemaId}`
**Description:** Get specific schema submission

#### `PUT /providers/{providerId}/schema-submissions/{schemaId}`
**Description:** Update schema submission (admin approval/rejection)

**Payload:**
```json
{
  "status": "approved|rejected|changes_required"
}
```

### Authentication

#### `POST /auth/exchange`
**Description:** Exchange API credentials for Asgardeo access token

**Payload:**
```json
{
  "apiKey": "string",
  "apiSecret": "string",
  "scope": "gov-dx-api"
}
```

**Response:**
```json
{
  "access_token": "string",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "gov-dx-api",
  "consumerId": "string"
}
```

#### `POST /auth/validate`
**Description:** Validate Asgardeo access token

**Payload:**
```json
{
  "token": "string"
}
```

**Response:**
```json
{
  "valid": true,
  "consumerId": "string"
}
```

### Admin

#### `GET /admin/metrics`
**Description:** Get system metrics

**Response:**
```json
{
  "total_consumer_apps": 0,
  "total_provider_submissions": 0,
  "total_providers": 0,
  "total_schemas": 0
}
```

#### `GET /admin/recent-activity`
**Description:** Get recent system activity

**Response:**
```json
[
  {
    "type": "string",
    "description": "string",
    "id": "string",
    "timestamp": "datetime"
  }
]
```

#### `GET /admin/statistics`
**Description:** Get detailed statistics by resource type

**Response:**
```json
{
  "consumer-apps": {
    "total": 0,
    "pending": 0,
    "approved": 0,
    "denied": 0
  },
  "provider-submissions": {
    "total": 0,
    "pending": 0,
    "approved": 0,
    "rejected": 0
  },
  "provider-schemas": {
    "total": 0,
    "draft": 0,
    "pending": 0,
    "approved": 0,
    "rejected": 0
  }
}
```

### Allow List Management

#### `GET /admin/fields/{fieldName}/allow-list`
**Description:** List consumers in allow_list for a field

**Response:**
```json
{
  "fieldName": "person.permanentAddress",
  "allowList": [
    {
      "consumerId": "string",
      "expires_at": 1757560679,
      "grant_duration": "30d",
      "reason": "Consent approved by data owner",
      "updated_by": "admin",
      "created_at": "datetime"
    }
  ]
}
```

#### `POST /admin/fields/{fieldName}/allow-list`
**Description:** Add consumer to allow_list for a field

**Payload:**
```json
{
  "consumerId": "string",
  "expires_at": 1757560679,
  "grant_duration": "30d",
  "reason": "Consent approved by data owner",
  "updated_by": "admin"
}
```

#### `PUT /admin/fields/{fieldName}/allow-list/{consumerId}`
**Description:** Update consumer in allow_list

**Payload:**
```json
{
  "consumerId": "string",
  "expires_at": 1757560679,
  "grant_duration": "60d",
  "reason": "Extended access period",
  "updated_by": "admin"
}
```

#### `DELETE /admin/fields/{fieldName}/allow-list/{consumerId}`
**Description:** Remove consumer from allow_list

## Schema Status Workflow

1. **Draft**: Initial status when schema is created
2. **Pending**: When provider submits draft for admin review (via PUT with status: "pending")
3. **Approved**: When admin approves the schema (schemaId is generated and provider-metadata.json is updated)
4. **Rejected**: When admin rejects the schema

## Important Notes

### Consumer Applications ID Handling
The `/consumer-applications/{id}` endpoint handles both consumer IDs and submission IDs based on ID format detection:
- IDs starting with `consumer_` are treated as consumer IDs
- IDs starting with `sub_` are treated as submission IDs
- For other formats, the HTTP method determines the behavior (POST/GET = consumer, PUT = submission)

### Provider Profile Creation
Provider profiles are created automatically when a provider submission is approved. There is no direct endpoint to create provider profiles - they are generated through the approval workflow.

### Schema Modification
To modify an existing approved schema, create a new schema submission with the `schema_id` field set to the original schema's ID. This creates a new submission that references the original schema for modification.

## Consent Management Workflow Integration

The API server integrates with the Consent Management Workflow through the **Consent Engine service** (running on port 8081). The consent workflow ensures that data consumers obtain proper consent before accessing personal data.

### Workflow Overview

1. **Data Consumer Request**: A data consumer (e.g., Passport Application) requests access to specific data fields
2. **Policy Decision**: The Policy Decision Point (PDP) determines if consent is required
3. **Consent Creation**: If consent is required, a consent record is created with "pending" status
4. **Data Owner Notification**: The data owner is notified via SMS OTP (simplified to "000000" for testing)
5. **Consent Decision**: The data owner grants or denies consent through the consent portal
6. **Allow List Update**: Approved consent adds the consumer to the allow list for the requested fields
7. **Data Access**: Once consent is approved, the data consumer can access the requested data

### Consent Engine Endpoints
The consent workflow is handled by the **Consent Engine service** (port 8081), not this API server:

#### `POST http://localhost:8081/consents`
**Description:** Initiate a consent workflow request

**Payload:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["person.permanentAddress"]
    }
  ],
  "purpose": "passport_application",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback"
}
```

**Response:**
```json
{
  "status": "pending",
  "redirect_url": "https://consent-portal.gov.lk/consent_abc123",
  "fields": ["person.permanentAddress"],
  "owner_id": "199512345678",
  "consent_id": "consent_abc123",
  "session_id": "session_123",
  "purpose": "passport_application",
  "message": "Consent required. Please visit the consent portal."
}
```

#### `GET http://localhost:8081/consents/{id}`
**Description:** Get consent workflow status

**Response:**
```json
{
  "consent_id": "consent_abc123",
  "owner_id": "199512345678",
  "data_consumer": "passport-app",
  "status": "approved",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:20:00Z",
  "expires_at": "2025-10-10T10:20:00Z",
  "fields": ["person.permanentAddress"],
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback"
}
```

### Service Integration

The API server works in conjunction with:
- **API Server (Port 3000)**: Manages consumers, providers, and allow lists
- **Orchestration Engine (Port 8080)**: Coordinates data requests and consent workflow
- **Consent Engine (Port 8081)**: Manages consent records and OTP verification
- **Policy Decision Point (Port 8082)**: Determines consent requirements and manages allow lists

### Simplified OTP for Testing

For testing purposes, the OTP verification is simplified:
- **OTP Value**: Always "000000"
- **Verification**: Any request with `"otp": "000000"` is automatically approved
- **SMS Simulation**: OTP is logged to console instead of sending actual SMS

## Example Usage

### Consumer Management
```bash
# Create a consumer
curl -X POST http://localhost:3000/consumers \
  -H "Content-Type: application/json" \
  -d '{
    "consumerName": "Test Consumer",
    "contactEmail": "test@example.com",
    "phoneNumber": "123-456-7890"
  }'

# Get all consumers
curl -X GET http://localhost:3000/consumers

# Update consumer
curl -X PUT http://localhost:3000/consumers/{consumerId} \
  -H "Content-Type: application/json" \
  -d '{
    "consumerName": "Updated Consumer Name"
  }'
```

### Consumer Application Management
```bash
# Create a consumer application for a specific consumer
curl -X POST http://localhost:3000/consumer-applications/{consumerId} \
  -H "Content-Type: application/json" \
  -d '{
    "required_fields": {
      "person.fullName": true,
      "person.email": true
    }
  }'

# Get all consumer applications (admin view)
curl -X GET http://localhost:3000/consumer-applications

# Get applications for specific consumer
curl -X GET http://localhost:3000/consumer-applications/{consumerId}

# Approve consumer application
curl -X PUT http://localhost:3000/consumer-applications/{submissionId} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved"
  }'
```

### Allow List Management Examples

```bash
# List consumers in allow_list for a field
curl -X GET http://localhost:3000/admin/fields/person.permanentAddress/allow-list

# Add consumer to allow_list for a field
curl -X POST http://localhost:3000/admin/fields/person.permanentAddress/allow-list \
  -H "Content-Type: application/json" \
  -d '{
    "consumerId": "passport-app",
    "expires_at": 1757560679,
    "grant_duration": "30d",
    "reason": "Consent approved by data owner",
    "updated_by": "admin"
  }'

# Update consumer in allow_list
curl -X PUT http://localhost:3000/admin/fields/person.permanentAddress/allow-list/passport-app \
  -H "Content-Type: application/json" \
  -d '{
    "consumerId": "passport-app",
    "expires_at": 1757560679,
    "grant_duration": "60d",
    "reason": "Extended access period",
    "updated_by": "admin"
  }'

# Remove consumer from allow_list
curl -X DELETE http://localhost:3000/admin/fields/person.permanentAddress/allow-list/passport-app
```

## Development

### Running
```bash
go run main.go
# Server starts on http://localhost:3000
```

### Building
```bash
go build -o api-server
./api-server
```

### Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./tests

# Run specific test
go test -v ./tests -run TestConsumerService
```

### Dependencies
- Uses shared utils from `github.com/gov-dx-sandbox/exchange/utils`
- Standard Go modules for HTTP, JSON, and logging

## Authentication Setup

For detailed authentication setup, token exchange, and Asgardeo integration, see [AUTH_SETUP.md](AUTH_SETUP.md).