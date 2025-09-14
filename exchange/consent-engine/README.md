# Consent Engine (CE)

Service that manages data owner consent workflows for data access requests with OTP verification and retry logic.

## Overview

- **Technology**: Go + In-memory storage
- **Port**: 8081
- **Purpose**: Consent management and workflow coordination
- **Role**: Receives aggregated data owner and field information from Orchestration Engine

## Quick Start

```bash
# Run locally
cd consent-engine && go run .

# Run tests
go test -v

# Run specific tests
go test -v -run TestConsentWorkflowIntegration

# Docker
docker build -t consent-engine . && docker run -p 8081:8081 consent-engine
```

## Integration with Other Services

- **Orchestration Engine**: Receives consent requests and returns simplified responses
- **Consent Portal**: Provides consent information and handles user decisions
- **Policy Decision Point**: May receive notifications about consent status changes

## Workflow

### 1. Consent Request Flow

1. **Consent Request**: Orchestration Engine calls `POST /consents` with data owner and field information
2. **Consent Creation**: A consent record is created with `pending` status
3. **Response**: Returns simplified response with `status` and `redirect_url` for orchestration engine

### 2. Consent Decision Flow

1. **User Decision**: Data owner visits consent portal and approves/rejects consent
2. **Status Update**: `PUT /consents/:consentId` updates status to `approved` or `rejected`
3. **OTP Flow** (if approved):
   - User receives OTP (hardcoded to "123456" for testing)
   - User enters OTP via `POST /consents/:consentId/otp`
   - 3 attempts allowed before automatic rejection
   - Successful OTP verification finalizes consent as `approved`

### 3. Status Transitions

- `pending` → `approved` (via PUT /consents/:consentId)
- `approved` → `approved` (via successful OTP verification)
- `approved` → `rejected` (via failed OTP after 3 attempts)
- `pending` → `rejected` (via direct rejection)

## API Endpoints

| Endpoint | Method | Description | Purpose |
|----------|--------|-------------|---------|
| `/consents` | POST | Create new consent request | Create consent from orchestration engine |
| `/consents/{id}` | GET | Get consent information | Retrieve consent details |
| `/consents/{id}` | PUT | Update consent status | User approves/rejects consent |
| `/consents/{id}/otp` | POST | Verify OTP | Verify OTP for approved consents |
| `/consents/{id}` | DELETE | Revoke consent | Revoke existing consent |
| `/consent-website` | GET | Serve consent website | Display consent form |
| `/consent-portal/` | GET | Get portal information | Portal data for consent management |
| `/consent-portal/` | POST | Process portal decisions | Handle consent decisions from portal |
| `/data-owner/{owner}` | GET | Get consents by data owner | List all consents for a data owner |
| `/consumer/{consumer}` | GET | Get consents by consumer | List all consents for a consumer |
| `/admin/expiry-check` | POST | Check consent expiry | Admin function to check expired consents |
| `/health` | GET | Health check | Service health status |



## Default Test Data

The consent engine includes a default hardcoded ConsentRecord for easier testing:

- **Consent ID**: `consent_03c134ae`
- **Owner ID**: `199512345678`
- **Status**: `pending`
- **Fields**: `["personInfo.permanentAddress"]`

This record is automatically available when the service starts.


## API Documentation

### 1. Create Consent Request

**Endpoint:** `POST /consents`  
**Description:** Create new consent request from orchestration engine  
**Purpose:** Create consent records for data access requests

**Request Payload:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["personInfo.permanentAddress"]
    }
  ],
  "purpose": "passport_application",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback"
}
```

**Response (Simplified for Orchestration Engine):**
```json
{
  "status": "pending",
  "redirect_url": "http://localhost:5173/?consent_id=consent_abc123"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "passport-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "199512345678",
        "fields": ["personInfo.permanentAddress"]
      }
    ],
    "purpose": "passport_application",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback"
  }'
