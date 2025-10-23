package shared

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
// Uses json.Marshal for proper JSON encoding to handle special characters safely
func ScopesToJSON(scopes []string) string {
	if len(scopes) == 0 {
		return "[]"
	}

	// Use json.Marshal for proper JSON encoding
	// This handles special characters like quotes, backslashes, etc. safely
	jsonBytes, err := json.Marshal(scopes)
	if err != nil {
		// Fallback to empty array if marshaling fails
		return "[]"
	}

	return string(jsonBytes)
}

// JSONToScopes converts a JSON string to a slice of scopes
// Uses json.Unmarshal for proper JSON parsing to handle special characters safely
func JSONToScopes(scopesJSON string) []string {
	if scopesJSON == "[]" || scopesJSON == "" {
		return []string{}
	}

	// Use json.Unmarshal for proper JSON parsing
	// This handles special characters like quotes, backslashes, etc. safely
	var scopes []string
	err := json.Unmarshal([]byte(scopesJSON), &scopes)
	if err != nil {
		// Fallback to empty slice if parsing fails
		return []string{}
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

// TestScopesWithSpecialCharacters demonstrates how the fixed functions handle special characters
// This function can be used for testing purposes
func TestScopesWithSpecialCharacters() {
	// Test cases with special characters that could cause issues with manual string concatenation
	testCases := [][]string{
		{"read:data", "write:data"},
		{"scope with spaces", "scope\"with\"quotes"},
		{"scope\\with\\backslashes", "scope\nwith\nnewlines"},
		{"scope\twith\ttabs", "scope/with/slashes"},
		{"scope with unicode: 测试", "scope with emoji: 🚀"},
		{"scope with special chars: !@#$%^&*()", "scope with brackets: [test]"},
	}

	for i, scopes := range testCases {
		// Convert to JSON
		jsonStr := ScopesToJSON(scopes)

		// Convert back from JSON
		parsedScopes := JSONToScopes(jsonStr)

		// Verify round-trip conversion works
		if len(scopes) != len(parsedScopes) {
			fmt.Printf("Test case %d failed: length mismatch\n", i)
			continue
		}

		for j, scope := range scopes {
			if scope != parsedScopes[j] {
				fmt.Printf("Test case %d failed: scope mismatch at index %d\n", i, j)
				break
			}
		}

		fmt.Printf("Test case %d passed: %v -> %s -> %v\n", i, scopes, jsonStr, parsedScopes)
	}
}
