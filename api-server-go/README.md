# API Server (Go)

A RESTful API server for government data exchange portal management. Built with Go and runs on port 3000.

## Overview

The API server provides RESTful endpoints for managing:
- Consumer applications
- Provider submissions and profiles
- Provider schemas (with SDL support)
- Admin functions

## Architecture

```
api-server-go/
├── main.go              # Server entry point
├── handlers/
│   └── server.go        # Generic HTTP handlers (unified)
├── models/              # Data structures
│   ├── consumer.go      # Consumer types
│   └── provider.go      # Provider types
├── services/            # Business logic
│   ├── consumer.go      # Consumer operations
│   ├── provider.go      # Provider operations
│   ├── schema_converter.go # SDL to provider metadata conversion
│   └── admin.go         # Admin dashboard
├── tests/               # Unit tests
└── go.mod              # Dependencies
```

**Key Features:**
- Generic handler pattern reduces code duplication
- Shared utils package for common operations
- In-memory storage with thread-safe operations
- Comprehensive input validation

## API Endpoints

### Consumer Management
- `GET /consumers` - List all consumers
- `POST /consumers` - Create new consumer
- `GET /consumers/{consumerId}` - Get specific consumer
- `PUT /consumers/{consumerId}` - Update consumer
- `DELETE /consumers/{consumerId}` - Delete consumer

### Consumer Applications
- `GET /consumer-applications` - List all consumer applications
- `POST /consumer-applications` - Create new consumer application
- `GET /consumer-applications/{submissionId}` - Get specific consumer application
- `PUT /consumer-applications/{submissionId}` - Update consumer application (admin approval)

### Provider Management
- `GET /provider-submissions` - List all provider submissions
- `POST /provider-submissions` - Create new provider submission
- `GET /provider-submissions/{submissionId}` - Get specific provider submission
- `PUT /provider-submissions/{submissionId}` - Update provider submission (admin approval)
- `GET /provider-profiles` - List all approved provider profiles
- `GET /provider-profiles/{providerId}` - Get specific provider profile

### RESTful Provider Schema Management
- `GET /providers/{providerId}/schemas` - List approved schemas (status=approved, schemaId not null)
- `GET /providers/{providerId}/schema-submissions` - List provider's schema submissions (all statuses)
- `POST /providers/{providerId}/schema-submissions` - Create new schema submission (status: draft) or modify existing
- `GET /providers/{providerId}/schema-submissions/{schemaId}` - Get specific schema submission
- `PUT /providers/{providerId}/schema-submissions/{schemaId}` - Update schema submission (admin approval/rejection)
- `POST /providers/{providerId}/schema-submissions/{schemaId}/submit` - Submit draft schema for admin review (draft → pending)

### Admin
- `GET /admin/dashboard` - Dashboard with metrics and recent activity

## Detailed API Documentation

### Health Check
- `GET /health` - Check server health

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

#### `POST /consumer-applications`
**Description:** Create new consumer application

**Payload:**
```json
{
  "consumerId": "string",
  "required_fields": {
    "fieldName": "boolean"
  }
}
```

#### `GET /consumer-applications`
**Description:** List all consumer applications

#### `GET /consumer-applications/{submissionId}`
**Description:** Get specific consumer application

#### `PUT /consumer-applications/{submissionId}`
**Description:** Update consumer application (admin approval)

**Payload:**
```json
{
  "status": "approved|denied",
  "required_fields": {
    "fieldName": "boolean"
  }
}
```

### Provider Management

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
  "providerType": "string",
  "status": "pending",
  "createdAt": "datetime"
}
```

#### `GET /provider-submissions`
**Description:** List all provider submissions

#### `GET /provider-submissions/{submissionId}`
**Description:** Get specific provider submission

#### `PUT /provider-submissions/{submissionId}`
**Description:** Update provider submission (admin approval)

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
  "providerType": "string",
  "status": "approved",
  "createdAt": "datetime",
  "providerId": "string"
}
```

#### `GET /provider-profiles`
**Description:** List all approved provider profiles

#### `GET /provider-profiles/{providerId}`
**Description:** Get specific provider profile

### RESTful Provider Schema Management

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

#### `POST /providers/{providerId}/schema-submissions/{schemaId}/submit`
**Description:** Submit draft schema for admin review (draft → pending)

### Admin

#### `GET /admin/dashboard`
**Description:** Get admin dashboard with metrics and recent activity

**Response:**
```json
{
  "overview": {
    "total_applications": 0,
    "total_profiles": 0,
    "total_schemas": 0,
    "total_submissions": 0
  },
  "recent_activity": [
    {
      "type": "string",
      "description": "string",
      "id": "string",
      "timestamp": "datetime"
    }
  ],
  "submissions": {
    "approved": 0,
    "pending": 0
  }
}
```

## Schema Status Workflow

1. **Draft**: Initial status when schema is created
2. **Pending**: When provider submits draft for admin review
3. **Approved**: When admin approves the schema (schemaId is generated)
4. **Rejected**: When admin rejects the schema

## Example Usage

### Provider Schema Submission with SDL
```bash
# Submit a schema with GraphQL SDL
curl -X POST http://localhost:3000/providers/drp/schemas \
  -H "Content-Type: application/json" \
  -d '{
    "sdl": "directive @accessControl(type: String!) on FIELD_DEFINITION\n\ndirective @source(value: String!) on FIELD_DEFINITION\n\ndirective @isOwner(value: Boolean!) on FIELD_DEFINITION\n\ndirective @description(value: String!) on FIELD_DEFINITION\n\ntype BirthInfo {\n  birthCertificateID: ID! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  birthPlace: String! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  birthDate: String! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n}\n\ntype User {\n  id: ID! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  name: String! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  email: String! @accessControl(type: \"public\") @source(value: \"authoritative\") @isOwner(value: false)\n  birthInfo: BirthInfo @accessControl(type: \"public\") @source(value: \"authoritative\") @description(value: \"Default Description\")\n}\n\ntype Query {\n  getUser(id: ID!): User @description(value: \"Default Description\")\n  listUsers: [User!]! @description(value: \"Default Description\")\n  getBirthInfo(userId: ID!): BirthInfo @description(value: \"Default Description\")\n  listUsersByBirthPlace(birthPlace: String!): [User!]! @description(value: \"Default Description\")\n  searchUsersByName(name: String!): [User!]! @description(value: \"Default Description\")\n}"
  }'
```

### Schema Approval (Updates provider-metadata.json)
```bash
# Approve a schema - this automatically updates provider-metadata.json
curl -X PUT http://localhost:3000/provider-schemas/{schema-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved"
  }'
```

### Consumer Application Management
```bash
# Create a consumer application
curl -X POST http://localhost:3000/consumers \
  -H "Content-Type: application/json" \
  -d '{
    "requiredFields": {
      "person.fullName": "required",
      "person.email": "required"
    }
  }'

# Get all consumer applications
curl -X GET http://localhost:3000/consumers

# Approve consumer application
curl -X PUT http://localhost:3000/consumers/{app-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved"
  }'
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

## Implementation Details

**Handler Pattern:** Generic `handleCollection` and `handleItem` functions reduce code duplication across all endpoints.

**Data Storage:** In-memory maps with `sync.RWMutex` for thread-safe concurrent access.

**Validation:** Server-side validation for required fields with detailed error messages.

**Response Format:** Consistent JSON responses with proper HTTP status codes and error handling.

**Admin Dashboard:** Dynamic generation of metrics, recent activity, and aggregated data from all services.