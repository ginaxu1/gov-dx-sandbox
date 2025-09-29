package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gov-dx-sandbox/audit-service/models"
)

// AuditService provides read-only access to audit logs
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

// GetAuditEvents retrieves all audit events with optional filtering (for Admin Portal)
func (s *AuditService) GetAuditEvents(filter *models.AuditFilter) (*models.AuditResponse, error) {
	query := `
		SELECT event_id, timestamp, consumer_id, provider_id, transaction_status, citizen_hash,
		       request_path, request_method
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
		slog.Error("Failed to query audit events", "error", err)
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []models.AuditEvent
	for rows.Next() {
		var event models.AuditEvent
		var requestPath, requestMethod sql.NullString
		err := rows.Scan(
			&event.EventID,
			&event.Timestamp,
			&event.ConsumerID,
			&event.ProviderID,
			&event.TransactionStatus,
			&event.CitizenHash,
			&requestPath,
			&requestMethod,
		)
		if err != nil {
			slog.Error("Failed to scan audit event", "error", err)
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Handle NULL values
		if requestPath.Valid {
			event.RequestPath = requestPath.String
		}
		if requestMethod.Valid {
			event.RequestMethod = requestMethod.String
		}

		events = append(events, event)
	}

	// Get total count for pagination
	total, err := s.getTotalCount(filter)
	if err != nil {
		return nil, err
	}

	return &models.AuditResponse{
		Events: events,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

// GetProviderAuditEvents retrieves audit events for a specific provider (for Provider Portal)
func (s *AuditService) GetProviderAuditEvents(providerID string, filter *models.AuditFilter) (*models.AuditResponse, error) {
	// Override provider filter to ensure security
	filter.ProviderID = providerID

	query := `
		SELECT event_id, timestamp, consumer_id, provider_id, transaction_status, citizen_hash,
		       request_path, request_method
		FROM audit_logs
		WHERE provider_id = $1
	`
	args := []interface{}{providerID}
	argIndex := 2

	// Add additional filters
	if filter.ConsumerID != "" {
		query += fmt.Sprintf(" AND consumer_id = $%d", argIndex)
		args = append(args, filter.ConsumerID)
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
		slog.Error("Failed to query provider audit events", "provider_id", providerID, "error", err)
		return nil, fmt.Errorf("failed to query provider audit events: %w", err)
	}
	defer rows.Close()

	var events []models.AuditEvent
	for rows.Next() {
		var event models.AuditEvent
		var requestPath, requestMethod sql.NullString
		err := rows.Scan(
			&event.EventID,
			&event.Timestamp,
			&event.ConsumerID,
			&event.ProviderID,
			&event.TransactionStatus,
			&event.CitizenHash,
			&requestPath,
			&requestMethod,
		)
		if err != nil {
			slog.Error("Failed to scan provider audit event", "error", err)
			return nil, fmt.Errorf("failed to scan provider audit event: %w", err)
		}

		// Handle NULL values
		if requestPath.Valid {
			event.RequestPath = requestPath.String
		}
		if requestMethod.Valid {
			event.RequestMethod = requestMethod.String
		}

		events = append(events, event)
	}

	// Get total count for pagination
	total, err := s.getProviderTotalCount(providerID, filter)
	if err != nil {
		return nil, err
	}

	return &models.AuditResponse{
		Events: events,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

// GetConsumerAuditEvents retrieves audit events for a specific consumer (for Consumer Portal)
func (s *AuditService) GetConsumerAuditEvents(consumerID string, filter *models.AuditFilter) (*models.AuditResponse, error) {
	// Override consumer filter to ensure security
	filter.ConsumerID = consumerID

	query := `
		SELECT event_id, timestamp, consumer_id, provider_id, transaction_status, citizen_hash,
		       request_path, request_method
		FROM audit_logs
		WHERE consumer_id = $1
	`
	args := []interface{}{consumerID}
	argIndex := 2

	// Add additional filters
	if filter.ProviderID != "" {
		query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
		args = append(args, filter.ProviderID)
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
		slog.Error("Failed to query consumer audit events", "consumer_id", consumerID, "error", err)
		return nil, fmt.Errorf("failed to query consumer audit events: %w", err)
	}
	defer rows.Close()

	var events []models.AuditEvent
	for rows.Next() {
		var event models.AuditEvent
		var requestPath, requestMethod sql.NullString
		err := rows.Scan(
			&event.EventID,
			&event.Timestamp,
			&event.ConsumerID,
			&event.ProviderID,
			&event.TransactionStatus,
			&event.CitizenHash,
			&requestPath,
			&requestMethod,
		)
		if err != nil {
			slog.Error("Failed to scan consumer audit event", "error", err)
			return nil, fmt.Errorf("failed to scan consumer audit event: %w", err)
		}

		// Handle NULL values
		if requestPath.Valid {
			event.RequestPath = requestPath.String
		}
		if requestMethod.Valid {
			event.RequestMethod = requestMethod.String
		}

		events = append(events, event)
	}

	// Get total count for pagination
	total, err := s.getConsumerTotalCount(consumerID, filter)
	if err != nil {
		return nil, err
	}

	return &models.AuditResponse{
		Events: events,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

// getTotalCount gets the total count of audit events matching the filter
func (s *AuditService) getTotalCount(filter *models.AuditFilter) (int64, error) {
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

	var total int64
	err := s.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return total, nil
}

// getProviderTotalCount gets the total count of audit events for a provider
func (s *AuditService) getProviderTotalCount(providerID string, filter *models.AuditFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE provider_id = $1"
	args := []interface{}{providerID}
	argIndex := 2

	// Add additional filters
	if filter.ConsumerID != "" {
		query += fmt.Sprintf(" AND consumer_id = $%d", argIndex)
		args = append(args, filter.ConsumerID)
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

	var total int64
	err := s.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get provider total count: %w", err)
	}

	return total, nil
}

// getConsumerTotalCount gets the total count of audit events for a consumer
func (s *AuditService) getConsumerTotalCount(consumerID string, filter *models.AuditFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE consumer_id = $1"
	args := []interface{}{consumerID}
	argIndex := 2

	// Add additional filters
	if filter.ProviderID != "" {
		query += fmt.Sprintf(" AND provider_id = $%d", argIndex)
		args = append(args, filter.ProviderID)
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

	var total int64
	err := s.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get consumer total count: %w", err)
	}

	return total, nil
}
