package models

import (
	"encoding/json"
	"time"
)

// CreateDataExchangeEventRequest represents the request payload for creating a data exchange event
type CreateDataExchangeEventRequest struct {
	Timestamp         string          `json:"timestamp" validate:"required"`
	Status            string          `json:"status" validate:"required"`
	ApplicationID     string          `json:"applicationId" validate:"required"`
	SchemaID          string          `json:"schemaId" validate:"required"`
	RequestedData     json.RawMessage `json:"requestedData" validate:"required"`
	OnBehalfOfOwnerID *string         `json:"onBehalfOfOwnerId,omitempty"`
	ConsumerID        *string         `json:"consumerId,omitempty"`
	ProviderID        *string         `json:"providerId,omitempty"`
	AdditionalInfo    json.RawMessage `json:"additionalInfo,omitempty"`
}

// DataExchangeEvent represents a data model for data exchange events
type DataExchangeEvent struct {
	ID                string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Timestamp         time.Time       `gorm:"type:timestamp with time zone;not null" json:"timestamp"`
	Status            string          `gorm:"type:varchar(10);not null;check:status IN ('success', 'failure')" json:"status"`
	ApplicationID     string          `gorm:"type:varchar(255);not null" json:"applicationId"`
	SchemaID          string          `gorm:"type:varchar(255);not null" json:"schemaId"`
	RequestedData     json.RawMessage `gorm:"type:jsonb;not null" json:"requestedData"`
	OnBehalfOfOwnerID *string         `gorm:"type:varchar(255);default:null" json:"onBehalfOfOwnerId,omitempty"`
	ConsumerID        *string         `gorm:"type:varchar(255);default:null" json:"consumerId,omitempty"`
	ProviderID        *string         `gorm:"type:varchar(255);default:null" json:"providerId,omitempty"`
	AdditionalInfo    json.RawMessage `gorm:"type:jsonb;default:null" json:"additionalInfo,omitempty"`
	CreatedAt         time.Time       `gorm:"type:timestamp with time zone;default:now()" json:"createdAt"`
}

// TableName sets the table name for DataExchangeEvent model
func (DataExchangeEvent) TableName() string {
	return "data_exchange_events"
}

// DataExchangeEventResponse represents the response payload for data exchange events
type DataExchangeEventResponse struct {
	ID                string          `json:"id"`
	Timestamp         string          `json:"timestamp"`
	Status            string          `json:"status"`
	ApplicationID     string          `json:"applicationId"`
	SchemaID          string          `json:"schemaId"`
	RequestedData     json.RawMessage `json:"requestedData"`
	OnBehalfOfOwnerID *string         `json:"onBehalfOfOwnerId,omitempty"`
	ConsumerID        *string         `json:"consumerId,omitempty"`
	ProviderID        *string         `json:"providerId,omitempty"`
	AdditionalInfo    json.RawMessage `json:"additionalInfo,omitempty"`
	CreatedAt         string          `json:"createdAt"`
}

// ToResponse converts DataExchangeEvent to DataExchangeEventResponse
func (e *DataExchangeEvent) ToResponse() *DataExchangeEventResponse {
	return &DataExchangeEventResponse{
		ID:                e.ID,
		Timestamp:         e.Timestamp.Format(time.RFC3339),
		Status:            e.Status,
		ApplicationID:     e.ApplicationID,
		SchemaID:          e.SchemaID,
		RequestedData:     e.RequestedData,
		OnBehalfOfOwnerID: e.OnBehalfOfOwnerID,
		ConsumerID:        e.ConsumerID,
		ProviderID:        e.ProviderID,
		AdditionalInfo:    e.AdditionalInfo,
		CreatedAt:         e.CreatedAt.Format(time.RFC3339),
	}
}

// DataExchangeEventListResponse represents a paginated list response for data exchange events
type DataExchangeEventListResponse struct {
	Total  int64                       `json:"total"`
	Limit  int                         `json:"limit"`
	Offset int                         `json:"offset"`
	Events []DataExchangeEventResponse `json:"events"`
}

// DataExchangeEventFilter represents filtering options for querying data exchange events
type DataExchangeEventFilter struct {
	ApplicationID *string    `json:"applicationId,omitempty"`
	SchemaID      *string    `json:"schemaId,omitempty"`
	ConsumerID    *string    `json:"consumerId,omitempty"`
	ProviderID    *string    `json:"providerId,omitempty"`
	Status        *string    `json:"status,omitempty"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"`
	Limit         int        `json:"limit,omitempty"`
	Offset        int        `json:"offset,omitempty"`
}
