package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// AuditService provides access to audit logs
type AuditService struct {
	db *sql.DB
}

// DB returns the database connection (for internal use)
func (s *AuditService) DB() *sql.DB {
	return s.db
}

// NewAuditService creates a new audit service
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// GetLogs retrieves logs with optional filtering
func (s *AuditService) GetLogs(filter *models.LogFilter) (*models.LogResponse, error) {
	query := `
		SELECT id, timestamp, status, requested_data, consumer_id, provider_id
		FROM audit_logs
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

	// Execute query
	rows, err := s.db.Query(query, args...)
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
	total, err := s.getLogsTotalCount(filter)
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
func (s *AuditService) CreateLog(logReq *models.LogRequest) (*models.Log, error) {
	query := `
		INSERT INTO audit_logs (status, requested_data, consumer_id, provider_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, timestamp, status, requested_data, consumer_id, provider_id
	`

	var log models.Log
	err := s.db.QueryRow(query, logReq.Status, logReq.RequestedData, logReq.ConsumerID, logReq.ProviderID).Scan(
		&log.ID,
		&log.Timestamp,
		&log.Status,
		&log.RequestedData,
		&log.ConsumerID,
		&log.ProviderID,
	)

	if err != nil {
		slog.Error("Failed to create log", "error", err)
		return nil, fmt.Errorf("failed to create log: %w", err)
	}

	return &log, nil
}

// getLogsTotalCount gets the total count of logs matching the filter
func (s *AuditService) getLogsTotalCount(filter *models.LogFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
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
	err := s.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get logs total count: %w", err)
	}

	return total, nil
}
