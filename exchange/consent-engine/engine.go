package main

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/utils"

	"github.com/google/uuid"
)

// DataField represents a request for specific data fields from a data owner
type DataField struct {
	OwnerType  string   `json:"owner_type"`
	OwnerID    string   `json:"owner_id"`
	OwnerEmail string   `json:"owner_email"`
	Fields     []string `json:"fields"`
}

// ConsentResponse defines the structure for consent API responses
type ConsentResponse struct {
	ConsentID        string   `json:"consent_id"`
	Status           string   `json:"status"`
	OwnerID          string   `json:"owner_id"`
	OwnerEmail       string   `json:"owner_email"`
	Fields           []string `json:"fields"`
	SessionID        string   `json:"session_id"`
	ConsentPortalURL string   `json:"consent_portal_url"`
	Purpose          string   `json:"purpose"`
	Message          string   `json:"message"`
}

// Consent-engine error messages
const (
	ErrConsentNotFound     = "consent record not found"
	ErrConsentCreateFailed = "failed to create consent record"
	ErrConsentUpdateFailed = "failed to update consent record"
	ErrConsentRevokeFailed = "failed to revoke consent record"
	ErrConsentGetFailed    = "failed to get consent records"
	ErrConsentExpiryFailed = "failed to check consent expiry"
	ErrPortalRequestFailed = "failed to process consent portal request"
)

// Consent-engine log operations
const (
	OpCreateConsent         = "create consent"
	OpUpdateConsent         = "update consent"
	OpRevokeConsent         = "revoke consent"
	OpGetConsentStatus      = "get consent status"
	OpGetConsentsByOwner    = "get consents by data owner"
	OpGetConsentsByConsumer = "get consents by consumer"
	OpCheckConsentExpiry    = "check consent expiry"
	OpProcessPortalRequest  = "process consent portal"
)

// ConsentStatus defines the possible states of a consent record
type ConsentStatus string

const (
	// StatusPending indicates that the consent request has been created but not yet actioned
	StatusPending ConsentStatus = "pending"
	// Statusapproved indicates that the data owner has approved the consent request
	StatusApproved ConsentStatus = "approved"
	// StatusRejected indicates that the data owner has rejected the consent request
	StatusRejected ConsentStatus = "rejected"
	// StatusExpired indicates that the consent has expired
	StatusExpired ConsentStatus = "expired"
	// StatusRevoked indicates that the consent has been revoked by the data owner
	StatusRevoked ConsentStatus = "revoked"
)

// ConsentType defines the type of consent mechanism
type ConsentType string

const (
	// ConsentTypeRealTime indicates real-time consent from the user
	ConsentTypeRealTime ConsentType = "realtime"
	// ConsentTypeOffline indicates offline consent from the data owner
	ConsentTypeOffline ConsentType = "offline"
)

// ConsentRecord stores the details and state of a consent request
// It represents a formal agreement from the data owner to the data consumer
type ConsentRecord struct {
	// ConsentID is the unique identifier for the consent record
	ConsentID string `json:"consent_id"`
	// OwnerID is the unique identifier for the data owner
	OwnerID string `json:"owner_id"`
	// OwnerEmail is the email address of the data owner
	OwnerEmail string `json:"owner_email"`
	// AppID is the unique identifier for the consumer application
	AppID string `json:"app_id"`
	// Status is the status of the consent record: pending, approved, rejected, expired, revoked
	Status string `json:"status"`
	// Type is the type of consent mechanism "realtime" or "offline"
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	// GrantDuration is the duration of consent grant (e.g., "30d", "1h")
	GrantDuration string `json:"grant_duration"`
	// Fields is the list of data fields that require consent
	Fields []string `json:"fields"`
	// SessionID is the session identifier for tracking the consent flow
	SessionID string `json:"session_id"`
	// ConsentPortalURL is the URL to redirect to for consent portal
	ConsentPortalURL string `json:"consent_portal_url"`
}

