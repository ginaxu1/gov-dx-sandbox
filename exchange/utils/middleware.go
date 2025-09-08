// Package utils provides common utility functions for the project
package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

// PanicRecoveryMiddleware is an HTTP middleware that recovers from panics
// It logs the error and stack trace, then returns a 500 Internal Server Error
func PanicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("handler panic recovered", "error", err, "stack", string(debug.Stack()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// HandlerFunc represents a function that returns a result, status code, and error
type HandlerFunc func() (interface{}, int, error)

// PathHandlerFunc represents a function that takes a path parameter and returns a result, status code, and error
type PathHandlerFunc func(string) (interface{}, int, error)

// GenericHandler handles common HTTP patterns with error handling
func GenericHandler(w http.ResponseWriter, r *http.Request, handler HandlerFunc) {
	result, statusCode, err := handler()
	if err != nil {
		slog.Error("Request failed", "error", err, "path", r.URL.Path, "method", r.Method)
		RespondWithJSON(w, statusCode, ErrorResponse{Error: err.Error()})
		return
	}
	RespondWithJSON(w, statusCode, result)
}

// JSONHandler handles requests requiring JSON body parsing
func JSONHandler(w http.ResponseWriter, r *http.Request, target interface{}, handler HandlerFunc) {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		slog.Warn("Invalid JSON body", "error", err, "path", r.URL.Path)
		RespondWithJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	GenericHandler(w, r, handler)
}

// PathHandler handles path-based operations with ID extraction
func PathHandler(w http.ResponseWriter, r *http.Request, prefix string, handler PathHandlerFunc) {
	id := strings.TrimPrefix(r.URL.Path, prefix)
	if id == "" {
		slog.Warn("Missing ID in path", "path", r.URL.Path)
		RespondWithJSON(w, http.StatusBadRequest, ErrorResponse{Error: "ID is required"})
		return
	}

	GenericHandler(w, r, func() (interface{}, int, error) {
		return handler(id)
	})
}
