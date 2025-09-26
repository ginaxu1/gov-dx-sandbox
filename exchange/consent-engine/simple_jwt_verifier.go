package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SimpleJWTVerifier handles JWT token validation without network calls
type SimpleJWTVerifier struct {
	expectedIssuer   string
	expectedAudience string
}

// NewSimpleJWTVerifier creates a new simple JWT verifier
func NewSimpleJWTVerifier(expectedIssuer, expectedAudience string) *SimpleJWTVerifier {
	return &SimpleJWTVerifier{
		expectedIssuer:   expectedIssuer,
		expectedAudience: expectedAudience,
	}
}

// VerifyToken verifies a JWT token without signature verification
func (j *SimpleJWTVerifier) VerifyToken(tokenString string) (*jwt.Token, error) {
	slog.Debug("Verifying JWT token without signature verification",
		"expected_issuer", j.expectedIssuer,
		"expected_audience", j.expectedAudience)

	// Parse the token
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, &jwt.MapClaims{})

	if err != nil {
		slog.Error("Failed to parse JWT token", "error", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate basic claims
	if err := j.validateClaims(token); err != nil {
		slog.Error("Token claims validation failed", "error", err)
		return nil, fmt.Errorf("token claims validation failed: %w", err)
	}

	slog.Debug("JWT token verification completed successfully")
	return token, nil
}

// validateClaims validates the iss, aud, and exp claims
func (j *SimpleJWTVerifier) validateClaims(token *jwt.Token) error {
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		slog.Error("Invalid token claims type")
		return fmt.Errorf("invalid token claims")
	}

	slog.Debug("Validating claims", "iss", (*claims)["iss"], "aud", (*claims)["aud"], "exp", (*claims)["exp"])

	// Validate issuer (iss)
	if iss, ok := (*claims)["iss"].(string); ok {
		if iss != j.expectedIssuer {
			slog.Error("Invalid issuer", "expected", j.expectedIssuer, "got", iss)
			return fmt.Errorf("invalid issuer: expected %s, got %s", j.expectedIssuer, iss)
		}
		slog.Debug("Issuer validation passed", "iss", iss)
	} else {
		slog.Error("Missing or invalid issuer claim", "iss", (*claims)["iss"])
		return fmt.Errorf("missing or invalid issuer claim")
	}

	// Validate audience (aud)
	if aud, ok := (*claims)["aud"]; ok {
		// aud can be a string or array of strings
		var audiences []string
		switch v := aud.(type) {
		case string:
			audiences = []string{v}
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					audiences = append(audiences, str)
				}
			}
		default:
			return fmt.Errorf("invalid audience claim type")
		}

		// Check if our expected audience is in the list
		found := false
		for _, audience := range audiences {
			if audience == j.expectedAudience {
				found = true
				break
			}
		}
		if !found {
			slog.Error("Invalid audience", "expected", j.expectedAudience, "got", audiences)
			return fmt.Errorf("invalid audience: expected %s, got %v", j.expectedAudience, audiences)
		}
		slog.Debug("Audience validation passed", "audiences", audiences)
	} else {
		slog.Error("Missing audience claim")
		return fmt.Errorf("missing audience claim")
	}

	// Validate expiry (exp)
	if exp, ok := (*claims)["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Now().After(expTime) {
			slog.Error("Token has expired", "expired_at", expTime)
			return fmt.Errorf("token has expired: expired at %v", expTime)
		}
		slog.Debug("Expiry validation passed", "expires_at", expTime)
	} else {
		slog.Error("Missing or invalid expiry claim", "exp", (*claims)["exp"])
		return fmt.Errorf("missing or invalid expiry claim")
	}

	return nil
}

// ExtractEmailFromToken extracts the email from a verified JWT token
func (j *SimpleJWTVerifier) ExtractEmailFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Try different possible email claim names
	emailFields := []string{"email", "sub", "preferred_username"}

	for _, field := range emailFields {
		if email, ok := (*claims)[field].(string); ok && email != "" {
			slog.Debug("Email extracted from token", "field", field, "email", email)
			return email, nil
		}
	}

	slog.Error("Email not found in token claims", "available_claims", getClaimKeys(claims))
	return "", fmt.Errorf("email not found in token claims")
}

// getClaimKeys returns all claim keys for debugging
func getClaimKeys(claims *jwt.MapClaims) []string {
	keys := make([]string, 0, len(*claims))
	for k := range *claims {
		keys = append(keys, k)
	}
	return keys
}

// VerifyAndExtractEmail verifies a JWT token and extracts the email
func (j *SimpleJWTVerifier) VerifyAndExtractEmail(tokenString string) (string, error) {
	token, err := j.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	email, err := j.ExtractEmailFromToken(token)
	if err != nil {
		return "", err
	}

	return email, nil
}
