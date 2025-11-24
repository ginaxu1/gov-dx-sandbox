package models

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	appDisplayName = cases.Title(language.English).String(appDisplayName)

	// Simple mapping for owner_id to a human-readable name.
	// In a real application, this would be a database lookup or API call.
	ownerName := strings.ReplaceAll(cr.OwnerID, "-", " ")
	ownerName = cases.Title(language.English).String(ownerName)

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
	AppID               string               `json:"app_id"`
	ConsentRequirements []ConsentRequirement `json:"consent_requirements"`
	GrantDuration       string               `json:"grant_duration,omitempty"`
}

// ConsentRequirement represents a consent requirement for a specific owner
type ConsentRequirement struct {
	Owner   string         `json:"owner"`
	OwnerID string         `json:"owner_id"`
	Fields  []ConsentField `json:"fields"`
}

// ConsentField represents a field that requires consent
type ConsentField struct {
	FieldName string `json:"fieldName"`
	SchemaID  string `json:"schemaId"`
}

// ConsentResponse represents the simplified response for consent operations
type ConsentResponse struct {
	ConsentID        string  `json:"consent_id"`
	Status           string  `json:"status"`
	ConsentPortalURL *string `json:"consent_portal_url,omitempty"` // Only present when status is pending
}

// ToConsentResponse converts a ConsentRecord to a simplified ConsentResponse
// Only includes consent_portal_url when status is pending and the URL is not empty
func (cr *ConsentRecord) ToConsentResponse() ConsentResponse {
	response := ConsentResponse{
		ConsentID: cr.ConsentID,
		Status:    cr.Status,
	}

	// Only include consent_portal_url when status is pending and URL is not empty
	if cr.Status == string(StatusPending) && cr.ConsentPortalURL != "" {
		portalURL := cr.ConsentPortalURL
		response.ConsentPortalURL = &portalURL
	}

	return response
}

// Legacy structures for backwards compatibility (deprecated)
type DataField struct {
	OwnerType  string   `json:"owner_type,omitempty"`
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
