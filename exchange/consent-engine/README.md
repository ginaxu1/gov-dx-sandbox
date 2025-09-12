# Consent Engine (CE)

Service that manages data owner consent workflows for data access requests.

## Overview

- **Technology**: Go + In-memory storage
- **Port**: 8081
- **Purpose**: Consent management and workflow coordination
- **Role**: Receives aggregated data owner and field information from Orchestration Engine

## Quick Start

```bash
# Run locally
cd consent-engine && go run main.go

# Run tests
go test -v

# Docker
docker build -t ce . && docker run -p 8081:8081 ce
```

## Workflow

1. **Consent Request**: Orchestration Engine aggregates data owners and fields, then calls CE
2. **Consent Creation**: A consent record is created with pending status
3. **Data Owner Notification**: Data owner is notified via SMS OTP (simplified to "000000" for testing)
4. **Consent Decision**: Data owner grants or denies consent through consent portal
5. **OTP Verification**: Data owner verifies identity using OTP "000000"
6. **Allow List Update**: Approved consent adds consumer to allow_list
7. **Consent Tracking**: System tracks status, expiry, and revocation

### Data Owner Aggregation

The Orchestration Engine handles data ownership aggregation and calls the Consent Engine with:

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
  "redirect_url": "https://passport-app.gov.lk/callback"
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/consent` | POST | Process new consent workflow request or update status |
| `/consent/{id}` | GET, POST, PUT, DELETE | Get, update, or revoke consent |
| `/consent/update` | POST | Update consent record with exact payload structure |
| `/consent-website` | GET | Serve consent website with OTP verification |
| `/consent-portal/` | GET, POST | Portal info and processing |
| `/data-owner/{owner}` | GET | Get consents by data owner |
| `/consumer/{consumer}` | GET | Get consents by consumer |
| `/admin/expiry-check` | POST | Check consent expiry |
| `/health` | GET | Health check |

## Key API Examples

### Create Consent Request

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

### Create Consent Record

```bash
curl -X POST http://localhost:8081/consent/update \
  -H "Content-Type: application/json" \
  -d '{
    "consent_id": "consent_abc123",
    "status": "pending",
    "type": "realtime",
    "owner_id": "199512345678",
    "data_consumer": "passport-app",
    "created_at": "2025-09-10T10:20:00Z",
    "updated_at": "2025-09-10T10:20:00Z",
    "expires_at": "2025-10-10T10:20:00Z",
    "fields": [
      "person.permanentAddress",
      "person.nic"
    ],
    "session_id": "session_123",
    "redirect_url": "https://passport-app.gov.lk/?consent_id=consent_abc123"
  }'
```

### Get Consent Information

```bash
curl -X GET http://localhost:8081/consent/consent_abc123
```

**Response:**
```json
{
  "consent_id": "consent_abc123",
  "owner_id": "199512345678",
  "data_consumer": "passport-app",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:20:00Z",
  "expires_at": "2025-10-10T10:20:00Z",
  "fields": [
    "person.permanentAddress",
    "person.nic"
  ],
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/?consent_id=consent_abc123"
}
```

### Update Consent Status

**User clicks Yes (Grant):**
```bash
curl -X PUT http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "citizen_199512345678",
    "reason": "Data Owner approved consent via portal",
    "otp": "000000"
  }'
```

**User clicks No (Reject):**
```bash
curl -X PUT http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "rejected",
    "updated_by": "citizen_199512345678",
    "reason": "Data Owner rejected consent via portal",
    "otp": "000000"
  }'
```

### Update Consent with OTP (Website)

**User approves with OTP:**
```bash
curl -X POST http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "otp": "123456"
  }'
```

**User rejects with OTP:**
```bash
curl -X POST http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "rejected",
    "otp": "123456"
  }'
```

### Get Consents by Data Owner

```bash
curl -X GET http://localhost:8081/data-owner/199512345678
```

### Revoke Consent

```bash
curl -X DELETE http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{"reason": "User requested data deletion"}'
```

## Consent States

- **`pending`**: Awaiting data owner decision
- **`approved`**: Data owner has approved consent (Yes button)
- **`rejected`**: Data owner has rejected consent (No button)
- **`expired`**: Consent has expired based on expires_at timestamp
- **`revoked`**: Consent has been revoked by data owner

## Allow List Integration

The Consent Engine integrates with the Policy Decision Point's allow_list system:

- **Consent Approval**: Automatically adds consumer to allow_list for requested fields
- **Consent Revocation**: Removes consumer from allow_list
- **Consent Expiry**: Removes consumer from allow_list when expired
- **Field-Specific Authorization**: Each field maintains its own allow_list

## Consent Website

The consent website provides a modern web interface for data owners with OTP verification:

### Website Endpoints
- `GET /consent-website` - Serve the consent website HTML
- `GET /consent/{id}` - Get consent information for the website
- `POST /consent/{id}` - Update consent status with OTP verification

### Website Workflow
1. **Step 1**: Data owner accesses the consent website with `consent_id` parameter
2. **Step 2**: Website displays consent information (app, fields, purpose)
3. **Step 3**: Data owner clicks "Approve Consents" or "Deny Consents"
4. **Step 4**: Website shows OTP verification form (hardcoded OTP: 123456)
5. **Step 5**: Data owner enters OTP and submits decision
6. **Step 6**: Website updates consent status and shows success message
7. **Step 7**: Data owner is redirected back to the consumer application

### Website Features
- **Modern UI**: Clean, responsive design with step indicators
- **OTP Verification**: Secure identity verification (simplified for testing)
- **Real-time Updates**: Live status updates and error handling
- **Mobile Friendly**: Responsive design for mobile devices
- **Accessibility**: Clear visual feedback and error messages

## Consent Portal

The consent portal provides a web interface for data owners:

### Portal Endpoints
- `GET /consent-portal/` - Get portal information for a consent
- `POST /consent-portal/` - Process consent decisions from the portal

### Portal Workflow
1. Data owner accesses the portal
2. Portal displays pending consent requests
3. Data owner can grant or deny each request
4. Portal updates consent status via API
5. Data owner can view and manage existing consents

## Configuration

### Environment Variables
- `PORT` - Service port (default: 8081)
- `ENVIRONMENT` - Environment (local/production)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)

## Integration

The Consent Engine integrates with:
- **Orchestration Engine**: Receives consent requests and provides consent status
- **Policy Decision Point**: Provides consent requirements for authorization decisions
- **Data Owner Portal**: Web interface for consent management

## Simplified OTP for Testing

For testing purposes, the OTP verification is simplified:

### OTP Configuration
- **OTP Value**: Always "000000"
- **Verification**: Any request with `"otp": "000000"` is automatically approved
- **SMS Simulation**: OTP is logged to console instead of sending actual SMS

### OTP Usage
```bash
# Send OTP (simplified)
curl -X POST http://localhost:8081/consent/{consent-id}/otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+94771234567"
  }'

# Response
{
  "success": true,
  "message": "OTP sent successfully (simplified for testing)",
  "consent_id": "consent_abc123",
  "phone_number": "+94771234567",
  "otp": "000000",
  "expires_at": "2025-09-10T10:25:00Z"
}
```

### OTP Verification
```bash
# Update consent with OTP verification
curl -X PUT http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "citizen_199512345678",
    "reason": "Data Owner approved consent via portal",
    "otp": "000000"
  }'
```

## Security

- Consent records are immutable once created (status can only be updated)
- Expired consents are automatically flagged
- Data owners can revoke consent at any time
- All consent operations are logged for audit purposes
- Consent records include purpose and expiry information for transparency
- OTP verification ensures data owner identity (simplified for testing)