package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring"
	"github.com/lib/pq"
)

// postgresConsentEngine implements ConsentEngine interface using PostgreSQL
type postgresConsentEngine struct {
	db               *sql.DB
	consentPortalURL string
	pendingCache     sync.Map
	consentCache     sync.Map
}

const (
	consentDBName    = "consent-db"
	pendingCacheName = "consent-engine.pending_cache"
	consentCacheName = "consent-engine.consents_cache"
)

func (pce *postgresConsentEngine) exec(operation string, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := pce.db.Exec(query, args...)
	monitoring.RecordDBLatency(context.Background(), consentDBName, operation, time.Since(start))
	return result, err
}

func (pce *postgresConsentEngine) queryRows(operation string, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := pce.db.Query(query, args...)
	monitoring.RecordDBLatency(context.Background(), consentDBName, operation, time.Since(start))
	return rows, err
}

func (pce *postgresConsentEngine) queryRow(operation string, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := pce.db.QueryRow(query, args...)
	monitoring.RecordDBLatency(context.Background(), consentDBName, operation, time.Since(start))
	return row
}

func cloneConsentRecord(record *ConsentRecord) *ConsentRecord {
	if record == nil {
		return nil
	}
	clone := *record
	if record.Fields != nil {
		clone.Fields = append([]string(nil), record.Fields...)
	}
	return &clone
}

func cacheKey(ownerID, appID string) string {
	return ownerID + "|" + appID
}

func (pce *postgresConsentEngine) loadPendingCache(ownerID, appID string) (*ConsentRecord, bool) {
	if value, ok := pce.pendingCache.Load(cacheKey(ownerID, appID)); ok {
		if record, ok := value.(*ConsentRecord); ok {
			monitoring.RecordCacheEvent(context.Background(), pendingCacheName, true)
			return cloneConsentRecord(record), true
		}
	}
	monitoring.RecordCacheEvent(context.Background(), pendingCacheName, false)
	return nil, false
}

func (pce *postgresConsentEngine) storePendingCache(ownerID, appID string, record *ConsentRecord) {
	key := cacheKey(ownerID, appID)
	if record == nil {
		pce.pendingCache.Delete(key)
		return
	}
	pce.pendingCache.Store(key, cloneConsentRecord(record))
}

func (pce *postgresConsentEngine) loadConsentCache(ownerID, appID string) (*ConsentRecord, bool) {
	if value, ok := pce.consentCache.Load(cacheKey(ownerID, appID)); ok {
		if record, ok := value.(*ConsentRecord); ok {
			monitoring.RecordCacheEvent(context.Background(), consentCacheName, true)
			return cloneConsentRecord(record), true
		}
	}
	monitoring.RecordCacheEvent(context.Background(), consentCacheName, false)
	return nil, false
}

func (pce *postgresConsentEngine) storeConsentCache(ownerID, appID string, record *ConsentRecord) {
	key := cacheKey(ownerID, appID)
	if record == nil {
		pce.consentCache.Delete(key)
		return
	}
	pce.consentCache.Store(key, cloneConsentRecord(record))
}

func (pce *postgresConsentEngine) deleteConsentCaches(ownerID, appID string) {
	pce.pendingCache.Delete(cacheKey(ownerID, appID))
	pce.consentCache.Delete(cacheKey(ownerID, appID))
}

// NewPostgresConsentEngine creates a new PostgreSQL-based consent engine
func NewPostgresConsentEngine(db *sql.DB, consentPortalURL string) ConsentEngine {
	return &postgresConsentEngine{
		db:               db,
		consentPortalURL: consentPortalURL,
	}
}

