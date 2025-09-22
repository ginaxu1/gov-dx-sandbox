package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

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

// ConsentStatus represents the status of a consent record
type ConsentStatus string

const (
	StatusPending  ConsentStatus = "pending"
	StatusApproved ConsentStatus = "approved"
	StatusRejected ConsentStatus = "rejected"
	StatusExpired  ConsentStatus = "expired"
	StatusRevoked  ConsentStatus = "revoked"
)

// ConsentRecord represents a consent record in the system
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
	// UpdatedBy identifies who last updated the consent (audit field)
	UpdatedBy string `json:"updated_by,omitempty"`
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
	AppID         string      `json:"app_id"`
	DataFields    []DataField `json:"data_fields"`
	SessionID     string      `json:"session_id"`
	GrantDuration string      `json:"grant_duration,omitempty"`
}

// DataField represents a data field that requires consent
type DataField struct {
	OwnerID    string   `json:"owner_id"`
	OwnerEmail string   `json:"owner_email"`
	Fields     []string `json:"fields"`
}

// UpdateConsentRequest defines the structure for updating a consent record
type UpdateConsentRequest struct {
	Status        ConsentStatus `json:"status"`
	UpdatedBy     string        `json:"updated_by,omitempty"`
	GrantDuration string        `json:"grant_duration,omitempty"`
	Fields        []string      `json:"fields,omitempty"`
	Reason        string        `json:"reason,omitempty"`
}

// ConsentPortalRequest defines the structure for consent portal interactions
type ConsentPortalRequest struct {
	ConsentID string `json:"consent_id"`
	Action    string `json:"action"` // "approve" or "reject"
	DataOwner string `json:"data_owner"`
	Reason    string `json:"reason,omitempty"`
}

// ConsentEngine defines the interface for consent management operations
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
	// StartBackgroundExpiryProcess starts the background process for checking consent expiry
	StartBackgroundExpiryProcess(ctx context.Context, interval time.Duration)
	// RevokeConsent revokes a consent record
	RevokeConsent(id string, reason string) (*ConsentRecord, error)
	// ProcessConsentRequest processes the new consent workflow request
	ProcessConsentRequest(req ConsentRequest) (*ConsentRecord, error)
	// StopBackgroundExpiryProcess stops the background expiry process
	StopBackgroundExpiryProcess()
}

// NewConsentEngine creates a new instance of the consent engine (PostgreSQL implementation)
func NewConsentEngine(consentPortalUrl string) ConsentEngine {
	// This function is deprecated - use NewPostgresConsentEngine instead
	panic("NewConsentEngine is deprecated. Use NewPostgresConsentEngine instead.")
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
	duration, err := utils.ParseExpiryTime(grantDuration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid grant duration format: %w", err)
	}
	return createdAt.Add(duration), nil
}

// getAllFields extracts all fields from data fields
func getAllFields(dataFields []DataField) []string {
	var allFields []string
	for _, field := range dataFields {
		allFields = append(allFields, field.Fields...)
	}
	return allFields
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(current, new ConsentStatus) bool {
	validTransitions := map[ConsentStatus][]ConsentStatus{
		StatusPending:  {StatusApproved, StatusRejected, StatusExpired},                // Initial decision
		StatusApproved: {StatusApproved, StatusRejected, StatusRevoked, StatusExpired}, // Direct approval flow: approved->approved (success), approved->rejected (direct rejection), approved->revoked (user revocation), approved->expired (expiry)
		StatusRejected: {StatusExpired},                                                // Terminal state - can only transition to expired
		StatusExpired:  {StatusExpired},                                                // Terminal state - can only stay expired
		StatusRevoked:  {StatusExpired},                                                // Terminal state - can only transition to expired
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
