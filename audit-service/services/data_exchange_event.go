package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
	"gorm.io/gorm"
)

// DataExchangeEventService provides access to audit logs
type DataExchangeEventService struct {
	db *gorm.DB
}

// NewDataExchangeEventService creates a new audit service
func NewDataExchangeEventService(db *gorm.DB) *DataExchangeEventService {
	return &DataExchangeEventService{db: db}
}

// CreateDataExchangeEvent creates a new data exchange event log
func (s *DataExchangeEventService) CreateDataExchangeEvent(ctx context.Context, req *models.CreateDataExchangeEventRequest) (*models.DataExchangeEventResponse, error) {
	// Parse and validate timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		slog.Error("Invalid timestamp format", "error", err)
		return nil, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Validate status
	if req.Status != "success" && req.Status != "failure" {
		slog.Error("Invalid status", "status", req.Status)
		return nil, fmt.Errorf("invalid status: %s. Must be success or failure", req.Status)
	}

	// Create the event record
	event := &models.DataExchangeEvent{
		ID:                uuid.New().String(),
		Timestamp:         timestamp,
		Status:            req.Status,
		ApplicationID:     req.ApplicationID,
		SchemaID:          req.SchemaID,
		RequestedData:     req.RequestedData,
		OnBehalfOfOwnerID: req.OnBehalfOfOwnerID,
		ConsumerID:        req.ConsumerID,
		ProviderID:        req.ProviderID,
		AdditionalInfo:    req.AdditionalInfo,
	}

	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		slog.Error("Failed to create data exchange event", "error", err)
		return nil, fmt.Errorf("failed to create data exchange event: %w", err)
	}

	slog.Info("Created data exchange event", "eventID", event.ID)
	return event.ToResponse(), nil
}

// GetDataExchangeEvent retrieves a data exchange event with optional filtering
func (s *DataExchangeEventService) GetDataExchangeEvents(ctx context.Context, filter *models.DataExchangeEventFilter) (*models.DataExchangeEventListResponse, error) {
	var events []models.DataExchangeEvent
	var total int64

	// Build query with filters
	query := s.db.WithContext(ctx).Model(&models.DataExchangeEvent{})

	// Apply filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.ApplicationID != nil {
		query = query.Where("application_id = ?", *filter.ApplicationID)
	}
	if filter.SchemaID != nil {
		query = query.Where("schema_id = ?", *filter.SchemaID)
	}
	if filter.ConsumerID != nil {
		query = query.Where("consumer_id = ?", *filter.ConsumerID)
	}
	if filter.ProviderID != nil {
		query = query.Where("provider_id = ?", *filter.ProviderID)
	}
	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", *filter.EndDate)
	}

	// Get total count for pagination
	if err := query.Count(&total).Error; err != nil {
		slog.Error("Failed to count data exchange events", "error", err)
		return nil, fmt.Errorf("failed to count data exchange events: %w", err)
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
		slog.Error("Failed to retrieve data exchange events", "error", err)
		return nil, fmt.Errorf("failed to retrieve data exchange events: %w", err)
	}

	// Convert to response format
	var eventResponses []models.DataExchangeEventResponse
	for _, event := range events {
		eventResponses = append(eventResponses, *event.ToResponse())
	}

	response := &models.DataExchangeEventListResponse{
		Events: eventResponses,
		Total:  total,
		Limit:  limit,
		Offset: filter.Offset,
	}

	slog.Debug("Retrieved data exchange events", "count", len(eventResponses))
	return response, nil
}
