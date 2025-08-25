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

| Endpoint | Method | Description | Purpose |
|----------|--------|-------------|---------|
| `/consent` | POST | Process new consent workflow request | Create consent request from orchestration engine |
| `/consent/{id}` | GET | Get consent information | Retrieve consent details for website/portal |
| `/consent/{id}` | POST | Update consent status with OTP | User approves/rejects with OTP verification |
| `/consent/{id}` | PUT | Update consent status | Admin or system updates consent status |
| `/consent/{id}` | DELETE | Revoke consent | Revoke existing consent |
| `/consent/update` | POST | Create/update consent record | Direct consent record management |
| `/consent-website` | GET | Serve consent website | Display consent form to data owner |
| `/consent-portal/` | GET | Get portal information | Portal data for consent management |
| `/consent-portal/` | POST | Process portal decisions | Handle consent decisions from portal |
| `/data-owner/{owner}` | GET | Get consents by data owner | List all consents for a data owner |
| `/consumer/{consumer}` | GET | Get consents by consumer | List all consents for a consumer |
| `/admin/expiry-check` | POST | Check consent expiry | Admin function to check expired consents |
| `/health` | GET | Health check | Service health status |

## Detailed API Documentation

### 1. Create Consent Request

**Endpoint:** `POST /consent`  
**Description:** Process new consent workflow request from orchestration engine  
**Purpose:** Create consent records for data access requests

**Request Payload:**
```json
{
  "app_id": "passport-app",
  "data_fields": [
    {
      "owner_type": "citizen",
      "owner_id": "199512345678",
      "fields": ["person.permanentAddress", "person.nic"]
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
  "status": "pending",
  "redirect_url": "http://localhost:8081/consent-website?consent_id=consent_abc123",
  "fields": ["person.permanentAddress", "person.nic"],
  "owner_id": "199512345678",
  "consent_id": "consent_abc123",
  "session_id": "session_123",
  "purpose": "passport_application",
  "message": "Consent required. Please visit the consent portal."
}
```

**cURL Example:**
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

### 2. Get Consent Information

**Endpoint:** `GET /consent/{id}`  
**Description:** Retrieve consent details for website/portal  
**Purpose:** Get consent information to display to data owner

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

**cURL Example:**
```bash
curl -X GET http://localhost:8081/consent/consent_abc123
```

### 3. Update Consent Status (User Decision with OTP)

**Endpoint:** `POST /consent/{id}`  
**Description:** Update consent status when user approves/rejects with OTP verification  
**Purpose:** Handle user decisions from consent website

**Request Payload (Approve):**
```json
{
  "status": "approved",
  "otp": "000000"
}
```

**Request Payload (Reject):**
```json
{
  "status": "rejected",
  "otp": "000000"
}
```

**Response:**
```json
{
  "consent_uuid": "consent_abc123",
  "status": "approved",
  "updated_at": "2025-09-10T10:25:00Z",
  "message": "Consent status updated successfully"
}
```

**cURL Examples:**
```bash
# User approves consent
curl -X POST http://localhost:8081/consent/consent_abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "otp": "000000"
  }'

# User rejects consent
curl -X POST http://localhost:8081/consent/consent_abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "rejected",
    "otp": "000000"
  }'
```

### 4. Update Consent Status (Admin/System)

**Endpoint:** `PUT /consent/{id}`  
**Description:** Update consent status for admin or system operations  
**Purpose:** Admin updates or system status changes

**Request Payload:**
```json
{
  "status": "approved",
  "updated_by": "citizen_199512345678",
  "reason": "Data Owner approved consent via portal",
  "otp": "000000"
}
```

**Response:**
```json
{
  "consent_uuid": "consent_abc123",
  "status": "approved",
  "updated_at": "2025-09-10T10:25:00Z",
  "message": "Consent status updated successfully"
}
```

**cURL Example:**
```bash
curl -X PUT http://localhost:8081/consent/consent_abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "citizen_199512345678",
    "reason": "Data Owner approved consent via portal",
    "otp": "000000"
  }'
```