// ConsentPortalView represents the user-facing consent object for the UI.
type ConsentPortalView struct {
	AppDisplayName string    `json:"app_display_name"`
	CreatedAt      time.Time `json:"created_at"`
	Fields         []string  `json:"fields"`
	OwnerName      string    `json:"owner_name"`
	OwnerEmail     string    `json:"owner_email"`
	Status         string    `json:"status"`
	Type           string    `json:"type"`
}

// ToConsentPortalView converts an internal ConsentRecord to a user-facing view.
func (cr *ConsentRecord) ToConsentPortalView() *ConsentPortalView {
	// Simple mapping for app_id to a human-readable name.
	// You may need a more robust mapping or database lookup for this in a real application.
	appDisplayName := strings.ReplaceAll(cr.AppID, "-", " ")
	appDisplayName = strings.Title(appDisplayName)

	// Simple mapping for owner_id to a human-readable name.
	// In a real application, this would be a database lookup or API call.
	ownerName := strings.ReplaceAll(cr.OwnerID, "-", " ")
	ownerName = strings.Title(ownerName)

	return &ConsentPortalView{
		AppDisplayName: appDisplayName,
		CreatedAt:      cr.CreatedAt,
		Fields:         cr.Fields,
		OwnerName:      ownerName,
		OwnerEmail:     cr.OwnerEmail,
		Status:         cr.Status,
		Type:           cr.Type,
	}
}

// ConsentRequest defines the structure for creating a consent record
type ConsentRequest struct {
	// AppID identifies the application requesting access to the data
	AppID string `json:"app_id"`
	// DataFields contains the data owner information and fields
	DataFields []DataField `json:"data_fields"`
	// Purpose describes the purpose for which consent is being requested
	Purpose string `json:"purpose"`
	// SessionID is the session identifier for tracking the consent flow
	SessionID string `json:"session_id"`
	// ExpiresAt is the expiry time as epoch timestamp (optional)
	ExpiresAt int64 `json:"expires_at,omitempty"`
	// GrantDuration is the duration of consent grant (e.g., "30d", "1h") (optional)
	GrantDuration string `json:"grant_duration,omitempty"`
}

