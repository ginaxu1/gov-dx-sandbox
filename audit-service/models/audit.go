package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONString represents a JSON string for JSONB fields in PostgreSQL
type JSONString string

// Value implements driver.Valuer for JSONString
func (j JSONString) Value() (driver.Value, error) {
	if j == "" {
		return json.Marshal(map[string]interface{}{})
	}
	// Check if already valid JSON
	var js interface{}
	if err := json.Unmarshal([]byte(j), &js); err == nil {
		// Already valid JSON, return as is
		return []byte(j), nil
	}
	// Not valid JSON, wrap as a JSON string
	escaped, _ := json.Marshal(string(j))
	return escaped, nil
}

// Scan implements sql.Scanner for JSONString
func (j *JSONString) Scan(value interface{}) error {
	if value == nil {
		*j = ""
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = JSONString(v)
		return nil
	case string:
		*j = JSONString(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into JSONString", value)
	}
}

// AuditLog represents the audit_logs table
type AuditLog struct {
	ID                string      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	EventID           string      `gorm:"type:uuid;not null" json:"event_id"`
	Timestamp         time.Time   `gorm:"type:timestamp with time zone;not null;default:now()" json:"timestamp"`
	ConsumerID        string      `gorm:"type:text;not null" json:"consumer_id"`
	ProviderID        string      `gorm:"type:text;not null" json:"provider_id"`
	RequestedData     JSONString  `gorm:"type:jsonb;not null" json:"requested_data"`
	ResponseData      *JSONString `gorm:"type:jsonb" json:"response_data"`
	TransactionStatus string      `gorm:"type:text;not null" json:"transaction_status"`
	CitizenHash       string      `gorm:"type:text;not null" json:"citizen_hash"`
	UserAgent         string      `gorm:"type:text;not null" json:"user_agent"`
	IPAddress         *string     `gorm:"type:inet" json:"ip_address"`
	ApplicationID     string      `gorm:"type:varchar(255);not null" json:"application_id"`
	SchemaID          string      `gorm:"type:varchar(255);not null" json:"schema_id"`
	Status            string      `gorm:"type:varchar(255);not null" json:"status"`
	// New fields for M2M vs User differentiation
	RequestType string    `gorm:"type:varchar(50);default:'unknown'" json:"request_type"`
	AuthMethod  string    `gorm:"type:varchar(50);default:'none'" json:"auth_method"`
	UserID      *string   `gorm:"type:varchar(255)" json:"user_id"`
	SessionID   *string   `gorm:"type:varchar(255)" json:"session_id"`
	CreatedAt   time.Time `gorm:"type:timestamp with time zone;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp with time zone;default:now()" json:"updated_at"`
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
	// New fields for M2M vs User differentiation
	RequestType string `json:"requestType,omitempty"`
	AuthMethod  string `json:"authMethod,omitempty"`
	UserID      string `json:"userId,omitempty"`
	SessionID   string `json:"sessionId,omitempty"`
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
