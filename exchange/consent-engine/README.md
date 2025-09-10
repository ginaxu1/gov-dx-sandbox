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
3. **Data Owner Notification**: The data owner is notified via the consent portal
4. **Consent Decision**: The data owner can grant or deny consent
5. **Consent Tracking**: The system tracks consent status, expiry, and revocation

### Consent States

- **`pending`**: Consent request created, awaiting data owner decision
- **`granted`**: Data owner has granted consent
- **`denied`**: Data owner has denied consent
- **`expired`**: Consent has expired based on expiry_time
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
| `/consent` | POST | Create consent record |
| `/consent/{id}` | GET, PUT, DELETE | Get, update, or revoke consent |
| `/consent/portal` | GET, POST | Portal info and processing |
| `/data-owner/{owner}` | GET | Get consents by data owner |
| `/consumer/{consumer}` | GET | Get consents by consumer |
| `/admin/expiry-check` | POST | Check consent expiry |
| `/health` | GET | Health check |

## Example Usage

### Create Consent
```bash
curl -X POST http://localhost:8081/consent \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "passport-app",
    "data_owner": "user-nuwan-fernando-456", 
    "data_fields": ["person.permanentAddress", "person.birthDate"],
    "purpose": "passport application",
    "expiry_days": 30
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
