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
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	sharedutils "github.com/gov-dx-sandbox/api-server-go/shared/utils"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	authutils "github.com/gov-dx-sandbox/api-server-go/v1/utils"
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

// JWTAuthMiddleware provides JWT authentication functionality
type JWTAuthMiddleware struct {
	jwksURL          string
	expectedIssuer   string
	expectedAudience string
	orgName          string
	httpClient       *http.Client
	keys             map[string]*rsa.PublicKey
	lastFetch        time.Time
}

// JWTAuthConfig contains configuration for JWT authentication
type JWTAuthConfig struct {
	JWKSURL          string
	ExpectedIssuer   string
	ExpectedAudience string
	OrgName          string
	Timeout          time.Duration
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(config JWTAuthConfig) *JWTAuthMiddleware {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &JWTAuthMiddleware{
		jwksURL:          config.JWKSURL,
		expectedIssuer:   config.ExpectedIssuer,
		expectedAudience: config.ExpectedAudience,
		orgName:          config.OrgName,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		keys: make(map[string]*rsa.PublicKey),
	}
}

// AuthenticateJWT returns a middleware function that validates JWT tokens
func (j *JWTAuthMiddleware) AuthenticateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health and debug endpoints
		if j.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		tokenString, err := authutils.ExtractBearerToken(r)
		if err != nil {
			slog.Warn("Failed to extract bearer token", "error", err, "path", r.URL.Path, "method", r.Method)
			sharedutils.RespondWithError(w, http.StatusUnauthorized, "Invalid or missing authorization header")
			return
		}

		// Validate and parse the token
		user, authCtx, err := j.validateToken(tokenString)
		if err != nil {
			slog.Warn("Token validation failed", "error", err, "path", r.URL.Path, "method", r.Method)
			sharedutils.RespondWithError(w, http.StatusUnauthorized, "Invalid access token")
			return
		}

		// Check if token is expired
		if user.IsTokenExpired() {
			slog.Warn("Token is expired", "expiry", user.ExpiresAt, "user", user.Email)
			sharedutils.RespondWithError(w, http.StatusUnauthorized, "Access token has expired")
			return
		}

		// Add user and auth context to request context
		ctx := authutils.SetAuthenticatedUser(r.Context(), user)
		ctx = authutils.SetAuthContext(ctx, authCtx)

		// Log successful authentication
		slog.Info("User authenticated successfully",
			"user_id", user.IdpUserID,
			"email", user.Email,
			"roles", user.Roles,
			"path", r.URL.Path,
			"method", r.Method)

		// Continue to the next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates a JWT token and returns the authenticated user
func (j *JWTAuthMiddleware) validateToken(tokenString string) (*models.AuthenticatedUser, *models.AuthContext, error) {
	// Ensure we have fresh JWKS keys
	if err := j.ensureKeysFresh(); err != nil {
		return nil, nil, fmt.Errorf("failed to ensure fresh keys: %w", err)
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &models.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing 'kid' in token header")
		}

		// Find the public key
		publicKey, exists := j.keys[kid]
		if !exists {
			// Try to refresh keys once
			slog.Info("Key not found, refreshing JWKS", "kid", kid)
			if err := j.fetchJWKS(); err != nil {
				return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
			}
			publicKey, exists = j.keys[kid]
			if !exists {
				return nil, fmt.Errorf("no public key found for kid: %s", kid)
			}
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*models.UserClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("invalid token claims")
	}

	// Validate standard claims
	if err := j.validateStandardClaims(claims); err != nil {
		return nil, nil, fmt.Errorf("claim validation failed: %w", err)
	}

	// Create authenticated user from claims
	user := models.NewAuthenticatedUser(claims)

	// Create auth context
	authCtx := &models.AuthContext{
		User:        user,
		Token:       tokenString,
		IssuedBy:    claims.Issuer,
		Audience:    claims.Audience,
		Permissions: user.GetPermissions(),
	}

	return user, authCtx, nil
}

// validateStandardClaims validates the standard JWT claims
func (j *JWTAuthMiddleware) validateStandardClaims(claims *models.UserClaims) error {
	now := time.Now()

	// Check if token is expired
	if !claims.ExpiresAt.IsZero() && now.After(claims.ExpiresAt) {
		return fmt.Errorf("token is expired")
	}

	// Check not before time
	if !claims.NotBefore.IsZero() && now.Before(claims.NotBefore) {
		return fmt.Errorf("token is not valid yet")
	}

	// Validate issuer
	if j.expectedIssuer != "" && claims.Issuer != j.expectedIssuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", j.expectedIssuer, claims.Issuer)
	}

	// Validate audience
	if j.expectedAudience != "" && !j.containsAudience(claims.Audience, j.expectedAudience) {
		return fmt.Errorf("invalid audience: expected %s, got %v", j.expectedAudience, claims.Audience)
	}

	// Validate organization name if configured
	if j.orgName != "" && claims.OrgName != j.orgName {
		return fmt.Errorf("invalid org_name: expected %s, got %s", j.orgName, claims.OrgName)
	}

	// Validate required fields
	if claims.Email == "" {
		return fmt.Errorf("email claim is missing")
	}

	if claims.IdpUserID == "" {
		return fmt.Errorf("subject claim is missing")
	}

	return nil
}

// containsAudience checks if the audience list contains the expected audience
func (j *JWTAuthMiddleware) containsAudience(audiences []string, expected string) bool {
	for _, aud := range audiences {
		if aud == expected {
			return true
		}
	}
	return false
}

// fetchJWKS fetches the JWKS from the configured endpoint
func (j *JWTAuthMiddleware) fetchJWKS() error {
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
func (j *JWTAuthMiddleware) buildRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
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
func (j *JWTAuthMiddleware) ensureKeysFresh() error {
	if len(j.keys) == 0 || time.Since(j.lastFetch) > time.Hour {
		return j.fetchJWKS()
	}
	return nil
}

// shouldSkipAuth determines if authentication should be skipped for this path
func (j *JWTAuthMiddleware) shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/debug",
		"/openapi.yaml",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// parseUnixTimestamp converts a numeric claim to time.Time
func parseUnixTimestamp(claim interface{}) (time.Time, error) {
	switch v := claim.(type) {
	case float64:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	case string:
		timestamp, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(timestamp, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}
}
