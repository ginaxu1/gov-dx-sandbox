# Consent Engine (CE)

The Consent Engine is a service that manages data owner consent workflows for data access requests. It creates, tracks, and manages consent records, enabling data owners to grant or revoke access to their personal data.

## How It Works

### Architecture
- **Technology**: Go + In-memory storage
- **Purpose**: Consent management and workflow coordination
- **Port**: 8081

### Consent Workflow

1. **Consent Request**: When the Policy Decision Point determines that consent is required, the Orchestration Engine calls the Consent Engine
2. **Consent Creation**: A consent record is created with pending status
3. **Data Owner Notification**: The data owner is notified via the consent portal or SMS OTP
4. **Consent Decision**: The data owner can grant or deny consent
5. **Allow List Update**: When consent is approved, the consumer is automatically added to the allow_list for the requested fields
6. **Consent Tracking**: The system tracks consent status, expiry, and revocation

### Allow List Integration

The Consent Engine integrates with the Policy Decision Point's allow_list system:

- **Consent Approval**: When a data owner approves consent, the consumer is automatically added to the allow_list for the requested fields
- **Consent Revocation**: When consent is revoked, the consumer is removed from the allow_list
- **Consent Expiry**: When consent expires, the consumer is removed from the allow_list
- **Field-Specific Authorization**: Each field maintains its own allow_list with authorized consumers

### Consent States

- **`pending`**: Consent request created, awaiting data owner decision
- **`approved`**: Data owner has approved consent
- **`denied`**: Data owner has denied consent
- **`expired`**: Consent has expired based on expires_at timestamp
- **`revoked`**: Consent has been revoked by data owner

### Data Storage

The Consent Engine uses in-memory storage for development and testing. In production, this would be replaced with a persistent database.

## Running

### Local Development
```bash
# Ensure you're in local development state
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/restore-local-build.sh

# Run the service
cd consent-engine
go run main.go
```

### Docker
```bash
# Build and run
docker build -t ce .
docker run -p 8081:8081 ce
```

