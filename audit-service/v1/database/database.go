package database

import (
	"context"

	"github.com/gov-dx-sandbox/audit-service/v1/models"
)

// AuditRepository defines the database-agnostic interface for audit log operations
// This allows the service to work with any database implementation (PostgreSQL, MongoDB, etc.)
type AuditRepository interface {
	// CreateAuditLog creates a new audit log entry
	CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)

	// GetAuditLogsByTraceID retrieves all audit logs for a given trace ID
	GetAuditLogsByTraceID(ctx context.Context, traceID string) ([]models.AuditLog, error)

	// GetAuditLogs retrieves audit logs with optional filtering
	GetAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]models.AuditLog, int64, error)
}

// AuditLogFilters represents query filters for retrieving audit logs
type AuditLogFilters struct {
	TraceID     *string
	EventType   *string
	EventAction *string
	Status      *string
	Limit       int
	Offset      int
}
