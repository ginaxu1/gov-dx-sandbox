package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
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
// Only includes consent_portal_url when status is pending
func (cr *ConsentRecord) ToConsentResponse() ConsentResponse {
	response := ConsentResponse{
		ConsentID: cr.ConsentID,
		Status:    cr.Status,
	}

	// Only include consent_portal_url when status is pending
	if cr.Status == string(StatusPending) {
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
	duration, err := parseISODuration(grantDuration)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid grant duration format: %w", err)
	}
	return createdAt.Add(duration), nil
}

// parseISODuration parses an ISO 8601 duration string and returns the duration
// Supports formats like: P30D, P1M, P1Y, PT1H, PT30M, P1Y2M3DT4H5M6S
func parseISODuration(duration string) (time.Duration, error) {
	if duration == "" {
		// Default to 1 hour if no duration specified
		return time.Hour, nil
	}

	// Check if it's ISO 8601 format (starts with 'P')
	if len(duration) > 0 && duration[0] == 'P' {
		return parseISO8601Duration(duration)
	}

	// Fallback to legacy format parsing
	return utils.ParseExpiryTime(duration)
}

// parseISO8601Duration parses an ISO 8601 duration string into a time.Duration
func parseISO8601Duration(duration string) (time.Duration, error) {
	// Validate ISO 8601 duration format
	if !isValidISODuration(duration) {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", duration)
	}

	// Remove the 'P' prefix
	if len(duration) == 0 || duration[0] != 'P' {
		return 0, fmt.Errorf("duration must start with 'P'")
	}
	duration = duration[1:]

	var total time.Duration
	var err error

	// Check if there's a time component (starts with 'T')
	timeIndex := -1
	for i, char := range duration {
		if char == 'T' {
			timeIndex = i
			break
		}
	}

	// Parse date components (before 'T' or entire string if no 'T')
	datePart := duration
	if timeIndex != -1 {
		datePart = duration[:timeIndex]
	}

	// Parse years
	years, datePart, err := parseComponent(datePart, "Y")
	if err != nil {
		return 0, err
	}
	total += time.Duration(years) * 365 * 24 * time.Hour

	// Parse months
	months, datePart, err := parseComponent(datePart, "M")
	if err != nil {
		return 0, err
	}
	total += time.Duration(months) * 30 * 24 * time.Hour // Approximate month as 30 days

	// Parse days
	days, _, err := parseComponent(datePart, "D")
	if err != nil {
		return 0, err
	}
	total += time.Duration(days) * 24 * time.Hour

	// Parse time components (after 'T')
	if timeIndex != -1 {
		timePart := duration[timeIndex+1:]

		// Parse hours
		hours, timePart, err := parseComponent(timePart, "H")
		if err != nil {
			return 0, err
		}
		total += time.Duration(hours) * time.Hour

		// Parse minutes
		minutes, timePart, err := parseComponent(timePart, "M")
		if err != nil {
			return 0, err
		}
		total += time.Duration(minutes) * time.Minute

		// Parse seconds
		seconds, _, err := parseComponent(timePart, "S")
		if err != nil {
			return 0, err
		}
		total += time.Duration(seconds) * time.Second
	}

	return total, nil
}

// isValidISODuration validates if a string is a valid ISO 8601 duration
func isValidISODuration(duration string) bool {
	// ISO 8601 duration pattern: P(\d+Y)?(\d+M)?(\d+D)?(T(\d+H)?(\d+M)?(\d+(\.\d+)?S)?)?
	pattern := `^P(?:\d+Y)?(?:\d+M)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`
	matched, _ := regexp.MatchString(pattern, duration)
	return matched
}

// parseComponent extracts a numeric component from a duration string
func parseComponent(duration, suffix string) (int, string, error) {
	pattern := fmt.Sprintf(`(\d+)%s`, suffix)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(duration)

	if len(matches) == 0 {
		return 0, duration, nil
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, duration, err
	}

	// Remove the matched part from the duration string
	remaining := re.ReplaceAllString(duration, "")
	return value, remaining, nil
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
