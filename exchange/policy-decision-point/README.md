# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Architecture

- **Technology**: Go + Open Policy Agent (OPA) + Rego policies
- **Port**: 8082
- **Purpose**: Attribute-based access control (ABAC) authorization with field-level access control

## Data Flow

1. **Input**: App ID + required fields from orchestration-engine
2. **Authorization**: Check `provider-metadata.json` for field permissions
3. **Output**: Allow/deny decision + consent requirements

### Access Control

- **Public fields**: Accessible to any app
- **Restricted fields**: Only apps in `allow_list` can access
- **Consent**: Required when `consent_required: true` AND `owner != provider`

### Example Response
```json
{
  "allow": true,
  "consent_required": true,
  "consent_required_fields": ["person.photo"]
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
    "app_id": "passport-app",
    "request_id": "req_123",
    "required_fields": ["person.fullName", "person.photo"]
  }'
```

## Data Files

- `data/provider-metadata.json` - Field permissions and consent requirements
- `policies/main.rego` - OPA authorization policies

### Field Access Control Logic

#### Access Control Types

1. **Public Fields** (`access_control_type: "public"`)
   - Any app can access
   - No allow_list restrictions
   - Consent only required if `consent_required: true` AND `provider != owner`

2. **Restricted Fields** (`access_control_type: "restricted"`)
   - Only apps in `allow_list` can access
   - Consent required if `consent_required: true` AND `provider != owner`

#### Consent Requirements

Consent is required when:
- `consent_required: true` AND
- `provider != owner` (data is cross-provider)

#### Allow List Construction

The `allow_list` is automatically constructed when a provider schema is approved and contains consumers that are authorized to access restricted fields. Here's how it works:

##### 1. Schema Submission and Approval Process

1. **Provider submits schema** via `POST /providers/{providerId}/schema-submissions`
2. **Admin approves schema** via `PUT /providers/{providerId}/schema-submissions/{schemaId}` with `status: "approved"`
3. **Schema converter** processes the GraphQL SDL and generates provider metadata
4. **Allow list entries** are created for each restricted field with authorized consumers

##### 2. Allow List Entry Structure

```json
{
  "consumerId": "passport-app",
  "expires_at": 1757560679,
  "grant_duration": "30d"
}
```

- **`consumerId`**: The application ID that's authorized to access the field
- **`expires_at`**: Epoch timestamp when the authorization expires
- **`grant_duration`**: Human-readable duration (e.g., "30d", "1h", "7d")

##### 3. Consumer Authorization Process

Consumers are added to the allow list through:

1. **Direct Admin Action**: Admin manually adds consumer to allow list
2. **Consent Approval**: When consent is granted, consumer is automatically added
3. **MOU (Memorandum of Understanding)**: Pre-approved consumers based on agreements
4. **API Integration**: Programmatic addition via admin APIs

##### 4. Allow List Management

```bash
# Add consumer to allow list for a specific field
curl -X POST http://localhost:8080/admin/fields/{fieldName}/allow-list \
  -H "Content-Type: application/json" \
  -d '{
    "consumerId": "passport-app",
    "expires_at": 1757560679,
    "grant_duration": "30d"
  }'

# Remove consumer from allow list
curl -X DELETE http://localhost:8080/admin/fields/{fieldName}/allow-list/{consumerId}

# List all consumers in allow list for a field
curl -X GET http://localhost:8080/admin/fields/{fieldName}/allow-list
```

##### 5. Allow List Lifecycle

1. **Creation**: When schema is approved or consumer is explicitly authorized
2. **Validation**: PDP checks if requesting app is in allow list
3. **Expiry**: Entries automatically expire based on `expires_at` timestamp
4. **Renewal**: Consumers can request renewal before expiry
5. **Revocation**: Admin can revoke access at any time

#### Field Authorization Examples

Each example demonstrates a different authorization scenario:

