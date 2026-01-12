package audit

import (
	"context"
	"log/slog"
	"sync"
)

// AuditMiddleware handles audit logging operations
type AuditMiddleware struct {
	client Auditor
}

// Global audit middleware instance for easy access from handlers
var (
	globalAuditMiddleware *AuditMiddleware
	globalAuditOnce       sync.Once
)

// NewAuditMiddleware creates a new audit middleware with thread-safe global initialization
// This function should typically only be called once during application startup.
// Subsequent calls will return a new instance but won't update the global instance.
//
// The client parameter should be an implementation of Auditor interface.
// When client is nil or IsEnabled() returns false, the middleware will skip all audit logging operations
// but services will continue to function normally.
func NewAuditMiddleware(client Auditor) *AuditMiddleware {
	middleware := &AuditMiddleware{client: client}

	globalAuditOnce.Do(func() {
		globalAuditMiddleware = middleware
	})

	return middleware
}

// Client returns the audit client instance
// This allows service-specific wrappers to access the client
func (m *AuditMiddleware) Client() Auditor {
	return m.client
}

// LogAuditEvent sends an audit event to the audit service API
// This function is used to log audit events using the unified audit log structure
func (m *AuditMiddleware) LogAuditEvent(ctx context.Context, auditRequest *AuditLogRequest) {
	if m.client == nil {
		return
	}
	m.client.LogEvent(ctx, auditRequest)
}

// LogAuditEvent logs an audit event using global audit middleware instance
// This is the public function that should be called from handlers and other components
func LogAuditEvent(ctx context.Context, auditRequest *AuditLogRequest) {
	if globalAuditMiddleware != nil {
		globalAuditMiddleware.LogAuditEvent(ctx, auditRequest)
	} else {
		slog.Warn("Global AuditMiddleware is not initialized; audit event not logged")
	}
}

// GetGlobalAuditMiddleware returns the global audit middleware instance
// This can be used by service-specific wrappers that need access to the global instance
func GetGlobalAuditMiddleware() *AuditMiddleware {
	return globalAuditMiddleware
}

// ResetGlobalAuditMiddleware is a helper function for tests to reset the global audit middleware instance
// This should only be used in tests to reset state between test cases
func ResetGlobalAuditMiddleware() {
	globalAuditOnce = sync.Once{}
	globalAuditMiddleware = nil
}
