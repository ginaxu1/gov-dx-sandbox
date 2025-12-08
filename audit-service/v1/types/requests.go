package types

import (
	"encoding/json"
)

// CreateAuditLogRequest represents the request payload for creating a generalized audit log
// This matches the audit-service v1 API structure
type CreateAuditLogRequest struct {
	// Trace & Correlation
	TraceID *string `json:"traceId,omitempty"` // UUID string, nullable for standalone events

	// Temporal
	Timestamp *string `json:"timestamp,omitempty"` // ISO 8601 format, optional (defaults to now)

	// Event Classification
	EventName string  `json:"eventName" validate:"required"` // POLICY_CHECK, CONSENT_CHECK, DATA_FETCH, MANAGEMENT_EVENT
	EventType *string `json:"eventType,omitempty"`           // CREATE, READ, UPDATE, DELETE (nullable for non-CRUD)
	Status    string  `json:"status" validate:"required"`    // SUCCESS or FAILURE

	// Actor (Flattened from ActorMetadata)
	ActorType        string          `json:"actorType" validate:"required"` // USER or SERVICE
	ActorServiceName *string         `json:"actorServiceName,omitempty"`    // Required for SERVICE, NULL for USER
	ActorUserID      *string         `json:"actorUserId,omitempty"`         // Required for USER, NULL for SERVICE (UUID string)
	ActorUserType    *string         `json:"actorUserType,omitempty"`       // ADMIN or MEMBER (for USER)
	ActorMetadata    json.RawMessage `json:"actorMetadata,omitempty"`       // Additional actor context

	// Target (Flattened from TargetMetadata)
	TargetType        string          `json:"targetType" validate:"required"` // RESOURCE or SERVICE
	TargetServiceName *string         `json:"targetServiceName,omitempty"`    // Required for SERVICE, NULL for RESOURCE
	TargetResource    *string         `json:"targetResource,omitempty"`       // Required for RESOURCE, NULL for SERVICE
	TargetResourceID  *string         `json:"targetResourceId,omitempty"`     // Optional UUID string
	TargetMetadata    json.RawMessage `json:"targetMetadata,omitempty"`       // Additional target context

	// Request/Response (PIA-free)
	RequestedData    json.RawMessage `json:"requestedData,omitempty"`    // Request payload
	ResponseMetadata json.RawMessage `json:"responseMetadata,omitempty"` // Response or error

	// Additional Context
	EventMetadata json.RawMessage `json:"eventMetadata,omitempty"` // Additional event-specific metadata
}
