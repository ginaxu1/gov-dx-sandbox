package models

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/audit-service/config"
	"gorm.io/gorm"
)

// Audit log status constants (not configurable via YAML as they are core to the system)
const (
	StatusSuccess = "SUCCESS"
	StatusFailure = "FAILURE"
	// Note: Partial failures should be logged as FAILURE with details in response_metadata or additional_metadata
	// Example: {"succeeded": 95, "failed": 5, "total": 100}
)

// Enum configuration (loaded from YAML config file)
// Uses config.AuditEnums to leverage O(1) validation lookups
var (
	enumConfig     *config.AuditEnums
	enumConfigOnce sync.Once
)

// SetEnumConfig sets the enum configuration (called at service startup)
// Accepts config.AuditEnums to use its efficient O(1) validation methods
func SetEnumConfig(enums *config.AuditEnums) {
	enumConfigOnce.Do(func() {
		enumConfig = enums
	})
}

// GetEnumConfig returns the current enum configuration
func GetEnumConfig() *config.AuditEnums {
	return enumConfig
}

// AuditLog represents a generalized audit log entry matching the SQL schema
// This model is designed to be reusable across different projects (opendif-core, opensuperapp)
type AuditLog struct {
	// Primary Key
	ID uuid.UUID `gorm:"primaryKey" json:"id"`

	// Temporal
	Timestamp time.Time `gorm:"not null;index:idx_audit_logs_timestamp" json:"timestamp"`

	// Trace & Correlation
	// Global trace ID for distributed requests. Provided by the client. Nullable for standalone events.
	TraceID *uuid.UUID `gorm:"index:idx_audit_logs_trace_id" json:"traceId,omitempty"`

	// Event Classification
	Status      string  `gorm:"type:varchar(20);not null;index:idx_audit_logs_status" json:"status"`
	EventType   *string `gorm:"type:varchar(50)" json:"eventType,omitempty"`   // e.g., POLICY_CHECK, MANAGEMENT_EVENT (user-defined custom names)
	EventAction *string `gorm:"type:varchar(50)" json:"eventAction,omitempty"` // e.g., CREATE, READ, UPDATE, DELETE

	// Actor Information (unified approach)
	ActorType string `gorm:"type:varchar(50);not null" json:"actorType"`
	ActorID   string `gorm:"type:varchar(255);not null" json:"actorId"` // email, uuid, or service-name

	// Target Information (unified approach)
	TargetType string  `gorm:"type:varchar(50);not null" json:"targetType"`
	TargetID   *string `gorm:"type:varchar(255)" json:"targetId,omitempty"` // resource_id or service_name

	// Metadata (Payload without PII/sensitive data)
	RequestMetadata    json.RawMessage `gorm:"type:text" json:"requestMetadata,omitempty"`    // Request payload without PII/sensitive data
	ResponseMetadata   json.RawMessage `gorm:"type:text" json:"responseMetadata,omitempty"`   // Response or Error details
	AdditionalMetadata json.RawMessage `gorm:"type:text" json:"additionalMetadata,omitempty"` // Additional context-specific data

	// BaseModel provides CreatedAt
	BaseModel
}

// TableName sets the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to set default values
func (l *AuditLog) BeforeCreate(tx *gorm.DB) error {
	// Generate ID if not set
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}

	// Timestamp should already be set by the service layer (required field)
	// This check ensures data integrity but should not be the primary source
	if l.Timestamp.IsZero() {
		// This should not happen if service validation is correct, but we set it as a safety fallback
		l.Timestamp = time.Now().UTC()
	}

	// Call BaseModel BeforeCreate to set CreatedAt
	return l.BaseModel.BeforeCreate(tx)
}

// Validate performs validation checks matching the database constraints
// Uses enum configuration if available, otherwise falls back to default constants
// Uses O(1) lookup methods from config.AuditEnums for efficient validation
func (l *AuditLog) Validate() error {
	// Validate status (not configurable, core system constant)
	if l.Status != StatusSuccess && l.Status != StatusFailure {
		return fmt.Errorf("invalid status: %s (must be %s or %s)", l.Status, StatusSuccess, StatusFailure)
	}

	// Validate actor_id is not empty (required for all actor types)
	if l.ActorID == "" {
		return fmt.Errorf("actorId is required")
	}

	// Validate actor_type using config's O(1) validation method if available
	if enumConfig != nil {
		if !enumConfig.IsValidActorType(l.ActorType) {
			return fmt.Errorf("invalid actorType: %s", l.ActorType)
		}
	} else {
		// Fallback to default validation when config is not loaded
		// Use config.DefaultEnums to avoid duplication (access fields directly to avoid copying sync.Once)
		if !contains(config.DefaultEnums.ActorTypes, l.ActorType) {
			return fmt.Errorf("invalid actorType: %s (must be one of: %v)", l.ActorType, config.DefaultEnums.ActorTypes)
		}
	}

	// Validate target_type using config's O(1) validation method if available
	if enumConfig != nil {
		if !enumConfig.IsValidTargetType(l.TargetType) {
			return fmt.Errorf("invalid targetType: %s", l.TargetType)
		}
	} else {
		// Fallback to default validation when config is not loaded
		// Use config.DefaultEnums to avoid duplication (access fields directly to avoid copying sync.Once)
		if !contains(config.DefaultEnums.TargetTypes, l.TargetType) {
			return fmt.Errorf("invalid targetType: %s (must be one of: %v)", l.TargetType, config.DefaultEnums.TargetTypes)
		}
	}

	// Validate event_type if provided (nullable field, using config's O(1) validation method)
	if l.EventType != nil && *l.EventType != "" {
		if enumConfig != nil {
			if !enumConfig.IsValidEventType(*l.EventType) {
				return fmt.Errorf("invalid eventType: %s", *l.EventType)
			}
		} else {
			// Fallback to default event types when config is not loaded
			// Use config.DefaultEnums to avoid duplication (access fields directly to avoid copying sync.Once)
			if !contains(config.DefaultEnums.EventTypes, *l.EventType) {
				return fmt.Errorf("invalid eventType: %s", *l.EventType)
			}
		}
	}

	// Validate event_action if provided (nullable field, using config's O(1) validation method)
	if l.EventAction != nil && *l.EventAction != "" {
		if enumConfig != nil {
			if !enumConfig.IsValidEventAction(*l.EventAction) {
				return fmt.Errorf("invalid eventAction: %s", *l.EventAction)
			}
		} else {
			// Fallback to default actions when config is not loaded
			// Use config.DefaultEnums to avoid duplication (access fields directly to avoid copying sync.Once)
			if !contains(config.DefaultEnums.EventActions, *l.EventAction) {
				return fmt.Errorf("invalid eventAction: %s", *l.EventAction)
			}
		}
	}

	return nil
}

// contains checks if a string slice contains a value
// Used only for fallback validation when config is not available
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
