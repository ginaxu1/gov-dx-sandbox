# Consent Engine (CE)

Service that manages data owner consent workflows for data access requests.

## Overview

- **Technology**: Go + In-memory storage
- **Port**: 8081
- **Purpose**: Consent management and workflow coordination

## Quick Start

```bash
# Run locally
cd consent-engine && go run main.go

# Run tests
go test -v

# Docker
docker build -t ce . && docker run -p 8081:8081 ce
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/consents` | POST | Create new consent |
| `/consents/{id}` | GET | Get consent information (user-facing) |
| `/consents/{id}` | PUT | Update consent status |
| `/consents/{id}` | DELETE | Revoke consent |
| `/health` | GET | Health check |

## Create Consent

**Endpoint:** `POST /consents`

**Request:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["person.permanentAddress", "person.photo"]
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
  "consent_id": "consent_122af00e",
  "owner_id": "199512345678",
  "app_id": "passport-app",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-15T14:51:14.395412+05:30",
  "updated_at": "2025-09-15T14:51:14.395412+05:30",
  "expires_at": "2025-10-15T14:51:14.395389+05:30",
  "grant_duration": "30d",
  "fields": ["person.permanentAddress", "person.photo"],
  "session_id": "session_123",
  "redirect_url": "http://localhost:5173/?consent_id=consent_122af00e",
  "otp_attempts": 0
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
        "fields": ["person.permanentAddress"]
      }
    ],
    "purpose": "passport_application",
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/callback"
  }'
```

## Get Consent Information

**Endpoint:** `GET /consents/{id}`

**Response (User-facing format):**
```json
{
  "consent_id": "consent_122af00e",
  "app_display_name": "Passport Application",
  "created_at": "2025-09-15T14:51:14.395412+05:30",
  "fields": ["person.permanentAddress", "person.photo"],
  "owner_name": "199512345678",
  "status": "pending",
  "type": "realtime"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8081/consents/consent_122af00e
```

## Update Consent Status

**Endpoint:** `PUT /consents/{id}`

**Request:**
```json
{
  "status": "approved",
  "owner_id": "199512345678",
  "message": "User approved consent"
}
```

**Response:**
```json
{
  "consent_id": "consent_122af00e",
  "status": "approved",
  "updated_at": "2025-09-15T14:55:00.000000+05:30",
  "message": "Consent status updated successfully"
}
```

**cURL Example:**
```bash
curl -X PUT http://localhost:8081/consents/consent_122af00e \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "owner_id": "199512345678",
    "message": "User approved consent"
  }'
```

## Revoke Consent

**Endpoint:** `DELETE /consents/{id}`

**Request:**
```json
{
  "reason": "User requested data deletion"
}
```

**Response:**
```json
{
  "consent_id": "consent_122af00e",
  "status": "revoked",
  "updated_at": "2025-09-15T14:55:00.000000+05:30",
  "message": "Consent revoked successfully"
}
```

**cURL Example:**
```bash
curl -X DELETE http://localhost:8081/consents/consent_122af00e \
  -H "Content-Type: application/json" \
  -d '{"reason": "User requested data deletion"}'
```

## Testing

### Test Complete Workflow

```bash
# 1. Create consent
CONSENT_RESPONSE=$(curl -s -X POST http://localhost:8081/consents \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "passport-app",
    "data_fields": [
      {
        "owner_type": "citizen",
        "owner_id": "test-123",
        "fields": ["person.fullName"]
      }
    ],
    "purpose": "testing",
    "session_id": "test-session",
    "redirect_url": "https://example.com"
  }')

# 2. Extract consent ID
CONSENT_ID=$(echo $CONSENT_RESPONSE | jq -r '.consent_id')

# 3. Get consent information
curl -X GET http://localhost:8081/consents/$CONSENT_ID

# 4. Update consent status
curl -X PUT http://localhost:8081/consents/$CONSENT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "owner_id": "test-123",
    "message": "User approved"
  }'

# 5. Revoke consent
curl -X DELETE http://localhost:8081/consents/$CONSENT_ID \
  -H "Content-Type: application/json" \
  -d '{"reason": "Test revocation"}'
```

## Consent States

- **`pending`**: Awaiting data owner decision
- **`approved`**: Data owner has approved consent
- **`rejected`**: Data owner has rejected consent
- **`expired`**: Consent has expired
- **`revoked`**: Consent has been revoked

## Health Check

```bash
curl http://localhost:8081/health
```

**Response:**
```json
{
  "service": "consent-engine",
  "status": "healthy"
}
```

## Configuration

### Environment Variables
- `PORT` - Service port (default: 8081)
- `CONSENT_PORTAL_URL` - Consent portal URL (default: http://localhost:5173)

## Integration

The Consent Engine integrates with:
- **Policy Decision Point**: Provides consent requirements for authorization decisions
- **Consent Portal**: Web interface for consent management
- **Data Consumer Applications**: Receives consent requests and provides status updates