package models

import "time"

// ConsentStatus represents the status of a consent record
type ConsentStatus string

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
	// ExpiresAt is the timestamp when the consent expires
	// Calculated by adding GrantDuration to the current time when consent is approved/denied
	// To check if consent has expired: compare ExpiresAt < current_time
	ExpiresAt time.Time `json:"expires_at"`
	// GrantDuration is the duration to add to current time when approving/denying consent (e.g., "P30D", "1h")
	// Used to calculate ExpiresAt: ExpiresAt = current_time + GrantDuration
	GrantDuration string `json:"grant_duration"`
	// Fields is the list of data fields that require consent with rich metadata
	Fields []ConsentField `json:"fields"`
	// SessionID is the session identifier for tracking the consent flow
	SessionID string `json:"session_id"`
	// ConsentPortalURL is the URL to redirect to for consent portal
	ConsentPortalURL string `json:"consent_portal_url"`
	// UpdatedBy identifies who last updated the consent (audit field)
	UpdatedBy *string `json:"updated_by,omitempty"`
}
