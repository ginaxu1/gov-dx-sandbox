package types

import (
	"encoding/json"
)

// CreateAuditLogRequest represents the request payload for creating a generalized audit log
// This is flexible enough to support both opendif-core and opensuperapp use cases
type CreateAuditLogRequest struct {
	// Trace & Correlation (for opendif-core distributed tracing)
	TraceID *string `json:"traceId,omitempty"` // UUID string, nullable for standalone events

	// Temporal
	Timestamp *string `json:"timestamp,omitempty"` // ISO 8601 format, optional (defaults to now)

	// Event Classification
	EventName string  `json:"eventName" validate:"required"` // Generic event name/type
	Action    *string `json:"action,omitempty"`              // For opensuperapp: CREATE, UPDATE, DELETE, ACCESS, etc.
	Status    string  `json:"status" validate:"required"`    // SUCCESS, FAILURE, PARTIAL_FAILURE

	// Actor Information
	ActorType string  `json:"actorType" validate:"required"` // USER, SERVICE, SYSTEM
	ActorID   *string `json:"actorId,omitempty"`             // email, client_id, user_id, etc.

	// Resource Information (for opensuperapp)
	ResourceType *string `json:"resourceType,omitempty"` // USER, MICROAPP, FILE, etc.
	ResourceID   *string `json:"resourceId,omitempty"`   // UUID string

	// Service Information (for opendif-core)
	SourceService *string `json:"sourceService,omitempty"` // Service reporting the event
	TargetService *string `json:"targetService,omitempty"` // Target service

	// Request Context (for opensuperapp)
	IPAddress  *string `json:"ipAddress,omitempty"`
	UserAgent  *string `json:"userAgent,omitempty"`
	RequestID  *string `json:"requestId,omitempty"`
	SessionID  *string `json:"sessionId,omitempty"`
	MicroappID *string `json:"microappId,omitempty"`
	Platform   *string `json:"platform,omitempty"` // WEB, ANDROID, IOS, MICROAPP_SERVICE

	// Change Tracking (for opensuperapp)
	Changes json.RawMessage `json:"changes,omitempty"` // {"field": {"from": "old", "to": "new"}}

	// Flexible Metadata (project-specific data)
	// For opendif-core: requestedData, responseMetadata, eventMetadata
	// For opensuperapp: additional context-specific data
	Metadata json.RawMessage `json:"metadata,omitempty"`

	// Error Information
	ErrorMessage *string `json:"errorMessage,omitempty"`
	ErrorCode    *string `json:"errorCode,omitempty"`

	// Security & Authorization (optional)
	AuthorizationGroups  json.RawMessage `json:"authorizationGroups,omitempty"`  // Array of group names
	AuthenticationMethod *string         `json:"authenticationMethod,omitempty"` // USER_TOKEN, SERVICE_TOKEN
}
