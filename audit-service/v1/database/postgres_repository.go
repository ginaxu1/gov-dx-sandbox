package database

import (
	"context"
	"fmt"

	"github.com/gov-dx-sandbox/audit-service/v1/models"
	"gorm.io/gorm"
)

// PostgresRepository implements AuditRepository using GORM/PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	// Auto-migrate the audit_logs table
	if err := db.AutoMigrate(&models.AuditLog{}); err != nil {
		// Log migration error but don't fail service creation
		// The actual database operation will fail later if schema is wrong
		fmt.Printf("Warning: Failed to auto-migrate audit_logs table: %v\n", err)
	}
	return &PostgresRepository{db: db}
}

// CreateAuditLog creates a new audit log entry
func (r *PostgresRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) (*models.AuditLog, error) {
	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", result.Error)
	}
	return log, nil
}

// GetAuditLogsByTraceID retrieves all audit logs for a given trace ID
func (r *PostgresRepository) GetAuditLogsByTraceID(ctx context.Context, traceID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	result := r.db.WithContext(ctx).
		Where("trace_id = ?", traceID).
		Order("timestamp ASC").
		Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve audit logs by trace ID: %w", result.Error)
	}
	if logs == nil {
		logs = []models.AuditLog{}
	}
	return logs, nil
}

// GetAuditLogsByEventName retrieves audit logs filtered by event name
func (r *PostgresRepository) GetAuditLogsByEventName(ctx context.Context, eventName string, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{}).Where("event_name = ?", eventName)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Apply pagination and ordering
	if err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, total, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (r *PostgresRepository) GetAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	// Apply filters
	if filters.TraceID != nil && *filters.TraceID != "" {
		query = query.Where("trace_id = ?", *filters.TraceID)
	}
	if filters.EventName != nil && *filters.EventName != "" {
		query = query.Where("event_name = ?", *filters.EventName)
	}
	if filters.Status != nil && *filters.Status != "" {
		query = query.Where("status = ?", *filters.Status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Apply pagination and ordering
	limit := filters.Limit
	if limit <= 0 {
		limit = 100 // default
	}
	if limit > 1000 {
		limit = 1000 // max
	}

	if err := query.Order("timestamp DESC").Limit(limit).Offset(filters.Offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, total, nil
}