// UpdateConsentRequest defines the structure for updating a consent record
type UpdateConsentRequest struct {
	// Status is the new status for the consent record
	Status ConsentStatus `json:"status"`
	// GrantDuration is the duration of consent grant (e.g., "30d", "1h") (optional)
	GrantDuration string `json:"grant_duration,omitempty"`
	// UpdatedBy identifies who is updating the consent (data owner, system, etc.)
	UpdatedBy string `json:"updated_by"`
	// Reason provides context for the status change
	Reason string `json:"reason,omitempty"`
	// Fields is the list of data fields that require consent
	Fields []string `json:"fields,omitempty"`
	// Metadata contains additional information about the update
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConsentPortalRequest defines the structure for consent portal interactions
type ConsentPortalRequest struct {
	// ConsentID is the ID of the consent record
	ConsentID string `json:"consent_id"`
	// Action is the action being performed (approve, deny, revoke)
	Action string `json:"action"`
	// DataOwner is the ID of the data owner performing the action
	DataOwner string `json:"data_owner"`
	// SessionID is the session identifier
	SessionID string `json:"session_id,omitempty"`
	// Reason provides context for the action
	Reason string `json:"reason,omitempty"`
}

// ConsentEngine defines the interface for managing consent records
type ConsentEngine interface {
	// CreateConsent creates a new consent record and stores it
	CreateConsent(req ConsentRequest) (*ConsentRecord, error)
	// FindExistingConsent finds an existing consent record by consumer app ID and owner ID
	FindExistingConsent(consumerAppID, ownerID string) *ConsentRecord
	// GetConsentStatus retrieves a consent record by its ID
	GetConsentStatus(id string) (*ConsentRecord, error)
	// UpdateConsent updates the status of a consent record
	UpdateConsent(id string, req UpdateConsentRequest) (*ConsentRecord, error)
	// ProcessConsentPortalRequest handles consent portal interactions
	ProcessConsentPortalRequest(req ConsentPortalRequest) (*ConsentRecord, error)
	// GetConsentsByDataOwner retrieves all consent records for a data owner
	GetConsentsByDataOwner(dataOwner string) ([]*ConsentRecord, error)
	// GetConsentsByConsumer retrieves all consent records for a consumer
	GetConsentsByConsumer(consumer string) ([]*ConsentRecord, error)
	// CheckConsentExpiry checks and updates expired consent records
	CheckConsentExpiry() ([]*ConsentRecord, error)
	// RevokeConsent revokes a consent record
	RevokeConsent(id string, reason string) (*ConsentRecord, error)
	// ProcessConsentRequest processes the new consent workflow request
	ProcessConsentRequest(req ConsentRequest) (*ConsentRecord, error)
	// StopBackgroundExpiryProcess stops the background expiry process
	StopBackgroundExpiryProcess()
}

// consentEngineImpl is the private, in-memory implementation of the ConsentEngine interface
type consentEngineImpl struct {
	consentRecords   map[string]*ConsentRecord
	consentPortalUrl string
	lock             sync.RWMutex
	stopChan         chan struct{}
}

// NewConsentEngine creates a new in-memory instance of the ConsentEngine
func NewConsentEngine(consentPortalUrl string) ConsentEngine {
	ce := &consentEngineImpl{
		consentRecords:   make(map[string]*ConsentRecord),
		consentPortalUrl: consentPortalUrl,
		stopChan:         make(chan struct{}),
	}

	// Start the background expiry process
	ce.startBackgroundExpiryProcess()

	return ce
}

// CreateConsent creates a new consent record and saves it in the in-memory store
func (ce *consentEngineImpl) CreateConsent(req ConsentRequest) (*ConsentRecord, error) {
	// Create ONE consent record for EACH data field
	if len(req.DataFields) == 0 {
		return nil, fmt.Errorf("at least one data field is required")
	}

	now := time.Now()

	ce.lock.Lock()
	defer ce.lock.Unlock()

	// Check if there's already a PENDING consent for this (app_id, owner_id, owner_email) combination
	if len(req.DataFields) > 0 {
		existingPendingRecord := ce.findExistingPendingConsentByEmailUnsafe(req.AppID, req.DataFields[0].OwnerID, req.DataFields[0].OwnerEmail)
		if existingPendingRecord != nil {
			// Update the existing pending record with new fields and updated_at
			now := time.Now()

			// Combine all fields from all data fields
			var allFields []string
			for _, dataField := range req.DataFields {
				allFields = append(allFields, dataField.Fields...)
			}

			// Update the existing record with all provided fields (PATCH-like behavior)
			existingPendingRecord.Fields = allFields
			existingPendingRecord.UpdatedAt = now
			existingPendingRecord.SessionID = req.SessionID

			// Update grant duration (use default "1h" if not provided)
			grantDuration := getDefaultGrantDuration(req.GrantDuration)
			existingPendingRecord.GrantDuration = grantDuration

			// Recalculate expires_at based on grant_duration and updated_at
			expiresAt, err := calculateExpiresAt(grantDuration, now)
			if err != nil {
				return nil, err
			}
			existingPendingRecord.ExpiresAt = expiresAt

			// Store the updated record
			ce.consentRecords[existingPendingRecord.ConsentID] = existingPendingRecord

			slog.Info("Updated existing pending consent record", "consent_id", existingPendingRecord.ConsentID, "owner_id", existingPendingRecord.OwnerID, "owner_email", existingPendingRecord.OwnerEmail, "app_id", existingPendingRecord.AppID, "status", existingPendingRecord.Status, "type", existingPendingRecord.Type, "created_at", existingPendingRecord.CreatedAt, "updated_at", existingPendingRecord.UpdatedAt, "expires_at", existingPendingRecord.ExpiresAt, "grant_duration", existingPendingRecord.GrantDuration, "fields", existingPendingRecord.Fields, "session_id", existingPendingRecord.SessionID, "consent_portal_url", existingPendingRecord.ConsentPortalURL)

			return existingPendingRecord, nil
		}
	}

	// Validate that we don't have multiple pending consents for this combination
	if len(req.DataFields) > 0 {
		if err := ce.validatePendingConsentUniqueness(req.AppID, req.DataFields[0].OwnerID, req.DataFields[0].OwnerEmail); err != nil {
			return nil, err
		}
	}

	// Create new consent record for this owner
	consentID := fmt.Sprintf("consent_%s", uuid.New().String()[:8])

	// Set default grant duration if not provided
	grantDuration := getDefaultGrantDuration(req.GrantDuration)

	// Calculate expires_at based on grant_duration and updated_at
	expiresAt, err := calculateExpiresAt(grantDuration, now)
	if err != nil {
		return nil, err
	}

	// Combine all fields from all data fields
	var allFields []string
	for _, dataField := range req.DataFields {
		allFields = append(allFields, dataField.Fields...)
	}
	slog.Info("Creating consent record", "consent_id", consentID, "owner_id", req.DataFields[0].OwnerID, "owner_email", req.DataFields[0].OwnerEmail, "app_id", req.AppID, "status", string(StatusPending), "type", string(ConsentTypeRealTime), "created_at", now, "updated_at", now, "expires_at", expiresAt, "grant_duration", grantDuration, "fields", allFields, "session_id", req.SessionID, "consent_portal_url", fmt.Sprintf("%s/?consent_id=%s", ce.consentPortalUrl, consentID))

	record := &ConsentRecord{
		ConsentID:        consentID,
		OwnerID:          req.DataFields[0].OwnerID,    // Use the first owner ID
		OwnerEmail:       req.DataFields[0].OwnerEmail, // Use the first owner email
		AppID:            req.AppID,
		Status:           string(StatusPending),
		Type:             string(ConsentTypeRealTime),
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        expiresAt,
		GrantDuration:    grantDuration,
		Fields:           allFields,
		SessionID:        req.SessionID,
		ConsentPortalURL: fmt.Sprintf("%s/?consent_id=%s", ce.consentPortalUrl, consentID),
	}

	ce.consentRecords[record.ConsentID] = record
	slog.Info("Consent record created", "consent_id", record.ConsentID)
	// Return the created record
	return record, nil
}

// FindExistingConsent finds an existing consent record by consumer app ID and owner ID
func (ce *consentEngineImpl) FindExistingConsent(consumerAppID, ownerID string) *ConsentRecord {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	return ce.findExistingConsentUnsafe(consumerAppID, ownerID)
}

// findExistingConsentUnsafe finds an existing consent record without acquiring a lock
// This should only be called when the caller already holds the appropriate lock
func (ce *consentEngineImpl) findExistingConsentUnsafe(consumerAppID, ownerID string) *ConsentRecord {
	for _, record := range ce.consentRecords {
		if record.AppID == consumerAppID && record.OwnerID == ownerID {
			return record
		}
	}
	return nil
}

// findExistingPendingConsentByEmailUnsafe finds an existing PENDING consent record by app_id, owner_id, and owner_email
// This should only be called when the caller already holds the appropriate lock
func (ce *consentEngineImpl) findExistingPendingConsentByEmailUnsafe(consumerAppID, ownerID, ownerEmail string) *ConsentRecord {
	for _, record := range ce.consentRecords {
		if record.AppID == consumerAppID && record.OwnerID == ownerID && record.OwnerEmail == ownerEmail && record.Status == string(StatusPending) {
			return record
		}
	}
	return nil
}

// validatePendingConsentUniqueness validates that there's only one pending consent record per (owner_id, owner_email, app_id) combination
// This should only be called when the caller already holds the appropriate lock
func (ce *consentEngineImpl) validatePendingConsentUniqueness(consumerAppID, ownerID, ownerEmail string) error {
	pendingCount := 0
	for _, record := range ce.consentRecords {
		if record.AppID == consumerAppID && record.OwnerID == ownerID && record.OwnerEmail == ownerEmail && record.Status == string(StatusPending) {
			pendingCount++
		}
	}

	if pendingCount > 1 {
		return fmt.Errorf("data integrity violation: found %d pending consent records for (app_id=%s, owner_id=%s, owner_email=%s), expected maximum 1", pendingCount, consumerAppID, ownerID, ownerEmail)
	}

	return nil
}

// GetConsentStatus retrieves a specific consent record from the in-memory store
func (ce *consentEngineImpl) GetConsentStatus(id string) (*ConsentRecord, error) {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	record, ok := ce.consentRecords[id]
	if !ok {
		return nil, fmt.Errorf("consent record with ID '%s' not found", id)
	}
	slog.Info("Consent record found", "consent_id", record.ConsentID, "owner_id", record.OwnerID, "owner_email", record.OwnerEmail, "app_id", record.AppID, "status", record.Status, "type", record.Type, "created_at", record.CreatedAt, "updated_at", record.UpdatedAt, "expires_at", record.ExpiresAt, "grant_duration", record.GrantDuration, "fields", record.Fields, "session_id", record.SessionID, "consent_portal_url", record.ConsentPortalURL)
	return record, nil
}

// UpdateConsent updates the status of a consent record
func (ce *consentEngineImpl) UpdateConsent(id string, req UpdateConsentRequest) (*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	record, ok := ce.consentRecords[id]
	if !ok {
		return nil, fmt.Errorf("consent record with ID '%s' not found", id)
	}

	// Validate status transition
	if !isValidStatusTransition(ConsentStatus(record.Status), req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", record.Status, string(req.Status))
	}

	// Update the record
	record.Status = string(req.Status)
	record.UpdatedAt = time.Now()

	// Update grant duration if provided, otherwise use existing or default
	if req.GrantDuration != "" {
		record.GrantDuration = req.GrantDuration
	} else {
		// Ensure we have a valid grant duration
		record.GrantDuration = getDefaultGrantDuration(record.GrantDuration)
	}

	// Update fields if provided
	if len(req.Fields) > 0 {
		record.Fields = req.Fields
	}

	// Recalculate expires_at based on grant_duration and updated_at
	expiresAt, err := calculateExpiresAt(record.GrantDuration, record.UpdatedAt)
	if err != nil {
		return nil, err
	}
	record.ExpiresAt = expiresAt

	ce.consentRecords[id] = record
	slog.Info("Consent record updated", "consent_id", record.ConsentID, "owner_id", record.OwnerID, "owner_email", record.OwnerEmail, "app_id", record.AppID, "status", record.Status, "type", record.Type, "created_at", record.CreatedAt, "updated_at", record.UpdatedAt, "expires_at", record.ExpiresAt, "grant_duration", record.GrantDuration, "fields", record.Fields, "session_id", record.SessionID, "consent_portal_url", record.ConsentPortalURL)
	return record, nil
}

// ProcessConsentPortalRequest handles consent portal interactions
func (ce *consentEngineImpl) ProcessConsentPortalRequest(req ConsentPortalRequest) (*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	record, ok := ce.consentRecords[req.ConsentID]
	if !ok {
		return nil, fmt.Errorf("consent record with ID '%s' not found", req.ConsentID)
	}

	// Validate data owner
	if record.OwnerID != req.DataOwner {
		return nil, fmt.Errorf("data owner mismatch for consent record %s", req.ConsentID)
	}

	// Process the action
	var newStatus string
	switch req.Action {
	case "approve", "grant":
		newStatus = string(StatusApproved)
	case "deny", "reject":
		newStatus = string(StatusRejected)
	case "revoke":
		newStatus = string(StatusRevoked)
	default:
		return nil, fmt.Errorf("invalid action: %s", req.Action)
	}

	// Validate status transition
	if !isValidStatusTransition(ConsentStatus(record.Status), ConsentStatus(newStatus)) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", record.Status, newStatus)
	}

	// Update the record
	record.Status = newStatus
	record.UpdatedAt = time.Now()

	ce.consentRecords[req.ConsentID] = record
	return record, nil
}

