package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	authutils "github.com/gov-dx-sandbox/api-server-go/v1/utils"
)

// Context keys for storing request information
type contextKey string

const (
	// Actor information
	contextKeyActorType contextKey = "actorType"
	contextKeyActorID   contextKey = "actorID"
	contextKeyActorRole contextKey = "actorRole"

	// Resource IDs extracted from path or body
	contextKeySchemaID                contextKey = "schemaId"
	contextKeyApplicationID           contextKey = "applicationId"
	contextKeyMemberID                contextKey = "memberId"
	contextKeySchemaSubmissionID      contextKey = "schemaSubmissionId"
	contextKeyApplicationSubmissionID contextKey = "applicationSubmissionId"

	// Target resource type (for audit logging)
	contextKeyTargetResource contextKey = "targetResource"
)

// RequestContextMiddleware extracts information from the request and sets it in context
// This middleware should run early in the chain, after authentication but before audit
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract actor information from authenticated user context (set by JWT middleware)
		// Fallback to headers if no authenticated user is found
		actorType, actorID, actorRole := extractActorInfo(r)
		if actorType != "" {
			ctx = context.WithValue(ctx, contextKeyActorType, actorType)
		}
		if actorID != nil {
			ctx = context.WithValue(ctx, contextKeyActorID, *actorID)
		}
		if actorRole != nil {
			ctx = context.WithValue(ctx, contextKeyActorRole, *actorRole)
		}

		// Extract resource IDs from path
		// This is much simpler than brute force - we know the route patterns
		if schemaID := extractSchemaIDFromPath(r.URL.Path); schemaID != "" {
			ctx = context.WithValue(ctx, contextKeySchemaID, schemaID)
		}
		if applicationID := extractApplicationIDFromPath(r.URL.Path); applicationID != "" {
			ctx = context.WithValue(ctx, contextKeyApplicationID, applicationID)
		}
		if memberID := extractMemberIDFromPath(r.URL.Path); memberID != "" {
			ctx = context.WithValue(ctx, contextKeyMemberID, memberID)
		}
		if submissionID := extractSchemaSubmissionIDFromPath(r.URL.Path); submissionID != "" {
			ctx = context.WithValue(ctx, contextKeySchemaSubmissionID, submissionID)
		}
		if submissionID := extractApplicationSubmissionIDFromPath(r.URL.Path); submissionID != "" {
			ctx = context.WithValue(ctx, contextKeyApplicationSubmissionID, submissionID)
		}

		// Determine target resource type from path
		if resourceType := determineResourceType(r.URL.Path, r.Method); resourceType != "" {
			ctx = context.WithValue(ctx, contextKeyTargetResource, resourceType)
		}

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractActorInfo extracts actor information from authenticated user context
// Falls back to headers if no authenticated user is found
func extractActorInfo(r *http.Request) (actorType string, actorID *string, actorRole *string) {
	// First, try to extract from authenticated user context (set by JWT middleware)
	user, err := authutils.GetAuthenticatedUser(r.Context())
	if err == nil && user != nil {
		// We have an authenticated user
		actorType = "USER"
		actorID = &user.IdpUserID

		// Map Role enum to audit service format
		// OpenDIF_Admin -> ADMIN, OpenDIF_Member -> MEMBER, OpenDIF_System -> SERVICE (no role)
		primaryRole := user.GetPrimaryRole()
		var roleStr string
		switch primaryRole {
		case models.RoleAdmin:
			roleStr = "ADMIN"
		case models.RoleMember:
			roleStr = "MEMBER"
		case models.RoleSystem:
			// System role is treated as SERVICE type, not USER
			actorType = "SERVICE"
			actorID = nil
			actorRole = nil
			return actorType, actorID, actorRole
		default:
			// Default to MEMBER for unknown roles
			roleStr = "MEMBER"
		}
		actorRole = &roleStr

		slog.Debug("Extracted actor info from authenticated user context",
			"actorType", actorType,
			"actorID", *actorID,
			"actorRole", *actorRole)

		return actorType, actorID, actorRole
	}

	// Fallback: Try to extract from headers (for backward compatibility or service-to-service calls)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.Header.Get("X-Auth-User-ID")
	}

	role := r.Header.Get("X-User-Role")
	if role == "" {
		role = r.Header.Get("X-Auth-Role")
	}

	// If we have user info from headers, use USER type, otherwise use SERVICE
	if userID != "" {
		actorType = "USER"
		actorID = &userID
		if role != "" && (role == "MEMBER" || role == "ADMIN") {
			actorRole = &role
		} else {
			// Default to MEMBER if role not specified
			defaultRole := "MEMBER"
			actorRole = &defaultRole
		}
		slog.Debug("Extracted actor info from headers (fallback)",
			"actorType", actorType,
			"actorID", *actorID,
			"actorRole", *actorRole)
	} else {
		actorType = "SERVICE"
		actorID = nil
		actorRole = nil
		slog.Debug("No actor info found, defaulting to SERVICE type")
	}

	return actorType, actorID, actorRole
}

