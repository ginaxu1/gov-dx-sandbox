package types

import (
	"github.com/gov-dx-sandbox/audit-service/v1/models"
)

// AuditLogResponse represents a single audit log entry in API responses
type AuditLogResponse struct {
	models.AuditLog
}

// GetAuditLogsResponse represents the response for querying audit logs
type GetAuditLogsResponse struct {
	Total  int64              `json:"total"`  // Total number of matching records
	Limit  int                `json:"limit"`  // Max results per page
	Offset int                `json:"offset"` // Current offset
	Events []AuditLogResponse `json:"events"` // Audit log entries
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`             // Error type
	Message string `json:"message,omitempty"` // Error message
}
