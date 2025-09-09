package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConsentStatus defines the possible states of a consent record
type ConsentStatus string

const (
	// StatusPending indicates that the consent request has been created but not yet actioned
	StatusPending ConsentStatus = "pending"
	// StatusApproved indicates that the data owner has approved the consent request
	StatusApproved ConsentStatus = "approved"
	// StatusDenied indicates that the data owner has denied the consent request
	StatusDenied ConsentStatus = "denied"
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
	// ID is the unique identifier for the consent record
	ID string `json:"id"`
	// Status indicates the current state of the consent
	Status ConsentStatus `json:"status"`
	// Type indicates whether this is real-time or offline consent
	Type ConsentType `json:"type"`
	// CreatedAt is the timestamp when the consent record was initially created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the consent record was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt is the timestamp when the consent expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// DataConsumer identifies the entity requesting access to the data
	DataConsumer string `json:"data_consumer"`
	// DataOwner identifies the user or entity granting consent
	DataOwner string `json:"data_owner"`
	// Fields is a list of specific data fields the consent applies to
	Fields []string `json:"fields"`
	// ConsentPortalURL is the URL where the user can provide consent
	ConsentPortalURL string `json:"consent_portal_url,omitempty"`
	// SessionID is the session identifier for tracking the consent flow
	SessionID string `json:"session_id,omitempty"`
	// RedirectURL is the URL to redirect to after consent is provided
	RedirectURL string `json:"redirect_url,omitempty"`
	// Metadata contains additional information about the consent request
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateConsentRequest defines the structure for creating a consent record
type CreateConsentRequest struct {
	// DataConsumer identifies the entity requesting access to the data
	DataConsumer string `json:"data_consumer"`
	// DataOwner identifies the user or entity that owns the data and can grant consent
	DataOwner string `json:"data_owner"`
	// Fields is the list of specific data fields for which access is being requested
	Fields []string `json:"fields"`
	// Type indicates whether this is real-time or offline consent
	Type ConsentType `json:"type"`
	// SessionID is the session identifier for tracking the consent flow
	SessionID string `json:"session_id,omitempty"`
	// RedirectURL is the URL to redirect to after consent is provided
	RedirectURL string `json:"redirect_url,omitempty"`
	// ExpiryTime is the duration for which consent is valid (e.g., "30d", "1h")
	ExpiryTime string `json:"expiry_time,omitempty"`
	// Metadata contains additional information about the consent request
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateConsentRequest defines the structure for updating a consent record
type UpdateConsentRequest struct {
	// Status is the new status for the consent record
	Status ConsentStatus `json:"status"`
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
	CreateConsent(req CreateConsentRequest) (*ConsentRecord, error)
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
}

// consentEngineImpl is the private, in-memory implementation of the ConsentEngine interface
type consentEngineImpl struct {
	consentRecords map[string]*ConsentRecord
	lock           sync.RWMutex
}

// NewConsentEngine creates a new in-memory instance of the ConsentEngine
func NewConsentEngine() ConsentEngine {
	return &consentEngineImpl{
		consentRecords: make(map[string]*ConsentRecord),
	}
}

// CreateConsent creates a new consent record and saves it in the in-memory store
func (ce *consentEngineImpl) CreateConsent(req CreateConsentRequest) (*ConsentRecord, error) {
	ce.lock.Lock()
	defer ce.lock.Unlock()

	now := time.Now()
	record := &ConsentRecord{
		ID:               uuid.New().String(),
		Status:           StatusPending,
		Type:             req.Type,
		CreatedAt:        now,
		UpdatedAt:        now,
		DataConsumer:     req.DataConsumer,
		DataOwner:        req.DataOwner,
		Fields:           req.Fields,
		SessionID:        req.SessionID,
		RedirectURL:      req.RedirectURL,
		ConsentPortalURL: fmt.Sprintf("/consent-portal/%s", uuid.New().String()),
		Metadata:         req.Metadata,
	}

	// Set expiry time if provided
	if req.ExpiryTime != "" {
		if expiry, err := parseExpiryTime(req.ExpiryTime); err == nil {
			expiryTime := now.Add(expiry)
			record.ExpiresAt = &expiryTime
		}
	}

	ce.consentRecords[record.ID] = record
	return record, nil
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
	if !isValidStatusTransition(record.Status, req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", record.Status, req.Status)
	}

	// Update the record
	record.Status = req.Status
	record.UpdatedAt = time.Now()

	// Update metadata
	if record.Metadata == nil {
		record.Metadata = make(map[string]interface{})
	}
	record.Metadata["updated_by"] = req.UpdatedBy
	record.Metadata["update_reason"] = req.Reason
	record.Metadata["last_updated"] = record.UpdatedAt

	// Merge additional metadata
	for k, v := range req.Metadata {
		record.Metadata[k] = v
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
	if record.DataOwner != req.DataOwner {
		return nil, fmt.Errorf("data owner mismatch for consent record %s", req.ConsentID)
	}

	// Process the action
	var newStatus ConsentStatus
	switch req.Action {
	case "approve":
		newStatus = StatusApproved
	case "deny":
		newStatus = StatusDenied
	case "revoke":
		newStatus = StatusRevoked
	default:
		return nil, fmt.Errorf("invalid action: %s", req.Action)
	}

	// Validate status transition
	if !isValidStatusTransition(record.Status, newStatus) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", record.Status, newStatus)
	}

	// Update the record
	record.Status = newStatus
	record.UpdatedAt = time.Now()

	// Update metadata
	if record.Metadata == nil {
		record.Metadata = make(map[string]interface{})
	}
	record.Metadata["portal_action"] = req.Action
	record.Metadata["action_reason"] = req.Reason
	record.Metadata["session_id"] = req.SessionID
	record.Metadata["last_updated"] = record.UpdatedAt

	ce.consentRecords[req.ConsentID] = record
	return record, nil
}

// GetConsentsByDataOwner retrieves all consent records for a data owner
func (ce *consentEngineImpl) GetConsentsByDataOwner(dataOwner string) ([]*ConsentRecord, error) {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	var records []*ConsentRecord
	for _, record := range ce.consentRecords {
		if record.DataOwner == dataOwner {
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
		if record.DataConsumer == consumer {
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
		if record.ExpiresAt != nil && record.ExpiresAt.Before(now) && record.Status == StatusApproved {
			record.Status = StatusExpired
			record.UpdatedAt = now

			if record.Metadata == nil {
				record.Metadata = make(map[string]interface{})
			}
			record.Metadata["expired_at"] = now
			record.Metadata["expiry_reason"] = "automatic_expiry"

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
	if !isValidStatusTransition(record.Status, StatusRevoked) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", record.Status, StatusRevoked)
	}

	record.Status = StatusRevoked
	record.UpdatedAt = time.Now()

	if record.Metadata == nil {
		record.Metadata = make(map[string]interface{})
	}
	record.Metadata["revoked_at"] = record.UpdatedAt
	record.Metadata["revocation_reason"] = reason

	ce.consentRecords[id] = record
	return record, nil
}

// Helper function to validate status transitions
func isValidStatusTransition(current, new ConsentStatus) bool {
	validTransitions := map[ConsentStatus][]ConsentStatus{
		StatusPending:  {StatusApproved, StatusDenied, StatusExpired},
		StatusApproved: {StatusRevoked, StatusExpired},
		StatusDenied:   {StatusPending}, // Allow retry
		StatusExpired:  {StatusPending}, // Allow renewal
		StatusRevoked:  {},              // No transitions from revoked
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
