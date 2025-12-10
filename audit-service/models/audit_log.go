package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AuditLog represents a generalized system event for tracking distributed request flows
type AuditLog struct {
	ID            string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	TraceID       string          `gorm:"type:uuid;index;not null" json:"traceId"`                          // Global trace ID for distributed requests
	Timestamp     time.Time       `gorm:"type:timestamp with time zone;not null" json:"timestamp"`          // Time of event
	SourceService string          `gorm:"type:varchar(50);not null" json:"sourceService"`                   // Service reporting the event (e.g., "orchestration-engine")
	TargetService string          `gorm:"type:varchar(50)" json:"targetService,omitempty"`                  // Target service (e.g., "pdp", "consent-engine")
	EventType     string          `gorm:"type:varchar(50);not null" json:"eventType"`                       // Event type (e.g., "DATA_REQUEST", "POLICY_CHECK")
	Status        string          `gorm:"type:varchar(20);not null;check:status IN ('SUCCESS', 'FAILURE')" json:"status"` // Event status
	ActorID       *string         `gorm:"type:varchar(255)" json:"actorId,omitempty"`                       // Who initiated it (User ID or System)
	Resources     json.RawMessage `gorm:"type:jsonb" json:"resources,omitempty"`                            // Affected resources (SchemaID, AppID, etc.)
	Metadata      json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`                             // Detailed payload (requested fields, error messages)
	CreatedAt     time.Time       `gorm:"type:timestamp with time zone;default:now()" json:"createdAt"`
}

// TableName sets the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to set default values if needed
func (l *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now()
	}
	return
}

// CreateAuditLogRequest represents the request payload for creating an audit log
type CreateAuditLogRequest struct {
	TraceID       string          `json:"traceId" validate:"required"`
	Timestamp     string          `json:"timestamp"` // Optional, defaults to now
	SourceService string          `json:"sourceService" validate:"required"`
	TargetService string          `json:"targetService,omitempty"`
	EventType     string          `json:"eventType" validate:"required"`
	Status        string          `json:"status" validate:"required"`
	ActorID       *string         `json:"actorId,omitempty"`
	Resources     json.RawMessage `json:"resources,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}
