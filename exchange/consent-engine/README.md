# Consent Engine - Consent Flow Implementation

## Overview

Manages data owner consent workflow. Creates consent records, provides consent portals, and tracks consent status for data access requests requiring explicit permission.

## Architecture

1. **Orchestration Engine** requests consent record creation
2. **Consent Engine** creates consent record and provides portal URL
3. **Data Owner** interacts with consent portal to approve/deny consent
4. **Consent Engine** manages consent status and expiry
5. **Orchestration Engine** checks consent status before data access

## Consent Flow States

### Consent Status
- **pending**: Consent request created but not yet actioned
- **approved**: Data owner has approved the consent request
- **denied**: Data owner has denied the consent request
- **expired**: Consent has expired based on expiry time
- **revoked**: Consent has been revoked by the data owner

### Consent Types
- **realtime**: Real-time consent from the user via consent portal
- **offline**: Offline consent from the data owner (pre-configured)

## API Endpoints

### Consent Management

#### Create Consent Request
```http
POST /consent
Content-Type: application/json

{
  "data_consumer": "passport-app",
  "data_owner": "user123",
  "fields": ["person.permanentAddress", "person.birthDate"],
  "type": "realtime",
  "session_id": "sess_12345",
  "redirect_url": "https://passport-app.gov.lk/callback",
  "expiry_time": "30d",
  "metadata": {
    "request_source": "official_portal"
  }
}
```

#### Get Consent Status
```http
GET /consent/{consent_id}
```

#### Update Consent Status
```http
PUT /consent/{consent_id}
Content-Type: application/json

{
  "status": "approved",
  "updated_by": "system",
  "reason": "automatic_approval",
  "metadata": {
    "approval_method": "automated"
  }
}
```

#### Revoke Consent
```http
DELETE /consent/{consent_id}
Content-Type: application/json

{
  "reason": "user_requested_revocation"
}
```

### Consent Portal Integration

#### Process Consent Portal Request
```http
POST /consent-portal/
Content-Type: application/json

{
  "consent_id": "consent_123",
  "action": "approve",
  "data_owner": "user123",
  "session_id": "sess_12345",
  "reason": "user_approved_via_portal"
}
```

#### Get Consent Portal Information
```http
GET /consent-portal/?consent_id={consent_id}
```

### Data Owner Operations

#### Get Consents by Data Owner
```http
GET /data-owner/{data_owner_id}
```

### Consumer Operations

#### Get Consents by Consumer
```http
GET /consumer/{consumer_id}
```

### Administrative Operations

#### Check Consent Expiry
```http
POST /admin/expiry-check
```

## Consent Workflow Examples

### Example 1: Real-time Consent Flow

1. **PDP Decision**: PDP determines consent is required for `person.permanentAddress`
2. **Create Consent**: Orchestration Engine creates consent request
3. **Redirect User**: User is redirected to consent portal
4. **User Action**: User approves/denies consent via portal
5. **Update Status**: Consent status is updated
6. **Data Access**: Orchestration Engine proceeds with data access

```bash
# Step 1: Create consent request
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "data_consumer": "passport-app",
    "data_owner": "user123",
    "fields": ["person.permanentAddress"],
    "type": "realtime",
    "expiry_time": "30d"
  }'

# Step 2: User approves via consent portal
curl -X POST http://localhost:8081/consent-portal/ \
  -H "Content-Type: application/json" \
  -d '{
    "consent_id": "consent_123",
    "action": "approve",
    "data_owner": "user123"
  }'

# Step 3: Check consent status
curl http://localhost:8081/consent/consent_123
```

### Example 2: Offline Consent Flow

1. **Pre-configured Consent**: Data owner has pre-approved certain data access
2. **Create Consent**: System creates consent with offline type
3. **Automatic Approval**: Consent is automatically approved
4. **Data Access**: Orchestration Engine proceeds immediately

```bash
# Create offline consent
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "data_consumer": "passport-app",
    "data_owner": "user123",
    "fields": ["person.fullName", "person.nic"],
    "type": "offline"
  }'

# Automatically approve offline consent
curl -X PUT http://localhost:8081/consent/consent_456 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "system",
    "reason": "offline_consent_pre_approved"
  }'
```

## Integration with Policy Decision Point

The Consent Engine integrates with the PDP through the following flow:

