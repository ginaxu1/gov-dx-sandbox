package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Owner represents the owner enum (matches PolicyDecisionPoint Owner type)
type Owner string

const (
	OwnerCitizen Owner = "citizen"
)

// ConsentRequest defines the structure for creating a consent record
// GrantDuration is optional - nil means not provided and will use default value
type ConsentRequest struct {
	AppID               string               `json:"app_id"`
	ConsentRequirements []ConsentRequirement `json:"consent_requirements"`
	GrantDuration       *string              `json:"grant_duration,omitempty"`
}

// ConsentRequirement represents a consent requirement for a specific owner
type ConsentRequirement struct {
	Owner   string         `json:"owner"`
	OwnerID string         `json:"owner_id"`
	Fields  []ConsentField `json:"fields"`
}

// ConsentField represents a field that requires consent
// Matches PolicyDecisionResponseFieldRecord structure from PolicyDecisionPoint
type ConsentField struct {
	FieldName   string  `json:"fieldName"`
	SchemaID    string  `json:"schemaId"`
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Owner       *Owner  `json:"owner,omitempty"`
}

// UpdateConsentRequest defines the structure for updating a consent record
// Status is optional (omitempty) to support PATCH operations where status may not be provided
// Optional fields use pointers to distinguish between "not provided" (nil) and "provided as empty" (pointer to empty value)
type UpdateConsentRequest struct {
	Status        ConsentStatus `json:"status,omitempty"`         // Required for PUT, optional for PATCH
	UpdatedBy     *string       `json:"updated_by,omitempty"`     // Optional - nil means not provided
	GrantDuration *string       `json:"grant_duration,omitempty"` // Optional - nil means not provided
	Fields        *[]string     `json:"fields,omitempty"`         // Optional - nil means not provided
	Reason        *string       `json:"reason,omitempty"`         // Optional - nil means not provided
}

// ConsentPortalRequest defines the structure for consent portal interactions
// Reason is optional - nil means not provided, pointer allows distinguishing from empty string
type ConsentPortalRequest struct {
	ConsentID uuid.UUID `json:"consent_id"`
	Action    string    `json:"action"` // "approve" or "reject"
	DataOwner string    `json:"data_owner"`
	Reason    *string   `json:"reason,omitempty"`
}

// ConsentResponse represents the simplified response for consent operations
type ConsentResponse struct {
	ConsentID        uuid.UUID `json:"consent_id"`
	Status           string    `json:"status"`
	ConsentPortalURL *string   `json:"consent_portal_url,omitempty"` // Only present when status is pending
}

// ConsentPortalView represents the user-facing consent object for the UI.
// Uses rich field information for better UX in the consent portal
type ConsentPortalView struct {
	AppDisplayName string         `json:"app_display_name"`
	CreatedAt      time.Time      `json:"created_at"`
	Fields         []ConsentField `json:"fields"` // Rich field information with display names and descriptions
	OwnerName      string         `json:"owner_name"`
	OwnerEmail     string         `json:"owner_email"`
	Status         string         `json:"status"`
	Type           string         `json:"type"`
}

// Legacy structures for backwards compatibility (deprecated)
type DataField struct {
	OwnerType  string   `json:"owner_type,omitempty"`
	OwnerID    string   `json:"owner_id"`
	OwnerEmail string   `json:"owner_email"`
	Fields     []string `json:"fields"`
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

// ToConsentPortalView converts an internal ConsentRecord to a user-facing view.
// Returns rich field information including display names and descriptions for better UX
func (cr *ConsentRecord) ToConsentPortalView() *ConsentPortalView {
	// Simple mapping for app_id to a human-readable name.
	appDisplayName := strings.ReplaceAll(cr.AppID, "-", " ")
	appDisplayName = cases.Title(language.English).String(appDisplayName)

	ownerName := strings.ReplaceAll(cr.OwnerID, "-", " ")
	ownerName = cases.Title(language.English).String(ownerName)

	return &ConsentPortalView{
		AppDisplayName: appDisplayName,
		CreatedAt:      cr.CreatedAt,
		Fields:         cr.Fields, // Now includes DisplayName, Description, and Owner for rich UI rendering
		OwnerName:      ownerName,
		OwnerEmail:     cr.OwnerEmail,
		Status:         cr.Status,
		Type:           cr.Type,
	}
}
