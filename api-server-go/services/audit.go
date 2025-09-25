package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/models"
)

// AuditService provides methods for managing audit logs
type AuditService struct {
	db *sql.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// CreateAuditLog creates a new audit log entry
func (s *AuditService) CreateAuditLog(req *models.AuditLogRequest) (*models.AuditLog, error) {
	auditLog := &models.AuditLog{
		EventID:           uuid.New(),
		Timestamp:         time.Now(),
		ConsumerID:        req.ConsumerID,
		ProviderID:        req.ProviderID,
		RequestedData:     req.RequestedData,
		ResponseData:      req.ResponseData,
		TransactionStatus: req.TransactionStatus,
		CitizenHash:       req.CitizenHash,
	}

	query := `
		INSERT INTO audit_logs (
			event_id, timestamp, consumer_id, provider_id,
			requested_data, response_data, transaction_status, citizen_hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := s.db.Exec(
		query,
		auditLog.EventID,
		auditLog.Timestamp,
		auditLog.ConsumerID,
		auditLog.ProviderID,
		auditLog.RequestedData,
		auditLog.ResponseData,
		auditLog.TransactionStatus,
		auditLog.CitizenHash,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	slog.Info("Audit log created successfully",
		"event_id", auditLog.EventID,
		"consumer_id", auditLog.ConsumerID,
		"provider_id", auditLog.ProviderID,
		"transaction_status", auditLog.TransactionStatus)

	return auditLog, nil
}

// GetAuditLogsByConsumerID retrieves audit logs for a specific consumer
func (s *AuditService) GetAuditLogsByConsumerID(consumerID string, limit, offset int) ([]*models.AuditLogResponse, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id,
			   requested_data, response_data, transaction_status, citizen_hash
		FROM audit_logs
		WHERE consumer_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, consumerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return s.scanAuditLogs(rows)
}

// GetAuditLogsByProviderID retrieves audit logs for a specific provider
func (s *AuditService) GetAuditLogsByProviderID(providerID string, limit, offset int) ([]*models.AuditLogResponse, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id,
			   requested_data, response_data, transaction_status, citizen_hash
		FROM audit_logs
		WHERE provider_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, providerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return s.scanAuditLogs(rows)
}

// GetAuditLogsByCitizenHash retrieves audit logs for a specific citizen (hashed ID)
func (s *AuditService) GetAuditLogsByCitizenHash(citizenHash string, limit, offset int) ([]*models.AuditLogResponse, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id,
			   requested_data, response_data, transaction_status, citizen_hash
		FROM audit_logs
		WHERE citizen_hash = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, citizenHash, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return s.scanAuditLogs(rows)
}

// GetAuditLogsForAdmin retrieves all audit logs for admin oversight (with pagination)
func (s *AuditService) GetAuditLogsForAdmin(limit, offset int) ([]*models.AuditLogResponse, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id,
			   requested_data, response_data, transaction_status, citizen_hash
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return s.scanAuditLogs(rows)
}

// GetAuditLogsWithFilter retrieves audit logs with custom filters
func (s *AuditService) GetAuditLogsWithFilter(filter *models.AuditLogFilter) ([]*models.AuditLogResponse, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id,
			   requested_data, response_data, transaction_status, citizen_hash
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

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

	if filter.CitizenHash != "" {
		query += fmt.Sprintf(" AND citizen_hash = $%d", argIndex)
		args = append(args, filter.CitizenHash)
		argIndex++
	}

	if filter.TransactionStatus != "" {
		query += fmt.Sprintf(" AND transaction_status = $%d", argIndex)
		args = append(args, filter.TransactionStatus)
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

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	return s.scanAuditLogs(rows)
}

// GetAuditLogSummary retrieves summary statistics for audit logs
func (s *AuditService) GetAuditLogSummary(startDate, endDate time.Time) (*models.AuditLogSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN transaction_status = 'SUCCESS' THEN 1 END) as successful_requests,
			COUNT(CASE WHEN transaction_status = 'FAILURE' THEN 1 END) as failed_requests,
			COUNT(DISTINCT consumer_id) as unique_consumers,
			COUNT(DISTINCT provider_id) as unique_providers
		FROM audit_logs
		WHERE timestamp >= $1 AND timestamp <= $2
	`

	var summary models.AuditLogSummary
	err := s.db.QueryRow(query, startDate, endDate).Scan(
		&summary.TotalRequests,
		&summary.SuccessfulRequests,
		&summary.FailedRequests,
		&summary.UniqueConsumers,
		&summary.UniqueProviders,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit log summary: %w", err)
	}

	summary.DateRange.Start = startDate
	summary.DateRange.End = endDate

	return &summary, nil
}

// DeleteOldAuditLogs deletes audit logs older than the specified duration
func (s *AuditService) DeleteOldAuditLogs(olderThan time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE timestamp < $1`

	result, err := s.db.Exec(query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	slog.Info("Deleted old audit logs",
		"older_than", olderThan,
		"rows_deleted", rowsAffected)

	return rowsAffected, nil
}

// scanAuditLogs is a helper function to scan audit log rows
func (s *AuditService) scanAuditLogs(rows *sql.Rows) ([]*models.AuditLogResponse, error) {
	var logs []*models.AuditLogResponse
	for rows.Next() {
		log := &models.AuditLogResponse{}
		err := rows.Scan(
			&log.EventID,
			&log.Timestamp,
			&log.ConsumerID,
			&log.ProviderID,
			&log.RequestedData,
			&log.ResponseData,
			&log.TransactionStatus,
			&log.CitizenHash,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}
