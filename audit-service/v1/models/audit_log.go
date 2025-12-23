package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Audit log status constants
const (
	StatusSuccess = "SUCCESS"
	StatusFailure = "FAILURE"
)

// Actor type constants
const (
	ActorTypeUser    = "USER"
	ActorTypeService = "SERVICE"
)

// Target type constants
const (
	TargetTypeResource = "RESOURCE"
	TargetTypeService  = "SERVICE"
)

// Event type constants (CRUD operations)
const (
	EventTypeCreate = "CREATE"
	EventTypeRead   = "READ"
	EventTypeUpdate = "UPDATE"
	EventTypeDelete = "DELETE"
)

// Event name constants
const (
	EventNamePolicyCheck     = "POLICY_CHECK"
	EventNameConsentCheck    = "CONSENT_CHECK"
	EventNameDataFetch       = "DATA_FETCH"
	EventNameManagementEvent = "MANAGEMENT_EVENT"
)

// AuditLog represents a generalized audit log entry matching the proposed SQL schema
type AuditLog struct {
	// Primary Key
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Temporal
	Timestamp time.Time `gorm:"type:timestamp with time zone;not null;index:idx_audit_logs_timestamp" json:"timestamp"`

	// Trace & Correlation
	TraceID *uuid.UUID `gorm:"type:uuid;index:idx_audit_logs_trace_id,where:trace_id IS NOT NULL" json:"traceId,omitempty"` // NULL for standalone events

	// Event Classification
	EventName string  `gorm:"type:varchar(100);not null;index:idx_audit_logs_event_name" json:"eventName"` // POLICY_CHECK, CONSENT_CHECK, DATA_FETCH, MANAGEMENT_EVENT
	EventType *string `gorm:"type:varchar(20)" json:"eventType,omitempty"`                                 // CREATE, READ, UPDATE, DELETE (nullable for non-CRUD)
	Status    string  `gorm:"type:varchar(10);not null;check:status IN ('SUCCESS', 'FAILURE');index:idx_audit_logs_status" json:"status"`

	// Actor (Flattened from ActorMetadata)
	ActorType        string          `gorm:"type:varchar(10);not null;check:actor_type IN ('USER', 'SERVICE')" json:"actorType"`
	ActorServiceName *string         `gorm:"type:varchar(100);index:idx_audit_logs_actor_service,where:actor_type = 'SERVICE'" json:"actorServiceName,omitempty"` // NULL for USER, required for SERVICE
	ActorUserID      *uuid.UUID      `gorm:"type:uuid;index:idx_audit_logs_actor_user_id,where:actor_type = 'USER'" json:"actorUserId,omitempty"`                 // NULL for SERVICE, required for USER
	ActorUserType    *string         `gorm:"type:varchar(20)" json:"actorUserType,omitempty"`                                                                     // NULL for SERVICE, 'ADMIN' or 'MEMBER' for USER
	ActorMetadata    json.RawMessage `gorm:"type:jsonb" json:"actorMetadata,omitempty"`                                                                           // Additional actor context

	// Target (Flattened from TargetMetadata)
	TargetType        string          `gorm:"type:varchar(10);not null;check:target_type IN ('RESOURCE', 'SERVICE')" json:"targetType"`
	TargetServiceName *string         `gorm:"type:varchar(100);index:idx_audit_logs_target_service,where:target_type = 'SERVICE'" json:"targetServiceName,omitempty"` // NULL for RESOURCE, required for SERVICE
	TargetResource    *string         `gorm:"type:varchar(100);index:idx_audit_logs_target_resource,where:target_type = 'RESOURCE'" json:"targetResource,omitempty"`  // NULL for SERVICE, required for RESOURCE
	TargetResourceID  *uuid.UUID      `gorm:"type:uuid" json:"targetResourceId,omitempty"`                                                                            // NULL for SERVICE, optional for RESOURCE
	TargetMetadata    json.RawMessage `gorm:"type:jsonb" json:"targetMetadata,omitempty"`                                                                             // Additional target context

	// Request/Response (PIA-free)
	RequestedData    json.RawMessage `gorm:"type:jsonb" json:"requestedData,omitempty"`    // Request payload
	ResponseMetadata json.RawMessage `gorm:"type:jsonb" json:"responseMetadata,omitempty"` // Response or error

	// Additional Context
	EventMetadata json.RawMessage `gorm:"type:jsonb" json:"eventMetadata,omitempty"` // Additional event-specific metadata
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
	return
}

// Validate performs validation checks matching the database constraints
func (l *AuditLog) Validate() error {
	// Validate status
	if l.Status != StatusSuccess && l.Status != StatusFailure {
		return gorm.ErrInvalidValue
	}

	// Validate actor_type constraint
	if l.ActorType == ActorTypeService {
		if l.ActorServiceName == nil || *l.ActorServiceName == "" {
			return gorm.ErrInvalidValue // actor_service_name required for SERVICE
		}
		if l.ActorUserID != nil {
			return gorm.ErrInvalidValue // actor_user_id must be NULL for SERVICE
		}
	} else if l.ActorType == ActorTypeUser {
		if l.ActorUserID == nil {
			return gorm.ErrInvalidValue // actor_user_id required for USER
		}
		if l.ActorServiceName != nil && *l.ActorServiceName != "" {
			return gorm.ErrInvalidValue // actor_service_name must be NULL for USER
		}
	} else {
		return gorm.ErrInvalidValue // actor_type must be USER or SERVICE
	}

	// Validate target_type constraint
	if l.TargetType == TargetTypeService {
		if l.TargetServiceName == nil || *l.TargetServiceName == "" {
			return gorm.ErrInvalidValue // target_service_name required for SERVICE
		}
		if l.TargetResource != nil && *l.TargetResource != "" {
			return gorm.ErrInvalidValue // target_resource must be NULL for SERVICE
		}
	} else if l.TargetType == TargetTypeResource {
		if l.TargetResource == nil || *l.TargetResource == "" {
			return gorm.ErrInvalidValue // target_resource required for RESOURCE
		}
		if l.TargetServiceName != nil && *l.TargetServiceName != "" {
			return gorm.ErrInvalidValue // target_service_name must be NULL for RESOURCE
		}
	} else {
		return gorm.ErrInvalidValue // target_type must be RESOURCE or SERVICE
	}

	return nil
}
