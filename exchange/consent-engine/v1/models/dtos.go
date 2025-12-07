package models

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ConsentField represents a field that requires consent
// Matches PolicyDecisionResponseFieldRecord DTO structure from PolicyDecisionPoint
type ConsentField struct {
	FieldName   string    `json:"fieldName"`
	SchemaID    string    `json:"schemaId"`
	DisplayName *string   `json:"displayName,omitempty"`
	Description *string   `json:"description,omitempty"`
	Owner       OwnerType `json:"owner"`
}

// ConsentRequirement represents a consent requirement for a specific owner
type ConsentRequirement struct {
	Owner      string         `json:"owner"`
	OwnerID    string         `json:"ownerId"`
	OwnerEmail string         `json:"ownerEmail"`
	Fields     []ConsentField `json:"fields"`
}

// CreateConsentRequest defines the structure for creating a consent record
// GrantDuration is optional - nil means not provided and will use default value
type CreateConsentRequest struct {
	AppID               string               `json:"appId"`
	AppName             *string              `json:"appName,omitempty"`
	ConsentRequirements []ConsentRequirement `json:"consentRequirements"`
	GrantDuration       *string              `json:"grantDuration,omitempty"`
}

// ConsentPortalActionRequest defines the structure for consent portal interactions
type ConsentPortalActionRequest struct {
	ConsentID string              `json:"consentId"`
	Action    ConsentPortalAction `json:"action"` // "approve" or "reject"
	UpdatedBy string              `json:"updatedBy"`
}

// ConsentResponseInternalView represents a simplified consent response structure for Internal API Responses
type ConsentResponseInternalView struct {
	ConsentID        string  `json:"consentId"`
	Status           string  `json:"status"`
	ConsentPortalURL *string `json:"consentPortalUrl,omitempty"` // Only present when status is pending
}

// ConsentResponsePortalView represents the user-facing consent object for the UI.
// Uses rich field information for better UX in the consent portal
type ConsentResponsePortalView struct {
	AppID      string         `json:"appId"`
	AppName    string         `json:"appName"`
	OwnerID    string         `json:"ownerId"`
	OwnerEmail string         `json:"ownerEmail"`
	Status     ConsentStatus  `json:"status"`
	Type       ConsentType    `json:"type"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	Fields     []ConsentField `json:"fields"` // Rich field information with display names and descriptions
}

// ConsentCreateResponse represents the detailed consent object returned upon creation of a consent record
// Not necessarily returned unless specifically requested
type ConsentCreateResponse struct {
	ConsentID        string         `json:"consentId"`
	OwnerID          string         `json:"ownerId"`
	OwnerEmail       string         `json:"ownerEmail"`
	AppID            string         `json:"appId"`
	AppName          *string        `json:"appName,omitempty"`
	Status           string         `json:"status"`
	Type             string         `json:"type"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	PendingExpiresAt *time.Time     `json:"pendingExpiresAt,omitempty"`
	GrantExpiresAt   *time.Time     `json:"grantExpiresAt,omitempty"`
	GrantDuration    string         `json:"grantDuration"`
	Fields           []ConsentField `json:"fields"`
	SessionID        *string        `json:"sessionId,omitempty"`
	ConsentPortalURL string         `json:"consentPortalUrl"`
	UpdatedBy        *string        `json:"updatedBy,omitempty"`
}

// ToConsentResponseInternalView converts a ConsentRecord to a simplified ConsentResponseInternalView.
// Only includes consent_portal_url when status is pending and the URL is not empty
func (cr *ConsentRecord) ToConsentResponseInternalView() ConsentResponseInternalView {
	response := ConsentResponseInternalView{
		ConsentID: cr.ConsentID.String(),
		Status:    cr.Status,
	}

	// Only include consent_portal_url when status is pending and URL is not empty
	if cr.Status == string(StatusPending) && cr.ConsentPortalURL != "" {
		portalURL := cr.ConsentPortalURL
		response.ConsentPortalURL = &portalURL
	}

	return response
}

// ToConsentResponsePortalView converts an internal ConsentRecord to a user-facing view.
// Returns rich field information including display names and descriptions for better UX
func (cr *ConsentRecord) ToConsentResponsePortalView() ConsentResponsePortalView {
	appDisplayName := cr.AppName
	if appDisplayName == nil || strings.TrimSpace(*appDisplayName) == "" {
		// Derive a display name from AppID by capitalizing and replacing underscores/dashes with spaces
		derivedName := strings.ReplaceAll(cr.AppID, "_", " ")
		derivedName = strings.ReplaceAll(derivedName, "-", " ")
		derivedName = cases.Title(language.English).String(derivedName)
		appDisplayName = &derivedName
	}

	return ConsentResponsePortalView{
		AppID:      cr.AppID,
		AppName:    *appDisplayName,
		OwnerID:    cr.OwnerID,
		OwnerEmail: cr.OwnerEmail,
		Status:     ConsentStatus(cr.Status),
		Type:       ConsentType(cr.Type),
		CreatedAt:  cr.CreatedAt,
		UpdatedAt:  cr.UpdatedAt,
		Fields:     cr.Fields, // Now includes DisplayName, Description, and Owner for rich UI rendering

	}
}

// ToConsentCreateResponse converts a ConsentRecord to a detailed ConsentCreateResponse
func (cr *ConsentRecord) ToConsentCreateResponse() *ConsentCreateResponse {
	return &ConsentCreateResponse{
		ConsentID:        cr.ConsentID.String(),
		OwnerID:          cr.OwnerID,
		OwnerEmail:       cr.OwnerEmail,
		AppID:            cr.AppID,
		AppName:          cr.AppName,
		Status:           cr.Status,
		Type:             cr.Type,
		CreatedAt:        cr.CreatedAt,
		UpdatedAt:        cr.UpdatedAt,
		PendingExpiresAt: cr.PendingExpiresAt,
		GrantExpiresAt:   cr.GrantExpiresAt,
		GrantDuration:    cr.GrantDuration,
		Fields:           cr.Fields,
		SessionID:        cr.SessionID,
		ConsentPortalURL: cr.ConsentPortalURL,
		UpdatedBy:        cr.UpdatedBy,
	}
}
