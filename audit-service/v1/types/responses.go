package types

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogResponse represents the response payload for an audit log entry
type AuditLogResponse struct {
	ID        uuid.UUID  `json:"id"`
	Timestamp time.Time  `json:"timestamp"`
	TraceID   *uuid.UUID `json:"traceId,omitempty"`

	EventName string  `json:"eventName"`
	Action    *string `json:"action,omitempty"`
	Status    string  `json:"status"`

	ActorType string  `json:"actorType"`
	ActorID   *string `json:"actorId,omitempty"`

	ResourceType *string    `json:"resourceType,omitempty"`
	ResourceID   *uuid.UUID `json:"resourceId,omitempty"`

	SourceService *string `json:"sourceService,omitempty"`
	TargetService *string `json:"targetService,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

// GetAuditLogsResponse represents the response for querying audit logs
type GetAuditLogsResponse struct {
	Logs  []AuditLogResponse `json:"logs"`
	Total int                `json:"total"`
}