// GetConsentsByDataOwner retrieves all consent records for a data owner
func (ce *consentEngineImpl) GetConsentsByDataOwner(dataOwner string) ([]*ConsentRecord, error) {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	var records []*ConsentRecord
	for _, record := range ce.consentRecords {
		if record.OwnerID == dataOwner {
			records = append(records, record)
		}
	}
	return records, nil
}

// GetConsentsByConsumer retrieves all consent records for a consumer
func (ce *consentEngineImpl) GetConsentsByConsumer(consumer string) ([]*ConsentRecord, error) {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	var records []*ConsentRecord
	for _, record := range ce.consentRecords {
		if record.AppID == consumer {
			records = append(records, record)
		}
	}
	return records, nil
}

// CheckConsentExpiry checks and updates expired consent records
func (ce *consentEngineImpl) CheckConsentExpiry() ([]*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	now := time.Now()
	var expiredRecords []*ConsentRecord

	for _, record := range ce.consentRecords {
		if record.ExpiresAt.Before(now) && record.Status == string(StatusApproved) {
			record.Status = string(StatusExpired)
			record.UpdatedAt = now
			expiredRecords = append(expiredRecords, record)
		}
	}

	return expiredRecords, nil
}

// RevokeConsent revokes a consent record
func (ce *consentEngineImpl) RevokeConsent(id string, reason string) (*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	record, ok := ce.consentRecords[id]
	if !ok {
		return nil, fmt.Errorf("consent record with ID '%s' not found", id)
	}

	// Validate status transition
	if !isValidStatusTransition(ConsentStatus(record.Status), StatusRevoked) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", record.Status, string(StatusRevoked))
	}

	record.Status = string(StatusRevoked)
	record.UpdatedAt = time.Now()

	ce.consentRecords[id] = record
	return record, nil
}