### 5. Create/Update Consent Record

**Endpoint:** `POST /consent/update`  
**Description:** Create or update consent record with exact payload structure  
**Purpose:** Direct consent record management

**Request Payload:**
```json
{
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
}
```

**Response:**
```json
{
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
}
```

**cURL Example:**
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

### 6. Get Consents by Data Owner

**Endpoint:** `GET /data-owner/{owner}`  
**Description:** List all consents for a specific data owner  
**Purpose:** Retrieve all consent records for a data owner

**Response:**
```json
{
  "count": 2,
  "items": [
    {
      "consent_id": "consent_abc123",
      "owner_id": "199512345678",
      "data_consumer": "passport-app",
      "status": "approved",
      "type": "realtime",
      "created_at": "2025-09-10T10:20:00Z",
      "updated_at": "2025-09-10T10:25:00Z",
      "expires_at": "2025-10-10T10:20:00Z",
      "fields": ["person.permanentAddress"],
      "session_id": "session_123"
    }
  ]
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8081/data-owner/199512345678
```

### 7. Get Consents by Consumer

**Endpoint:** `GET /consumer/{consumer}`  
**Description:** List all consents for a specific consumer  
**Purpose:** Retrieve all consent records for a consumer application

**Response:**
```json
{
  "count": 1,
  "items": [
    {
      "consent_id": "consent_abc123",
      "owner_id": "199512345678",
      "data_consumer": "passport-app",
      "status": "approved",
      "type": "realtime",
      "created_at": "2025-09-10T10:20:00Z",
      "updated_at": "2025-09-10T10:25:00Z",
      "expires_at": "2025-10-10T10:20:00Z",
      "fields": ["person.permanentAddress"],
      "session_id": "session_123"
    }
  ]
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8081/consumer/passport-app
```

### 8. Revoke Consent

**Endpoint:** `DELETE /consent/{id}`  
**Description:** Revoke existing consent  
**Purpose:** Allow data owner to revoke previously granted consent

**Request Payload:**
```json
{
  "reason": "User requested data deletion"
}
```

**Response:**
```json
{
  "consent_id": "consent_abc123",
  "status": "revoked",
  "updated_at": "2025-09-10T10:30:00Z",
  "message": "Consent revoked successfully"
}
```

**cURL Example:**
```bash
curl -X DELETE http://localhost:8081/consent/consent_abc123 \
  -H "Content-Type: application/json" \
  -d '{"reason": "User requested data deletion"}'
```

### 9. Consent Website

**Endpoint:** `GET /consent-website`  
**Description:** Serve consent website with OTP verification  
**Purpose:** Display consent form to data owner

**Query Parameters:**
- `consent_id` (required): The consent ID to display

**Response:** HTML page with consent form

**cURL Example:**
```bash
curl -X GET "http://localhost:8081/consent-website?consent_id=consent_abc123"
```

### 10. Consent Portal Information

**Endpoint:** `GET /consent-portal/`  
**Description:** Get portal information for a consent  
**Purpose:** Portal data for consent management

**Query Parameters:**
- `consent_id` (required): The consent ID

**Response:**
```json
{
  "consentId": "consent_abc123",
  "status": "pending",
  "dataConsumer": "passport-app",
  "dataOwner": "199512345678",
  "fields": ["person.permanentAddress"],
  "consentPortalUrl": "/consent-portal/consent_abc123",
  "expiresAt": "2025-10-10T10:20:00Z",
  "createdAt": "2025-09-10T10:20:00Z"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8081/consent-portal/?consent_id=consent_abc123"
```

### 11. Health Check

**Endpoint:** `GET /health`  
**Description:** Check service health status  
**Purpose:** Monitor service availability

**Response:**
```json
{
  "service": "consent-engine",
  "status": "healthy",
  "timestamp": "2025-09-10T10:20:00Z"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8081/health
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