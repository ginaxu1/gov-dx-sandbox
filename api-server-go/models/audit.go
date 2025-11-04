package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLogRequest represents the request structure for creating audit logs
type AuditLogRequest struct {
	EventID           uuid.UUID       `json:"event_id"`
	ConsumerID        string          `json:"consumer_id"`
	ProviderID        string          `json:"provider_id"`
	ApplicationID     string          `json:"application_id,omitempty"`
	SchemaID          string          `json:"schema_id,omitempty"`
	RequestedData     json.RawMessage `json:"requested_data"`
	ResponseData      json.RawMessage `json:"response_data,omitempty"`
	TransactionStatus string          `json:"transaction_status"` // SUCCESS or FAILURE
	UserAgent         string          `json:"user_agent,omitempty"`
	IPAddress         string          `json:"ip_address,omitempty"`
}

// AuditLogResponse represents the response from audit-service
type AuditLogResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

// AuditContext holds audit information for a request
type AuditContext struct {
	EventID       uuid.UUID
	ConsumerID    string
	ProviderID    string
	ApplicationID string // Extracted from request path or body
	SchemaID      string // Extracted from request path or body
	RequestData   json.RawMessage
	ResponseData  json.RawMessage
	Status        string
	StartTime     time.Time
	EndTime       time.Time
	UserAgent     string
	IPAddress     string
}

// NewAuditContext creates a new audit context for a request
func NewAuditContext() *AuditContext {
	return &AuditContext{
		EventID:   uuid.New(),
		StartTime: time.Now(),
	}
}

// ToAuditLogRequest converts AuditContext to AuditLogRequest
func (ac *AuditContext) ToAuditLogRequest() *AuditLogRequest {
	return &AuditLogRequest{
		EventID:           ac.EventID,
		ConsumerID:        ac.ConsumerID,
		ProviderID:        ac.ProviderID,
		ApplicationID:     ac.ApplicationID,
		SchemaID:          ac.SchemaID,
		RequestedData:     ac.RequestData,
		ResponseData:      ac.ResponseData,
		TransactionStatus: ac.Status,
		UserAgent:         ac.UserAgent,
		IPAddress:         ac.IPAddress,
	}
}

// Log represents a simplified audit log entry for API responses (matches audit-service)
type Log struct {
	ID            string    `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	RequestedData string    `json:"requestedData"`
	ConsumerID    string    `json:"consumerId"`
	ProviderID    string    `json:"providerId"`
}

// LogRequest represents the request structure for creating logs (matches audit-service)
type LogRequest struct {
	Status        string `json:"status"`
	RequestedData string `json:"requestedData"`
	ApplicationID string `json:"applicationId"`
	SchemaID      string `json:"schemaId"`
}
