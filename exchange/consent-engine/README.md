# Consent Engine (CE)

Service that manages data owner consent workflows for data access requests.

## Overview

- **Technology**: Go + In-memory storage
- **Port**: 8081
- **Purpose**: Consent management and workflow coordination
- **Authentication**: JWT tokens from Asgardeo with email-based authorization

## Quick Start

### 1. Environment Setup

First, set up your environment variables:

```bash
# Run the setup script to create .env.local
./setup-env.sh

# Edit .env.local with your Asgardeo credentials
nano .env.local
```

The service will automatically load environment variables from `.env.local` if it exists. You can also set them directly in your shell or use the example file:

```bash
# Copy the example file
cp env.example .env.local

# Edit with your values
nano .env.local
```

### 2. Run the Service

```bash
# Run locally
cd consent-engine && go run main.go engine.go jwt_verifier.go

# Run tests
go test -v

# Docker
docker build -t ce . && docker run -p 8081:8081 ce
```

## API Endpoints

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/consents` | POST | Create new consent | None |
| `/consents/{id}` | GET | Get consent information (user-facing) | **JWT Required** |
| `/consents/{id}` | PUT | Update consent status | **JWT Required** |
| `/consents/{id}` | POST | Update consent status (alternative) | **JWT Required** |
| `/consents/{id}` | DELETE | Revoke consent | **JWT Required** |
| `/data-info/{id}` | GET | Get data owner information | None |
| `/consent-portal` | POST | Create consent via portal | None |
| `/consent-portal` | PUT | Update consent via portal | None |
| `/consent-portal` | GET | Get consent portal info | None |
| `/consent-website` | GET | Serve consent portal website | None |
| `/data-owner/{id}` | GET | Get consents by data owner | None |
| `/consumer/{id}` | GET | Get consents by consumer | None |
| `/admin/expiry-check` | POST | Check expired consents | None |
| `/health` | GET | Health check | None |

## Authentication

The Consent Engine uses JWT tokens from Asgardeo for authentication on protected endpoints. The JWT must contain a valid email claim that matches the consent owner's email.

### JWT Configuration

- **JWKS URL**: `https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks`
- **Issuer**: `https://api.asgardeo.io/t/lankasoftwarefoundation`
- **Audience**: `lankasoftwarefoundation`

### Environment Variables

- `ASGARDEO_JWKS_URL` - JWKS endpoint URL (default: https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks)
- `ASGARDEO_ISSUER` - JWT issuer URL (default: https://api.asgardeo.io/t/lankasoftwarefoundation)
- `ASGARDEO_AUDIENCE` - JWT audience (default: lankasoftwarefoundation)

### JWT Token Format

```bash
Authorization: Bearer <jwt_token>
```

The JWT token must contain an email claim in one of these fields:
- `email`
- `sub` (subject)
- `preferred_username`

### Email Authorization

For protected endpoints (`/consents/{id}`), the JWT email must match the consent owner's email. If they don't match, the request will be rejected with a 403 Forbidden response.

## Data Owner Information

**Endpoint:** `GET /data-info/{consentId}`

Retrieves only the owner ID and email for a specific consent record. This endpoint does not require authentication.

**Response:**
```json
{
  "owner_id": "test-owner-123",
  "owner_email": "owner@example.com"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8081/data-info/consent_122af00e
```

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
**Authentication:** JWT Required (email must match consent owner)

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
curl -X GET http://localhost:8081/consents/consent_122af00e \
  -H "Authorization: Bearer <jwt_token>"
```

**Error Responses:**
- `403 Forbidden` - JWT token invalid or email doesn't match consent owner
- `404 Not Found` - Consent record not found

## Update Consent Status

**Endpoint:** `PUT /consents/{id}`  
**Authentication:** JWT Required (email must match consent owner)

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
  -H "Authorization: Bearer <jwt_token>" \
  -d '{
    "status": "approved",
    "owner_id": "199512345678",
    "message": "User approved consent"
  }'
```

**Error Responses:**
- `403 Forbidden` - JWT token invalid or email doesn't match consent owner
- `404 Not Found` - Consent record not found

## Revoke Consent

**Endpoint:** `DELETE /consents/{id}`  
**Authentication:** JWT Required (email must match consent owner)

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
  -H "Authorization: Bearer <jwt_token>" \
  -d '{"reason": "User requested data deletion"}'
```

**Error Responses:**
- `403 Forbidden` - JWT token invalid or email doesn't match consent owner
- `404 Not Found` - Consent record not found

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

#### Required (for JWT authentication)
- `ASGARDEO_BASE_URL` - Your Asgardeo organization URL (e.g., https://api.asgardeo.io/t/lankasoftwarefoundation)
- `ASGARDEO_CLIENT_ID` - Your Asgardeo application client ID
- `ASGARDEO_CLIENT_SECRET` - Your Asgardeo application client secret

#### Optional (with defaults)
- `ASGARDEO_JWKS_URL` - JWKS endpoint URL (default: https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks)
- `ASGARDEO_ISSUER` - JWT issuer URL (default: https://api.asgardeo.io/t/lankasoftwarefoundation)
- `ASGARDEO_AUDIENCE` - JWT audience
- `PORT` - Service port (default: 8081)
- `CONSENT_PORTAL_URL` - Consent portal URL (default: http://localhost:5173)

### Getting Asgardeo Credentials

1. Go to [Asgardeo Console](https://console.asgardeo.io/)
2. Navigate to your organization
3. Go to Applications → Your App → Settings
4. Copy the following values:
   - **Base URL**: Your organization URL
   - **Client ID**: From the application settings
   - **Client Secret**: From the application settings (if using confidential client)

## Integration

The Consent Engine integrates with:
- **Policy Decision Point**: Provides consent requirements for authorization decisions
- **Consent Portal**: Web interface for consent management
- **Data Consumer Applications**: Receives consent requests and provides status updates