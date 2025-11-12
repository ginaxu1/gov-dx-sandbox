package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	sharedutils "github.com/gov-dx-sandbox/api-server-go/shared/utils"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	authutils "github.com/gov-dx-sandbox/api-server-go/v1/utils"
)

// AuthorizationMiddleware provides role-based access control functionality
type AuthorizationMiddleware struct {
	// Future: could add configuration for different authorization modes
}

// NewAuthorizationMiddleware creates a new authorization middleware
func NewAuthorizationMiddleware() *AuthorizationMiddleware {
	return &AuthorizationMiddleware{}
}

// AuthorizeRequest returns a middleware function that checks user permissions for endpoints
func (a *AuthorizationMiddleware) AuthorizeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authorization for endpoints that don't require authentication
		if a.shouldSkipAuthorization(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get authenticated user from context (should be set by JWT middleware)
		user, err := authutils.RequireAuthentication(r)
		if err != nil {
			slog.Warn("Authorization failed: user not authenticated", "path", r.URL.Path, "method", r.Method, "error", err)
			sharedutils.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Find the endpoint permission requirement
		endpointPermission, found := authutils.FindEndpointPermission(r.Method, r.URL.Path)
		if !found {
			// If no specific permission is defined, allow admin and system users, deny others
			if user.IsAdmin() || user.IsSystem() {
				slog.Info("Access granted to undefined endpoint", "user", user.Email, "role", user.GetPrimaryRole(), "path", r.URL.Path, "method", r.Method)
				next.ServeHTTP(w, r)
				return
			}

			slog.Warn("Access denied to undefined endpoint", "user", user.Email, "role", user.GetPrimaryRole(), "path", r.URL.Path, "method", r.Method)
			sharedutils.RespondWithError(w, http.StatusForbidden, "Access denied")
			return
		}

		// Check if user has the required permission
		if !user.HasPermission(endpointPermission.Permission) {
			slog.Warn("Access denied: insufficient permissions",
				"user", user.Email,
				"role", user.GetPrimaryRole(),
				"required_permission", endpointPermission.Permission,
				"path", r.URL.Path,
				"method", r.Method)
			sharedutils.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}

		// For endpoints that require ownership, we need to check resource ownership
		// This will be handled at the handler level since we need to extract resource IDs
		// For now, we just ensure the user has the permission

		slog.Info("Access granted",
			"user", user.Email,
			"role", user.GetPrimaryRole(),
			"permission", endpointPermission.Permission,
			"path", r.URL.Path,
			"method", r.Method)

		next.ServeHTTP(w, r)
	})
}

// RequireRole returns a middleware that requires a specific role
func (a *AuthorizationMiddleware) RequireRole(requiredRole models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authutils.RequireRole(r, requiredRole)
			if err != nil {
				slog.Warn("Role requirement not met",
					"required_role", requiredRole,
					"path", r.URL.Path,
					"method", r.Method,
					"error", err)
				sharedutils.RespondWithError(w, http.StatusForbidden, "Insufficient privileges")
				return
			}

			slog.Info("Role requirement satisfied",
				"user", user.Email,
				"required_role", requiredRole,
				"user_roles", user.Roles,
				"path", r.URL.Path,
				"method", r.Method)

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole returns a middleware that requires any of the specified roles
func (a *AuthorizationMiddleware) RequireAnyRole(requiredRoles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authutils.RequireAnyRole(r, requiredRoles...)
			if err != nil {
				roleNames := make([]string, len(requiredRoles))
				for i, role := range requiredRoles {
					roleNames[i] = role.String()
				}

				slog.Warn("Role requirement not met",
					"required_roles", strings.Join(roleNames, ", "),
					"path", r.URL.Path,
					"method", r.Method,
					"error", err)
				sharedutils.RespondWithError(w, http.StatusForbidden, "Insufficient privileges")
				return
			}

			slog.Info("Role requirement satisfied",
				"user", user.Email,
				"required_roles", requiredRoles,
				"user_roles", user.Roles,
				"path", r.URL.Path,
				"method", r.Method)

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission returns a middleware that requires a specific permission
func (a *AuthorizationMiddleware) RequirePermission(requiredPermission models.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authutils.RequirePermission(r, requiredPermission)
			if err != nil {
				slog.Warn("Permission requirement not met",
					"required_permission", requiredPermission,
					"path", r.URL.Path,
					"method", r.Method,
					"error", err)
				sharedutils.RespondWithError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			slog.Info("Permission requirement satisfied",
				"user", user.Email,
				"required_permission", requiredPermission,
				"user_permissions", user.GetPermissions(),
				"path", r.URL.Path,
				"method", r.Method)

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdminRole is a convenience middleware that requires admin role
func (a *AuthorizationMiddleware) RequireAdminRole() func(http.Handler) http.Handler {
	return a.RequireRole(models.RoleAdmin)
}

// RequireMemberRole is a convenience middleware that requires member role
func (a *AuthorizationMiddleware) RequireMemberRole() func(http.Handler) http.Handler {
	return a.RequireRole(models.RoleMember)
}

// RequireSystemRole is a convenience middleware that requires system role
func (a *AuthorizationMiddleware) RequireSystemRole() func(http.Handler) http.Handler {
	return a.RequireRole(models.RoleSystem)
}

// RequireAdminOrSystemRole requires either admin or system role
func (a *AuthorizationMiddleware) RequireAdminOrSystemRole() func(http.Handler) http.Handler {
	return a.RequireAnyRole(models.RoleAdmin, models.RoleSystem)
}

// CheckResourceOwnership is a helper function to be used in handlers to verify resource ownership
func (a *AuthorizationMiddleware) CheckResourceOwnership(user *models.AuthenticatedUser, resourceOwnerIdpUserId string, permission models.Permission) bool {
	return authutils.CanAccessResource(user, permission, resourceOwnerIdpUserId)
}

// shouldSkipAuthorization determines if authorization should be skipped for this path
func (a *AuthorizationMiddleware) shouldSkipAuthorization(path string) bool {
	skipPaths := []string{
		"/health",
		"/debug",
		"/openapi.yaml",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// GetUserFromRequest is a helper to extract the authenticated user from request context
func GetUserFromRequest(r *http.Request) (*models.AuthenticatedUser, error) {
	return authutils.GetAuthenticatedUser(r.Context())
}

// GetAuthContextFromRequest is a helper to extract the auth context from request context
func GetAuthContextFromRequest(r *http.Request) (*models.AuthContext, error) {
	return authutils.GetAuthContext(r.Context())
}
