package audit

// InitializeGlobalAudit initializes the global audit middleware instance.
// This should be called once during application startup.
// Subsequent calls will be ignored (safe to call multiple times).
//
// The client parameter should be an implementation of Auditor interface.
// When client is nil or IsEnabled() returns false, audit logging will be skipped
// but services will continue to function normally.
func InitializeGlobalAudit(client Auditor) {
	globalAuditOnce.Do(func() {
		globalAuditMiddleware = &AuditMiddleware{client: client}
	})
}
