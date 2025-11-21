package handlers

import (
	"net/http"

	"github.com/gov-dx-sandbox/exchange/consent-engine/models"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// validConsentStatuses is a list of valid consent status values
var validConsentStatuses = []string{
	string(models.StatusPending),
	string(models.StatusApproved),
	string(models.StatusRejected),
	string(models.StatusExpired),
	string(models.StatusRevoked),
}

// isValidConsentStatus checks if a status string is a valid consent status
func isValidConsentStatus(status string) bool {
	for _, validStatus := range validConsentStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// validateConsentStatus validates a status string and returns an error response if invalid
func validateConsentStatus(w http.ResponseWriter, status string) bool {
	if status != "" && !isValidConsentStatus(status) {
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{
			Error: "status must be one of: pending, approved, rejected, expired, revoked",
		})
		return false
	}
	return true
}

// getDefaultReason returns a default reason message based on status
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