// extractSchemaIDFromPath extracts schema ID from URL path
// Paths: /api/v1/schemas/{schemaId} or /api/v1/schemas/{schemaId}/...
func extractSchemaIDFromPath(path string) string {
	// Simple pattern matching - much more reliable than brute force
	// Pattern: /api/v1/schemas/{id}
	const prefix = "/api/v1/schemas/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		// Get the ID part (everything after the prefix until next /)
		id := path[len(prefix):]
		if idx := findFirstSlash(id); idx > 0 {
			return id[:idx]
		}
		return id
	}
	return ""
}

// extractApplicationIDFromPath extracts application ID from URL path
func extractApplicationIDFromPath(path string) string {
	const prefix = "/api/v1/applications/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		id := path[len(prefix):]
		if idx := findFirstSlash(id); idx > 0 {
			return id[:idx]
		}
		return id
	}
	return ""
}

// extractMemberIDFromPath extracts member ID from URL path
func extractMemberIDFromPath(path string) string {
	const prefix = "/api/v1/members/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		id := path[len(prefix):]
		if idx := findFirstSlash(id); idx > 0 {
			return id[:idx]
		}
		return id
	}
	return ""
}

// extractSchemaSubmissionIDFromPath extracts schema submission ID from URL path
func extractSchemaSubmissionIDFromPath(path string) string {
	const prefix = "/api/v1/schema-submissions/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		id := path[len(prefix):]
		if idx := findFirstSlash(id); idx > 0 {
			return id[:idx]
		}
		return id
	}
	return ""
}

// extractApplicationSubmissionIDFromPath extracts application submission ID from URL path
func extractApplicationSubmissionIDFromPath(path string) string {
	const prefix = "/api/v1/application-submissions/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		id := path[len(prefix):]
		if idx := findFirstSlash(id); idx > 0 {
			return id[:idx]
		}
		return id
	}
	return ""
}

// determineResourceType determines the resource type from the path and method
func determineResourceType(path, method string) string {
	// Only determine resource type for write operations
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
		return ""
	}

	// Simple pattern matching
	if contains(path, "/api/v1/schemas/") {
		return "SCHEMAS"
	}
	if contains(path, "/api/v1/applications/") {
		return "APPLICATIONS"
	}
	if contains(path, "/api/v1/members/") {
		return "MEMBERS"
	}
	if contains(path, "/api/v1/schema-submissions/") {
		return "SCHEMA-SUBMISSIONS"
	}
	if contains(path, "/api/v1/application-submissions/") {
		return "APPLICATION-SUBMISSIONS"
	}

	// Check collection endpoints (POST to create)
	if method == http.MethodPost {
		if path == "/api/v1/schemas" || path == "/api/v1/schemas/" {
			return "SCHEMAS"
		}
		if path == "/api/v1/applications" || path == "/api/v1/applications/" {
			return "APPLICATIONS"
		}
		if path == "/api/v1/members" || path == "/api/v1/members/" {
			return "MEMBERS"
		}
		if path == "/api/v1/schema-submissions" || path == "/api/v1/schema-submissions/" {
			return "SCHEMA-SUBMISSIONS"
		}
		if path == "/api/v1/application-submissions" || path == "/api/v1/application-submissions/" {
			return "APPLICATION-SUBMISSIONS"
		}
	}

	return ""
}

// Helper functions

func findFirstSlash(s string) int {
	for i, c := range s {
		if c == '/' {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Context getter functions (for use in handlers and middleware)

// GetActorType retrieves actor type from context
func GetActorType(ctx context.Context) string {
	if val := ctx.Value(contextKeyActorType); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetActorID retrieves actor ID from context
func GetActorID(ctx context.Context) *string {
	if val := ctx.Value(contextKeyActorID); val != nil {
		if s, ok := val.(string); ok {
			return &s
		}
	}
	return nil
}

// GetActorRole retrieves actor role from context
func GetActorRole(ctx context.Context) *string {
	if val := ctx.Value(contextKeyActorRole); val != nil {
		if s, ok := val.(string); ok {
			return &s
		}
	}
	return nil
}

// GetSchemaID retrieves schema ID from context
func GetSchemaID(ctx context.Context) string {
	if val := ctx.Value(contextKeySchemaID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetApplicationID retrieves application ID from context
func GetApplicationID(ctx context.Context) string {
	if val := ctx.Value(contextKeyApplicationID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetMemberID retrieves member ID from context
func GetMemberID(ctx context.Context) string {
	if val := ctx.Value(contextKeyMemberID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetTargetResource retrieves target resource type from context
func GetTargetResource(ctx context.Context) string {
	if val := ctx.Value(contextKeyTargetResource); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetResourceID sets a resource ID in context (for use in handlers after operations)
func SetResourceID(ctx context.Context, resourceType, resourceID string) context.Context {
	switch resourceType {
	case "SCHEMAS":
		return context.WithValue(ctx, contextKeySchemaID, resourceID)
	case "APPLICATIONS":
		return context.WithValue(ctx, contextKeyApplicationID, resourceID)
	case "MEMBERS":
		return context.WithValue(ctx, contextKeyMemberID, resourceID)
	case "SCHEMA-SUBMISSIONS":
		return context.WithValue(ctx, contextKeySchemaSubmissionID, resourceID)
	case "APPLICATION-SUBMISSIONS":
		return context.WithValue(ctx, contextKeyApplicationSubmissionID, resourceID)
	}
	return ctx
}
