package middleware

import (
	"context"

	auditpkg "github.com/gov-dx-sandbox/shared/audit"
)

// traceIDKey is the context key for trace ID
type traceIDKey struct{}

// Re-export types and functions from shared/audit for convenience
type (
	AuditClient     = auditpkg.AuditClient
	AuditLogRequest = auditpkg.AuditLogRequest
)

var (
	NewAuditMiddleware         = auditpkg.NewAuditMiddleware
	LogAuditEvent              = auditpkg.LogAuditEvent
	ResetGlobalAuditMiddleware = auditpkg.ResetGlobalAuditMiddleware
)

// Audit log status constants (re-exported from shared/audit)
const (
	AuditStatusSuccess = auditpkg.StatusSuccess
	AuditStatusFailure = auditpkg.StatusFailure
)

// GetTraceIDFromContext retrieves the trace ID from the context
// Returns empty string if trace ID is not found in context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}
