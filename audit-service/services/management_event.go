package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"gorm.io/gorm"
)

// ManagementEventService provides access to management events
type ManagementEventService struct {
	db *gorm.DB
}

// NewManagementEventService creates a new management event service
func NewManagementEventService(db *gorm.DB) *ManagementEventService {
	return &ManagementEventService{db: db}
}

// CreateManagementEvent creates a new management event
func (s *ManagementEventService) CreateManagementEvent(ctx context.Context, req *models.CreateManagementEventRequest) (*models.ManagementEvent, error) {
	// Parse and validate timestamp
	if req.Timestamp == "" {
		return nil, fmt.Errorf("timestamp is required")
	}
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Validate status
	if req.Status != "success" && req.Status != "failure" {
		return nil, fmt.Errorf("invalid status: %s. Must be success or failure", req.Status)
	}

	// Validate event type
	if req.EventType != "CREATE" && req.EventType != "UPDATE" && req.EventType != "DELETE" {
		return nil, fmt.Errorf("invalid event type: %s. Must be CREATE, UPDATE, or DELETE", req.EventType)
	}

	// Validate actor type
	if req.Actor.Type != "USER" && req.Actor.Type != "SERVICE" {
		return nil, fmt.Errorf("invalid actor type: %s. Must be USER or SERVICE", req.Actor.Type)
	}

	// Validate actor role if actor is USER
	if req.Actor.Type == "USER" {
		if req.Actor.Role == nil {
			return nil, fmt.Errorf("actor role is required when actor type is USER")
		}
		if *req.Actor.Role != "MEMBER" && *req.Actor.Role != "ADMIN" {
			return nil, fmt.Errorf("invalid actor role: %s. Must be MEMBER or ADMIN", *req.Actor.Role)
		}
	}

	// Validate target resource
	validResources := map[string]bool{
		"MEMBERS":                 true,
		"SCHEMAS":                 true,
		"SCHEMA-SUBMISSIONS":      true,
		"APPLICATIONS":            true,
		"APPLICATION-SUBMISSIONS": true,
		"POLICY-METADATA":         true,
	}
	if !validResources[req.Target.Resource] {
		return nil, fmt.Errorf("invalid target resource: %s", req.Target.Resource)
	}

	// Validate target resource ID presence for non-CREATE events
	if req.EventType != "CREATE" && req.Target.ResourceID == nil {
		return nil, fmt.Errorf("target resource ID is required for event type %s", req.EventType)
	}

	event := models.ManagementEvent{
		EventType:        req.EventType,
		Timestamp:        timestamp,
		ActorType:        req.Actor.Type,
		ActorID:          req.Actor.ID,
		ActorRole:        req.Actor.Role,
		TargetResource:   req.Target.Resource,
		TargetResourceID: req.Target.ResourceID,
		Status:           req.Status,
	}

	// Set metadata if provided
	if req.Metadata != nil {
		event.Metadata = (*models.Metadata)(req.Metadata)
	}

	// Insert the event using GORM
	if err := s.db.WithContext(ctx).Create(&event).Error; err != nil {
		slog.Error("Failed to create management event", "error", err)
		return nil, fmt.Errorf("failed to create management event: %w", err)
	}

	slog.Info("Management event created",
		"eventID", event.ID,
		"eventType", req.EventType,
		"targetResource", req.Target.Resource,
		"targetResourceId", req.Target.ResourceID)

	return &event, nil
}

// GetManagementEvents retrieves management events with optional filtering
func (s *ManagementEventService) GetManagementEvents(ctx context.Context, filter *models.ManagementEventFilter) (*models.ManagementEventResponse, error) {
	var events []models.ManagementEvent
	var total int64

	// Build query with filters
	query := s.db.WithContext(ctx).Model(&models.ManagementEvent{})

	// Apply filters
	if filter.EventType != nil {
		query = query.Where("event_type = ?", *filter.EventType)
	}

	if filter.ActorType != nil {
		query = query.Where("actor_type = ?", *filter.ActorType)
	}

	if filter.ActorID != nil {
		query = query.Where("actor_id = ?", *filter.ActorID)
	}

	if filter.ActorRole != nil {
		query = query.Where("actor_role = ?", *filter.ActorRole)
	}

	if filter.TargetResource != nil {
		query = query.Where("target_resource = ?", *filter.TargetResource)
	}

	if filter.TargetResourceID != nil {
		query = query.Where("target_resource_id = ?", *filter.TargetResourceID)
	}

	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", *filter.StartDate)
	}

	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", *filter.EndDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		slog.Error("Failed to get total count", "error", err)
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Apply ordering and pagination
	query = query.Order("timestamp DESC")

	limit := filter.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}
	query = query.Limit(limit)

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Execute query
	if err := query.Find(&events).Error; err != nil {
		slog.Error("Failed to query management events", "error", err)
		return nil, fmt.Errorf("failed to query management events: %w", err)
	}

	response := &models.ManagementEventResponse{
		Events: events,
		Total:  total,
		Limit:  limit,
		Offset: filter.Offset,
	}

	slog.Debug("Retrieved management events", "count", len(events))
	return response, nil
}