1. **PDP Evaluation**: PDP evaluates request and determines consent requirements
2. **Consent Creation**: If consent required, Orchestration Engine creates consent record
3. **Portal Redirect**: User is redirected to consent portal
4. **Consent Decision**: Data owner provides consent decision
5. **Status Update**: Consent status is updated
6. **Data Access**: Orchestration Engine checks consent status before data access

### Current Integration Status

**PDP Implementation**: Successfully implemented with working ABAC authorization
**Consent Engine**: Fully implemented with all consent workflow features
**Data Loading**: OPA data loading mechanism working correctly
**Authorization Flow**: PDP correctly identifies consent requirements
**Consent Workflow**: Complete consent lifecycle management implemented

### Integration Testing

To test the complete PDP → Consent Engine integration:

1. **Test PDP Authorization** (PDP determines consent is required):
```bash
curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {"id": "passport-app"},
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.permanentAddress"],
      "data_owner": "drp"
    }
  }'
```

2. **Create Consent Record** (Orchestration Engine creates consent):
```bash
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "data_consumer": "passport-app",
    "data_owner": "user123",
    "fields": ["person.permanentAddress"],
    "type": "realtime",
    "expiry_time": "30d"
  }'
```

3. **Process Consent** (Data owner approves via portal):
```bash
curl -X POST http://localhost:8081/consent-portal/ \
  -H "Content-Type: application/json" \
  -d '{
    "consent_id": "consent_abc123",
    "action": "approve",
    "data_owner": "user123"
  }'
```

### PDP Response with Consent Required
```json
{
  "allow": true,
  "deny_reason": null,
  "consent_required": true,
  "consent_required_fields": ["person.permanentAddress", "person.birthDate"],
  "data_owner": "drp",
  "expiry_time": "30d",
  "conditions": {
    "consumer_verified": true,
    "resource_authorized": true,
    "time_valid": true,
    "context_valid": true
  }
}
```

## Consent Expiry Management

The Consent Engine supports automatic expiry management:

- **Expiry Time**: Set during consent creation (e.g., "30d", "1h", "7d")
- **Automatic Expiry**: System automatically marks consents as expired
- **Expiry Check**: Administrative endpoint to check and update expired consents
- **Renewal**: Expired consents can be renewed by creating new consent requests

### Expiry Time Formats
- `30d` - 30 days
- `1h` - 1 hour
- `7d` - 7 days
- `30m` - 30 minutes
- `60s` - 60 seconds

## Security Considerations

- **Data Owner Validation**: All consent portal actions validate data owner identity
- **Status Transition Validation**: Only valid status transitions are allowed
- **Session Management**: Session IDs are tracked for audit purposes
- **Audit Logging**: All consent actions are logged with timestamps and reasons
- **Expiry Enforcement**: Expired consents are automatically invalidated

## Status Transition Rules

```
pending → approved, denied, expired
approved → revoked, expired
denied → pending (allow retry)
expired → pending (allow renewal)
revoked → (no transitions allowed)
```

## Recent Updates and Fixes

### API Routing Improvements
- **Fixed consent endpoint routing**: Both `/consent` and `/consent/` endpoints now work correctly
- **Improved error handling**: Better handling of malformed requests and missing parameters
- **Enhanced response consistency**: Standardized JSON responses across all endpoints

### Test Organization
- **Centralized test scripts**: All test scripts moved to `/tests` directory for better organization
- **Comprehensive test coverage**: Added complete consent flow integration tests
- **Automated test runner**: `run-all-tests.sh` executes all test suites in sequence

### Container Deployment
- **Updated Docker containers**: Rebuilt containers with latest routing fixes
- **Improved build process**: Better separation of build and runtime stages
- **Enhanced security**: Non-root user execution in containers

## Error Handling

The Consent Engine provides comprehensive error handling:

- **Invalid Status Transitions**: Returns error for invalid status changes
- **Data Owner Mismatch**: Validates data owner identity for portal actions
- **Missing Records**: Returns appropriate HTTP status codes
- **Invalid Input**: Validates request structure and required fields

## Testing

### Quick Test
```bash
cd ../tests && ./test-complete-consent-flow.sh
```

### Manual API Test
```bash
# Create consent
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "data_consumer": "passport-app",
    "data_owner": "user123", 
    "fields": ["person.permanentAddress"],
    "type": "realtime"
  }'

# Update consent status
curl -X PUT http://localhost:8081/consent/{consent_id} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "updated_by": "user123"
  }'
```
