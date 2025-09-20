package main

import (
	"fmt"
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
	// CreateOrUpdateConsentRecord creates a new consent record or updates an existing one
	CreateOrUpdateConsentRecord(req ConsentRecord) (*ConsentRecord, error)
	// UpdateConsentRecord updates an existing consent record directly
	UpdateConsentRecord(record *ConsentRecord) error
}

// consentEngineImpl is the private, in-memory implementation of the ConsentEngine interface
type consentEngineImpl struct {
	consentRecords   map[string]*ConsentRecord
	consentPortalUrl string
	lock             sync.RWMutex
}

// NewConsentEngine creates a new in-memory instance of the ConsentEngine
func NewConsentEngine(consentPortalUrl string) ConsentEngine {
	ce := &consentEngineImpl{
		consentRecords:   make(map[string]*ConsentRecord),
		consentPortalUrl: consentPortalUrl,
	}

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

	// Check if there's already an existing consent for this (app_id, owner_id, owner_email) combination
	// We only need to check the first data field since all should have the same owner
	if len(req.DataFields) > 0 {
		existingRecord := ce.findExistingConsentByEmailUnsafe(req.AppID, req.DataFields[0].OwnerID, req.DataFields[0].OwnerEmail)
		if existingRecord != nil {
			// Return the existing record instead of creating a new one
			return existingRecord, nil
		}
	}

	// Create new consent record for this owner
	consentID := fmt.Sprintf("consent_%s", uuid.New().String()[:8])
	expiresAt := now.Add(30 * 24 * time.Hour)

	// Set default grant duration if not provided
	grantDuration := req.GrantDuration
	if grantDuration == "" {
		grantDuration = "30d"
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

	ce.consentRecords[record.ConsentID] = record

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

// findExistingConsentByEmailUnsafe finds an existing consent record by app_id, owner_id, and owner_email
// This should only be called when the caller already holds the appropriate lock
func (ce *consentEngineImpl) findExistingConsentByEmailUnsafe(consumerAppID, ownerID, ownerEmail string) *ConsentRecord {
	for _, record := range ce.consentRecords {
		if record.AppID == consumerAppID && record.OwnerID == ownerID && record.OwnerEmail == ownerEmail {
			return record
		}
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

	// Update grant duration if provided
	if req.GrantDuration != "" {
		record.GrantDuration = req.GrantDuration

		// Recalculate expires_at if grant duration is provided
		duration, err := utils.ParseExpiryTime(req.GrantDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid grant duration format: %w", err)
		}
		record.ExpiresAt = time.Now().Add(duration)
	}

	ce.consentRecords[id] = record
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

// GetAllConsentRecords retrieves all consent records (for debugging/admin purposes)
func (ce *consentEngineImpl) GetAllConsentRecords() ([]*ConsentRecord, error) {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	var records []*ConsentRecord
	for _, record := range ce.consentRecords {
		records = append(records, record)
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
		StatusPending:  {StatusApproved, StatusRejected, StatusExpired}, // Initial decision
		StatusApproved: {StatusApproved, StatusRejected, StatusRevoked}, // Direct approval flow: approved->approved (success), approved->rejected (direct rejection), approved->revoked (user revocation)
		StatusRejected: {StatusPending, StatusApproved},                 // Allow retry from rejected
		StatusExpired:  {StatusPending},                                 // Allow renewal
		StatusRevoked:  {StatusPending},                                 // Allow revocation reconsideration (must go through pending first)
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

	// Create expiry time (30 days from now)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	ce.lock.Lock()
	defer ce.lock.Unlock()

	// Check if there's already an existing consent for this (app_id, owner_id, owner_email) combination
	// We only need to check the first data field since all should have the same owner
	if len(req.DataFields) > 0 {
		existingRecord := ce.findExistingConsentByEmailUnsafe(req.AppID, req.DataFields[0].OwnerID, req.DataFields[0].OwnerEmail)
		if existingRecord != nil {
			// Return the existing record instead of creating a new one
			return existingRecord, nil
		}
	}

	// Create new consent record for this owner
	consentID := fmt.Sprintf("consent_%s", uuid.New().String()[:8])

	// Set default grant duration if not provided
	grantDuration := req.GrantDuration
	if grantDuration == "" {
		grantDuration = "30d"
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
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
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

// Helper function to parse expiry time strings like "30d", "1h", "7d"
func parseExpiryTime(expiryStr string) (time.Duration, error) {
	if len(expiryStr) < 2 {
		return 0, fmt.Errorf("invalid expiry time format")
	}

	unit := expiryStr[len(expiryStr)-1:]
	value := expiryStr[:len(expiryStr)-1]

	var duration time.Duration
	switch unit {
	case "d":
		duration = 24 * time.Hour
	case "h":
		duration = time.Hour
	case "m":
		duration = time.Minute
	case "s":
		duration = time.Second
	default:
		return 0, fmt.Errorf("unsupported time unit: %s", unit)
	}

	// Parse the numeric value
	var multiplier int
	if _, err := fmt.Sscanf(value, "%d", &multiplier); err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", value)
	}

	return time.Duration(multiplier) * duration, nil
}

// CreateOrUpdateConsentRecord creates a new consent record or updates an existing one
func (ce *consentEngineImpl) CreateOrUpdateConsentRecord(req ConsentRecord) (*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	// Check if record already exists
	if existingRecord, exists := ce.consentRecords[req.ConsentID]; exists {
		// Update existing record
		existingRecord.Status = req.Status
		existingRecord.Type = req.Type
		existingRecord.OwnerID = req.OwnerID
		existingRecord.OwnerEmail = req.OwnerEmail
		existingRecord.AppID = req.AppID
		existingRecord.UpdatedAt = time.Now()
		existingRecord.Fields = req.Fields
		existingRecord.SessionID = req.SessionID
		existingRecord.ConsentPortalURL = fmt.Sprintf("%s/?consent_id=%s", ce.consentPortalUrl, existingRecord.ConsentID)
		existingRecord.ExpiresAt = req.ExpiresAt
		existingRecord.GrantDuration = req.GrantDuration

		return existingRecord, nil
	}

	// Set default grant duration if not provided
	grantDuration := req.GrantDuration
	if grantDuration == "" {
		grantDuration = "30d"
	}

	// Create new record
	record := &ConsentRecord{
		ConsentID:        req.ConsentID,
		OwnerID:          req.OwnerID,
		OwnerEmail:       req.OwnerEmail,
		AppID:            req.AppID,
		Status:           req.Status,
		Type:             req.Type,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		ExpiresAt:        req.ExpiresAt,
		GrantDuration:    grantDuration,
		Fields:           req.Fields,
		SessionID:        req.SessionID,
		ConsentPortalURL: fmt.Sprintf("%s/?consent_id=%s", ce.consentPortalUrl, req.ConsentID),
	}

	// Store the record
	ce.consentRecords[req.ConsentID] = record

	return record, nil
}

// UpdateConsentRecord updates an existing consent record directly
func (ce *consentEngineImpl) UpdateConsentRecord(record *ConsentRecord) error {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	// Check if the record exists
	if _, exists := ce.consentRecords[record.ConsentID]; !exists {
		return fmt.Errorf("consent record with ID '%s' not found", record.ConsentID)
	}

	// Update the record
	ce.consentRecords[record.ConsentID] = record

	return nil
}
