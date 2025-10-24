package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditLog represents the audit_logs table
type AuditLog struct {
	ID                string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	EventID           string    `gorm:"type:uuid;not null" json:"event_id"`
	Timestamp         time.Time `gorm:"type:timestamp with time zone;not null;default:now()" json:"timestamp"`
	ConsumerID        string    `gorm:"type:text;not null" json:"consumer_id"`
	ProviderID        string    `gorm:"type:text;not null" json:"provider_id"`
	RequestedData     string    `gorm:"type:jsonb;not null" json:"requested_data"`
	ResponseData      *string   `gorm:"type:jsonb" json:"response_data"`
	TransactionStatus string    `gorm:"type:text;not null" json:"transaction_status"`
	CitizenHash       string    `gorm:"type:text;not null" json:"citizen_hash"`
	UserAgent         string    `gorm:"type:text;not null" json:"user_agent"`
	IPAddress         *string   `gorm:"type:inet" json:"ip_address"`
	ApplicationID     string    `gorm:"type:varchar(255);not null" json:"application_id"`
	SchemaID          string    `gorm:"type:varchar(255);not null" json:"schema_id"`
	Status            string    `gorm:"type:varchar(255);not null" json:"status"`
	CreatedAt         time.Time `gorm:"type:timestamp with time zone;default:now()" json:"created_at"`
	UpdatedAt         time.Time `gorm:"type:timestamp with time zone;default:now()" json:"updated_at"`
}

// TableName returns the table name for the AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook to generate UUID for EventID if not provided
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.EventID == "" {
		a.EventID = uuid.New().String()
	}
	return nil
}

// LogRequest represents the request structure for creating audit logs
type LogRequest struct {
	Status        string `json:"status" validate:"required,oneof=success failure"`
	RequestedData string `json:"requestedData" validate:"required"`
	ApplicationID string `json:"applicationId" validate:"required"`
	SchemaID      string `json:"schemaId" validate:"required"`
}

// Log represents the response structure for audit logs
type Log struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	RequestedData string    `json:"requestedData"`
	ApplicationID string    `json:"applicationId"`
	SchemaID      string    `json:"schemaId"`
	ConsumerID    string    `json:"consumerId"`
	ProviderID    string    `json:"providerId"`
	CreatedAt     time.Time `json:"createdAt"`
}

// LogFilter represents filters for querying logs
type LogFilter struct {
	ConsumerID string    `json:"consumerId,omitempty"`
	ProviderID string    `json:"providerId,omitempty"`
	Status     string    `json:"status,omitempty"`
	StartDate  time.Time `json:"startDate,omitempty"`
	EndDate    time.Time `json:"endDate,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
}

// LogResponse represents the API response structure for logs
type LogResponse struct {
	Logs   []Log `json:"logs"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
