package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditLogResponse represents the response payload for an audit log entry
type AuditLogResponse struct {
	ID        uuid.UUID  `json:"id"`
	Timestamp time.Time  `json:"timestamp"`
	TraceID   *uuid.UUID `json:"traceId,omitempty"`

	EventType   *string `json:"eventType,omitempty"`
	EventAction *string `json:"eventAction,omitempty"`
	Status      string  `json:"status"`

	ActorType string `json:"actorType"`
	ActorID   string `json:"actorId"`

	TargetType string  `json:"targetType"`
	TargetID   *string `json:"targetId,omitempty"`

	RequestMetadata    json.RawMessage `json:"requestMetadata,omitempty"`
	ResponseMetadata   json.RawMessage `json:"responseMetadata,omitempty"`
	AdditionalMetadata json.RawMessage `json:"additionalMetadata,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

// GetAuditLogsResponse represents the response for querying audit logs
type GetAuditLogsResponse struct {
	Logs   []AuditLogResponse `json:"logs"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// ToAuditLogResponse converts an AuditLog model to an AuditLogResponse
// This encapsulates the mapping logic to keep handlers clean and reduce maintenance risk
func ToAuditLogResponse(log AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:                 log.ID,
		Timestamp:          log.Timestamp,
		TraceID:            log.TraceID,
		EventType:          log.EventType,
		EventAction:        log.EventAction,
		Status:             log.Status,
		ActorType:          log.ActorType,
		ActorID:            log.ActorID,
		TargetType:         log.TargetType,
		TargetID:           log.TargetID,
		RequestMetadata:    log.RequestMetadata,
		ResponseMetadata:   log.ResponseMetadata,
		AdditionalMetadata: log.AdditionalMetadata,
		CreatedAt:          log.CreatedAt,
	}
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}
