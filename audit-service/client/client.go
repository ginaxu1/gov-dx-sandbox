package client

import (
	"context"
	"log/slog"
)

// AuditClient provides methods for logging audit events
// This interface allows for easy mocking in tests and different implementations
type AuditClient interface {
	// LogDataExchange logs a data exchange event (Case 1) from Orchestration Engine
	// This is called asynchronously (fire-and-forget) to avoid blocking
	LogDataExchange(ctx context.Context, event DataExchangeEvent) error
}

// NewAuditClient creates a new audit client
// If auditServiceURL is empty, returns nil
// Callers should check for nil before using the client
func NewAuditClient(auditServiceURL string) AuditClient {
	if auditServiceURL == "" {
		slog.Info("Audit service URL not provided, audit client will be nil")
		return nil
	}
	return &httpClient{
		baseURL:    auditServiceURL,
		httpClient: newHTTPClient(),
	}
}
