package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/audit-service/v1/models"
	"github.com/gov-dx-sandbox/audit-service/v1/types"
	"gorm.io/gorm"
)

// AuditService interface defines methods for handling audit logs
type AuditService interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error)
	GetAuditLogs(ctx context.Context, filter *types.GetAuditLogsRequest) (*types.GetAuditLogsResponse, error)
}

// auditService implementation
type auditService struct {
	db *gorm.DB
}

// NewAuditService creates a new instance of AuditService
func NewAuditService(db *gorm.DB) AuditService {
	// Auto-migrate the table
	if err := db.AutoMigrate(&models.AuditLog{}); err != nil {
		// Log migration error but don't fail service creation
		// This allows the service to start even if migration fails
		// The actual database operation will fail later if schema is wrong
		slog.Warn("Failed to auto-migrate audit_logs table", "error", err)
	}
	return &auditService{db: db}
}

// CreateAuditLog creates a new audit log entry
func (s *auditService) CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Validate using model's Validate method
	if err := log.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Additional validation for required fields
	if log.EventName == "" {
		return nil, errors.New("eventName is required")
	}
	if log.Status != models.StatusSuccess && log.Status != models.StatusFailure {
		return nil, fmt.Errorf("invalid status: %s (must be SUCCESS or FAILURE)", log.Status)
	}

	result := s.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", result.Error)
	}

	return log, nil
}

// GetAuditLogs retrieves audit logs with flexible filtering
func (s *auditService) GetAuditLogs(ctx context.Context, filter *types.GetAuditLogsRequest) (*types.GetAuditLogsResponse, error) {
	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Build query
	query := s.db.WithContext(ctx).Model(&models.AuditLog{})

	// Apply filters
	if filter.TraceID != nil {
		query = query.Where("trace_id = ?", *filter.TraceID)
	}
	if filter.EventName != nil && *filter.EventName != "" {
		query = query.Where("event_name = ?", *filter.EventName)
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.ActorServiceName != nil && *filter.ActorServiceName != "" {
		query = query.Where("actor_service_name = ? AND actor_type = ?", *filter.ActorServiceName, models.ActorTypeService)
	}
	if filter.ActorUserID != nil {
		query = query.Where("actor_user_id = ? AND actor_type = ?", *filter.ActorUserID, models.ActorTypeUser)
	}
	if filter.TargetServiceName != nil && *filter.TargetServiceName != "" {
		query = query.Where("target_service_name = ? AND target_type = ?", *filter.TargetServiceName, models.TargetTypeService)
	}
	if filter.TargetResource != nil && *filter.TargetResource != "" {
		query = query.Where("target_resource = ? AND target_type = ?", *filter.TargetResource, models.TargetTypeResource)
	}

	// Time range filters
	if filter.StartTime != nil && *filter.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, *filter.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid startTime format: %w", err)
		}
		query = query.Where("timestamp >= ?", startTime)
	}
	if filter.EndTime != nil && *filter.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, *filter.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid endTime format: %w", err)
		}
		query = query.Where("timestamp <= ?", endTime)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Apply pagination
	limit := 100 // default
	if filter.Limit != nil {
		if *filter.Limit > 0 && *filter.Limit <= 1000 {
			limit = *filter.Limit
		} else if *filter.Limit > 1000 {
			limit = 1000 // max limit
		}
	}
	offset := 0 // default
	if filter.Offset != nil && *filter.Offset > 0 {
		offset = *filter.Offset
	}

	// Fetch logs
	var logs []models.AuditLog
	result := query.
		Order("timestamp DESC"). // Most recent first
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve audit logs: %w", result.Error)
	}

	// Return empty slice instead of nil if no logs found
	if logs == nil {
		logs = []models.AuditLog{}
	}

	// Convert to response type
	events := make([]types.AuditLogResponse, len(logs))
	for i := range logs {
		events[i] = types.AuditLogResponse{AuditLog: logs[i]}
	}

	return &types.GetAuditLogsResponse{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Events: events,
	}, nil
}
