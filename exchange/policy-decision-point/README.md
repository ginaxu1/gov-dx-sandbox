# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Architecture

- **Technology**: Go + Open Policy Agent (OPA) + Rego policies
- **Port**: 8082
- **Purpose**: ABAC authorization with field-level access control

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

## Schema Conversion

Converts GraphQL SDL to provider metadata:

```graphql
type User {
  id: ID! @accessControl(type: "public") @isOwner(value: true)
  email: String! @accessControl(type: "restricted") @isOwner(value: false)
}
```

→ Generates field-level permissions in `provider-metadata.json`
