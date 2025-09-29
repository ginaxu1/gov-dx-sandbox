package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a single audit log entry from the database
type AuditLog struct {
	EventID           uuid.UUID       `json:"event_id" db:"event_id"`
	Timestamp         time.Time       `json:"timestamp" db:"timestamp"`
	ConsumerID        string          `json:"consumer_id" db:"consumer_id"`
	ProviderID        string          `json:"provider_id" db:"provider_id"`
	RequestedData     json.RawMessage `json:"requested_data" db:"requested_data"`
	ResponseData      json.RawMessage `json:"response_data,omitempty" db:"response_data"`
	TransactionStatus string          `json:"transaction_status" db:"transaction_status"`
	CitizenHash       string          `json:"citizen_hash" db:"citizen_hash"`
	RequestPath       string          `json:"request_path" db:"request_path"`
	RequestMethod     string          `json:"request_method" db:"request_method"`
	UserAgent         string          `json:"user_agent" db:"user_agent"`
	IPAddress         string          `json:"ip_address" db:"ip_address"`
}

// AuditEvent represents the simplified audit event for API responses
type AuditEvent struct {
	EventID           uuid.UUID `json:"event_id"`
	Timestamp         time.Time `json:"timestamp"`
	ConsumerID        string    `json:"consumer_id"`
	ProviderID        string    `json:"provider_id"`
	TransactionStatus string    `json:"transaction_status"`
	CitizenHash       string    `json:"citizen_hash"`
	RequestPath       string    `json:"request_path"`
	RequestMethod     string    `json:"request_method"`
	UserAgent         string    `json:"user_agent"`
	IPAddress         string    `json:"ip_address"`
	// RequestedData and ResponseData are intentionally omitted for security
	// They contain sensitive information that should not be exposed via API
}

// AuditFilter represents filters for querying audit logs
type AuditFilter struct {
	ConsumerID        string    `json:"consumer_id,omitempty"`
	ProviderID        string    `json:"provider_id,omitempty"`
	TransactionStatus string    `json:"transaction_status,omitempty"`
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	Limit             int       `json:"limit,omitempty"`
	Offset            int       `json:"offset,omitempty"`
}

// AuditResponse represents the API response structure
type AuditResponse struct {
	Events []AuditEvent `json:"events"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// AuditLogRequest represents the request to create an audit log
type AuditLogRequest struct {
	EventID           uuid.UUID       `json:"event_id,omitempty"`
	ConsumerID        string          `json:"consumer_id" validate:"required"`
	ProviderID        string          `json:"provider_id" validate:"required"`
	RequestedData     json.RawMessage `json:"requested_data" validate:"required"`
	ResponseData      json.RawMessage `json:"response_data,omitempty"`
	TransactionStatus string          `json:"transaction_status" validate:"required,oneof=SUCCESS FAILURE"`
	CitizenHash       string          `json:"citizen_hash,omitempty"` // Optional - will be auto-generated
	RequestPath       string          `json:"request_path" validate:"required"`
	RequestMethod     string          `json:"request_method" validate:"required"`
	UserAgent         string          `json:"user_agent,omitempty"`
	IPAddress         string          `json:"ip_address,omitempty"`
}

// JWTClaims represents the JWT token claims for authentication
type JWTClaims struct {
	ConsumerID string `json:"consumer_id,omitempty"`
	ProviderID string `json:"provider_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	Email      string `json:"email,omitempty"`
	Exp        int64  `json:"exp"`
	Iat        int64  `json:"iat"`
}

// Constants for transaction status
const (
	TransactionStatusSuccess = "SUCCESS"
	TransactionStatusFailure = "FAILURE"
)