// Helper function to validate status transitions
func isValidStatusTransition(current, new ConsentStatus) bool {
	validTransitions := map[ConsentStatus][]ConsentStatus{
		StatusPending:  {StatusApproved, StatusRejected, StatusExpired},                               // Initial decision
		StatusApproved: {StatusPending, StatusApproved, StatusRejected, StatusRevoked, StatusExpired}, // Direct approval flow: approved->approved (success), approved->rejected (direct rejection), approved->revoked (user revocation), approved->expired (expiry)
		StatusRejected: {StatusPending, StatusExpired},                                                // Terminal state - can only transition to expired
		StatusExpired:  {StatusExpired},                                                               // Terminal state - can only stay expired
		StatusRevoked:  {StatusExpired},                                                               // Terminal state - can only transition to expired
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

// ProcessConsentRequest processes the new consent workflow request
func (ce *consentEngineImpl) ProcessConsentRequest(req ConsentRequest) (*ConsentRecord, error) {
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

	ce.lock.Lock()
	defer ce.lock.Unlock()

	// Check if there's already a PENDING consent for this (app_id, owner_id, owner_email) combination
	// We only need to check the first data field since all should have the same owner
	if len(req.DataFields) > 0 {
		existingPendingRecord := ce.findExistingPendingConsentByEmailUnsafe(req.AppID, req.DataFields[0].OwnerID, req.DataFields[0].OwnerEmail)
		if existingPendingRecord != nil {
			// Update the existing pending record with new fields and updated_at
			now := time.Now()

			// Combine all fields from all data fields
			var allFields []string
			for _, dataField := range req.DataFields {
				allFields = append(allFields, dataField.Fields...)
			}

			// Update the existing record with all provided fields
			existingPendingRecord.Fields = allFields
			existingPendingRecord.UpdatedAt = now
			existingPendingRecord.SessionID = req.SessionID

			// Update grant duration (use default "1h" if not provided)
			grantDuration := getDefaultGrantDuration(req.GrantDuration)
			existingPendingRecord.GrantDuration = grantDuration

			// Recalculate expires_at based on grant_duration and updated_at
			expiresAt, err := calculateExpiresAt(grantDuration, now)
			if err != nil {
				return nil, err
			}
			existingPendingRecord.ExpiresAt = expiresAt

			// Store the updated record
			ce.consentRecords[existingPendingRecord.ConsentID] = existingPendingRecord

			slog.Info("Updated existing pending consent record", "consent_id", existingPendingRecord.ConsentID, "owner_id", existingPendingRecord.OwnerID, "owner_email", existingPendingRecord.OwnerEmail, "app_id", existingPendingRecord.AppID, "status", existingPendingRecord.Status, "type", existingPendingRecord.Type, "created_at", existingPendingRecord.CreatedAt, "updated_at", existingPendingRecord.UpdatedAt, "expires_at", existingPendingRecord.ExpiresAt, "grant_duration", existingPendingRecord.GrantDuration, "fields", existingPendingRecord.Fields, "session_id", existingPendingRecord.SessionID, "consent_portal_url", existingPendingRecord.ConsentPortalURL)

			return existingPendingRecord, nil
		}
	}

	// Validate that we don't have multiple pending consents for this combination
	if len(req.DataFields) > 0 {
		if err := ce.validatePendingConsentUniqueness(req.AppID, req.DataFields[0].OwnerID, req.DataFields[0].OwnerEmail); err != nil {
			return nil, err
		}
	}

	// Create new consent record for this owner
	consentID := fmt.Sprintf("consent_%s", uuid.New().String()[:8])

	// Set default grant duration if not provided
	grantDuration := getDefaultGrantDuration(req.GrantDuration)

	// Calculate expires_at based on grant_duration and updated_at
	now := time.Now()
	expiresAt, err := calculateExpiresAt(grantDuration, now)
	if err != nil {
		return nil, err
	}

	// Combine all fields from all data fields
	var allFields []string
	for _, dataField := range req.DataFields {
		allFields = append(allFields, dataField.Fields...)
	}

	record := &ConsentRecord{
		ConsentID:        consentID,
		OwnerID:          req.DataFields[0].OwnerID,    // Use the first owner ID
		OwnerEmail:       req.DataFields[0].OwnerEmail, // Use the first owner email
		AppID:            req.AppID,
		Status:           string(StatusPending),
		Type:             string(ConsentTypeRealTime),
		CreatedAt:        now,
		UpdatedAt:        now,
		ExpiresAt:        expiresAt,
		GrantDuration:    grantDuration,
		Fields:           allFields,
		SessionID:        req.SessionID,
		ConsentPortalURL: fmt.Sprintf("%s/?consent_id=%s", ce.consentPortalUrl, consentID),
	}

	// Store the record
	ce.consentRecords[consentID] = record

	return record, nil
}

// getDefaultGrantDuration returns the grant duration with a default value if empty
func getDefaultGrantDuration(grantDuration string) string {
	if grantDuration == "" {
		return "1h"
	}
	return grantDuration
}

// calculateExpiresAt calculates the expiration time based on grant duration and updated time
func calculateExpiresAt(grantDuration string, updatedAt time.Time) (time.Time, error) {
	duration, err := utils.ParseExpiryTime(grantDuration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid grant duration format: %w", err)
	}
	return updatedAt.Add(duration), nil
}

// startBackgroundExpiryProcess starts a background goroutine that runs every 24 hours
// to check and update expired consent records
func (ce *consentEngineImpl) startBackgroundExpiryProcess() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		slog.Info("Background expiry process started", "interval", "24h")

		for {
			select {
			case <-ticker.C:
				// Run the expiry check
				expiredRecords, err := ce.CheckConsentExpiry()
				if err != nil {
					slog.Error("Background expiry check failed", "error", err)
				} else if len(expiredRecords) > 0 {
					slog.Info("Background expiry process completed", "expired_count", len(expiredRecords))
				} else {
					slog.Debug("Background expiry process completed", "expired_count", 0)
				}
			case <-ce.stopChan:
				slog.Info("Background expiry process stopped")
				return
			}
		}
	}()
}

// StopBackgroundExpiryProcess stops the background expiry process
func (ce *consentEngineImpl) StopBackgroundExpiryProcess() {
	close(ce.stopChan)
}
