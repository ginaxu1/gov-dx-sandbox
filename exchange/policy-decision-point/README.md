# Policy Decision Point (PDP)

Authorization service using Open Policy Agent (OPA) that evaluates data access requests and determines consent requirements.

## Overview

- **Technology**: Go + Open Policy Agent (OPA) + Rego policies
- **Port**: 8082
- **Purpose**: Attribute-based access control (ABAC) with field-level permissions

## Quick Start

```bash
# Run locally
cd policy-decision-point && go run main.go

# Run tests
go test -v

# Docker
docker build -t pdp . && docker run -p 8082:8082 pdp
```

## API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/decide` | POST | Authorization decision |
| `/health` | GET | Health check |

### Authorization Request

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

### Response Format

```json
{
  "allow": true,
  "consent_required": true,
  "consent_required_fields": ["person.photo"]
}
```

## Access Control Logic

### Field Types

1. **Public Fields** (`access_control_type: "public"`)
   - Any app can access
   - Consent only required if `consent_required: true` AND `provider != owner`

2. **Restricted Fields** (`access_control_type: "restricted"`)
   - Only apps in `allow_list` can access
   - Consent required if `consent_required: true` AND `provider != owner`

### Consent Requirements

Consent is required when:
- `consent_required: true` AND
- `provider != owner` (cross-provider data)

### Allow List Management

The `allow_list` is automatically constructed when provider schemas are approved:

```json
{
  "consumerId": "passport-app",
  "expires_at": 1757560679,
  "grant_duration": "30d"
}
```

**Management APIs:**
```bash
# Add consumer to allow list
curl -X POST http://localhost:8080/admin/fields/{fieldName}/allow-list \
  -d '{"consumerId": "passport-app", "expires_at": 1757560679, "grant_duration": "30d"}'

# Remove consumer
curl -X DELETE http://localhost:8080/admin/fields/{fieldName}/allow-list/{consumerId}

# List consumers
curl -X GET http://localhost:8080/admin/fields/{fieldName}/allow-list
```

## Schema Conversion

The Policy Decision Point converts GraphQL SDL schemas into provider metadata format for authorization. The PDP only needs to understand access control types and consumer authorization lists - data ownership is handled by the Orchestration Engine.

### What is Schema Conversion?

When a provider submits a GraphQL schema, the PDP needs to extract:
- **Access Control Types**: Whether fields are public or restricted
- **Consumer Authorization Lists**: Which consumers can access restricted fields
- **Field Metadata**: Basic field information for authorization decisions

**Note**: Data ownership and consent requirements are determined by the Orchestration Engine, not the PDP.

### Approach 1: @accessControl Directive (Self-Contained SDL)

**What it means**: All authorization information is embedded directly in the GraphQL SDL using access control directives.

**When to use**: When you want to keep all authorization metadata in the schema itself.

**Example:**
```graphql
type PersonInfo {
  fullName: String! @accessControl(type: "public")
  nic: String! @accessControl(type: "public")
  photo: String! @accessControl(type: "restricted")
  permanentAddress: String! @accessControl(type: "restricted")
  birthDate: String! @accessControl(type: "restricted")
}
```

**How it works:**
- `@accessControl(type: "public")` - Field is accessible to any consumer
- `@accessControl(type: "restricted")` - Field requires consumer to be in allow_list

**Benefits:**
- Self-contained and portable
- Clear access control in schema
- No external configuration needed
- Version control friendly

### Approach 2: Separate Authorization Config (Clean SDL + External Config)

**What it means**: Keep the GraphQL SDL clean and provide authorization information in a separate JSON configuration.

**When to use**: When you want to keep the schema clean and manage authorization separately, or when you need dynamic authorization updates without changing the schema.

**Example:**
```json
{
  "sdl": "type PersonInfo { fullName: String! }",
  "authorization": {
    "person.fullName": {
      "access_control_type": "public"
    },
    "person.permanentAddress": {
      "access_control_type": "restricted",
      "allowed_consumers": [
        {"consumerId": "passport-app", "expires_at": 1757560679, "grant_duration": "30d"}
      ]
    }
  }
}
```

**How it works:**
- **`sdl`**: Clean GraphQL schema without authorization directives
- **`authorization`**: Maps field names to access control types and allowed consumers

**Benefits:**
- Clean, readable GraphQL schema
- Flexible authorization management
- Easy to update without schema changes
- Better separation of concerns

### Conversion Process

Both approaches result in the same provider metadata format:

```json
{
  "fields": {
    "person.fullName": {
      "access_control_type": "public",
      "allow_list": []
    },
    "person.permanentAddress": {
      "access_control_type": "restricted",
      "allow_list": [
        {
          "consumerId": "passport-app",
          "expires_at": 1757560679,
          "grant_duration": "30d"
        }
      ]
    }
  }
}
```

### Choosing Between Approaches

| Factor | Approach 1 (@accessControl) | Approach 2 (Separate Config) |
|--------|----------------------------|-------------------------------|
| **Schema Clarity** | More verbose | Clean and simple |
| **Portability** | Self-contained | Requires external config |
| **Flexibility** | Schema changes needed | Easy config updates |
| **Version Control** | All in one place | Split across files |
| **Dynamic Updates** | Requires schema change | Config-only updates |
| **Team Workflow** | Schema-focused | Config-focused |

## Data Files

- `data/provider-metadata.json` - Field permissions and consent requirements
- `policies/main.rego` - OPA authorization policies

## Authorization Examples

**Example 1: Public Field (No Restrictions)**
- Field: `person.fullName`
- Access Control: `public`
- Owner: `citizen`, Provider: `drp`
- Consumer: `any-app`
- Decision: **ALLOWED** (immediate access)

**Example 2: Restricted Field (Allow List + Consent)**
- Field: `person.permanentAddress`
- Access Control: `restricted`
- Owner: `citizen`, Provider: `drp`
- Consumer: `passport-app` (in allow_list)
- Consent Required: `true` (owner ≠ provider)
- Decision: **ALLOWED** (but consent required)

**Example 3: Unauthorized Consumer**
- Field: `person.nic`
- Access Control: `restricted`
- Consumer: `unauthorized-app` (NOT in allow_list)
- Decision: **DENIED**