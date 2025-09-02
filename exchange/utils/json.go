// Package utils provides common utility functions for the project
package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse defines the structure for a standard JSON error message
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondWithJSON is a utility function to write a JSON response
// It sets the Content-Type header, writes the HTTP status code, and encodes the payload
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}
