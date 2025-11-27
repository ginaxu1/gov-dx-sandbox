package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/lib/pq"
)

// PostgresConsentEngine implements ConsentEngine interface using PostgreSQL
type PostgresConsentEngine struct {
	db               *sql.DB
	consentPortalURL string
}

// NewPostgresConsentEngine creates a new PostgreSQL-based consent engine
func NewPostgresConsentEngine(db *sql.DB, consentPortalURL string) ConsentEngine {
	return &PostgresConsentEngine{
		db:               db,
		consentPortalURL: consentPortalURL,
	}
}

// ProcessConsentRequest processes a consent request and creates a consent record
func (pce *PostgresConsentEngine) ProcessConsentRequest(req models.ConsentRequest) (*models.ConsentRecord, error) {
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
	var consentFields []models.ConsentField
	for _, requirement := range req.ConsentRequirements {
		for _, field := range requirement.Fields {
			// Store as "fieldName" format
			allFields = append(allFields, field.FieldName)
			consentFields = append(consentFields, field)
		}
	}

	// Use default grant duration if not provided
	var grantDurationStr string
	if req.GrantDuration != nil {
		grantDurationStr = *req.GrantDuration
	}
	grantDuration := getDefaultGrantDuration(grantDurationStr)

	// Calculate pending_expires_at for pending status
	pendingExpiresAt, err := calculateExpiresAt(models.DefaultPendingTimeoutDuration, now)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate pending expiry time: %w", err)
	}

	// Generate consent portal URL using the configured base URL
	consentPortalURL := fmt.Sprintf("%s/?consent_id=%s", pce.consentPortalURL, consentID)

	// Insert new consent record (session_id is optional, use empty string)
	insertSQL := `
		INSERT INTO consent_records (
			consent_id, owner_id, owner_email, app_id, status, type, 
			created_at, updated_at, pending_expires_at, grant_duration, fields, 
			session_id, consent_portal_url, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	updatedByStr := ownerID
	_, err = pce.db.Exec(insertSQL,
		consentID, ownerID, ownerEmail, req.AppID, string(models.StatusPending), "realtime",
		now, now, pendingExpiresAt, grantDuration, pq.Array(allFields),
		"", // Intentionally passing empty string for session_id in the new format (session information not required)
		consentPortalURL, updatedByStr)

	if err != nil {
		return nil, fmt.Errorf("failed to create consent record: %w", err)
	}

	record := &models.ConsentRecord{
		ConsentID:        consentID,
		OwnerID:          ownerID,
		OwnerEmail:       ownerEmail,
		AppID:            req.AppID,
		Status:           string(models.StatusPending),
		Type:             "realtime",
		CreatedAt:        now,
		UpdatedAt:        now,
		PendingExpiresAt: &pendingExpiresAt,
		GrantDuration:    grantDuration,
		Fields:           consentFields,
		SessionID:        "",
		ConsentPortalURL: consentPortalURL,
		UpdatedBy:        &updatedByStr,
	}

	slog.Info("Consent record created",
		"consent_id", record.ConsentID,
		"owner_id", record.OwnerID,
		"owner_email", record.OwnerEmail,
		"app_id", record.AppID)

	return record, nil
}

// CreateConsent creates a new consent record (alias for ProcessConsentRequest)
func (pce *PostgresConsentEngine) CreateConsent(req models.ConsentRequest) (*models.ConsentRecord, error) {
	return pce.ProcessConsentRequest(req)
}

// GetConsentStatus retrieves a consent record by ID
func (pce *PostgresConsentEngine) GetConsentStatus(consentID string) (*models.ConsentRecord, error) {
	// Parse string ID to UUID
	consentUUID, err := uuid.Parse(consentID)
	if err != nil {
		return nil, fmt.Errorf("invalid consent ID format: %w", err)
	}

	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE consent_id = $1
	`

	row := pce.db.QueryRow(querySQL, consentUUID)

	var record models.ConsentRecord
	var pendingExpiresAt, grantExpiresAt sql.NullTime
	var updatedBy sql.NullString
	var fieldsArray []string

	err = row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
		&record.SessionID, &record.ConsentPortalURL, &updatedBy)

	if pendingExpiresAt.Valid {
		record.PendingExpiresAt = &pendingExpiresAt.Time
	}
	if grantExpiresAt.Valid {
		record.GrantExpiresAt = &grantExpiresAt.Time
	}
	if updatedBy.Valid {
		updatedByStr := updatedBy.String
		record.UpdatedBy = &updatedByStr
	}

	// Convert string array to ConsentField array (simplified - just field names)
	record.Fields = make([]models.ConsentField, len(fieldsArray))
	for i, fieldName := range fieldsArray {
		record.Fields[i] = models.ConsentField{
			FieldName: fieldName,
		}
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("consent record with ID '%s' not found", consentID)
		}
		return nil, fmt.Errorf("failed to retrieve consent record: %w", err)
	}

	return &record, nil
}

