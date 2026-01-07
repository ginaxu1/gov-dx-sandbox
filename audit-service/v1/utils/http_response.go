package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gov-dx-sandbox/audit-service/v1/models"
)

// RespondWithJSON sends a JSON response with the given status code
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, log it but don't try to send another response
		// as headers have already been written
		slog.Error("Failed to encode JSON response", "error", err, "statusCode", statusCode)
		return
	}
}

// RespondWithError sends a JSON error response with the given status code
func RespondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorResp := models.ErrorResponse{
		Error: message,
	}
	if err != nil {
		errorResp.Details = err.Error()
	}

	RespondWithJSON(w, statusCode, errorResp)
}
