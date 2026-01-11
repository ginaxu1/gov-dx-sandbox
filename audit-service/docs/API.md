# API Documentation

Complete API reference for the Audit Service.

## Base URL

```
http://localhost:3001
```

## Endpoints Overview

| Method | Endpoint          | Description                        |
| ------ | ----------------- | ---------------------------------- |
| POST   | `/api/audit-logs` | Create a new audit log entry       |
| GET    | `/api/audit-logs` | Retrieve audit logs with filtering |
| GET    | `/health`         | Service health check               |
| GET    | `/version`        | Service version information        |

---

## Audit Log Operations

### Create Audit Log

Create a new audit log entry for tracking events across services.

**Endpoint:** `POST /api/audit-logs`

**Request Headers:**

```
Content-Type: application/json
```

**Request Body:**

| Field                | Type          | Required | Description                                                  |
| -------------------- | ------------- | -------- | ------------------------------------------------------------ |
| `timestamp`          | string        | ✅       | ISO 8601 timestamp (RFC3339 format)                          |
| `status`             | string        | ✅       | Event status: `SUCCESS` or `FAILURE`                         |
| `actorType`          | string        | ✅       | Actor type: `SERVICE`, `ADMIN`, `MEMBER`, `SYSTEM`           |
| `actorId`            | string        | ✅       | Actor identifier (email, UUID, service name)                 |
| `targetType`         | string        | ✅       | Target type: `SERVICE` or `RESOURCE`                         |
| `traceId`            | string (UUID) | ❌       | Trace ID for distributed tracing (null for standalone)       |
| `eventType`          | string        | ❌       | Custom event type (e.g., `POLICY_CHECK`, `MANAGEMENT_EVENT`) |
| `eventAction`        | string        | ❌       | Action: `CREATE`, `READ`, `UPDATE`, `DELETE`                 |
| `targetId`           | string        | ❌       | Target identifier (resource ID, service name)                |
| `requestMetadata`    | object        | ❌       | Request payload (without PII/sensitive data)                 |
| `responseMetadata`   | object        | ❌       | Response or error details                                    |
| `additionalMetadata` | object        | ❌       | Additional context-specific data                             |

**Example Request:**

```bash
curl -X POST http://localhost:3001/api/audit-logs \
  -H "Content-Type: application/json" \
  -d '{
    "traceId": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-01-20T10:00:00Z",
    "eventType": "POLICY_CHECK",
    "eventAction": "READ",
    "status": "SUCCESS",
    "actorType": "SERVICE",
    "actorId": "orchestration-engine",
    "targetType": "SERVICE",
    "targetId": "policy-decision-point",
    "requestMetadata": {
      "schemaId": "schema-123",
      "requestedFields": ["name", "address"]
    },
    "responseMetadata": {
      "decision": "ALLOWED",
      "policyId": "policy-456"
    }
  }'
```

**Success Response: 201 Created**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "traceId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2024-01-20T10:00:00Z",
  "eventType": "POLICY_CHECK",
  "eventAction": "READ",
  "status": "SUCCESS",
  "actorType": "SERVICE",
  "actorId": "orchestration-engine",
  "targetType": "SERVICE",
  "targetId": "policy-decision-point",
  "requestMetadata": {
    "schemaId": "schema-123",
    "requestedFields": ["name", "address"]
  },
  "responseMetadata": {
    "decision": "ALLOWED",
    "policyId": "policy-456"
  },
  "additionalMetadata": null,
  "createdAt": "2024-01-20T10:00:00.123456Z"
}
```

**Error Response: 400 Bad Request**

```json
{
  "error": "Validation error: invalid timestamp format, expected RFC3339"
}
```

---

### Get Audit Logs

Retrieve audit logs with optional filtering and pagination.

**Endpoint:** `GET /api/audit-logs`

**Query Parameters:**

| Parameter     | Type          | Required | Default | Description                               |
| ------------- | ------------- | -------- | ------- | ----------------------------------------- |
| `traceId`     | string (UUID) | ❌       | -       | Filter by trace ID                        |
| `eventType`   | string        | ❌       | -       | Filter by event type                      |
| `eventAction` | string        | ❌       | -       | Filter by event action                    |
| `status`      | string        | ❌       | -       | Filter by status (`SUCCESS` or `FAILURE`) |
| `limit`       | integer       | ❌       | 100     | Max results per page (1-1000)             |
| `offset`      | integer       | ❌       | 0       | Number of results to skip                 |

**Example Requests:**

```bash
# Get all audit logs (paginated)
curl http://localhost:3001/api/audit-logs

# Filter by trace ID
curl http://localhost:3001/api/audit-logs?traceId=550e8400-e29b-41d4-a716-446655440000

# Filter by event type
curl http://localhost:3001/api/audit-logs?eventType=POLICY_CHECK

# Filter by status
curl http://localhost:3001/api/audit-logs?status=FAILURE

# Pagination
curl http://localhost:3001/api/audit-logs?limit=50&offset=100