> **For complete deployment instructions, see [Main README](../README.md#building-for-local-development)**

## Testing

### Unit Tests
```bash
# All tests
go test -v

# Specific test suites
go test -v -run TestConsentEngine
go test -v -run TestConsentWorkflow
go test -v -run TestConsentExpiry
```

### Test Coverage
The test suite covers:

1. **Consent Creation**: Creating new consent records
2. **Consent Retrieval**: Getting consent by ID, data owner, or consumer
3. **Consent Updates**: Updating consent status (grant/deny/revoke)
4. **Consent Expiry**: Handling expired consents
5. **Consent Portal**: Portal functionality for data owners
6. **Error Handling**: Invalid requests and edge cases

> **For integration testing, see [Main README](../README.md#testing)**

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/consent` | POST | Process new consent workflow request |
| `/consent/{id}` | GET, PUT, DELETE | Get, update, or revoke consent |
| `/consent-portal/` | GET, POST | Portal info and processing |
| `/data-owner/{owner}` | GET | Get consents by data owner |
| `/consumer/{consumer}` | GET | Get consents by consumer |
| `/admin/expiry-check` | POST | Check consent expiry |
| `/health` | GET | Health check |

## API Documentation

### 1. Process Consent Request

**Endpoint:** `POST /consent`  
**Description:** Process a new consent workflow request from the Orchestration Engine

**Request Payload:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["personInfo.address"]
    }
  ],
  "purpose": "passport_application",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "expires_at": 1757560679,
  "grant_duration": "30d"
}
```

**Response:**
```json
{
  "id": "consent_abc123",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:20:00Z",
  "expires_at": "2025-10-10T10:20:00Z",
  "data_consumer": "passport-app",
  "data_owner": "199512345678",
  "fields": ["personInfo.address"],
  "consent_portal_url": "/consent-portal/consent_abc123",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "metadata": {
    "purpose": "passport_application",
    "request_id": "req_abc123"
  }
}
```

### 2. Get Consent Status

**Endpoint:** `GET /consent/{id}`  
**Description:** Retrieve a specific consent record by ID

**Response:**
```json
{
  "id": "consent_abc123",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:20:00Z",
  "expires_at": "2025-10-10T10:20:00Z",
  "data_consumer": "passport-app",
  "data_owner": "199512345678",
  "fields": ["person.permanentAddress"],
  "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "metadata": {
    "purpose": "passport_application",
    "app_id": "passport-app",
    "data_owners": ["199512345678"],
    "grant_duration": "30d"
  }
}
```

### 3. Update Consent Status

**Endpoint:** `PUT /consent/{id}`  
**Description:** Update consent status (approve/deny)

**Request Payload:**
```json
{
  "status": "approved",
  "updated_by": "citizen_199512345678",
  "reason": "User granted consent via portal"
}
```

**Response:**
```json
{
  "success": true,
  "operation": "consent_updated",
  "data": {
    "id": "consent_abc123",
    "status": "approved",
    "type": "realtime",
    "created_at": "2025-09-10T10:20:00Z",
    "updated_at": "2025-09-10T10:25:00Z",
    "expires_at": "2025-10-10T10:20:00Z",
    "data_consumer": "passport-app",
    "data_owner": "199512345678",
    "fields": ["person.permanentAddress"],
    "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback",
    "metadata": {
      "purpose": "passport_application",
      "app_id": "passport-app",
      "updated_by": "citizen_199512345678",
      "update_reason": "User granted consent via portal",
      "last_updated": "2025-09-10T10:25:00Z"
    }
  },
  "metadata": {
    "id": "consent_abc123",
    "status": "approved"
  }
}
```

### 4. Revoke Consent

**Endpoint:** `DELETE /consent/{id}`  
**Description:** Revoke a consent record

**Request Payload:**
```json
{
  "reason": "User requested data deletion"
}
```

**Response:**
```json
{
  "success": true,
  "operation": "consent_revoked",
  "data": {
    "id": "consent_abc123",
    "status": "revoked",
    "type": "realtime",
    "created_at": "2025-09-10T10:20:00Z",
    "updated_at": "2025-09-10T10:30:00Z",
    "expires_at": "2025-10-10T10:20:00Z",
    "data_consumer": "passport-app",
    "data_owner": "199512345678",
    "fields": ["person.permanentAddress"],
    "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback",
    "metadata": {
      "purpose": "passport_application",
      "app_id": "passport-app",
      "updated_by": "citizen_199512345678",
      "update_reason": "User requested data deletion",
      "last_updated": "2025-09-10T10:30:00Z"
    }
  },
  "metadata": {
    "id": "consent_abc123",
    "reason": "User requested data deletion"
  }
}
```

### 5. Get Consents by Data Owner

**Endpoint:** `GET /data-owner/{owner_id}`  
**Description:** Retrieve all consent records for a specific data owner

**Response:**
```json
{
  "data_owner": "199512345678",
  "consents": [
    {
      "id": "consent_abc123",
      "status": "approved",
      "type": "realtime",
      "created_at": "2025-09-10T10:20:00Z",
      "updated_at": "2025-09-10T10:25:00Z",
      "expires_at": "2025-10-10T10:20:00Z",
      "data_consumer": "passport-app",
      "data_owner": "199512345678",
      "fields": ["person.permanentAddress"],
      "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
      "session_id": "session_123",
      "redirect_url": "https://passport-app.gov.lk/callback"
    }
  ],
  "count": 1
}
```

### 6. Get Consents by Consumer

**Endpoint:** `GET /consumer/{consumer_id}`  
**Description:** Retrieve all consent records for a specific data consumer

**Response:**
```json
{
  "consumer": "passport-app",
  "consents": [
    {
      "id": "consent_abc123",
      "status": "approved",
      "type": "realtime",
      "created_at": "2025-09-10T10:20:00Z",
      "updated_at": "2025-09-10T10:25:00Z",
      "expires_at": "2025-10-10T10:20:00Z",
      "data_consumer": "passport-app",
      "data_owner": "199512345678",
      "fields": ["person.permanentAddress"],
      "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
      "session_id": "session_123",
      "redirect_url": "https://passport-app.gov.lk/callback"
    }
  ],
  "count": 1
}
```

### 7. Consent Portal Info

**Endpoint:** `GET /consent-portal/`  
**Description:** Get consent portal information for a specific consent

**Query Parameters:**
- `consent_id` (required): The consent ID to get portal info for

**Response:**
```json
{
  "consent_id": "consent_abc123",
  "status": "pending",
  "data_consumer": "passport-app",
  "data_owner": "199512345678",
  "fields": ["person.permanentAddress"],
  "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
  "expires_at": "2025-10-10T10:20:00Z",
  "created_at": "2025-09-10T10:20:00Z"
}
```

### 8. Process Consent Portal Request

**Endpoint:** `POST /consent-portal/`  
**Description:** Process consent decisions from the portal

**Request Payload:**
```json
{
  "consent_id": "consent_abc123",
  "action": "approve",
  "updated_by": "citizen_199512345678",
  "reason": "User granted consent"
}
```

**Response:**
```json
{
  "success": true,
  "operation": "portal_request_processed",
  "data": {
    "id": "consent_abc123",
    "status": "approved",
    "type": "realtime",
    "created_at": "2025-09-10T10:20:00Z",
    "updated_at": "2025-09-10T10:25:00Z",
    "expires_at": "2025-10-10T10:20:00Z",
    "data_consumer": "passport-app",
    "data_owner": "199512345678",
    "fields": ["person.permanentAddress"],
    "consent_portal_url": "https://consent-portal.gov.lk/consent/consent_abc123",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback"
  },
  "metadata": {
    "id": "consent_abc123",
    "action": "approve",
    "status": "approved"
  }
}
```

### 9. Check Consent Expiry

**Endpoint:** `POST /admin/expiry-check`  
**Description:** Check and update expired consent records

**Response:**
```json
{
  "success": true,
  "operation": "consent_expiry_checked",
  "data": {
    "expired_records": [
      {
        "id": "consent_xyz789",
        "status": "expired",
        "expires_at": "2025-09-05T10:20:00Z",
        "data_consumer": "old-app",
        "data_owner": "199512345678"
      }
    ],
    "count": 1,
    "checked_at": "2025-09-10T10:30:00Z"
  },
  "metadata": {
    "expired_count": 1
  }
}
```

### 10. Health Check

**Endpoint:** `GET /health`  
**Description:** Service health check

**Response:**
```json
{
  "status": "healthy",
  "service": "consent-engine",
  "timestamp": "2025-09-10T10:30:00Z",
  "version": "1.0.0"
}
```

## Example Usage

### Process New Consent Request
```bash
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
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
    "redirect_url": "https://passport-app.gov.lk/callback",
    "expires_at": 1757560679,
    "grant_duration": "30d"
  }'
```

### Get Consent by ID
```bash
curl -X GET http://localhost:8081/consent/{consent-id}
```

### Get Consents by Data Owner
```bash
curl -X GET http://localhost:8081/data-owner/user-nuwan-fernando-456
```

### Get Consents by Consumer
```bash
curl -X GET http://localhost:8081/consumer/passport-app
```

### Update Consent Status
```bash
curl -X PUT http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "granted"
  }'
```

### Revoke Consent
```bash
curl -X DELETE http://localhost:8081/consent/{consent-id}
```

### Check Consent Expiry
```bash
curl -X POST http://localhost:8081/admin/expiry-check
```

## Configuration

### Environment Variables
- `PORT` - Service port (default: 8081)
- `ENVIRONMENT` - Environment (local/production)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)

### Consent Record Structure
```json
{
  "id": "consent-123",
  "consumer_id": "passport-app",
  "data_owner": "user-nuwan-fernando-456",
  "data_fields": ["person.permanentAddress", "person.birthDate"],
  "purpose": "passport application",
  "status": "granted",
  "created_at": "2025-09-10T10:00:00Z",
  "updated_at": "2025-09-10T10:05:00Z",
  "expires_at": "2025-10-10T10:00:00Z"
}
```

## Consent Portal

The consent portal provides a web interface for data owners to manage their consents:

### Portal Endpoints
- `GET /consent/portal` - Get portal information and pending consents
- `POST /consent/portal` - Process consent decisions from the portal

### Portal Workflow
1. Data owner accesses the portal
2. Portal displays pending consent requests
3. Data owner can grant or deny each request
4. Portal updates consent status via API
5. Data owner can view and manage existing consents

## Integration

The Consent Engine integrates with:
- **Orchestration Engine**: Receives consent requests and provides consent status
- **Policy Decision Point**: Provides consent requirements for authorization decisions
- **Data Owner Portal**: Web interface for consent management

## Security Considerations

- Consent records are immutable once created (status can only be updated)
- Expired consents are automatically flagged
- Data owners can revoke consent at any time
- All consent operations are logged for audit purposes
- Consent records include purpose and expiry information for transparency
