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

	// LogManagementEvent logs a management event (Case 2) from API Server
	// This is called asynchronously (fire-and-forget) to avoid blocking
	LogManagementEvent(ctx context.Context, event ManagementEvent) error
}

// NewAuditClient creates a new audit client
// If auditServiceURL is empty, returns a no-op client that does nothing
// This allows services to work even when audit service is not configured
func NewAuditClient(auditServiceURL string) AuditClient {
	if auditServiceURL == "" {
		slog.Info("Audit service URL not provided, using no-op client")
		return &noOpClient{}
	}
	return &httpClient{
		baseURL:    auditServiceURL,
		httpClient: newHTTPClient(),
	}
}