```

### 2. Get Consent Information

**Endpoint:** `GET /consents/{id}`  
**Description:** Retrieve consent details  
**Purpose:** Get consent information for portal display

**Response:**
```json
{
  "consent_uuid": "consent_abc123",
  "owner_id": "199512345678",
  "data_consumer": "passport-app",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00+05:30",
  "updated_at": "2025-09-10T10:20:00+05:30",
  "expires_at": "2025-10-10T10:20:00+05:30",
  "fields": ["personInfo.permanentAddress"],
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "purpose": "passport_application",
  "message": "Consent required. Please visit the consent portal."
}
```

### 3. Update Consent Status

**Endpoint:** `PUT /consents/{id}`  
**Description:** Update consent status (approve/reject)  
**Purpose:** Handle user consent decisions

**Request Payload:**
```json
{
  "status": "approved",
  "owner_id": "199512345678",
  "message": "Approved via consent portal"
}
```

**Response:**
```json
{
  "consent_id": "consent_abc123",
  "owner_id": "199512345678",
  "app_id": "passport-app",
  "status": "approved",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00+05:30",
  "updated_at": "2025-09-10T10:25:00+05:30",
  "expires_at": "2025-10-10T10:20:00+05:30",
  "fields": ["personInfo.permanentAddress"],
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "purpose": "passport_application",
  "message": "User approved consent via portal - OTP verification required",
  "otp_attempts": 0
}
```

### 4. Verify OTP

**Endpoint:** `POST /consents/{id}/otp`  
**Description:** Verify OTP for approved consents  
**Purpose:** Final step in consent approval process

**Request Payload:**
```json
{
  "otp_code": "123456"
}
```

**Success Response:**
```json
{
  "success": true,
  "consent_id": "consent_abc123",
  "status": "approved",
  "message": "OTP verified successfully. Consent has been approved.",
  "updated_at": "2025-09-10T10:30:00+05:30"
}
```

**Error Response (Wrong OTP):**
```json
{
  "error": "Invalid OTP code. 2 attempts remaining."
}
```

**Error Response (After 3 Failed Attempts):**
```json
{
  "error": "OTP verification failed after 3 attempts. Consent has been rejected."
}
```

### 5. Revoke Consent

**Endpoint:** `DELETE /consents/{id}`  
**Description:** Revoke existing consent  
**Purpose:** Allow data owners to revoke previously granted consent

**Request Payload:**
```json
{
  "reason": "User requested revocation"
}
```

## Testing

### Unit Tests

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestConsentWorkflowIntegration

# Run with coverage
go test -v -cover
```

### Integration Tests

```bash
# Start the consent engine
go run . &

# Test the complete workflow
curl -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "passport-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "199512345678",
        "fields": ["personInfo.permanentAddress"]
      }
    ],
    "purpose": "passport_application",
    "session_id": "session_test",
    "redirect_url": "https://passport-app.gov.lk"
  }'

# Get the consent_id from response, then test approval
curl -X PUT http://localhost:8081/consents/consent_03c134ae \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "owner_id": "199512345678",
    "message": "Approved via consent portal"
  }'

# Test OTP verification
curl -X POST http://localhost:8081/consents/consent_03c134ae/otp \
  -H "Content-Type: application/json" \
  -d '{"otp_code": "123456"}'
```

### Scenarios

1. **Basic Flow**: Create consent → Approve → Verify OTP → Success
2. **Rejection Flow**: Create consent → Reject → No OTP needed
3. **OTP Retry Flow**: Create consent → Approve → Wrong OTP (3 times) → Auto-rejection
4. **Default Record**: Test with pre-existing `consent_03c134ae`

## OTP Configuration

- **Test OTP**: `123456` (hardcoded for testing)
- **Max Attempts**: 3
- **Timeout**: No timeout (manual retry)
- **Retry Logic**: After 3 failed attempts, consent is automatically rejected

## Data Structures

### ConsentRecord

```go
type ConsentRecord struct {
    ConsentID   string        `json:"consent_id"`
    OwnerID     string        `json:"owner_id"`
    AppID       string        `json:"app_id"`
    Status      ConsentStatus `json:"status"`
    Type        ConsentType   `json:"type"`
    CreatedAt   time.Time     `json:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at"`
    ExpiresAt   time.Time     `json:"expires_at"`
    Fields      []string      `json:"fields"`
    SessionID   string        `json:"session_id"`
    RedirectURL string        `json:"redirect_url"`
    Purpose     string        `json:"purpose"`
    Message     string        `json:"message"`
    OTPAttempts int           `json:"otp_attempts"`
}
```

### ConsentStatus Values

- `pending`: Initial state, awaiting user decision
- `approved`: User approved, may require OTP verification
- `rejected`: User rejected or OTP verification failed

## Error Handling

- **400 Bad Request**: Invalid JSON, missing required fields, invalid status transitions
- **404 Not Found**: Consent record not found
- **500 Internal Server Error**: Server-side errors

## Logging

The service logs all operations with structured logging:

```
time=2025-09-14T13:00:18.607+05:30 level=INFO msg="OTP verified successfully" consentId=consent_0fc1947b ownerId=1991111111 attempts=1
time=2025-09-14T12:59:59.676+05:30 level=INFO msg="OTP verification failed after 3 attempts" consentId=consent_03c134ae ownerId=199512345678
```
