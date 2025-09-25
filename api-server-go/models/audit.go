package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a single audit log entry
type AuditLog struct {
	EventID           uuid.UUID       `json:"event_id" db:"event_id"`
	Timestamp         time.Time       `json:"timestamp" db:"timestamp"`
	ConsumerID        string          `json:"consumer_id" db:"consumer_id"`
	ProviderID        string          `json:"provider_id" db:"provider_id"`
	RequestedData     json.RawMessage `json:"requested_data" db:"requested_data"`
	ResponseData      json.RawMessage `json:"response_data,omitempty" db:"response_data"`
	TransactionStatus string          `json:"transaction_status" db:"transaction_status"`
	CitizenHash       string          `json:"citizen_hash" db:"citizen_hash"`
}

// AuditLogRequest represents the request to create an audit log
type AuditLogRequest struct {
	ConsumerID        string          `json:"consumer_id" validate:"required"`
	ProviderID        string          `json:"provider_id" validate:"required"`
	RequestedData     json.RawMessage `json:"requested_data" validate:"required"`
	ResponseData      json.RawMessage `json:"response_data,omitempty"`
	TransactionStatus string          `json:"transaction_status" validate:"required,oneof=SUCCESS FAILURE"`
	CitizenHash       string          `json:"citizen_hash" validate:"required"`
}

// AuditLogResponse represents the response for audit log queries
type AuditLogResponse struct {
	EventID           uuid.UUID       `json:"event_id"`
	Timestamp         time.Time       `json:"timestamp"`
	ConsumerID        string          `json:"consumer_id"`
	ProviderID        string          `json:"provider_id"`
	RequestedData     json.RawMessage `json:"requested_data"`
	ResponseData      json.RawMessage `json:"response_data,omitempty"`
	TransactionStatus string          `json:"transaction_status"`
	CitizenHash       string          `json:"citizen_hash"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	ConsumerID        string    `json:"consumer_id,omitempty"`
	ProviderID        string    `json:"provider_id,omitempty"`
	CitizenHash       string    `json:"citizen_hash,omitempty"`
	TransactionStatus string    `json:"transaction_status,omitempty"`
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	Limit             int       `json:"limit,omitempty"`
	Offset            int       `json:"offset,omitempty"`
}

// AuditLogSummary represents summary statistics for audit logs
type AuditLogSummary struct {
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`
	UniqueConsumers    int64 `json:"unique_consumers"`
	UniqueProviders    int64 `json:"unique_providers"`
	DateRange          struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	} `json:"date_range"`
}

// Constants for audit log status values
const (
	// Transaction Status
	TransactionStatusSuccess = "SUCCESS"
	TransactionStatusFailure = "FAILURE"
)