// UpdateConsent updates a consent record
func (pce *PostgresConsentEngine) UpdateConsent(consentID string, req models.UpdateConsentRequest) (*models.ConsentRecord, error) {
	// Get existing record
	existingRecord, err := pce.GetConsentStatus(consentID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !isValidStatusTransition(models.ConsentStatus(existingRecord.Status), req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", existingRecord.Status, string(req.Status))
	}

	// Update the record
	now := time.Now()

	// Update grant duration if provided, otherwise use existing or default
	var grantDuration string
	if req.GrantDuration != nil && *req.GrantDuration != "" {
		grantDuration = *req.GrantDuration
	} else {
		grantDuration = getDefaultGrantDuration(existingRecord.GrantDuration)
	}

	// Parse string ID to UUID
	consentUUID, err := uuid.Parse(consentID)
	if err != nil {
		return nil, fmt.Errorf("invalid consent ID format: %w", err)
	}

	// Convert fields to string array for database storage
	var fieldsArray []string
	if req.Fields != nil && len(*req.Fields) > 0 {
		// Convert ConsentField array to string array
		for _, field := range *req.Fields {
			fieldsArray = append(fieldsArray, field.FieldName)
		}
	} else {
		// Convert existing ConsentField array to string array
		for _, field := range existingRecord.Fields {
			fieldsArray = append(fieldsArray, field.FieldName)
		}
	}

	// Determine which expiry field to set based on new status
	var pendingExpiresAt *time.Time
	var grantExpiresAt *time.Time

	if req.Status == models.StatusPending {
		// Set pending expiry
		expiresAt, err := calculateExpiresAt(models.DefaultPendingTimeoutDuration, now)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate pending expiry time: %w", err)
		}
		pendingExpiresAt = &expiresAt
	} else if req.Status == models.StatusApproved {
		// Set grant expiry
		expiresAt, err := calculateExpiresAt(grantDuration, now)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate grant expiry time: %w", err)
		}
		grantExpiresAt = &expiresAt
		// Clear pending expiry when approved
		pendingExpiresAt = nil
	} else {
		// For other statuses, clear both expiries
		pendingExpiresAt = nil
		grantExpiresAt = nil
	}

	updateSQL := `
		UPDATE consent_records 
		SET status = $1, updated_at = $2, pending_expires_at = $3, grant_expires_at = $4, 
		    grant_duration = $5, fields = $6, updated_by = $7
		WHERE consent_id = $8
	`

	var updatedByStr string
	if req.UpdatedBy != nil {
		updatedByStr = *req.UpdatedBy
	} else {
		updatedByStr = existingRecord.OwnerID // Default to owner ID if not provided
	}
	_, err = pce.db.Exec(updateSQL,
		string(req.Status), now, pendingExpiresAt, grantExpiresAt, grantDuration,
		pq.Array(fieldsArray), updatedByStr, consentUUID)

	if err != nil {
		return nil, fmt.Errorf("failed to update consent record: %w", err)
	}

	// Convert fieldsArray back to ConsentField array
	fields := make([]models.ConsentField, len(fieldsArray))
	for i, fieldName := range fieldsArray {
		fields[i] = models.ConsentField{
			FieldName: fieldName,
		}
	}

	// Return updated record
	updatedRecord := *existingRecord
	updatedRecord.Status = string(req.Status)
	updatedRecord.UpdatedAt = now
	updatedRecord.PendingExpiresAt = pendingExpiresAt
	updatedRecord.GrantExpiresAt = grantExpiresAt
	updatedRecord.GrantDuration = grantDuration
	updatedRecord.Fields = fields
	if req.UpdatedBy != nil {
		updatedRecord.UpdatedBy = req.UpdatedBy
	}

	slog.Info("Consent record updated",
		"consent_id", updatedRecord.ConsentID,
		"owner_id", updatedRecord.OwnerID,
		"status", updatedRecord.Status)

	return &updatedRecord, nil
}

