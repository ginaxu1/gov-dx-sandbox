// Package utils provides common utility functions for the project
package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
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

// RespondWithError is a utility function to write a JSON error response
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, ErrorResponse{Error: message})
}

// RespondWithSuccess is a utility function to write a JSON success response
func RespondWithSuccess(w http.ResponseWriter, code int, data interface{}) {
	RespondWithJSON(w, code, data)
}

// ExtractIDFromPath extracts the ID from a URL path by taking the last segment
func ExtractIDFromPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// ParseJSONRequest parses a JSON request body into the target struct
func ParseJSONRequest(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// CreateCollectionResponse creates a standardized collection response with count
func CreateCollectionResponse(items interface{}, count int) map[string]interface{} {
	return map[string]interface{}{
		"items": items,
		"count": count,
	}
}
