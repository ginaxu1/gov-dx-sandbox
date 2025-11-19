package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ManagementEventRequest represents the request structure for management events
type ManagementEventRequest struct {
	EventID string `json:"eventId"` // UUID
	// Timestamp is optional - if not provided, will use current time
	Timestamp *string                 `json:"timestamp,omitempty"` // ISO 8601 timestamp
	EventType string                  `json:"eventType"`           // "CREATE", "UPDATE", "DELETE"
	Status    string                  `json:"status"`              // "SUCCESS", "FAILURE"
	Actor     Actor                   `json:"actor"`
	Target    Target                  `json:"target"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty"` // Optional additional context
}

// Actor represents the actor who performed the action
type Actor struct {
	Type string  `json:"type"` // "USER" or "SERVICE"
	ID   *string `json:"id"`   // User ID (null if SERVICE type)
	Role *string `json:"role"` // "MEMBER" or "ADMIN" (null if SERVICE type)
}

// Target represents the resource that was acted upon
type Target struct {
	Resource   string `json:"resource"`   // "MEMBERS", "SCHEMAS", etc.
	ResourceID string `json:"resourceId"` // The ID of the resource
}

// ManagementEvent represents the database model for management events
type ManagementEvent struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	EventID          string    `gorm:"type:uuid;uniqueIndex;not null" json:"eventId"`
	EventType        string    `gorm:"type:varchar(10);not null;check:event_type IN ('CREATE', 'READ', 'UPDATE', 'DELETE')" json:"eventType"`
	Status           string    `gorm:"type:varchar(10);not null;check:status IN ('SUCCESS', 'FAILURE')" json:"status"`
	Timestamp        time.Time `gorm:"type:timestamp with time zone;default:now()" json:"timestamp"`
	ActorType        string    `gorm:"type:varchar(10);not null;check:actor_type IN ('USER', 'SERVICE')" json:"actorType"`
	ActorID          *string   `gorm:"type:varchar(255)" json:"actorId"`
	ActorRole        *string   `gorm:"type:varchar(10);check:actor_role IN ('MEMBER', 'ADMIN')" json:"actorRole"`
	TargetResource   string    `gorm:"type:varchar(50);not null;check:target_resource IN ('MEMBERS', 'SCHEMAS', 'SCHEMA-SUBMISSIONS', 'APPLICATIONS', 'APPLICATION-SUBMISSIONS', 'POLICY-METADATA')" json:"targetResource"`
	TargetResourceID string    `gorm:"type:varchar(255);not null" json:"targetResourceId"`
	Metadata         *Metadata `gorm:"type:jsonb" json:"metadata"`
	CreatedAt        time.Time `gorm:"type:timestamp with time zone;default:now()" json:"createdAt"`
}


// TableName specifies the table name for ManagementEvent
func (ManagementEvent) TableName() string {
	return "management_events"
}

// Metadata is a JSONB type for storing additional context
type Metadata map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), m)
	}

	return json.Unmarshal(bytes, m)
}

// ManagementEventFilter represents filter parameters for querying management events
type ManagementEventFilter struct {
	EventType        *string    `json:"eventType,omitempty"`
	Status           *string    `json:"status,omitempty"`
	ActorType        *string    `json:"actorType,omitempty"`
	ActorID          *string    `json:"actorId,omitempty"`
	ActorRole        *string    `json:"actorRole,omitempty"`
	TargetResource   *string    `json:"targetResource,omitempty"`
	TargetResourceID *string    `json:"targetResourceId,omitempty"`
	StartDate        *time.Time `json:"startDate,omitempty"`
	EndDate          *time.Time `json:"endDate,omitempty"`
	Limit            int        `json:"limit"`
	Offset           int        `json:"offset"`
}

// ManagementEventResponse represents the API response for management events
type ManagementEventResponse struct {
	Events []ManagementEvent `json:"events"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
