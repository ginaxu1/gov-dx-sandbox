# Policy Decision Point (PDP) - ABAC Implementation

## Overview

Implements Attribute-Based Access Control (ABAC) using Open Policy Agent (OPA) to evaluate data access requests and determine consent requirements.

## Architecture

1. **Orchestration Engine** sends data request to **PDP**
2. **PDP** evaluates request against ABAC policies
3. **PDP** returns authorization decision and consent requirements
4. **Orchestration Engine** coordinates with **Consent Engine** if needed

## ABAC Model

### Subject (Consumer) Attributes
- Consumer ID and approved data fields
- Consumer authorization based on grants data

### Resource Attributes
- Data field definitions and metadata
- Consent requirements per field
- Data ownership information

### Action Attributes
- Supported actions (currently: "read")
- Action-specific authorization policies

## Request Format

```json
{
  "consumer": {
    "id": "passport-app",
    "name": "Passport Application Service",
    "type": "government_service"
  },
  "request": {
    "resource": "person_data",
    "action": "read",
    "data_fields": ["person.fullName", "person.nic", "person.photo"],
    "data_owner": "drp"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Response Format

### Successful Authorization (No Consent Required)
```json
{
  "allow": true,
  "deny_reason": null,
  "consent_required": false,
  "consent_required_fields": [],
  "data_owner": "",
  "expiry_time": "",
  "conditions": {
    "consumer_verified": true,
    "resource_authorized": true,
    "action_authorized": true
  }
}
```

### Successful Authorization (Consent Required)
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
    "action_authorized": true
  }
}
```

### Denied Authorization
```json
{
  "allow": false,
  "deny_reason": "Consumer not authorized for requested fields",
  "consent_required": false,
  "consent_required_fields": [],
  "data_owner": "",
  "expiry_time": "",
  "conditions": {}
}
```

## Policy Evaluation Flow

1. **Input Validation**: Validate request structure and required fields
2. **Consumer Authorization**: Check if consumer exists and has access to requested fields
3. **Resource Authorization**: Verify requested fields exist in metadata
4. **Action Authorization**: Validate requested action is permitted
5. **Consent Analysis**: Determine which fields require consent
6. **Decision Generation**: Return structured authorization decision

## Integration with Consent Flow

When `consent_required: true` is returned:

1. The Orchestration Engine receives the PDP decision
2. Consent Engine is triggered for the specified data owner
3. User is redirected to consent portal
4. Consent Service manages the consent workflow
5. Only after consent is granted, data access proceeds

## Configuration Files

### Consumer Grants (`data/consumer-grants.json`)
Defines which consumers can access which data fields:
```json
{
  "passport-app": {
    "approved_fields": [
      "person.fullName",
      "person.nic", 
      "person.photo"
    ]
  }
}
```

### Provider Metadata (`data/provider-metadata.json`)
Defines field-level consent requirements and ownership:
```json
{
  "fields": {
    "person.fullName": { "consent_required": false, "owner": "drp" },
    "person.permanentAddress": {
      "consent_required": true,
      "owner": "drp",
      "expiry_time": "30d"
    }
  }
}
```

## Testing

### Quick Test
```bash
cd ../tests && ./test-pdp.sh
```

### Manual API Test
```bash
# Test no consent required
curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {"id": "passport-app"},
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic"]
    }
  }'

# Test consent required
curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {"id": "passport-app"},
    "request": {
      "resource": "person_data", 
      "action": "read",
      "data_fields": ["person.permanentAddress"]
    }
  }'
```


## Data Loading Mechanism

The PDP uses an explicit data loading mechanism to ensure JSON configuration files are properly loaded into the OPA policy engine:

### Implementation Details
- **Explicit JSON Loading**: JSON files are read and parsed using `loadJSONFile()` function
- **Embedded Data Module**: Data is embedded directly into Rego modules using `rego.Module()`
- **Direct Data Access**: Policies access data as `consumer_grants` and `provider_metadata` variables
- **No File System Dependencies**: Data is loaded at startup and embedded in the policy evaluation context

### Data Files
- `data/consumer-grants.json` - Consumer authorization data
- `data/provider-metadata.json` - Field metadata and consent requirements

### Loading Process
1. JSON files are read from the filesystem
2. Data is parsed and validated
3. Data is embedded as Rego module variables
4. OPA policy engine can access data during evaluation

## Recent Updates and Fixes

### Policy Logic Improvements
- **Fixed non-consent-required field handling**: PDP now correctly allows access to fields that don't require consent
- **Improved consent field analysis**: Better handling of mixed consent/non-consent field requests
- **Enhanced error handling**: More robust handling of edge cases in policy evaluation

### Test Organization
- **Centralized test scripts**: All test scripts moved to `/tests` directory for better organization
- **Comprehensive test coverage**: Added complete consent flow integration tests
- **Automated test runner**: `run-all-tests.sh` executes all test suites in sequence

### Container Deployment
- **Updated Docker containers**: Rebuilt containers with latest policy fixes
- **Improved build process**: Better separation of build and runtime stages
- **Enhanced security**: Non-root user execution in containers

## Security Considerations

- All requests are logged with decision outcomes
- Input validation prevents malformed requests
- Default deny policy ensures secure-by-default behavior
- Consent requirements are enforced at the field level
- Consumer authorization is verified against grants data
- Resource access is validated against metadata
- Data loading is performed at startup with validation
