package shared

import (
	"fmt"
	"net/http"
	"strings"
)

// ExtractAccessToken extracts the access token from the Authorization header
func ExtractAccessToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("missing access token")
	}

	return token, nil
}

// HasRequiredScope checks if the user has the required scope
func HasRequiredScope(userScopes []string, requiredScope string) bool {
	for _, scope := range userScopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the user has any of the required scopes
func HasAnyScope(userScopes []string, requiredScopes ...string) bool {
	for _, requiredScope := range requiredScopes {
		if HasRequiredScope(userScopes, requiredScope) {
			return true
		}
	}
	return false
}

// HasAllScopes checks if the user has all of the required scopes
func HasAllScopes(userScopes []string, requiredScopes ...string) bool {
	for _, requiredScope := range requiredScopes {
		if !HasRequiredScope(userScopes, requiredScope) {
			return false
		}
	}
	return true
}
