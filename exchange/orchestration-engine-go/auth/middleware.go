package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Validates access tokens for GraphQL requests
func AuthMiddleware(authClient *Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health check and OPTIONS requests
		if r.URL.Path == "/health" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithAuthError(w, "Authorization header is required")
			return
		}

		// Validate token format (Bearer <token>)
		if !strings.HasPrefix(authHeader, "Bearer ") {
			respondWithAuthError(w, "Invalid authorization header format. Expected 'Bearer <token>'")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token with API server
		response, err := authClient.ValidateToken(token)
		if err != nil {
			fmt.Printf("DEBUG: Token validation error: %v\n", err)
			respondWithAuthError(w, "Failed to validate token: "+err.Error())
			return
		}

		fmt.Printf("DEBUG: Token validation result: Valid=%v, ConsumerID=%s, Error=%s\n", response.Valid, response.ConsumerID, response.Error)

		if !response.Valid {
			errorMsg := response.Error
			if errorMsg == "" {
				errorMsg = "Token validation failed"
			}
			respondWithAuthError(w, errorMsg)
			return
		}

		// Add consumer ID to request context for potential use in resolvers
		r.Header.Set("X-Consumer-ID", response.ConsumerID)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// Sends an authentication error response
func respondWithAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	errorResponse := map[string]interface{}{
		"errors": []map[string]interface{}{
			{
				"message": message,
				"extensions": map[string]interface{}{
					"code": "UNAUTHENTICATED",
				},
			},
		},
	}

	json.NewEncoder(w).Encode(errorResponse)
}
