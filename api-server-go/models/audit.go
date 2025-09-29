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
	RequestedData     json.RawMessage `json:"requested_data"`
	ResponseData      json.RawMessage `json:"response_data,omitempty"`
	TransactionStatus string          `json:"transaction_status"` // SUCCESS or FAILURE
	RequestPath       string          `json:"request_path"`
	RequestMethod     string          `json:"request_method"`
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
	RequestData   json.RawMessage
	ResponseData  json.RawMessage
	Status        string
	StartTime     time.Time
	EndTime       time.Time
	RequestPath   string
	RequestMethod string
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
		RequestedData:     ac.RequestData,
		ResponseData:      ac.ResponseData,
		TransactionStatus: ac.Status,
		RequestPath:       ac.RequestPath,
		RequestMethod:     ac.RequestMethod,
	}
}
