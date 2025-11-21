package handlers

import (
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// handleConsentError handles common consent-related errors and returns appropriate HTTP responses
func handleConsentError(w http.ResponseWriter, err error, operation string) {
	if err == nil {
		return
	}

	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "not found"):
		utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{
			Error: "Consent record not found",
		})
	case strings.Contains(errMsg, "invalid action"):
		utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{
			Error: errMsg,
		})
	default:
		utils.RespondWithJSON(w, http.StatusInternalServerError, utils.ErrorResponse{
			Error: "Failed to " + operation + ": " + errMsg,
		})
	}
}

// extractConsentIDFromPath extracts consent ID from URL path
func extractConsentIDFromPath(path, prefix string) string {
	id := strings.TrimPrefix(path, prefix)
	id = strings.TrimPrefix(id, "/")
	// Remove any trailing slashes or additional path segments
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	return id
}
