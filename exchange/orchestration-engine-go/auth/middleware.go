package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TrustedAuthMiddleware handles pre-authenticated requests from Choreo Gateway
// The Gateway has already validated the JWT and extracted the consumer ID
func TrustedAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health check and OPTIONS requests
		if r.URL.Path == "/health" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract consumer ID from X-Consumer-ID header set by Choreo Gateway
		consumerID := r.Header.Get("X-Consumer-ID")
		if consumerID == "" {
			respondWithAuthError(w, "X-Consumer-ID header is required - request not properly authenticated by Choreo Gateway")
			return
		}

		// Log the authenticated request for debugging
		fmt.Printf("DEBUG: Trusted request from consumer: %s\n", consumerID)
		fmt.Printf("DEBUG: Request headers: Authorization=%s, X-Consumer-ID=%s, User-Agent=%s, RemoteAddr=%s\n",
			r.Header.Get("Authorization"), consumerID, r.Header.Get("User-Agent"), r.RemoteAddr)

		// The request is already authenticated by Choreo Gateway
		// We trust that the Gateway has validated the JWT properly
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