**Example 1: Public Field (No Restrictions)**
- Field: `person.fullName`
- Access Control Type: `public`
- Owner: `citizen` (data owner)
- Provider: `drp` (data provider)
- Consumer: `any-app` (any consumer can access)
- Consent Required: `false` (public fields don't require consent)
- Decision: **ALLOWED** (immediate access)

**Example 2: Restricted Field with Allow List (No Consent)**
- Field: `person.birthDate`
- Access Control Type: `restricted`
- Owner: `rgd` (different from provider)
- Provider: `drp` (data provider)
- Consumer: `driver-app` (in allow_list)
- Consent Required: `false` (consent_required = false)
- Decision: **ALLOWED** (consumer authorized via allow_list)

**Example 3: Restricted Field with Allow List (Consent Required)**
- Field: `person.permanentAddress`
- Access Control Type: `restricted`
- Owner: `rgd` (different from provider)
- Provider: `drp` (data provider)
- Consumer: `passport-app` (in allow_list)
- Consent Required: `true` (consent_required = true AND owner ≠ provider)
- Decision: **ALLOWED** (but consent required from data owner)

**Example 4: Restricted Field (Unauthorized Consumer)**
- Field: `person.nic`
- Access Control Type: `restricted`
- Owner: `rgd` (different from provider)
- Provider: `drp` (data provider)
- Consumer: `unauthorized-app` (NOT in allow_list)
- Consent Required: `false` (not applicable - access denied)
- Decision: **DENIED** (consumer not in allow_list)

## Enhanced Schema Conversion

The platform supports converting GraphQL SDL schemas to provider metadata format with two approaches:

### Approach 1: @owner Directive in GraphQL SDL

**GraphQL SDL with @owner directive:**
```graphql
directive @accessControl(type: String!) on FIELD_DEFINITION
directive @source(value: String!) on FIELD_DEFINITION
directive @isOwner(value: Boolean!) on FIELD_DEFINITION
directive @owner(value: String!) on FIELD_DEFINITION
directive @description(value: String!) on FIELD_DEFINITION

type PersonInfo {
  fullName: String! @accessControl(type: "public") @isOwner(value: false) @owner(value: "citizen")
  nic: String! @accessControl(type: "public") @isOwner(value: true)
  photo: String! @accessControl(type: "restricted") @isOwner(value: true)
  permanentAddress: String! @accessControl(type: "restricted") @isOwner(value: false) @owner(value: "citizen")
  birthDate: String! @accessControl(type: "restricted") @isOwner(value: false) @owner(value: "rgd")
}

type Query {
  getPerson(nic: String!): PersonInfo
}
```

**Converted Provider Metadata:**
```json
{
  "fields": {
    "person.fullName": {
      "consent_required": false,
      "owner": "citizen",
      "provider": "drp",
      "access_control_type": "public",
      "allow_list": []
    },
    "person.nic": {
      "consent_required": false,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "public",
      "allow_list": []
    },
    "person.photo": {
      "consent_required": false,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": []
    },
    "person.permanentAddress": {
      "consent_required": true,
      "owner": "citizen",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": []
    },
    "person.birthDate": {
      "consent_required": true,
      "owner": "rgd",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": []
    }
  }
}
```

### Approach 2: Separate Authorization Configuration

**JSON Payload with SDL and Authorization:**
```json
{
  "sdl": "directive @accessControl(type: String!) on FIELD_DEFINITION\n\ndirective @source(value: String!) on FIELD_DEFINITION\n\ndirective @isOwner(value: Boolean!) on FIELD_DEFINITION\n\ndirective @description(value: String!) on FIELD_DEFINITION\n\ntype PersonInfo {\n  fullName: String! @accessControl(type: \"public\") @isOwner(value: false)\n  nic: String! @accessControl(type: \"public\") @isOwner(value: true)\n  photo: String! @accessControl(type: \"restricted\") @isOwner(value: true)\n  permanentAddress: String! @accessControl(type: \"restricted\") @isOwner(value: false)\n  birthDate: String! @accessControl(type: \"restricted\") @isOwner(value: false)\n}\n\ntype Query {\n  getPerson(nic: String!): PersonInfo\n}",
  "field_owners": {
    "person.fullName": "citizen",
    "person.permanentAddress": "citizen",
    "person.birthDate": "rgd"
  },
  "authorization": {
    "person.permanentAddress": {
      "allowed_consumers": [
        {
          "consumerId": "passport-app",
          "expires_at": 1757560679,
          "grant_duration": "30d"
        }
      ]
    },
    "person.birthDate": {
      "allowed_consumers": [
        {
          "consumerId": "driver-app",
          "expires_at": 1757560679,
          "grant_duration": "7d"
        }
      ]
    }
  }
}
```

**Converted Provider Metadata:**
```json
{
  "fields": {
    "person.fullName": {
      "consent_required": false,
      "owner": "citizen",
      "provider": "drp",
      "access_control_type": "public",
      "allow_list": []
    },
    "person.nic": {
      "consent_required": false,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "public",
      "allow_list": []
    },
    "person.photo": {
      "consent_required": false,
      "owner": "drp",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": []
    },
    "person.permanentAddress": {
      "consent_required": true,
      "owner": "citizen",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": [
        {
          "consumerId": "passport-app",
          "expires_at": 1757560679,
          "grant_duration": "30d"
        }
      ]
    },
    "person.birthDate": {
      "consent_required": true,
      "owner": "rgd",
      "provider": "drp",
      "access_control_type": "restricted",
      "allow_list": [
        {
          "consumerId": "driver-app",
          "expires_at": 1757560679,
          "grant_duration": "7d"
        }
      ]
    }
  }
}
```

### Owner Resolution Priority

The system determines field ownership using this priority order:

1. **@owner directive** in GraphQL SDL (highest priority)
2. **field_owners** in authorization configuration
3. **@isOwner(value: true)** → provider is owner
4. **@isOwner(value: false)** → unknown owner (fallback)

### Consent Logic

Consent is required when:
- **Owner ≠ Provider** (cross-provider data)
- **AND** `access_control_type: "restricted"`

This ensures that data owned by citizens or other agencies requires explicit consent, while provider-owned data can be accessed based on allow_list authorization.
