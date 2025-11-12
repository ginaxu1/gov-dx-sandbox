# Authentication and Authorization System

This document describes the comprehensive authentication and authorization system implemented for the API server.

## Overview

The system provides JWT-based authentication with Asgardeo and role-based access control (RBAC) for all API endpoints. It consists of:

1. **JWT Authentication Middleware** - Validates JWT tokens from Asgardeo
2. **Authorization Middleware** - Enforces role-based access control
3. **Role and Permission System** - Defines user capabilities
4. **Resource Ownership Validation** - Ensures users can only access their own resources

## Architecture

### Authentication Flow
```
Request → JWT Auth Middleware → Authorization Middleware → Handler
     ↓                      ↓                         ↓
Extract & Validate JWT → Check Permissions → Access Resource
```

### Components

#### 1. JWT Authentication (`/v1/middleware/jwt_auth.go`)
- Extracts Bearer tokens from Authorization headers
- Validates JWT signatures using JWKS from Asgardeo
- Verifies standard JWT claims (exp, iss, aud, etc.)
- Creates authenticated user context

#### 2. Authorization (`/v1/middleware/authorization.go`)
- Checks user permissions against endpoint requirements
- Enforces role-based access control
- Provides helper methods for different authorization patterns

#### 3. Models (`/v1/models/`)
- `authorization.go` - Roles, permissions, and endpoint mappings
- `authentication.go` - User claims and authentication context

#### 4. Utilities (`/v1/utils/auth_utils.go`)
- Token extraction and validation helpers
- Context management for authenticated users
- Resource ownership checking utilities

## Roles and Permissions

### User Roles

| Role | Description | Access Level |
|------|-------------|--------------|
| `admin` | System administrators | Full access to all resources |
| `member` | Regular users | Access to own resources only |
| `system` | Internal services | Read-only access for system operations |

### Permission Model

Permissions follow the pattern: `resource:action[:scope]`

Examples:
- `schema:create` - Create schemas
- `schema:read:all` - Read all schemas
- `member:update` - Update member information
- `application_submission:approve` - Approve applications

### Endpoint Protection

Each endpoint is mapped to required permissions:

```go
// Example endpoint permission mapping
{"GET", "/api/v1/schemas", PermissionReadSchema, false},
{"POST", "/api/v1/schemas", PermissionCreateSchema, false},
{"GET", "/api/v1/schemas/*", PermissionReadSchema, true}, // Requires ownership
```

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# Asgardeo JWT Authentication Configuration
ASGARDEO_BASE_URL=https://api.asgardeo.io/t/{organization}
ASGARDEO_CLIENT_ID=your_client_id
ASGARDEO_JWKS_URL=https://api.asgardeo.io/t/{organization}/oauth2/jwks
ASGARDEO_ORG_NAME=your_organization_name
```

### JWT Token Requirements

The JWT token must include these claims:

```json
{
  "sub": "user_id_from_idp",
  "email": "user@example.com",
  "given_name": "John",
  "family_name": "Doe",
  "roles": ["member"],
  "org_name": "your_organization",
  "iss": "https://api.asgardeo.io/t/{organization}/oauth2/token",
  "aud": ["your_client_id"],
  "exp": 1234567890,
  "iat": 1234567890
}
```

## Usage Examples

### Making Authenticated Requests

Include the JWT token in the Authorization header:

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -X GET https://api.example.com/api/v1/members
```

### Handler Implementation

Handlers automatically receive authenticated user context:

```go
func (h *Handler) handleProtectedResource(w http.ResponseWriter, r *http.Request) {
    // Get authenticated user from context
    user, err := middleware.GetUserFromRequest(r)
    if err != nil {
        utils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
        return
    }
    
    // Check specific permissions
    if !user.HasPermission(models.PermissionReadSchema) {
        utils.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
        return
    }
    
    // Check resource ownership
    if !user.IsAdmin() && resource.OwnerID != user.IdpUserID {
        utils.RespondWithError(w, http.StatusForbidden, "Access denied")
        return
    }
    
    // Process request
    // ...
}
```

## Resource Ownership

The system implements resource-level authorization:

1. **Admin users** can access all resources
2. **System users** have read-only access to most resources
3. **Regular members** can only access resources they own

Ownership is determined by comparing the authenticated user's `IdpUserID` with the resource's owner ID.

## Security Features

### Token Validation
- JWT signature verification using RSA public keys
- Expiration time validation
- Issuer and audience verification
- Organization name validation

### Key Management
- Automatic JWKS key fetching and caching
- Key rotation support with automatic refresh
- 1-hour cache TTL for security keys

### Protection Against
- Token tampering (signature validation)
- Token replay attacks (expiration validation)
- Cross-organization access (org_name validation)
- Unauthorized resource access (ownership validation)

## Error Handling

The system provides clear HTTP status codes:

- `401 Unauthorized` - Missing, invalid, or expired token
- `403 Forbidden` - Valid token but insufficient permissions
- `404 Not Found` - Resource doesn't exist or user can't access it

## Testing

### Unit Tests
Test authentication and authorization components:

```bash
cd api-server-go
go test ./v1/middleware/...
go test ./v1/utils/...
```

### Integration Tests
Use valid JWT tokens for testing protected endpoints:

```bash
# Set environment variables
export TEST_JWT_TOKEN="your_test_token"

# Run integration tests
go test -tags=integration ./...
```

## Best Practices

### For Developers

1. **Always use middleware**: Don't bypass authentication in handlers
2. **Check ownership**: Verify resource ownership for sensitive operations
3. **Principle of least privilege**: Grant minimum required permissions
4. **Log security events**: Log authentication failures and permission denials

### For Administrators

1. **Regular token rotation**: Implement token refresh mechanisms
2. **Monitor failed attempts**: Set up alerts for authentication failures
3. **Audit permissions**: Regularly review user roles and permissions
4. **Secure key storage**: Protect JWKS URLs and ensure HTTPS

## Troubleshooting

### Common Issues

1. **"Invalid authorization header"**
   - Ensure header format: `Authorization: Bearer <token>`
   - Check for extra spaces or missing "Bearer" prefix

2. **"Token validation failed"**
   - Verify token hasn't expired (`exp` claim)
   - Check issuer URL matches configuration
   - Ensure JWKS URL is accessible

3. **"Insufficient permissions"**
   - Verify user has required role
   - Check if endpoint requires specific permissions
   - Confirm user is assigned proper roles in Asgardeo

4. **"Access denied to resource"**
   - Verify resource ownership
   - Check if user has admin privileges for cross-user access
   - Ensure resource exists and user has access

### Debug Mode

Enable detailed logging by setting log level:

```bash
export LOG_LEVEL=debug
```

This will log authentication attempts, permission checks, and authorization decisions.

## Maintenance

### Updating Permissions
1. Modify `RolePermissions` map in `authorization.go`
2. Update `EndpointPermissions` for new endpoints
3. Test with different user roles
4. Deploy with proper database migrations if needed

### Adding New Roles
1. Define new role in `authorization.go`
2. Add to `RolePermissions` mapping
3. Update permission checking logic if needed
4. Document role capabilities

### Monitoring
- Track authentication success/failure rates
- Monitor permission denial patterns
- Watch for unusual access patterns
- Set up alerts for security violations

## Migration Guide

If upgrading from a system without authentication:

1. **Phase 1**: Deploy authentication middleware in permissive mode
2. **Phase 2**: Gradually enable authorization checks
3. **Phase 3**: Remove legacy access patterns
4. **Phase 4**: Enable full enforcement

This ensures zero-downtime migration while maintaining security.