# Multiple filters
curl http://localhost:3001/api/audit-logs?eventType=POLICY_CHECK&status=SUCCESS&limit=20
```

**Success Response: 200 OK**

```json
{
  "logs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "traceId": "550e8400-e29b-41d4-a716-446655440000",
      "timestamp": "2024-01-20T10:00:00Z",
      "eventType": "POLICY_CHECK",
      "eventAction": "READ",
      "status": "SUCCESS",
      "actorType": "SERVICE",
      "actorId": "orchestration-engine",
      "targetType": "SERVICE",
      "targetId": "policy-decision-point",
      "requestMetadata": {
        "schemaId": "schema-123"
      },
      "responseMetadata": {
        "decision": "ALLOWED"
      },
      "additionalMetadata": null,
      "createdAt": "2024-01-20T10:00:00.123456Z"
    }
  ],
  "total": 100,
  "limit": 100,
  "offset": 0
}
```

**Empty Response:**

```json
{
  "logs": [],
  "total": 0,
  "limit": 100,
  "offset": 0
}
```

**Error Response: 400 Bad Request**

```json
{
  "error": "Invalid traceId format, expected UUID"
}
```

---

## System Endpoints

### Health Check

Check if the service is running and responsive.

**Endpoint:** `GET /health`

**Example Request:**

```bash
curl http://localhost:3001/health
```

**Success Response: 200 OK**

```json
{
  "service": "audit-service",
  "status": "healthy"
}
```

**Note:** Database connectivity is verified at startup, not during health checks.

---

### Version Information

Get service version and build information.

**Endpoint:** `GET /version`

**Example Request:**

```bash
curl http://localhost:3001/version
```

**Success Response: 200 OK**

```json
{
  "service": "audit-service",
  "version": "1.0.0",
  "buildTime": "2024-01-20T10:00:00Z",
  "gitCommit": "abc123def456"
}
```

---

## Data Types & Enums

### Event Status

- `SUCCESS` - Event completed successfully
- `FAILURE` - Event failed or encountered an error

### Actor Types

- `SERVICE` - Internal service (e.g., orchestration-engine, consent-engine)
- `ADMIN` - Administrator user from admin portal
- `MEMBER` - End user/member from member portal
- `SYSTEM` - System-level operations (e.g., scheduled jobs)

### Target Types

- `SERVICE` - Target is a service
- `RESOURCE` - Target is a resource (e.g., data schema, policy)

### Event Actions (Optional)

- `CREATE` - Resource creation
- `READ` - Data retrieval
- `UPDATE` - Resource modification
- `DELETE` - Resource deletion

### Event Types (Examples)

Custom event types can be defined per use case:

- `POLICY_CHECK` - Policy evaluation
- `MANAGEMENT_EVENT` - Administrative action
- `DATA_ACCESS` - Data retrieval operation
- `CONSENT_CHANGE` - Consent modification

---

## Best Practices

### Timestamp Format

Always use RFC3339 format (ISO 8601):

```
2024-01-20T10:00:00Z          ✅ Correct
2024-01-20T10:00:00.123456Z   ✅ Correct (with microseconds)
2024-01-20 10:00:00           ❌ Wrong
01/20/2024 10:00 AM           ❌ Wrong
```

### Trace IDs

- Use UUIDs (RFC 4122) for trace IDs
- Generate at the entry point of a distributed flow
- Pass through all services in the request chain
- Use `null` for standalone events

### Metadata Guidelines

**DO:**

- ✅ Include operation context in `requestMetadata`
- ✅ Include decision/result in `responseMetadata`
- ✅ Use `additionalMetadata` for service-specific context

**DON'T:**

- ❌ Store PII (Personally Identifiable Information)
- ❌ Store sensitive data (passwords, tokens, keys)
- ❌ Store full response payloads with user data

**Good Example:**

```json
{
  "requestMetadata": {
    "schemaId": "schema-123",
    "requestedFields": ["name", "address"],
    "purpose": "data_access_request"
  },
  "responseMetadata": {
    "decision": "ALLOWED",
    "fieldsReturned": 2
  }
}
```

**Bad Example:**

```json
{
  "requestMetadata": {
    "userEmail": "john.doe@example.com", // ❌ PII
    "password": "secret123" // ❌ Sensitive data
  },
  "responseMetadata": {
    "fullUserData": {
      // ❌ Contains PII
      "name": "John Doe",
      "ssn": "123-45-6789"
    }
  }
}
```

### Error Handling

Always log failed operations:

```json
{
  "status": "FAILURE",
  "responseMetadata": {
    "error": "policy_evaluation_failed",
    "errorMessage": "Policy not found",
    "errorCode": "POL_404"
  }
}
```

### Pagination

For large result sets, use pagination:

```bash
# First page
GET /api/audit-logs?limit=100&offset=0

# Second page
GET /api/audit-logs?limit=100&offset=100

# Third page
GET /api/audit-logs?limit=100&offset=200
```

Use the `total` field in the response to calculate total pages.

---

## OpenAPI Specification

For the complete OpenAPI 3.0 specification, see [openapi.yaml](../openapi.yaml).

You can use tools like Swagger UI or Postman to import and explore the API interactively.
