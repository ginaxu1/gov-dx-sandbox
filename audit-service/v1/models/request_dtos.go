package models

import (
	"encoding/json"
)

// CreateAuditLogRequest represents the request payload for creating a generalized audit log
// This matches the final SQL schema with unified actor/target approach
type CreateAuditLogRequest struct {
	// Trace & Correlation
	TraceID *string `json:"traceId,omitempty"` // UUID string, nullable for standalone events

	// Temporal
	Timestamp string `json:"timestamp" validate:"required"` // ISO 8601 format, required

	// Event Classification
	EventType   *string `json:"eventType,omitempty"`        // POLICY_CHECK, MANAGEMENT_EVENT (user-defined custom names)
	EventAction *string `json:"eventAction,omitempty"`      // CREATE, READ, UPDATE, DELETE
	Status      string  `json:"status" validate:"required"` // SUCCESS, FAILURE

	// Actor Information (unified approach)
	ActorType string `json:"actorType" validate:"required"` // SERVICE, ADMIN, MEMBER, SYSTEM
	ActorID   string `json:"actorId" validate:"required"`   // email, uuid, or service-name (required)

	// Target Information (unified approach)
	TargetType string  `json:"targetType" validate:"required"` // SERVICE, RESOURCE
	TargetID   *string `json:"targetId,omitempty"`             // resource_id or service_name

	// Metadata (Payload without PII/sensitive data)
	RequestMetadata    json.RawMessage `json:"requestMetadata,omitempty"`    // Request payload without PII/sensitive data
	ResponseMetadata   json.RawMessage `json:"responseMetadata,omitempty"`   // Response or Error details
	AdditionalMetadata json.RawMessage `json:"additionalMetadata,omitempty"` // Additional context-specific data
}
