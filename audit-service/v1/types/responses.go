package types

import (
	"time"

	"github.com/google/uuid"
)

// AuditLogResponse represents the response payload for an audit log entry
type AuditLogResponse struct {
	ID uuid.UUID `json:"id"`

	Timestamp time.Time  `json:"timestamp"`
	TraceID   *uuid.UUID `json:"traceId,omitempty"`

	EventName string  `json:"eventName"`
	EventType *string `json:"eventType,omitempty"`
	Status    string  `json:"status"`

	ActorType        string     `json:"actorType"`
	ActorServiceName *string    `json:"actorServiceName,omitempty"`
	ActorUserID      *uuid.UUID `json:"actorUserId,omitempty"`
	ActorUserType    *string    `json:"actorUserType,omitempty"`

	TargetType        string     `json:"targetType"`
	TargetServiceName *string    `json:"targetServiceName,omitempty"`
	TargetResource    *string    `json:"targetResource,omitempty"`
	TargetResourceID  *uuid.UUID `json:"targetResourceId,omitempty"`
}

// GetAuditLogsResponse represents the response for querying audit logs
type GetAuditLogsResponse struct {
	Logs  []AuditLogResponse `json:"logs"`
	Total int                `json:"total"`
}