// ProcessConsentRequest processes a consent request and creates a consent record
func (pce *postgresConsentEngine) ProcessConsentRequest(req ConsentRequest) (*ConsentRecord, error) {
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

	_, err = pce.exec("insert_consent", insertSQL,
		consentID, ownerID, ownerEmail, req.AppID, string(StatusPending), "realtime",
		now, now, expiresAt, grantDuration, pq.Array(allFields),
		"", // Intentionally passing empty string for session_id in the new format (session information not required)
		consentPortalURL, ownerID)

	if err != nil {
		return nil, fmt.Errorf("failed to create consent record: %w", err)
	}

	record := &ConsentRecord{
		ConsentID:        consentID,
		OwnerID:          ownerID,
		OwnerEmail:       ownerEmail,
		AppID:            req.AppID,
		Status:           string(StatusPending),
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

	pce.storePendingCache(ownerID, req.AppID, record)
	pce.storeConsentCache(ownerID, req.AppID, record)

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

	row := pce.queryRow("get_consent_status", querySQL, consentID)

	var record ConsentRecord

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

	if record.Status == string(StatusPending) {
		pce.storePendingCache(record.OwnerID, record.AppID, &record)
	} else {
		pce.storePendingCache(record.OwnerID, record.AppID, nil)
	}
	pce.storeConsentCache(record.OwnerID, record.AppID, &record)
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

	_, err = pce.exec("update_consent", updateSQL,
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

	if req.Status == StatusPending {
		pce.storePendingCache(existingRecord.OwnerID, existingRecord.AppID, &updatedRecord)
	} else {
		pce.storePendingCache(existingRecord.OwnerID, existingRecord.AppID, nil)
	}
	pce.storeConsentCache(existingRecord.OwnerID, existingRecord.AppID, &updatedRecord)

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

	// Find expired approved records
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE expires_at < $1 AND status = 'approved'
	`

	rows, err := pce.queryRows("select_expired_consents", querySQL, now)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired records: %w", err)
	}
	defer rows.Close()

	var deletedRecords []*ConsentRecord

	for rows.Next() {
		var record ConsentRecord

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
		_, err = pce.exec("delete_expired_consent", deleteSQL, record.ConsentID)
		if err != nil {
			slog.Error("Failed to delete expired consent", "consent_id", record.ConsentID, "error", err)
			continue
		}

		deletedRecords = append(deletedRecords, &record)
		pce.deleteConsentCaches(record.OwnerID, record.AppID)
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
func (pce *postgresConsentEngine) findExistingConsentByOwnerID(ownerID, appID string) (*ConsentRecord, error) {
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

	if cached, ok := pce.loadConsentCache(ownerID, appID); ok {
		return cached, nil
	}

	row := pce.queryRow("find_existing_consent", querySQL, ownerID, appID)

	var record ConsentRecord

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

	pce.storeConsentCache(ownerID, appID, &record)
	return &record, nil
}

// findExistingPendingConsentByOwnerID finds existing pending consent for the given owner_id and app
// This enforces the constraint that only one pending record can exist per (appId, ownerId) tuple
func (pce *postgresConsentEngine) findExistingPendingConsentByOwnerID(ownerID, appID string) (*ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE owner_id = $1 AND app_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 1
	`

	if cached, ok := pce.loadPendingCache(ownerID, appID); ok {
		return cached, nil
	}

	row := pce.queryRow("find_pending_consent", querySQL, ownerID, appID)

	var record ConsentRecord

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

	pce.storePendingCache(ownerID, appID, &record)
	return &record, nil
}

// updateExistingConsentNewFormat updates an existing consent record with new format (consent_requirements)
func (pce *postgresConsentEngine) updateExistingConsentNewFormat(existingConsent *ConsentRecord, req ConsentRequest) (*ConsentRecord, error) {
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

	_, err = pce.exec("update_existing_consent", updateSQL,
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

	pce.storeConsentCache(existingConsent.OwnerID, existingConsent.AppID, &updatedRecord)
	if updatedRecord.Status == string(StatusPending) {
		pce.storePendingCache(existingConsent.OwnerID, existingConsent.AppID, &updatedRecord)
	} else {
		pce.storePendingCache(existingConsent.OwnerID, existingConsent.AppID, nil)
	}

	return &updatedRecord, nil
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

	if cached, ok := pce.loadConsentCache(ownerID, consumerAppID); ok {
		return cached
	}

	row := pce.queryRow("find_consent_consumer_owner", querySQL, consumerAppID, ownerID)

	var record ConsentRecord

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

	pce.storeConsentCache(ownerID, consumerAppID, &record)
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

	rows, err := pce.queryRows("consents_by_owner", querySQL, dataOwner)
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
func (pce *postgresConsentEngine) GetConsentsByConsumer(consumer string) ([]*ConsentRecord, error) {
	querySQL := `
		SELECT consent_id, owner_id, owner_email, app_id, status, type,
		       created_at, updated_at, expires_at, grant_duration, fields,
		       session_id, consent_portal_url, updated_by
		FROM consent_records 
		WHERE app_id = $1
		ORDER BY created_at DESC
	`

	rows, err := pce.queryRows("consents_by_consumer", querySQL, consumer)
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
