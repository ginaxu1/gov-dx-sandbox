# Audit Service Enum Configuration

This directory contains the YAML configuration file that defines the allowed enum values for audit log fields.

## Configuration File

**File:** `enums.yaml`

This file defines the allowed values for:
- **Event Types**: User-defined custom names for event classification (e.g., POLICY_CHECK, MANAGEMENT_EVENT)
- **Event Actions**: CRUD operations (e.g., CREATE, READ, UPDATE, DELETE)
- **Actor Types**: Types of actors that can perform actions (SERVICE, ADMIN, MEMBER, SYSTEM)
- **Target Types**: Types of targets that actions can be performed on (SERVICE, RESOURCE)

## Usage

The configuration is automatically loaded at service startup. The service will:
1. Look for the config file at `config/enums.yaml` (default)
2. If not found, use default values defined in `config/config.go`
3. If the file exists but is invalid, log a warning and use defaults

### Custom Configuration Path

You can specify a custom path using the `AUDIT_ENUMS_CONFIG` environment variable:

```bash
AUDIT_ENUMS_CONFIG=/path/to/custom/enums.yaml ./audit-service
```

## Adding New Enum Values

To add new enum values, simply edit `enums.yaml` and add them to the appropriate list:

```yaml
eventTypes:
  - POLICY_CHECK
  - MANAGEMENT_EVENT
  - YOUR_NEW_EVENT_TYPE  # Add here
```

The service will automatically pick up the new values on the next restart.

## Validation

The audit service validates all enum fields against the configured values:
- Invalid values will be rejected during validation
- Empty/nullable fields (like `event_type` and `event_action`) are allowed
- Required fields (like `actor_type` and `target_type`) must match configured values
