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
	httpClient *http.Client
	keys       map[string]*rsa.PublicKey
	lastFetch  time.Time
}

// NewJWTVerifier creates a new JWT verifier instance
func NewJWTVerifier(jwksURL string) *JWTVerifier {
	return &JWTVerifier{
		jwksURL: jwksURL,
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

	req, err := http.NewRequestWithContext(ctx, "GET", j.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
	// Ensure we have fresh keys
	if err := j.ensureKeysFresh(); err != nil {
		return nil, fmt.Errorf("failed to ensure fresh keys: %w", err)
	}

	// Parse the token with custom claims validation that skips exp check
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid kid in token header")
		}

		// Get the public key for this kid
		publicKey, exists := j.keys[kid]
		if !exists {
			return nil, fmt.Errorf("no public key found for kid: %s", kid)
		}

		return publicKey, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return token, nil
}

// ExtractEmailFromToken extracts the email from a verified JWT token
func (j *JWTVerifier) ExtractEmailFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	// Try different possible email claim names
	emailFields := []string{"email", "sub", "preferred_username"}

	for _, field := range emailFields {
		if email, ok := claims[field].(string); ok && email != "" {
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
