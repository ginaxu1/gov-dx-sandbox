package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// postgresConsentEngine implements ConsentEngine interface using PostgreSQL
type postgresConsentEngine struct {
	db *sql.DB
}

// NewPostgresConsentEngine creates a new PostgreSQL-based consent engine
func NewPostgresConsentEngine(db *sql.DB) ConsentEngine {
	return &postgresConsentEngine{db: db}
}

// ProcessConsentRequest processes a consent request and creates a consent record
func (pce *postgresConsentEngine) ProcessConsentRequest(req ConsentRequest) (*ConsentRecord, error) {
	// Validate required fields
	if req.AppID == "" {
		return nil, fmt.Errorf("app_id is required")
	}

	if len(req.DataFields) == 0 {
		return nil, fmt.Errorf("data_fields is required")
	}

	// Validate each data field
	for i, dataField := range req.DataFields {
		if dataField.OwnerID == "" {
			return nil, fmt.Errorf("data_fields[%d].owner_id is required", i)
		}
		if len(dataField.Fields) == 0 {
			return nil, fmt.Errorf("data_fields[%d].fields is required", i)
		}
	}

	// Use the first data field for owner information
	firstDataField := req.DataFields[0]

	// Check for existing pending consent for the same owner and app
	existingConsent, err := pce.findExistingPendingConsentByEmail(firstDataField.OwnerEmail, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing consent: %w", err)
	}

	if existingConsent != nil {
		return nil, fmt.Errorf("a pending consent already exists for owner_email '%s' and app_id '%s' (consent_id: %s)",
			firstDataField.OwnerEmail, req.AppID, existingConsent.ConsentID)
	}

	// Create new consent record
	consentID := generateConsentID()
	now := time.Now()

	// Combine all fields from all data fields
	var allFields []string
	for _, dataField := range req.DataFields {
		allFields = append(allFields, dataField.Fields...)
	}

	// Use default grant duration if not provided
	grantDuration := getDefaultGrantDuration(req.GrantDuration)

	// Calculate expires_at
	expiresAt, err := calculateExpiresAt(grantDuration, now)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expiry time: %w", err)
	}

	// Generate consent portal URL
	consentPortalURL := fmt.Sprintf("http://localhost:5173/?consent_id=%s", consentID)

	// Insert new consent record
	insertSQL := `
		INSERT INTO consent_records (
			consent_id, owner_id, owner_email, app_id, status, type, 
			created_at, updated_at, expires_at, grant_duration, fields, 
			session_id, consent_portal_url, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = pce.db.Exec(insertSQL,
		consentID, firstDataField.OwnerID, firstDataField.OwnerEmail, req.AppID, string(StatusPending), "realtime",
		now, now, expiresAt, grantDuration, pq.Array(allFields),
		req.SessionID, consentPortalURL, firstDataField.OwnerID)

	if err != nil {
		return nil, fmt.Errorf("failed to create consent record: %w", err)
	}

	record := &ConsentRecord{
		ConsentID:        consentID,
		OwnerID:          firstDataField.OwnerID,
		OwnerEmail:       firstDataField.OwnerEmail,
		AppID:            req.AppID,
		Status:           string(StatusPending),
		Type:             "realtime",
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        expiresAt,
		GrantDuration:    grantDuration,
		Fields:           allFields,
		SessionID:        req.SessionID,
		ConsentPortalURL: consentPortalURL,
		UpdatedBy:        firstDataField.OwnerID,
	}

	slog.Info("Consent record created",
		"consent_id", record.ConsentID,
		"owner_id", record.OwnerID,
		"owner_email", record.OwnerEmail,
		"app_id", record.AppID)

	return record, nil
}

// CreateConsent creates a new consent record (alias for ProcessConsentRequest)
func (pce *postgresConsentEngine) CreateConsent(req ConsentRequest) (*ConsentRecord, error) {
	return pce.ProcessConsentRequest(req)
}

// GetConsentStatus retrieves a consent record by ID
func (pce *postgresConsentEngine) GetConsentStatus(consentID string) (*ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE consent_id = $1
	`

	row := pce.db.QueryRow(querySQL, consentID)

	var record ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, &record.Fields,
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
func (pce *postgresConsentEngine) UpdateConsent(consentID string, req UpdateConsentRequest) (*ConsentRecord, error) {
	// Get existing record
	existingRecord, err := pce.GetConsentStatus(consentID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !isValidStatusTransition(ConsentStatus(existingRecord.Status), req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", existingRecord.Status, string(req.Status))
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
	updatedRecord.Status = string(req.Status)
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
func (pce *postgresConsentEngine) RevokeConsent(consentID string, reason string) (*ConsentRecord, error) {
	updateReq := UpdateConsentRequest{
		Status:    StatusRevoked,
		UpdatedBy: "system", // Could be enhanced to get from context
		Reason:    reason,
	}
	return pce.UpdateConsent(consentID, updateReq)
}

// CheckConsentExpiry checks for and updates expired consent records
func (pce *postgresConsentEngine) CheckConsentExpiry() ([]*ConsentRecord, error) {
	now := time.Now()

	// Find expired records that are not already marked as expired
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE expires_at < $1 AND status != 'expired'
	`

	rows, err := pce.db.Query(querySQL, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired records: %w", err)
	}
	defer rows.Close()

	var expiredRecords []*ConsentRecord

	for rows.Next() {
		var record ConsentRecord

		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&record.ExpiresAt, &record.GrantDuration, &record.Fields,
			&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

		if err != nil {
			return nil, fmt.Errorf("failed to scan expired record: %w", err)
		}

		// Update status to expired
		updateReq := UpdateConsentRequest{
			Status:    StatusExpired,
			UpdatedBy: "system",
			Reason:    "Consent expired automatically",
		}

		_, err = pce.UpdateConsent(record.ConsentID, updateReq)
		if err != nil {
			slog.Error("Failed to update expired consent", "consent_id", record.ConsentID, "error", err)
			continue
		}

		record.Status = string(StatusExpired)
		expiredRecords = append(expiredRecords, &record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired records: %w", err)
	}

	slog.Info("Checked consent expiry", "expired_count", len(expiredRecords))
	return expiredRecords, nil
}

// StartBackgroundExpiryProcess starts the background process for checking consent expiry
func (pce *postgresConsentEngine) StartBackgroundExpiryProcess(interval time.Duration) {
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

// findExistingPendingConsentByEmail finds existing pending consent for the given email and app
func (pce *postgresConsentEngine) findExistingPendingConsentByEmail(ownerEmail, appID string) (*ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_email = $1 AND app_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := pce.db.QueryRow(querySQL, ownerEmail, appID)

	var record ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, &record.Fields,
		&record.SessionID, &record.ConsentPortalURL, &record.UpdatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No existing pending consent found
		}
		return nil, fmt.Errorf("failed to query existing consent: %w", err)
	}

	return &record, nil
}

// FindExistingConsent finds an existing consent record by consumer app ID and owner ID
func (pce *postgresConsentEngine) FindExistingConsent(consumerAppID, ownerID string) *ConsentRecord {
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

	var record ConsentRecord

	err := row.Scan(
		&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
		&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
		&record.ExpiresAt, &record.GrantDuration, &record.Fields,
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
func (pce *postgresConsentEngine) ProcessConsentPortalRequest(req ConsentPortalRequest) (*ConsentRecord, error) {
	// Map action to status
	var status ConsentStatus
	switch req.Action {
	case "approve":
		status = StatusApproved
	case "deny":
		status = StatusRejected
	case "revoke":
		status = StatusRevoked
	default:
		return nil, fmt.Errorf("invalid action: %s", req.Action)
	}

	// Update the record based on portal action
	updateReq := UpdateConsentRequest{
		Status:    status,
		UpdatedBy: req.DataOwner,
		Reason:    req.Reason,
	}

	return pce.UpdateConsent(req.ConsentID, updateReq)
}

// GetConsentsByDataOwner retrieves all consent records for a data owner
func (pce *postgresConsentEngine) GetConsentsByDataOwner(dataOwner string) ([]*ConsentRecord, error) {
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

	var records []*ConsentRecord

	for rows.Next() {
		var record ConsentRecord
		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&record.ExpiresAt, &record.GrantDuration, &record.Fields,
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
func (pce *postgresConsentEngine) GetConsentsByConsumer(consumer string) ([]*ConsentRecord, error) {
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

	var records []*ConsentRecord

	for rows.Next() {
		var record ConsentRecord
		err := rows.Scan(
			&record.ConsentID, &record.OwnerID, &record.OwnerEmail, &record.AppID,
			&record.Status, &record.Type, &record.CreatedAt, &record.UpdatedAt,
			&record.ExpiresAt, &record.GrantDuration, &record.Fields,
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

// generateConsentID generates a unique consent ID
func generateConsentID() string {
	return "consent_" + strings.ReplaceAll(uuid.New().String(), "-", "")
}