// RevokeConsent revokes a consent record
func (pce *PostgresConsentEngine) RevokeConsent(consentID string, reason string) (*models.ConsentRecord, error) {
	updatedBy := "system" // Could be enhanced to get from context
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	updateReq := models.UpdateConsentRequest{
		Status:    models.StatusRevoked,
		UpdatedBy: &updatedBy,
		Reason:    reasonPtr,
	}
	return pce.UpdateConsent(consentID, updateReq)
}

// CheckConsentExpiry checks for and updates expired consent records
func (pce *PostgresConsentEngine) CheckConsentExpiry() ([]*models.ConsentRecord, error) {
	now := time.Now()

	// Find expired records: pending records with expired pending_expires_at, or approved records with expired grant_expires_at
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE (status = 'pending' AND pending_expires_at < $1)
		   OR (status = 'approved' AND grant_expires_at < $1)
	`

	rows, err := pce.db.Query(querySQL, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired records: %w", err)
	}
	defer rows.Close()

	var deletedRecords []*models.ConsentRecord

	for rows.Next() {
		var record models.ConsentRecord
		var pendingExpiresAt, grantExpiresAt sql.NullTime
		var updatedBy sql.NullString
		var fieldsArray []string

		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
			&record.SessionID, &record.ConsentPortalURL, &updatedBy)

		if pendingExpiresAt.Valid {
			record.PendingExpiresAt = &pendingExpiresAt.Time
		}
		if grantExpiresAt.Valid {
			record.GrantExpiresAt = &grantExpiresAt.Time
		}
		if updatedBy.Valid {
			updatedByStr := updatedBy.String
			record.UpdatedBy = &updatedByStr
		}

		// Convert string array to ConsentField array
		record.Fields = make([]models.ConsentField, len(fieldsArray))
		for i, fieldName := range fieldsArray {
			record.Fields[i] = models.ConsentField{
				FieldName: fieldName,
			}
		}

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
func (pce *PostgresConsentEngine) StartBackgroundExpiryProcess(ctx context.Context, interval time.Duration) {
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
func (pce *PostgresConsentEngine) StopBackgroundExpiryProcess() {
	// For PostgreSQL implementation, we don't need to stop anything
	// as the goroutine will stop when the application shuts down
	slog.Info("Background expiry process stopped")
}

// findExistingConsentByOwnerID finds existing consent for the given owner_id and app
// Returns the most recent non-expired consent record (pending, approved, or rejected)
func (pce *PostgresConsentEngine) findExistingConsentByOwnerID(ownerID, appID string) (*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_id = $1 AND app_id = $2 
		AND status IN ('pending', 'approved', 'rejected')
		AND ((status = 'pending' AND pending_expires_at > NOW())
		     OR (status = 'approved' AND grant_expires_at > NOW())
		     OR (status = 'rejected'))
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, ownerID, appID)

	var record models.ConsentRecord
	var pendingExpiresAt, grantExpiresAt sql.NullTime
	var updatedBy sql.NullString
	var fieldsArray []string

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
		&record.SessionID, &record.ConsentPortalURL, &updatedBy)

	if pendingExpiresAt.Valid {
		record.PendingExpiresAt = &pendingExpiresAt.Time
	}
	if grantExpiresAt.Valid {
		record.GrantExpiresAt = &grantExpiresAt.Time
	}
	if updatedBy.Valid {
		updatedByStr := updatedBy.String
		record.UpdatedBy = &updatedByStr
	}

	// Convert string array to ConsentField array
	record.Fields = make([]models.ConsentField, len(fieldsArray))
	for i, fieldName := range fieldsArray {
		record.Fields[i] = models.ConsentField{
			FieldName: fieldName,
		}
	}

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
func (pce *PostgresConsentEngine) findExistingPendingConsentByOwnerID(ownerID, appID string) (*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_id = $1 AND app_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, ownerID, appID)

	var record models.ConsentRecord
	var pendingExpiresAt, grantExpiresAt sql.NullTime
	var updatedBy sql.NullString
	var fieldsArray []string

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
		&record.SessionID, &record.ConsentPortalURL, &updatedBy)

	if pendingExpiresAt.Valid {
		record.PendingExpiresAt = &pendingExpiresAt.Time
	}
	if grantExpiresAt.Valid {
		record.GrantExpiresAt = &grantExpiresAt.Time
	}
	if updatedBy.Valid {
		updatedByStr := updatedBy.String
		record.UpdatedBy = &updatedByStr
	}

	// Convert string array to ConsentField array
	record.Fields = make([]models.ConsentField, len(fieldsArray))
	for i, fieldName := range fieldsArray {
		record.Fields[i] = models.ConsentField{
			FieldName: fieldName,
		}
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No existing pending consent found
		}
		return nil, fmt.Errorf("failed to query existing consent: %w", err)
	}

	return &record, nil
}

// updateExistingConsentNewFormat updates an existing consent record with new format (consent_requirements)
func (pce *PostgresConsentEngine) updateExistingConsentNewFormat(existingConsent *models.ConsentRecord, req models.ConsentRequest) (*models.ConsentRecord, error) {
	// Convert ConsentFields to string array for storage (fieldName format)
	var allFields []string
	var consentFields []models.ConsentField
	for _, requirement := range req.ConsentRequirements {
		for _, field := range requirement.Fields {
			// Store as "fieldName" format
			allFields = append(allFields, field.FieldName)
			consentFields = append(consentFields, field)
		}
	}

	// Use default grant duration if not provided
	var grantDurationStr string
	if req.GrantDuration != nil {
		grantDurationStr = *req.GrantDuration
	}
	grantDuration := getDefaultGrantDuration(grantDurationStr)

	// Calculate pending_expires_at if status is pending
	now := time.Now()
	var pendingExpiresAt *time.Time
	if existingConsent.Status == string(models.StatusPending) {
		expiresAt, err := calculateExpiresAt(models.DefaultPendingTimeoutDuration, now)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate pending expiry time: %w", err)
		}
		pendingExpiresAt = &expiresAt
	}

	// Update the existing consent record (session_id is not in new format, keep existing or empty)
	updateSQL := `
		UPDATE consent_records 
		SET fields = $1, updated_at = $2, grant_duration = $3, pending_expires_at = $4, updated_by = $5
		WHERE consent_id = $6
	`

	updatedByStr := existingConsent.OwnerID
	if existingConsent.UpdatedBy != nil {
		updatedByStr = *existingConsent.UpdatedBy
	}

	_, err := pce.db.Exec(updateSQL,
		pq.Array(allFields), now, grantDuration, pendingExpiresAt, updatedByStr, existingConsent.ConsentID)

	if err != nil {
		return nil, fmt.Errorf("failed to update existing consent record: %w", err)
	}

	// Return updated consent record
	updatedRecord := *existingConsent
	updatedRecord.Fields = consentFields
	updatedRecord.GrantDuration = grantDuration
	updatedRecord.PendingExpiresAt = pendingExpiresAt
	updatedRecord.UpdatedAt = now

	return &updatedRecord, nil
}

// FindExistingConsent finds an existing consent record by consumer app ID and owner ID
func (pce *PostgresConsentEngine) FindExistingConsent(consumerAppID, ownerID string) *models.ConsentRecord {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE app_id = $1 AND owner_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, consumerAppID, ownerID)

	var record models.ConsentRecord
	var pendingExpiresAt, grantExpiresAt sql.NullTime
	var updatedBy sql.NullString
	var fieldsArray []string

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
		&record.SessionID, &record.ConsentPortalURL, &updatedBy)

	if pendingExpiresAt.Valid {
		record.PendingExpiresAt = &pendingExpiresAt.Time
	}
	if grantExpiresAt.Valid {
		record.GrantExpiresAt = &grantExpiresAt.Time
	}
	if updatedBy.Valid {
		updatedByStr := updatedBy.String
		record.UpdatedBy = &updatedByStr
	}

	// Convert string array to ConsentField array
	record.Fields = make([]models.ConsentField, len(fieldsArray))
	for i, fieldName := range fieldsArray {
		record.Fields[i] = models.ConsentField{
			FieldName: fieldName,
		}
	}

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
func (pce *PostgresConsentEngine) ProcessConsentPortalRequest(req models.ConsentPortalRequest) (*models.ConsentRecord, error) {
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
		UpdatedBy: &req.DataOwner,
		Reason:    req.Reason,
	}

	// Convert UUID to string for the interface
	return pce.UpdateConsent(req.ConsentID.String(), updateReq)
}

// GetConsentsByDataOwner retrieves all consent records for a data owner
func (pce *PostgresConsentEngine) GetConsentsByDataOwner(dataOwner string) ([]*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
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
		var pendingExpiresAt, grantExpiresAt sql.NullTime
		var updatedBy sql.NullString
		var fieldsArray []string

		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
			&record.SessionID, &record.ConsentPortalURL, &updatedBy)

		if pendingExpiresAt.Valid {
			record.PendingExpiresAt = &pendingExpiresAt.Time
		}
		if grantExpiresAt.Valid {
			record.GrantExpiresAt = &grantExpiresAt.Time
		}
		if updatedBy.Valid {
			updatedByStr := updatedBy.String
			record.UpdatedBy = &updatedByStr
		}

		// Convert string array to ConsentField array
		record.Fields = make([]models.ConsentField, len(fieldsArray))
		for i, fieldName := range fieldsArray {
			record.Fields[i] = models.ConsentField{
				FieldName: fieldName,
			}
		}

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
func (pce *PostgresConsentEngine) GetConsentsByConsumer(consumer string) ([]*models.ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, pending_expires_at, grant_expires_at, grant_duration, fields,
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
		var pendingExpiresAt, grantExpiresAt sql.NullTime
		var updatedBy sql.NullString
		var fieldsArray []string

		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&pendingExpiresAt, &grantExpiresAt, &record.GrantDuration, pq.Array(&fieldsArray),
			&record.SessionID, &record.ConsentPortalURL, &updatedBy)

		if pendingExpiresAt.Valid {
			record.PendingExpiresAt = &pendingExpiresAt.Time
		}
		if grantExpiresAt.Valid {
			record.GrantExpiresAt = &grantExpiresAt.Time
		}
		if updatedBy.Valid {
			updatedByStr := updatedBy.String
			record.UpdatedBy = &updatedByStr
		}

		// Convert string array to ConsentField array
		record.Fields = make([]models.ConsentField, len(fieldsArray))
		for i, fieldName := range fieldsArray {
			record.Fields[i] = models.ConsentField{
				FieldName: fieldName,
			}
		}

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
