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

// AuditLog represents a generalized system event for tracking distributed request flows
type AuditLog struct {
	ID            uuid.UUID       `gorm:"primaryKey" json:"id"`
	TraceID       uuid.UUID       `gorm:"index;not null" json:"traceId"`                                         // Global trace ID for distributed requests
	Timestamp     time.Time       `gorm:"not null" json:"timestamp"`                                             // Time of event
	SourceService string          `gorm:"size:50;not null" json:"sourceService"`                                 // Service reporting the event (e.g., "orchestration-engine")
	TargetService string          `gorm:"size:50" json:"targetService,omitempty"`                                // Target service (e.g., "pdp", "consent-engine")
	EventType     string          `gorm:"size:50;not null" json:"eventType"`                                     // Event type (e.g., "DATA_REQUEST", "POLICY_CHECK")
	Status        string          `gorm:"size:20;not null;check:status IN ('SUCCESS', 'FAILURE')" json:"status"` // Event status
	ActorID       *string         `gorm:"size:255" json:"actorId,omitempty"`                                     // Who initiated it (User ID or System)
	Resources     json.RawMessage `gorm:"type:bytes;serializer:json" json:"resources,omitempty"`                 // Affected resources (SchemaID, AppID, etc.)
	Metadata      json.RawMessage `gorm:"type:bytes;serializer:json" json:"metadata,omitempty"`                  // Detailed payload (requested fields, error messages)
	CreatedAt     time.Time       `json:"createdAt"`
}

// TableName sets the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to set default values if needed
func (l *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now()
	}
	return
}

// CreateAuditLogRequest represents the request payload for creating an audit log
type CreateAuditLogRequest struct {
	TraceID       string          `json:"traceId" validate:"required,uuid"`
	Timestamp     string          `json:"timestamp"` // Optional, defaults to now
	SourceService string          `json:"sourceService" validate:"required"`
	TargetService string          `json:"targetService,omitempty"`
	EventType     string          `json:"eventType" validate:"required"`
	Status        string          `json:"status" validate:"required"`
	ActorID       *string         `json:"actorId,omitempty"`
	Resources     json.RawMessage `json:"resources,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}
