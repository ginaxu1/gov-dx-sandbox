package models

import (
	"time"
)

// Log represents a simplified audit log entry for API responses
type Log struct {
	ID            string    `json:"id" db:"id"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	Status        string    `json:"status" db:"status"`
	RequestedData string    `json:"requestedData" db:"requested_data"`
	ApplicationID string    `json:"applicationId" db:"application_id"`
	SchemaID      string    `json:"schemaId" db:"schema_id"`
	ConsumerID    string    `json:"consumerId" db:"consumer_id"`
	ProviderID    string    `json:"providerId" db:"provider_id"`
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

// LogRequest represents the request to create a log entry
type LogRequest struct {
	Status        string `json:"status" validate:"required,oneof=success failure"`
	RequestedData string `json:"requestedData" validate:"required"`
	ApplicationID string `json:"applicationId" validate:"required"`
	SchemaID      string `json:"schemaId" validate:"required"`
}
