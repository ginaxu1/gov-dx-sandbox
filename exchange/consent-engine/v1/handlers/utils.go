package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// respondWithJSON sends a JSON response with the given status code
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, log it but don't try to send another response
		// as headers have already been written
		slog.Error("Failed to encode JSON response", "error", err, "statusCode", statusCode)
		return
	}
}

// respondWithError sends a JSON error response with the given status code
func respondWithError(w http.ResponseWriter, statusCode int, errorCode models.ConsentErrorCode, message string) {
	response := ErrorResponse{}
	response.Error.Code = string(errorCode)
	response.Error.Message = message

	respondWithJSON(w, statusCode, response)
}

// containsError checks if an error message contains a specific substring
func containsError(err error, substr string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), substr)
}
