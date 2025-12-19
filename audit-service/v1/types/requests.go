package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

// CreateAuditLogRequest represents the request payload for creating an audit log
type CreateAuditLogRequest struct {
	// Trace & Correlation
	TraceID *string `json:"traceId,omitempty"` // UUID string, nullable for standalone events

	// Temporal
	Timestamp *string `json:"timestamp,omitempty"` // ISO 8601 format, optional (defaults to now)

	// Event Classification
	EventName string  `json:"eventName" validate:"required"` // POLICY_CHECK, CONSENT_CHECK, DATA_FETCH, MANAGEMENT_EVENT
	EventType *string `json:"eventType,omitempty"`           // CREATE, READ, UPDATE, DELETE (nullable for non-CRUD)
	Status    string  `json:"status" validate:"required"`    // SUCCESS or FAILURE

	// Actor
	ActorType        string          `json:"actorType" validate:"required"` // USER or SERVICE
	ActorServiceName *string         `json:"actorServiceName,omitempty"`    // Required for SERVICE, NULL for USER
	ActorUserID      *string         `json:"actorUserId,omitempty"`         // Required for USER, NULL for SERVICE (UUID string)
	ActorUserType    *string         `json:"actorUserType,omitempty"`       // ADMIN or MEMBER (for USER)
	ActorMetadata    json.RawMessage `json:"actorMetadata,omitempty"`       // Additional actor context

	// Target
	TargetType        string          `json:"targetType" validate:"required"` // RESOURCE or SERVICE
	TargetServiceName *string         `json:"targetServiceName,omitempty"`    // Required for SERVICE, NULL for RESOURCE
	TargetResource    *string         `json:"targetResource,omitempty"`       // Required for RESOURCE, NULL for SERVICE
	TargetResourceID  *string         `json:"targetResourceId,omitempty"`     // Optional UUID string
	TargetMetadata    json.RawMessage `json:"targetMetadata,omitempty"`       // Additional target context

	// Request/Response
	RequestedData    json.RawMessage `json:"requestedData,omitempty"`    // Request payload without PIA
	ResponseMetadata json.RawMessage `json:"responseMetadata,omitempty"` // Response or error without PIA

	// Additional Context
	EventMetadata json.RawMessage `json:"eventMetadata,omitempty"` // Additional event-specific metadata
}

// GetAuditLogsRequest represents query parameters for retrieving audit logs
type GetAuditLogsRequest struct {
	TraceID           *uuid.UUID `json:"traceId,omitempty"`           // Filter by trace ID
	EventName         *string    `json:"eventName,omitempty"`         // Filter by event name
	Status            *string    `json:"status,omitempty"`            // Filter by status
	ActorServiceName  *string    `json:"actorServiceName,omitempty"`  // Filter by actor service
	ActorUserID       *uuid.UUID `json:"actorUserId,omitempty"`       // Filter by user ID
	TargetServiceName *string    `json:"targetServiceName,omitempty"` // Filter by target service
	TargetResource    *string    `json:"targetResource,omitempty"`    // Filter by target resource
	StartTime         *string    `json:"startTime,omitempty"`         // Start timestamp (ISO 8601)
	EndTime           *string    `json:"endTime,omitempty"`           // End timestamp (ISO 8601)
	Limit             *int       `json:"limit,omitempty"`             // Max results (default: 100, max: 1000)
	Offset            *int       `json:"offset,omitempty"`            // Pagination offset (default: 0)
}
