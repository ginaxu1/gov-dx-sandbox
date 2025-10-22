package shared

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateRandomString generates a cryptographically random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateToken generates a random token (32 bytes by default)
func GenerateToken() (string, error) {
	return GenerateRandomString(32)
}

// ScopesToJSON converts a slice of scopes to JSON string
func ScopesToJSON(scopes []string) string {
	if len(scopes) == 0 {
		return "[]"
	}
	return fmt.Sprintf(`["%s"]`, strings.Join(scopes, `","`))
}

// JSONToScopes converts a JSON string to a slice of scopes
func JSONToScopes(scopesJSON string) []string {
	if scopesJSON == "[]" || scopesJSON == "" {
		return []string{}
	}
	// Simple JSON parsing for scopes array
	scopesJSON = strings.Trim(scopesJSON, "[]")
	if scopesJSON == "" {
		return []string{}
	}
	scopes := strings.Split(scopesJSON, ",")
	for i, scope := range scopes {
		scopes[i] = strings.Trim(scope, `"`)
	}
	return scopes
}

// IsJWTToken checks if a token is a JWT (has 3 parts separated by dots)
func IsJWTToken(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 3
}

// CreateTokenResponse creates a standardized token response
func CreateTokenResponse(accessToken, refreshToken, scope string, expiresIn int) map[string]interface{} {
	response := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"scope":        scope,
	}

	if refreshToken != "" {
		response["refresh_token"] = refreshToken
	}

	return response
}
