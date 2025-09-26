package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
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

// GetAuditLogsSummaryForProvider retrieves simplified audit logs for a specific provider
func (s *AuditService) GetAuditLogsSummaryForProvider(providerID string, limit, offset int) ([]*models.AuditLogSummaryResponse, error) {
	logs, err := s.GetAuditLogsByProviderID(providerID, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.ConvertToSummaryResponse(logs), nil
}

// GetAuditLogsSummaryForAdmin retrieves simplified audit logs for admin oversight
func (s *AuditService) GetAuditLogsSummaryForAdmin(limit, offset int) ([]*models.AuditLogSummaryResponse, error) {
	logs, err := s.GetAuditLogsForAdmin(limit, offset)
	if err != nil {
		return nil, err
	}

	return s.ConvertToSummaryResponse(logs), nil
}

// GetAuditLogsSummaryForCitizen retrieves simplified audit logs for a specific citizen
func (s *AuditService) GetAuditLogsSummaryForCitizen(citizenHash string, limit, offset int) ([]*models.AuditLogSummaryResponse, error) {
	logs, err := s.GetAuditLogsByCitizenHash(citizenHash, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.ConvertToSummaryResponse(logs), nil
}

// ConvertToSummaryResponse converts detailed audit logs to simplified summary format
func (s *AuditService) ConvertToSummaryResponse(logs []*models.AuditLogResponse) []*models.AuditLogSummaryResponse {
	var summaryLogs []*models.AuditLogSummaryResponse

	for _, log := range logs {
		summary := &models.AuditLogSummaryResponse{
			ConsumerApp: log.ConsumerID,
			Citizen:     log.CitizenHash,
			Providers:   []string{log.ProviderID},
			Timestamp:   log.Timestamp.Format(time.RFC3339),
			Status:      log.TransactionStatus,
		}

		// Extract fields from requested_data
		fields := s.extractFieldsFromRequestedData(log.RequestedData)
		summary.Fields = fields

		summaryLogs = append(summaryLogs, summary)
	}

	return summaryLogs
}

// extractFieldsFromRequestedData extracts field names from GraphQL query in requested_data
func (s *AuditService) extractFieldsFromRequestedData(requestedData json.RawMessage) []string {
	var fields []string

	// Parse the requested data JSON
	var data map[string]interface{}
	if err := json.Unmarshal(requestedData, &data); err != nil {
		slog.Warn("Failed to parse requested_data JSON", "error", err)
		return fields
	}

	// Extract query from the data
	query, ok := data["query"].(string)
	if !ok {
		return fields
	}

	// Extract field names from GraphQL query
	fields = s.extractFieldsFromGraphQLQuery(query)

	return fields
}

// extractFieldsFromGraphQLQuery extracts field names from a GraphQL query string
func (s *AuditService) extractFieldsFromGraphQLQuery(query string) []string {
	var fields []string
	fieldSet := make(map[string]bool)

	// Remove comments and normalize whitespace
	query = s.removeGraphQLComments(query)
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")

	// Find field patterns in GraphQL query
	// Pattern 1: fieldName { ... } - nested fields
	nestedFieldRegex := regexp.MustCompile(`(\w+)\s*\{`)
	matches := nestedFieldRegex.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if len(match) > 1 {
			fieldSet[match[1]] = true
		}
	}

	// Pattern 2: fieldName - simple fields (not inside braces)
	// This is more complex as we need to avoid fields inside nested objects
	simpleFieldRegex := regexp.MustCompile(`\b(\w+)(?:\s*\([^)]*\))?\s*(?:\{|,|$)`)
	allMatches := simpleFieldRegex.FindAllStringSubmatch(query, -1)

	// Filter out GraphQL keywords and fields that are likely nested
	graphqlKeywords := map[string]bool{
		"query": true, "mutation": true, "subscription": true,
		"fragment": true, "on": true, "type": true, "interface": true,
		"union": true, "enum": true, "input": true, "scalar": true,
		"directive": true, "schema": true, "extend": true,
	}

	for _, match := range allMatches {
		if len(match) > 1 {
			fieldName := match[1]
			if !graphqlKeywords[fieldName] && !fieldSet[fieldName] {
				// Check if this field is not inside a nested object
				if s.isTopLevelField(query, fieldName) {
					fieldSet[fieldName] = true
				}
			}
		}
	}

	// Convert set to slice
	for field := range fieldSet {
		fields = append(fields, field)
	}

	return fields
}

// removeGraphQLComments removes comments from GraphQL query
func (s *AuditService) removeGraphQLComments(query string) string {
	// Remove # comments
	lines := strings.Split(query, "\n")
	var cleanLines []string
	for _, line := range lines {
		if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
			line = line[:commentIndex]
		}
		cleanLines = append(cleanLines, strings.TrimSpace(line))
	}
	return strings.Join(cleanLines, "\n")
}

// isTopLevelField checks if a field is at the top level of the query
func (s *AuditService) isTopLevelField(query, fieldName string) bool {
	// Find all occurrences of the field name
	fieldRegex := regexp.MustCompile(`\b` + regexp.QuoteMeta(fieldName) + `\b`)
	matches := fieldRegex.FindAllStringIndex(query, -1)

	for _, match := range matches {
		start := match[0]

		// Count opening and closing braces before this position
		beforeText := query[:start]
		openBraces := strings.Count(beforeText, "{")
		closeBraces := strings.Count(beforeText, "}")

		// If we're at the top level (no unmatched opening braces)
		if openBraces == closeBraces {
			return true
		}
	}

	return false
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
