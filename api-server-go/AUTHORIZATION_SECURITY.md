# Authorization Security Configuration

This API server implements configurable authorization behavior for undefined endpoints (endpoints without explicit permission mappings). This addresses the security concern of "fail open" vs "fail closed" behavior when new endpoints are added.

## Security Modes

### 1. `fail_closed` (Most Secure - Recommended for Production)
- **Behavior**: Denies all access to undefined endpoints, regardless of user role
- **Use Case**: Maximum security environments where all endpoints must be explicitly defined
- **Security Level**: ⭐⭐⭐⭐⭐ (Highest)

```bash
AUTHORIZATION_MODE=fail_closed
```

**Advantages:**
- Prevents accidental exposure of new endpoints
- Forces explicit security review for all endpoints
- Zero risk of privilege escalation through undefined endpoints

**Considerations:**
- Requires all endpoints to have explicit permission mappings
- May require more maintenance as new endpoints are added

### 2. `fail_open_admin` (Secure)
- **Behavior**: Allows only admin users to access undefined endpoints
- **Use Case**: Organizations where admin users need broad access for system management
- **Security Level**: ⭐⭐⭐⭐ (High)

```bash
AUTHORIZATION_MODE=fail_open_admin
```

**Advantages:**
- Balances security with administrative flexibility
- Prevents regular users from accessing undefined endpoints
- Suitable for most production environments

### 3. `fail_open_admin_system` (Legacy - Default)
- **Behavior**: Allows admin and system users to access undefined endpoints
- **Use Case**: Backward compatibility and environments with system-to-system communication
- **Security Level**: ⭐⭐⭐ (Medium)

```bash
AUTHORIZATION_MODE=fail_open_admin_system
```

**Advantages:**
- Maintains backward compatibility
- Allows system services to function without explicit mappings
- Good for development and staging environments

## Strict Mode

Enable strict mode to log warnings about undefined endpoint access. This helps identify endpoints that should have explicit permission mappings.

```bash
AUTHORIZATION_STRICT_MODE=true
```

**When enabled, logs like this will be generated:**
```
WARN SECURITY: Undefined endpoint accessed - consider adding explicit permission mapping
user=admin@example.com role=OpenDIF_Admin path=/api/v1/new-endpoint method=GET mode=fail_open_admin
```

## Configuration Examples

### Production (Maximum Security)
```bash
AUTHORIZATION_MODE=fail_closed
AUTHORIZATION_STRICT_MODE=true
```

### Production (Balanced Security)
```bash
AUTHORIZATION_MODE=fail_open_admin
AUTHORIZATION_STRICT_MODE=true
```

### Development/Staging
```bash
AUTHORIZATION_MODE=fail_open_admin_system
AUTHORIZATION_STRICT_MODE=false
```

## Adding New Endpoints

When adding new endpoints, always add explicit permission mappings to `v1/models/authorization.go`:

```go
// Add to EndpointPermissions array
{"GET", "/api/v1/new-endpoint", PermissionReadNewResource, false},
{"POST", "/api/v1/new-endpoint", PermissionCreateNewResource, false},
```

This ensures:
1. Explicit security review of each endpoint
2. Proper permission-to-role mapping
3. Consistent authorization behavior
4. No reliance on fallback authorization modes

## Security Best Practices

1. **Use `fail_closed` in production** for maximum security
2. **Enable strict mode** to identify undefined endpoints
3. **Add explicit permission mappings** for all endpoints
4. **Regular security audits** of endpoint permissions
5. **Monitor logs** for undefined endpoint access attempts

## Monitoring and Alerting

Set up alerts for:
- Undefined endpoint access warnings (when strict mode is enabled)
- Failed authorization attempts
- Privilege escalation patterns

Example log queries:
```
# Undefined endpoint access
level=WARN msg="Access denied to undefined endpoint"

# Failed authorization 
level=WARN msg="Access denied: insufficient permissions"

# Strict mode warnings
level=WARN msg="SECURITY: Undefined endpoint accessed"
```