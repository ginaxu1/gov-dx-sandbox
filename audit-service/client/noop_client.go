package client

import (
	"context"
)

// noOpClient implements AuditClient but does nothing
// Used when audit service is not configured or disabled
type noOpClient struct{}

// LogDataExchange is a no-op implementation
func (c *noOpClient) LogDataExchange(ctx context.Context, event DataExchangeEvent) error {
	// Do nothing - audit logging is disabled
	return nil
}

// LogManagementEvent is a no-op implementation
func (c *noOpClient) LogManagementEvent(ctx context.Context, event ManagementEvent) error {
	// Do nothing - audit logging is disabled
	return nil
}

