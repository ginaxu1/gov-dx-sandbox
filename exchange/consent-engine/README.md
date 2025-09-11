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
3. **Data Owner Notification**: Data owner is notified via consent portal or SMS OTP
4. **Consent Decision**: Data owner grants or denies consent
5. **Allow List Update**: Approved consent adds consumer to allow_list
6. **Consent Tracking**: System tracks status, expiry, and revocation

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
| `/consent` | POST | Process new consent workflow request |
| `/consent/{id}` | GET, PUT, DELETE | Get, update, or revoke consent |
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
  "id": "consent_abc123",
  "status": "pending",
  "type": "realtime",
  "created_at": "2025-09-10T10:20:00Z",
  "updated_at": "2025-09-10T10:20:00Z",
  "expires_at": "1757560679",
  "grant_duration": "30d",
  "data_consumer": "passport-app",
  "data_owner": "199512345678",
  "fields": ["person.permanentAddress"],
  "consent_portal_url": "/consent-portal/consent_abc123",
  "session_id": "session_123",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "metadata": {
    "purpose": "passport_application",
    "request_id": "req_abc123"
  }
}
```

### Update Consent Status

```bash
curl -X PUT http://localhost:8081/consent/{consent-id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "citizen_199512345678",
    "reason": "User granted consent via portal"
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
- **`approved`**: Data owner has approved consent
- **`denied`**: Data owner has denied consent
- **`expired`**: Consent has expired based on expires_at timestamp
- **`revoked`**: Consent has been revoked by data owner

## Allow List Integration

The Consent Engine integrates with the Policy Decision Point's allow_list system:

- **Consent Approval**: Automatically adds consumer to allow_list for requested fields
- **Consent Revocation**: Removes consumer from allow_list
- **Consent Expiry**: Removes consumer from allow_list when expired
- **Field-Specific Authorization**: Each field maintains its own allow_list

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

## Security

- Consent records are immutable once created (status can only be updated)
- Expired consents are automatically flagged
- Data owners can revoke consent at any time
- All consent operations are logged for audit purposes
- Consent records include purpose and expiry information for transparency