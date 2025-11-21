package middleware

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

// JWK represents a single JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// CustomClaims includes the standard JWT claims plus your application-specific ones.
type CustomClaims struct {
	jwt.RegisteredClaims
	OrgName string `json:"org_name"`
	Email   string `json:"email"`
}

// JWTVerifier handles JWT token verification.
type JWTVerifier struct {
	jwksURL    string
	audience   string
	orgName    string
	httpClient *http.Client
	keys       map[string]*rsa.PublicKey
	lastFetch  time.Time
}

// NewJWTVerifier creates a new JWT verifier instance.
func NewJWTVerifier(jwksURL, audience, orgName string) *JWTVerifier {
	return &JWTVerifier{
		jwksURL:  jwksURL,
		audience: audience,
		orgName:  orgName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		keys: make(map[string]*rsa.PublicKey),
	}
}

// VerifyTokenAndMatchEmail performs all validation steps and compares the token's email
// with the one from the consent record.
func (j *JWTVerifier) VerifyTokenAndMatchEmail(tokenString string, ownerEmail string) (bool, error) {
	// Step 1 & 2: Parse and Verify the Token Signature
	token, err := j.parseAndVerifySignature(tokenString)
	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return false, fmt.Errorf("token is invalid or has invalid claims")
	}

	// Step 3: Check Custom Claims (`aud` and `org_name`)
	// The `exp` and `iss` claims are validated automatically by the parser with default options.
	if !j.verifyAudience(claims.Audience, j.audience) {
		return false, fmt.Errorf("invalid audience: expected %s, got %v", j.audience, claims.Audience)
	}
	if claims.OrgName != j.orgName {
		return false, fmt.Errorf("invalid org_name: expected %s, got %s", j.orgName, claims.OrgName)
	}

	// Step 4: Extract and Compare Email
	if claims.Email == "" {
		return false, fmt.Errorf("email claim not found in token")
	}
	if claims.Email != ownerEmail {
		return false, fmt.Errorf("token email (%s) does not match consent owner email (%s)", claims.Email, ownerEmail)
	}

	slog.Info("Successfully verified token and matched email", "email", claims.Email)
	return true, nil
}

// parseAndVerifySignature handles the complex part: fetching keys and checking the signature.
func (j *JWTVerifier) parseAndVerifySignature(tokenString string) (*jwt.Token, error) {
	if err := j.ensureKeysFresh(); err != nil {
		return nil, fmt.Errorf("failed to ensure fresh keys: %w", err)
	}

	return jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing 'kid' in token header")
		}
		publicKey, exists := j.keys[kid]
		if !exists {
			// Refresh keys and try again in case the key was just rotated.
			slog.Info("Key not found, attempting to refresh JWKS", "kid", kid)
			if err := j.fetchJWKS(); err != nil {
				return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
			}
			publicKey, exists = j.keys[kid]
			if !exists {
				return nil, fmt.Errorf("no public key found for kid after refresh: %s", kid)
			}
		}
		return publicKey, nil
	})
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

// verifyAudience checks if the token audience contains the expected audience
func (j *JWTVerifier) verifyAudience(tokenAudience []string, expectedAudience string) bool {
	if len(tokenAudience) == 0 {
		return false
	}
	for _, aud := range tokenAudience {
		if aud == expectedAudience {
			return true
		}
	}
	return false
}

// ensureKeysFresh ensures we have fresh JWKS keys (refreshes if older than 1 hour)
func (j *JWTVerifier) ensureKeysFresh() error {
	if len(j.keys) == 0 || time.Since(j.lastFetch) > time.Hour {
		return j.fetchJWKS()
	}
	return nil
}

// VerifyToken verifies a JWT token and returns the claims (legacy method for backward compatibility)
func (j *JWTVerifier) VerifyToken(tokenString string) (*jwt.Token, error) {
	return j.parseAndVerifySignature(tokenString)
}

// ExtractEmailFromToken extracts the email from a verified JWT token (legacy method for backward compatibility)
func (j *JWTVerifier) ExtractEmailFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	if claims.Email == "" {
		return "", fmt.Errorf("email claim not found in token")
	}

	return claims.Email, nil
}

// VerifyAndExtractEmail verifies a JWT token and extracts the email (legacy method for backward compatibility)
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
