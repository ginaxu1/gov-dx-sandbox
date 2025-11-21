package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/models"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
	"github.com/lib/pq"
)

// postgresConsentEngine implements ConsentEngine interface using PostgreSQL
type postgresConsentEngine struct {
	db               *sql.DB
	consentPortalURL string
}

// NewPostgresConsentEngine creates a new PostgreSQL-based consent engine
func NewPostgresConsentEngine(db *sql.DB, consentPortalURL string) ConsentEngine {
	return &postgresConsentEngine{
		db:               db,
		consentPortalURL: consentPortalURL,
	}
}

// ProcessConsentRequest processes a consent request and creates a consent record
func (pce *postgresConsentEngine) ProcessConsentRequest(req models.ConsentRequest) (*models.ConsentRecord, error) {
	// Validate required fields
	if req.AppID == "" {
		return nil, fmt.Errorf("app_id is required")
	}

	if len(req.ConsentRequirements) == 0 {
		return nil, fmt.Errorf("consent_requirements is required")
	}

	// Use the first consent requirement for owner information
	// In the new format, owner_id is the email address
	firstRequirement := req.ConsentRequirements[0]
	ownerID := firstRequirement.OwnerID
	ownerEmail := firstRequirement.OwnerID // In new format, owner_id is the email

	// First, check for existing pending consent - only one pending record allowed per (appId, ownerId)
	existingPendingConsent, err := pce.findExistingPendingConsentByOwnerID(ownerID, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing pending consent: %w", err)
	}

	// If pending consent exists, update it and return (enforce only 1 pending per tuple)
	if existingPendingConsent != nil {
		slog.Info("Found existing pending consent record, updating with new request",
			"consent_id", existingPendingConsent.ConsentID,
			"owner_id", existingPendingConsent.OwnerID,
			"app_id", existingPendingConsent.AppID)

		// Update the existing pending consent record with new fields
		updatedConsent, err := pce.updateExistingConsentNewFormat(existingPendingConsent, req)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing pending consent: %w", err)
		}

		return updatedConsent, nil
	}

	// Check for existing non-pending consent (approved, rejected) for the same owner and app
	existingConsent, err := pce.findExistingConsentByOwnerID(ownerID, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing consent: %w", err)
	}

	// If existing non-pending consent found, update it and return
	if existingConsent != nil {
		slog.Info("Found existing non-pending consent record, updating with new request",
			"consent_id", existingConsent.ConsentID,
			"owner_id", existingConsent.OwnerID,
			"app_id", existingConsent.AppID,
			"current_status", existingConsent.Status)

		// Update the existing consent record with new fields
		updatedConsent, err := pce.updateExistingConsentNewFormat(existingConsent, req)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing consent: %w", err)
		}

		return updatedConsent, nil
	}

	// Create new consent record
	consentID := generateConsentID()
	now := time.Now()

	// Convert ConsentFields to string array for storage (fieldName format)
	var allFields []string
	for _, requirement := range req.ConsentRequirements {
		for _, field := range requirement.Fields {
			// Store as "fieldName" format
			allFields = append(allFields, field.FieldName)
		}
	}

	// Use default grant duration if not provided
	grantDuration := getDefaultGrantDuration(req.GrantDuration)

	// Calculate expires_at
	expiresAt, err := calculateExpiresAt(grantDuration, now)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expiry time: %w", err)
	}

	// Generate consent portal URL using the configured base URL
	consentPortalURL := fmt.Sprintf("%s/?consent_id=%s", pce.consentPortalURL, consentID)

	// Insert new consent record (session_id is optional, use empty string)
	insertSQL := `
		INSERT INTO consent_records (
			consent_id, owner_id, owner_email, app_id, status, type, 
			created_at, updated_at, expires_at, grant_duration, fields, 
			session_id, consent_portal_url, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = pce.db.Exec(insertSQL,
		consentID, ownerID, ownerEmail, req.AppID, string(models.StatusPending), "realtime",
		now, now, expiresAt, grantDuration, pq.Array(allFields),
		"", // Intentionally passing empty string for session_id in the new format (session information not required)
		consentPortalURL, ownerID)

	if err != nil {
		return nil, fmt.Errorf("failed to create consent record: %w", err)
	}

	record := &models.ConsentRecord{
		ConsentID:        consentID,
		OwnerID:          ownerID,
		OwnerEmail:       ownerEmail,
		AppID:            req.AppID,
		Status:           models.StatusPending,
		Type:             "realtime",
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        expiresAt,
		GrantDuration:    grantDuration,
		Fields:           allFields,
		SessionID:        "",
		ConsentPortalURL: consentPortalURL,
		UpdatedBy:        ownerID,
	}

	slog.Info("Consent record created",
		"consent_id", record.ConsentID,
		"owner_id", record.OwnerID,
		"owner_email", record.OwnerEmail,
		"app_id", record.AppID)

	return record, nil
}

// CreateConsent creates a new consent record (alias for ProcessConsentRequest)
func (pce *postgresConsentEngine) CreateConsent(req models.ConsentRequest) (*models.ConsentRecord, error) {
	return pce.ProcessConsentRequest(req)
}

// GetConsentStatus retrieves a consent record by ID
func (pce *postgresConsentEngine) GetConsentStatus(consentID string) (*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE consent_id = $1
	`

	row := pce.db.QueryRow(querySQL, consentID)

	var record models.ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
		&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("consent record with ID '%s' not found", consentID)
		}
		return nil, fmt.Errorf("failed to retrieve consent record: %w", err)
	}

	return &record, nil
}

