package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWKS represents the JSON Web Key Set structure
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWTVerifier handles JWT token verification using Asgardeo JWKS
type JWTVerifier struct {
	jwksURL    string
	issuer     string
	audience   string
	httpClient *http.Client
	keys       map[string]*rsa.PublicKey
	lastFetch  time.Time
}

// NewJWTVerifier creates a new JWT verifier instance
func NewJWTVerifier(jwksURL, issuer, audience string) *JWTVerifier {
	return &JWTVerifier{
		jwksURL:  jwksURL,
		issuer:   issuer,
		audience: audience,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		keys: make(map[string]*rsa.PublicKey),
	}
}

// fetchJWKS fetches the JWKS from the Asgardeo endpoint
func (j *JWTVerifier) fetchJWKS() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("Fetching JWKS", "url", j.jwksURL)
	req, err := http.NewRequestWithContext(ctx, "GET", j.jwksURL, nil)
	if err != nil {
		slog.Error("Failed to create JWKS request", "url", j.jwksURL, "error", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		slog.Error("Failed to fetch JWKS", "url", j.jwksURL, "error", err)
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("JWKS response received", "url", j.jwksURL, "status_code", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		slog.Error("JWKS endpoint returned non-200 status", "url", j.jwksURL, "status_code", resp.StatusCode)
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Clear existing keys
	j.keys = make(map[string]*rsa.PublicKey)

	// Process each key
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			publicKey, err := j.buildRSAPublicKey(key.N, key.E)
			if err != nil {
				slog.Warn("Failed to build RSA public key", "kid", key.Kid, "error", err)
				continue
			}
			j.keys[key.Kid] = publicKey
		}
	}

	j.lastFetch = time.Now()
	slog.Info("Successfully fetched JWKS", "keys_count", len(j.keys))
	return nil
}

// buildRSAPublicKey constructs an RSA public key from modulus and exponent
func (j *JWTVerifier) buildRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	// Decode base64url encoded modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode base64url encoded exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Validate exponent
	if !e.IsInt64() || e.Int64() < 2 {
		return nil, fmt.Errorf("invalid exponent")
	}

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// ensureKeysFresh ensures we have fresh JWKS keys (refreshes if older than 1 hour)
func (j *JWTVerifier) ensureKeysFresh() error {
	if len(j.keys) == 0 || time.Since(j.lastFetch) > time.Hour {
		return j.fetchJWKS()
	}
	return nil
}

// VerifyToken verifies a JWT token and returns the claims
func (j *JWTVerifier) VerifyToken(tokenString string) (*jwt.Token, error) {
	slog.Debug("Starting JWT verification", "jwks_url", j.jwksURL, "issuer", j.issuer, "audience", j.audience)

	// Ensure we have fresh keys
	if err := j.ensureKeysFresh(); err != nil {
		slog.Error("Failed to ensure fresh keys", "error", err)
		return nil, fmt.Errorf("failed to ensure fresh keys: %w", err)
	}
	slog.Debug("Keys are fresh", "keys_count", len(j.keys))

	// Parse the token with custom claims validation
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		slog.Debug("Parsing token", "alg", token.Header["alg"], "kid", token.Header["kid"])

		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			slog.Error("Unexpected signing method", "alg", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			slog.Error("Missing or invalid kid in token header", "kid", token.Header["kid"])
			return nil, fmt.Errorf("missing or invalid kid in token header")
		}

		// Get the public key for this kid
		publicKey, exists := j.keys[kid]
		if !exists {
			slog.Error("No public key found for kid", "kid", kid, "available_keys", len(j.keys))
			return nil, fmt.Errorf("no public key found for kid: %s", kid)
		}

		slog.Debug("Found public key for kid", "kid", kid)
		return publicKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		slog.Error("Token parsing failed", "error", err)
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if !token.Valid {
		slog.Error("Token is not valid")
		return nil, fmt.Errorf("token is not valid")
	}

	slog.Debug("Token parsed successfully, validating claims")

	// Validate custom claims
	if err := j.validateClaims(token); err != nil {
		slog.Error("Token claims validation failed", "error", err)
		return nil, fmt.Errorf("token claims validation failed: %w", err)
	}

	slog.Debug("Token verification completed successfully")
	return token, nil
}

// validateClaims validates the iss, aud, and exp claims
func (j *JWTVerifier) validateClaims(token *jwt.Token) error {
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		slog.Error("Invalid token claims type")
		return fmt.Errorf("invalid token claims")
	}

	slog.Debug("Validating claims", "iss", (*claims)["iss"], "aud", (*claims)["aud"], "exp", (*claims)["exp"])

	// Validate issuer (iss)
	if iss, ok := (*claims)["iss"].(string); ok {
		if iss != j.issuer {
			slog.Error("Invalid issuer", "expected", j.issuer, "got", iss)
			return fmt.Errorf("invalid issuer: expected %s, got %s", j.issuer, iss)
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
			if audience == j.audience {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid audience: expected %s, got %v", j.audience, audiences)
		}
	} else {
		return fmt.Errorf("missing audience claim")
	}

	// Validate expiry (exp)
	if exp, ok := (*claims)["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Now().After(expTime) {
			return fmt.Errorf("token has expired: expired at %v", expTime)
		}
	} else {
		return fmt.Errorf("missing or invalid expiry claim")
	}

	return nil
}

// ExtractEmailFromToken extracts the email from a verified JWT token
func (j *JWTVerifier) ExtractEmailFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Try different possible email claim names
	emailFields := []string{"email", "sub", "preferred_username"}

	for _, field := range emailFields {
		if email, ok := (*claims)[field].(string); ok && email != "" {
			return email, nil
		}
	}

	return "", fmt.Errorf("email not found in token claims")
}

// VerifyAndExtractEmail verifies a JWT token and extracts the email
func (j *JWTVerifier) VerifyAndExtractEmail(tokenString string) (string, error) {
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
