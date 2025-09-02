// Package utils provides common utility functions for the project
package utils

import (
	"log/slog"
	"net/http"
	"runtime/debug"
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