// UpdateConsent updates a consent record
func (pce *postgresConsentEngine) UpdateConsent(consentID string, req models.UpdateConsentRequest) (*models.ConsentRecord, error) {
	// Get existing record
	existingRecord, err := pce.GetConsentStatus(consentID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !isValidStatusTransition(existingRecord.Status, req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", string(existingRecord.Status), string(req.Status))
	}

	// Update the record
	now := time.Now()

	// Update grant duration if provided, otherwise use existing or default
	grantDuration := req.GrantDuration
	if grantDuration == "" {
		grantDuration = getDefaultGrantDuration(existingRecord.GrantDuration)
	}

	// Update fields if provided
	fields := req.Fields
	if len(fields) == 0 {
		fields = existingRecord.Fields
	}

	// Recalculate expires_at based on grant_duration and updated_at
	expiresAt, err := calculateExpiresAt(grantDuration, now)
	if err != nil {
		return nil, err
	}

	updateSQL := `
		UPDATE consent_records 
		SET status = $1, updated_at = $2, expires_at = $3, grant_duration = $4, 
		    fields = $5, updated_by = $6
		WHERE consent_id = $7
	`

	_, err = pce.db.Exec(updateSQL,
		string(req.Status), now, expiresAt, grantDuration,
		pq.Array(fields), req.UpdatedBy, consentID)

	if err != nil {
		return nil, fmt.Errorf("failed to update consent record: %w", err)
	}

	// Return updated record
	updatedRecord := *existingRecord
	updatedRecord.Status = req.Status
	updatedRecord.UpdatedAt = now
	updatedRecord.ExpiresAt = expiresAt
	updatedRecord.GrantDuration = grantDuration
	updatedRecord.Fields = fields
	updatedRecord.UpdatedBy = req.UpdatedBy

	slog.Info("Consent record updated",
		"consent_id", updatedRecord.ConsentID,
		"owner_id", updatedRecord.OwnerID,
		"status", updatedRecord.Status)

	return &updatedRecord, nil
}

// RevokeConsent revokes a consent record
func (pce *postgresConsentEngine) RevokeConsent(consentID string, reason string) (*models.ConsentRecord, error) {
	updateReq := models.UpdateConsentRequest{
		Status:    models.StatusRevoked,
		UpdatedBy: "system", // Could be enhanced to get from context
		Reason:    reason,
	}
	return pce.UpdateConsent(consentID, updateReq)
}

// CheckConsentExpiry checks for and updates expired consent records
func (pce *postgresConsentEngine) CheckConsentExpiry() ([]*models.ConsentRecord, error) {
	now := time.Now()

	// Find expired approved records
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE expires_at < $1 AND status = 'approved'
	`

	rows, err := pce.db.Query(querySQL, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired records: %w", err)
	}
	defer rows.Close()

	var deletedRecords []*models.ConsentRecord

	for rows.Next() {
		var record models.ConsentRecord

		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
			&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

		if err != nil {
			return nil, fmt.Errorf("failed to scan expired record: %w", err)
		}

		// Delete the expired record
		deleteSQL := `DELETE FROM consent_records WHERE consent_id = $1`
		_, err = pce.db.Exec(deleteSQL, record.ConsentID)
		if err != nil {
			slog.Error("Failed to delete expired consent", "consent_id", record.ConsentID, "error", err)
			continue
		}

		deletedRecords = append(deletedRecords, &record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired records: %w", err)
	}

	if len(deletedRecords) > 0 {
		slog.Info("Deleted expired consent records", "count", len(deletedRecords))
	}
	return deletedRecords, nil
}

// StartBackgroundExpiryProcess starts the background process for checking consent expiry
func (pce *postgresConsentEngine) StartBackgroundExpiryProcess(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.Info("Background expiry process started", "interval", interval)

		for {
			select {
			case <-ticker.C:
				_, err := pce.CheckConsentExpiry()
				if err != nil {
					slog.Error("Error checking consent expiry", "error", err)
				}
			case <-ctx.Done():
				slog.Info("Background expiry process stopped due to context cancellation")
				return
			}
		}
	}()
}

// StopBackgroundExpiryProcess stops the background expiry process
func (pce *postgresConsentEngine) StopBackgroundExpiryProcess() {
	// For PostgreSQL implementation, we don't need to stop anything
	// as the goroutine will stop when the application shuts down
	slog.Info("Background expiry process stopped")
}

// findExistingConsentByOwnerID finds existing consent for the given owner_id and app
// Returns the most recent non-expired consent record (pending, approved, or rejected)
func (pce *postgresConsentEngine) findExistingConsentByOwnerID(ownerID, appID string) (*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_id = $1 AND app_id = $2 
		AND status IN ('pending', 'approved', 'rejected')
		AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, ownerID, appID)

	var record models.ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
		&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No existing consent found
		}
		return nil, fmt.Errorf("failed to query existing consent: %w", err)
	}

	return &record, nil
}

// findExistingPendingConsentByOwnerID finds existing pending consent for the given owner_id and app
// This enforces the constraint that only one pending record can exist per (appId, ownerId) tuple
func (pce *postgresConsentEngine) findExistingPendingConsentByOwnerID(ownerID, appID string) (*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_id = $1 AND app_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, ownerID, appID)

	var record models.ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
		&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No existing pending consent found
		}
		return nil, fmt.Errorf("failed to query existing consent: %w", err)
	}

	return &record, nil
}

// updateExistingConsentNewFormat updates an existing consent record with new format (consent_requirements)
func (pce *postgresConsentEngine) updateExistingConsentNewFormat(existingConsent *models.ConsentRecord, req models.ConsentRequest) (*models.ConsentRecord, error) {
	// Convert ConsentFields to string array for storage (fieldName format)
	var allFields []string
	for _, requirement := range req.ConsentRequirements {
		for _, field := range requirement.Fields {
			// Store as "fieldName" format
			allFields = append(allFields, field.FieldName)
		}
	}

	// Use default grant duration if not provided
	grantDuration := getDefaultGrantDuration(req.GrantDuration)

	// Calculate new expires_at
	now := time.Now()
	expiresAt, err := calculateExpiresAt(grantDuration, now)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expiry time: %w", err)
	}

	// Update the existing consent record (session_id is not in new format, keep existing or empty)
	updateSQL := `
		UPDATE consent_records 
		SET fields = $1, updated_at = $2, grant_duration = $3, expires_at = $4, updated_by = $5
		WHERE consent_id = $6
	`

	_, err = pce.db.Exec(updateSQL,
		pq.Array(allFields), now, grantDuration, expiresAt, existingConsent.OwnerID, existingConsent.ConsentID)

	if err != nil {
		return nil, fmt.Errorf("failed to update existing consent record: %w", err)
	}

	// Return updated consent record
	updatedRecord := *existingConsent
	updatedRecord.Fields = allFields
	updatedRecord.GrantDuration = grantDuration
	updatedRecord.ExpiresAt = expiresAt
	updatedRecord.UpdatedAt = now

	return &updatedRecord, nil
}

// FindExistingConsent finds an existing consent record by consumer app ID and owner ID
func (pce *postgresConsentEngine) FindExistingConsent(consumerAppID, ownerID string) *models.ConsentRecord {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE app_id = $1 AND owner_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, consumerAppID, ownerID)

	var record models.ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
		&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		slog.Error("Failed to query existing consent", "error", err)
		return nil
	}

	return &record
}

// ProcessConsentPortalRequest handles consent portal interactions
func (pce *postgresConsentEngine) ProcessConsentPortalRequest(req models.ConsentPortalRequest) (*models.ConsentRecord, error) {
	// Map action to status
	var status models.ConsentStatus
	switch req.Action {
	case "approve":
		status = models.StatusApproved
	case "deny":
		status = models.StatusRejected
	case "revoke":
		status = models.StatusRevoked
	default:
		return nil, fmt.Errorf("invalid action: %s", req.Action)
	}

	// Update the record based on portal action
	updateReq := models.UpdateConsentRequest{
		Status:    status,
		UpdatedBy: req.DataOwner,
		Reason:    req.Reason,
	}

	return pce.UpdateConsent(req.ConsentID, updateReq)
}

// GetConsentsByDataOwner retrieves all consent records for a data owner
func (pce *postgresConsentEngine) GetConsentsByDataOwner(dataOwner string) ([]*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := pce.db.Query(querySQL, dataOwner)
	if err != nil {
		return nil, fmt.Errorf("failed to query consents by data owner: %w", err)
	}
	defer rows.Close()

	var records []*models.ConsentRecord

	for rows.Next() {
		var record models.ConsentRecord
		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
			&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

		if err != nil {
			return nil, fmt.Errorf("failed to scan consent record: %w", err)
		}

		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating consent records: %w", err)
	}

	return records, nil
}

// GetConsentsByConsumer retrieves all consent records for a consumer
func (pce *postgresConsentEngine) GetConsentsByConsumer(consumer string) ([]*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE app_id = $1
		ORDER BY created_at DESC
	`

	rows, err := pce.db.Query(querySQL, consumer)
	if err != nil {
		return nil, fmt.Errorf("failed to query consents by consumer: %w", err)
	}
	defer rows.Close()

	var records []*models.ConsentRecord

	for rows.Next() {
		var record models.ConsentRecord
		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&record.ExpiresAt, &record.GrantDuration, pq.Array(&record.Fields),
			&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

		if err != nil {
			return nil, fmt.Errorf("failed to scan consent record: %w", err)
		}

		records = append(records, &record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating consent records: %w", err)
	}

	return records, nil
}

// Utility functions for consent management

// generateConsentID generates a unique consent ID
func generateConsentID() string {
	return fmt.Sprintf("consent_%s", uuid.New().String()[:8])
}

// getDefaultGrantDuration returns the default grant duration
func getDefaultGrantDuration(duration string) string {
	if duration == "" {
		return "1h" // Default to 1 hour
	}
	return duration
}

// calculateExpiresAt calculates the expiry time based on grant duration
func calculateExpiresAt(grantDuration string, createdAt time.Time) (time.Time, error) {
	duration, err := parseISODuration(grantDuration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid grant duration format: %w", err)
	}
	return createdAt.Add(duration), nil
}

// parseISODuration parses an ISO 8601 duration string and returns the duration
// Supports formats like: P30D, P1M, P1Y, PT1H, PT30M, P1Y2M3DT4H5M6S
func parseISODuration(duration string) (time.Duration, error) {
	if duration == "" {
		// Default to 1 hour if no duration specified
		return time.Hour, nil
	}

	// Check if it's ISO 8601 format (starts with 'P')
	if len(duration) > 0 && duration[0] == 'P' {
		return parseISO8601Duration(duration)
	}

	// Fallback to legacy format parsing
	return utils.ParseExpiryTime(duration)
}

// parseISO8601Duration parses an ISO 8601 duration string into a time.Duration
func parseISO8601Duration(duration string) (time.Duration, error) {
	// Validate ISO 8601 duration format
	if !isValidISODuration(duration) {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", duration)
	}

	// Remove the 'P' prefix
	if len(duration) == 0 || duration[0] != 'P' {
		return 0, fmt.Errorf("duration must start with 'P'")
	}
	duration = duration[1:]

	var total time.Duration
	var err error

	// Check if there's a time component (starts with 'T')
	timeIndex := -1
	for i, char := range duration {
		if char == 'T' {
			timeIndex = i
			break
		}
	}

	// Parse date components (before 'T' or entire string if no 'T')
	datePart := duration
	if timeIndex != -1 {
		datePart = duration[:timeIndex]
	}

	// Parse years
	years, datePart, err := parseComponent(datePart, "Y")
	if err != nil {
		return 0, err
	}
	total += time.Duration(years) * 365 * 24 * time.Hour

	// Parse months
	months, datePart, err := parseComponent(datePart, "M")
	if err != nil {
		return 0, err
	}
	total += time.Duration(months) * 30 * 24 * time.Hour // Approximate month as 30 days

	// Parse days
	days, _, err := parseComponent(datePart, "D")
	if err != nil {
		return 0, err
	}
	total += time.Duration(days) * 24 * time.Hour

	// Parse time components (after 'T')
	if timeIndex != -1 {
		timePart := duration[timeIndex+1:]

		// Parse hours
		hours, timePart, err := parseComponent(timePart, "H")
		if err != nil {
			return 0, err
		}
		total += time.Duration(hours) * time.Hour

		// Parse minutes
		minutes, timePart, err := parseComponent(timePart, "M")
		if err != nil {
			return 0, err
		}
		total += time.Duration(minutes) * time.Minute

		// Parse seconds
		seconds, _, err := parseComponent(timePart, "S")
		if err != nil {
			return 0, err
		}
		total += time.Duration(seconds) * time.Second
	}

	return total, nil
}

// isValidISODuration validates if a string is a valid ISO 8601 duration
func isValidISODuration(duration string) bool {
	// ISO 8601 duration pattern: P(\d+Y)?(\d+M)?(\d+D)?(T(\d+H)?(\d+M)?(\d+(\.\d+)?S)?)?
	pattern := `^P(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`
	matched, _ := regexp.MatchString(pattern, duration)
	return matched
}

// parseComponent extracts a numeric component from a duration string
func parseComponent(duration, suffix string) (int, string, error) {
	pattern := fmt.Sprintf(`(\d+)%s`, suffix)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(duration)

	if len(matches) == 0 {
		return 0, duration, nil
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, duration, err
	}

	// Remove the matched part from the duration string
	remaining := re.ReplaceAllString(duration, "")
	return value, remaining, nil
}

// getAllFields extracts all fields from data fields
func getAllFields(dataFields []models.DataField) []string {
	var allFields []string
	for _, field := range dataFields {
		allFields = append(allFields, field.Fields...)
	}
	return allFields
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(current, new models.ConsentStatus) bool {
	validTransitions := map[models.ConsentStatus][]models.ConsentStatus{
		models.StatusPending:  {models.StatusApproved, models.StatusRejected, models.StatusExpired},                // Initial decision
		models.StatusApproved: {models.StatusApproved, models.StatusRejected, models.StatusRevoked, models.StatusExpired}, // Direct approval flow: approved->approved (success), approved->rejected (direct rejection), approved->revoked (user revocation), approved->expired (expiry)
		models.StatusRejected: {models.StatusExpired},                                                // Terminal state - can only transition to expired
		models.StatusExpired:  {models.StatusExpired},                                                // Terminal state - can only stay expired
		models.StatusRevoked:  {models.StatusExpired},                                                // Terminal state - can only transition to expired
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}
