package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Audit log status constants
const (
	StatusSuccess        = "SUCCESS"
	StatusFailure        = "FAILURE"
	StatusPartialFailure = "PARTIAL_FAILURE"
)

// Actor type constants
const (
	ActorTypeUser    = "USER"
	ActorTypeService = "SERVICE"
	ActorTypeSystem  = "SYSTEM"
)

// AuditLog represents a generalized audit log entry that can be used across different projects
// This model is flexible enough to support both opendif-core and opensuperapp use cases
type AuditLog struct {
	// Primary Key
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Temporal
	Timestamp time.Time `gorm:"type:timestamp with time zone;not null;index:idx_audit_logs_timestamp" json:"timestamp"`

	// Trace & Correlation (for distributed tracing - opendif-core)
	TraceID *uuid.UUID `gorm:"type:uuid;index:idx_audit_logs_trace_id,where:trace_id IS NOT NULL" json:"traceId,omitempty"` // NULL for standalone events

	// Event Classification
	// For opendif-core: eventName (POLICY_CHECK, CONSENT_CHECK, DATA_FETCH, MANAGEMENT_EVENT)
	// For opensuperapp: eventType (USER_MANAGEMENT, MICROAPP_MANAGEMENT, AUTHENTICATION, etc.)
	EventName string `gorm:"type:varchar(100);not null;index:idx_audit_logs_event_name" json:"eventName"` // Generic event name/type

	// For opensuperapp: action (CREATE, UPDATE, DELETE, ACCESS, LOGIN, etc.)
	// For opendif-core: eventType (CREATE, READ, UPDATE, DELETE) - stored in EventMetadata if needed
	Action *string `gorm:"type:varchar(50);index:idx_audit_logs_action,where:action IS NOT NULL" json:"action,omitempty"`

	// Status
	Status string `gorm:"type:varchar(20);not null;check:status IN ('SUCCESS', 'FAILURE', 'PARTIAL_FAILURE');index:idx_audit_logs_status" json:"status"`

	// Actor Information
	ActorType string  `gorm:"type:varchar(20);not null;check:actor_type IN ('USER', 'SERVICE', 'SYSTEM')" json:"actorType"`
	ActorID   *string `gorm:"type:varchar(255);index:idx_audit_logs_actor_id,where:actor_id IS NOT NULL" json:"actorId,omitempty"` // email, client_id, user_id, etc.

	// Resource Information (for opensuperapp)
	ResourceType *string    `gorm:"type:varchar(50);index:idx_audit_logs_resource_type,where:resource_type IS NOT NULL" json:"resourceType,omitempty"` // USER, MICROAPP, FILE, etc.
	ResourceID   *uuid.UUID `gorm:"type:uuid;index:idx_audit_logs_resource_id,where:resource_id IS NOT NULL" json:"resourceId,omitempty"`

	// Service Information (for opendif-core)
	SourceService *string `gorm:"type:varchar(100);index:idx_audit_logs_source_service,where:source_service IS NOT NULL" json:"sourceService,omitempty"` // Service reporting the event
	TargetService *string `gorm:"type:varchar(100);index:idx_audit_logs_target_service,where:target_service IS NOT NULL" json:"targetService,omitempty"` // Target service

	// Request Context (for opensuperapp)
	IPAddress  *string `gorm:"type:varchar(45)" json:"ipAddress,omitempty"` // IPv4 or IPv6
	UserAgent  *string `gorm:"type:varchar(500)" json:"userAgent,omitempty"`
	RequestID  *string `gorm:"type:varchar(100)" json:"requestId,omitempty"`
	SessionID  *string `gorm:"type:varchar(100)" json:"sessionId,omitempty"`
	MicroappID *string `gorm:"type:varchar(100)" json:"microappId,omitempty"`
	Platform   *string `gorm:"type:varchar(20)" json:"platform,omitempty"` // WEB, ANDROID, IOS, MICROAPP_SERVICE

	// Change Tracking (for opensuperapp)
	Changes json.RawMessage `gorm:"type:jsonb" json:"changes,omitempty"` // {"field": {"from": "old", "to": "new"}}

	// Flexible Metadata (project-specific data)
	// For opendif-core: requestedData, responseMetadata, eventMetadata
	// For opensuperapp: additional context-specific data
	Metadata json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Error Information
	ErrorMessage *string `gorm:"type:text" json:"errorMessage,omitempty"`
	ErrorCode    *string `gorm:"type:varchar(50)" json:"errorCode,omitempty"`

	// Security & Authorization (optional)
	AuthorizationGroups  json.RawMessage `gorm:"type:jsonb" json:"authorizationGroups,omitempty"`        // Array of group names
	AuthenticationMethod *string         `gorm:"type:varchar(50)" json:"authenticationMethod,omitempty"` // USER_TOKEN, SERVICE_TOKEN

	// Timestamps
	CreatedAt time.Time `gorm:"type:timestamp with time zone;not null;default:now()" json:"createdAt"`
}

// TableName sets the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to set default values
func (l *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now().UTC()
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	return
}

// Validate performs validation checks
func (l *AuditLog) Validate() error {
	// Validate status
	if l.Status != StatusSuccess && l.Status != StatusFailure && l.Status != StatusPartialFailure {
		return gorm.ErrInvalidValue
	}

	// Validate actor_type
	if l.ActorType != ActorTypeUser && l.ActorType != ActorTypeService && l.ActorType != ActorTypeSystem {
		return gorm.ErrInvalidValue
	}

	// Validate eventName is not empty
	if l.EventName == "" {
		return gorm.ErrInvalidValue
	}

	return nil
}
