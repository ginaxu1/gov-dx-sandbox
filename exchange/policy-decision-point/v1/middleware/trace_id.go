package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// TraceIDKey is the context key for Trace ID
type TraceIDKey struct{}

const (
	// TraceIDHeader is the HTTP header name for trace ID
	TraceIDHeader = "X-Trace-ID"
)

// GetTraceIDFromContext retrieves the trace ID from the context
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}

// TraceIDMiddleware extracts or generates a trace ID and adds it to the request context
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace ID from header or generate new one
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			// Generate new trace ID if not provided
			traceID = generateTraceID()
		}

		// Add trace ID to context
		ctx := context.WithValue(r.Context(), TraceIDKey{}, traceID)

		// Add trace ID to response header for client visibility
		w.Header().Set(TraceIDHeader, traceID)

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateTraceID generates a UUID trace ID
func generateTraceID() string {
	return uuid.New().String()
}
