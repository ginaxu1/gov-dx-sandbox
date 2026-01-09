package middleware

import (
	"context"
	"log/slog"
	"sync"

	auditclient "github.com/gov-dx-sandbox/shared/audit"
	tracemiddleware "github.com/gov-dx-sandbox/exchange/shared/middleware"
)

// Re-export trace ID functions from shared middleware for convenience
var (
	GetTraceIDFromContext    = tracemiddleware.GetTraceIDFromContext
	ExtractTraceIDFromRequest = tracemiddleware.ExtractTraceIDFromRequest
)

// AuditMiddleware handles audit logging
type AuditMiddleware struct {
	client *auditclient.Client
}

// Global audit middleware instance for easy access from handlers
var (
	globalAuditMiddleware *AuditMiddleware
	globalAuditOnce       sync.Once
)

// NewAuditMiddleware creates a new audit middleware with thread-safe global initialization
// Audit can be disabled by:
//   - Setting ENABLE_AUDIT=false environment variable
//   - Providing an empty auditServiceURL
//
// When disabled, the middleware will skip all audit logging operations but services
// will continue to function normally.
func NewAuditMiddleware(auditServiceURL string) *AuditMiddleware {
	client := auditclient.NewClient(auditServiceURL)
	middleware := &AuditMiddleware{client: client}

	globalAuditOnce.Do(func() {
		globalAuditMiddleware = middleware
	})

	return middleware
}

// LogGeneralizedAuditEvent logs a generalized audit event using global audit middleware instance
func LogGeneralizedAuditEvent(ctx context.Context, auditRequest *auditclient.AuditLogRequest) {
	if globalAuditMiddleware != nil {
		globalAuditMiddleware.client.LogEvent(ctx, auditRequest)
	} else {
		slog.Warn("Global AuditMiddleware is not initialized; audit event not logged")
	}
}
