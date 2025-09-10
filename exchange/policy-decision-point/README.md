# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Architecture

- **Technology**: Go + Open Policy Agent (OPA) + Rego policies
- **Port**: 8082
- **Purpose**: Attribute-based access control (ABAC) authorization with field-level access control

## Data Flow

1. **Input**: Consumer ID + required fields from orchestration-engine
2. **Authorization**: Check `provider-metadata.json` for field permissions
3. **Output**: Allow/deny decision + consent requirements

### Access Control

- **Public fields**: Accessible to any consumer
- **Restricted fields**: Only consumers in `allow_list` can access
- **Consent**: Required when `consent_required: true`

### Example Response
```json
{
  "allow": true,
  "consent_required": true,
  "consent_required_fields": ["person.photo"],
  "data_owner": "user-123",
  "expiry_time": "7d"
}
```

## Running

```bash
# Local development
cd /Users/tmp/gov-dx-sandbox/exchange && ./scripts/restore-local-build.sh
cd policy-decision-point && go run main.go

# Docker
docker build -t pdp . && docker run -p 8082:8082 pdp
```

## Testing

```bash
# All tests
go test -v

# New format tests
go test -v -run TestPolicyEvaluator_Authorize_NewFormat

# Schema conversion
go test -v -run TestSchemaConverter
```

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/decide` | POST | Authorization decision |
| `/health` | GET | Health check |
| `/debug` | GET | Debug info |

### Example Request
```bash
curl -X POST http://localhost:8082/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer_id": "passport-app",
    "required_fields": ["person.fullName", "person.photo"]
  }'
```

## Data Files

- `data/provider-metadata.json` - Field permissions and consent requirements
- `policies/main.rego` - OPA authorization policies

### Field Access Control Logic

#### Access Control Types

1. **Public Fields** (`access_control_type: "public"`)
   - Any consumer can access
   - No allow_list restrictions
   - Consent only required if `consent_required: true` AND `provider != owner`

2. **Restricted Fields** (`access_control_type: "restricted"`)
   - Only consumers in `allow_list` can access
   - Consent required if `consent_required: true` AND `provider != owner`

#### Consent Requirements

Consent is required when:
- `consent_required: true` AND
- `provider != owner` (data is cross-provider)

## Schema Conversion

The platform supports converting GraphQL SDL schemas to provider metadata format:

### GraphQL SDL Input
```graphql
type User {
  id: ID! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true)
  name: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: true)
  email: String! @accessControl(type: "restricted") @source(value: "authoritative") @isOwner(value: false)
}
```

### Converted Provider Metadata
```json
{
  "fields": {
    "user.id": {
      "consent_required": false,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "public",
      "allow_list": []
    },
    "user.name": {
      "consent_required": false,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "public",
      "allow_list": []
    },
    "user.email": {
      "consent_required": true,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": []
    }
  }
}
```
