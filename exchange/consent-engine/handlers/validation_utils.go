package handlers

import (
	"fmt"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
)

var validConsentStatuses = map[models.ConsentStatus]struct{}{
	models.StatusPending:  {},
	models.StatusApproved: {},
	models.StatusRejected: {},
	models.StatusExpired:  {},
	models.StatusRevoked:  {},
}

// isValidConsentStatus checks if a consent status is one of the supported values.
func isValidConsentStatus(status models.ConsentStatus) bool {
	_, ok := validConsentStatuses[status]
	return ok
}

// validateConsentStatus validates a string status and returns a typed version.
// Returns an empty status when no status was provided.
func validateConsentStatus(status string) (models.ConsentStatus, error) {
	if status == "" {
		return "", nil
	}

	typedStatus := models.ConsentStatus(status)
	if !isValidConsentStatus(typedStatus) {
		return "", fmt.Errorf("status must be one of: pending, approved, rejected, expired, revoked")
	}

	return typedStatus, nil
}

// getDefaultReason maps a status to a user-friendly default reason.
func getDefaultReason(status models.ConsentStatus) string {
	switch status {
	case models.StatusApproved:
		return "Consent approved via API"
	case models.StatusRejected:
		return "Consent rejected via API"
	case models.StatusExpired:
		return "Consent expired via API"
	case models.StatusRevoked:
		return "Consent revoked via API"
	case models.StatusPending:
		return "Consent reset to pending via API"
	default:
		return "Consent updated via API"
	}
}
