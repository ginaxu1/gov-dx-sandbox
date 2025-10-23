package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// AuditService provides access to audit logs
type AuditService struct {
	db *sql.DB
}

// AuditLogFilter represents filter criteria for audit logs
type AuditLogFilter struct {
	ConsumerID string
	ProviderID string
	Status     string
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

// AuditLogRequest represents a request to create an audit log
type AuditLogRequest struct {
	EventID           string
	ConsumerID        string
	ProviderID        string
	RequestedData     []byte
	ResponseData      []byte
	TransactionStatus string
	UserAgent         string
	IPAddress         string
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                string
	CreatedAt         time.Time
	TransactionStatus string
	RequestedData     string
	ResponseData      string
	ApplicationID     string
	SchemaID          string
	ConsumerID        string
	ProviderID        string
}

// NewAuditService creates a new audit service
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// GetLogs retrieves logs with optional filtering
func (s *AuditService) GetLogs(ctx context.Context, filter *models.LogFilter) (*models.LogResponse, error) {
	query := `
		SELECT id, timestamp, status, requested_data, application_id, schema_id, consumer_id, provider_id
		FROM audit_logs_with_provider_consumer
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if filter.ConsumerID != "" {
		query += fmt.Sprintf(" AND consumer_id = $%d", argIndex)
		args = append(args, filter.ConsumerID)
		argIndex++
	}

	if filter.ProviderID != "" {
		query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
		args = append(args, filter.ProviderID)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if !filter.StartDate.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, filter.StartDate)
		argIndex++
	}

	if !filter.EndDate.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, filter.EndDate)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	// Execute query with context
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("Failed to query logs", "error", err)
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []models.Log
	for rows.Next() {
		var log models.Log
		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Status,
			&log.RequestedData,
			&log.ApplicationID,
			&log.SchemaID,
			&log.ConsumerID,
			&log.ProviderID,
		)
		if err != nil {
			slog.Error("Failed to scan log", "error", err)
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}

	// Get total count for pagination
	total, err := s.getLogsTotalCount(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &models.LogResponse{
		Logs:   logs,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

// CreateLog creates a new log entry
func (s *AuditService) CreateLog(ctx context.Context, logReq *models.LogRequest) (*models.Log, error) {
	// First insert the log entry
	insertQuery := `
		INSERT INTO audit_logs (status, requested_data, application_id, schema_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var logID string
	err := s.db.QueryRowContext(ctx, insertQuery, logReq.Status, logReq.RequestedData, logReq.ApplicationID, logReq.SchemaID).Scan(&logID)
	if err != nil {
		slog.Error("Failed to create log", "error", err)
		return nil, fmt.Errorf("failed to create log: %w", err)
	}

	// Then fetch the complete log entry with joined data
	fetchQuery := `
		SELECT id, timestamp, status, requested_data, application_id, schema_id, consumer_id, provider_id
		FROM audit_logs_with_provider_consumer
		WHERE id = $1
	`

	var log models.Log
	err = s.db.QueryRowContext(ctx, fetchQuery, logID).Scan(
		&log.ID,
		&log.Timestamp,
		&log.Status,
		&log.RequestedData,
		&log.ApplicationID,
		&log.SchemaID,
		&log.ConsumerID,
		&log.ProviderID,
	)

	if err != nil {
		slog.Error("Failed to fetch created log", "error", err)
		return nil, fmt.Errorf("failed to fetch created log: %w", err)
	}

	return &log, nil
}

// getLogsTotalCount gets the total count of logs matching the filter
func (s *AuditService) getLogsTotalCount(ctx context.Context, filter *models.LogFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM audit_logs_with_provider_consumer WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Add filters (same as main query)
	if filter.ConsumerID != "" {
		query += fmt.Sprintf(" AND consumer_id = $%d", argIndex)
		args = append(args, filter.ConsumerID)
		argIndex++
	}

	if filter.ProviderID != "" {
		query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
		args = append(args, filter.ProviderID)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if !filter.StartDate.IsZero() {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, filter.StartDate)
		argIndex++
	}

	if !filter.EndDate.IsZero() {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, filter.EndDate)
		argIndex++
	}

	var total int64
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get logs total count: %w", err)
	}

	return total, nil
}

// GetLogs queries the database for audit logs based on filters
func (s *AuditService) GetAuditLogs(filter AuditLogFilter) ([]AuditLog, int, error) {
	ctx := context.Background()

	// Build the base query
	query := `
		SELECT id, created_at, transaction_status, requested_data, response_data, 
		       application_id, schema_id, consumer_id, provider_id
		FROM audit_logs_with_provider_consumer
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if filter.ConsumerID != "" {
		query += fmt.Sprintf(" AND consumer_id = $%d", argIndex)
		args = append(args, filter.ConsumerID)
		argIndex++
	}

	if filter.ProviderID != "" {
		query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
		args = append(args, filter.ProviderID)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND transaction_status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.Limit, filter.Offset)

	// Execute the query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID,
			&log.CreatedAt,
			&log.TransactionStatus,
			&log.RequestedData,
			&log.ResponseData,
			&log.ApplicationID,
			&log.SchemaID,
			&log.ConsumerID,
			&log.ProviderID,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	// Get total count for pagination
	countQuery := "SELECT COUNT(*) FROM audit_logs_with_provider_consumer WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1

	// Add the same filters for count
	if filter.ConsumerID != "" {
		countQuery += fmt.Sprintf(" AND consumer_id = $%d", countArgIndex)
		countArgs = append(countArgs, filter.ConsumerID)
		countArgIndex++
	}

	if filter.ProviderID != "" {
		countQuery += fmt.Sprintf(" AND provider_id = $%d", countArgIndex)
		countArgs = append(countArgs, filter.ProviderID)
		countArgIndex++
	}

	if filter.Status != "" {
		countQuery += fmt.Sprintf(" AND transaction_status = $%d", countArgIndex)
		countArgs = append(countArgs, filter.Status)
		countArgIndex++
	}

	if filter.StartDate != nil {
		countQuery += fmt.Sprintf(" AND created_at >= $%d", countArgIndex)
		countArgs = append(countArgs, *filter.StartDate)
		countArgIndex++
	}

	if filter.EndDate != nil {
		countQuery += fmt.Sprintf(" AND created_at <= $%d", countArgIndex)
		countArgs = append(countArgs, *filter.EndDate)
		countArgIndex++
	}

	var total int
	err = s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return logs, total, nil
}
