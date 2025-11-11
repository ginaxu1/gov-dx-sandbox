package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestVerifyAudience(t *testing.T) {
	verifier := NewJWTVerifier("http://example.com/jwks", "test-audience", "test-org")

	tests := []struct {
		name             string
		tokenAudience    []string
		expectedAudience string
		want             bool
	}{
		{
			name:             "Match - single audience",
			tokenAudience:    []string{"test-audience"},
			expectedAudience: "test-audience",
			want:             true,
		},
		{
			name:             "Match - multiple audiences",
			tokenAudience:    []string{"other-audience", "test-audience"},
			expectedAudience: "test-audience",
			want:             true,
		},
		{
			name:             "No match",
			tokenAudience:    []string{"other-audience"},
			expectedAudience: "test-audience",
			want:             false,
		},
		{
			name:             "Empty audience",
			tokenAudience:    []string{},
			expectedAudience: "test-audience",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.verifyAudience(tt.tokenAudience, tt.expectedAudience)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractEmailFromToken(t *testing.T) {
	verifier := NewJWTVerifier("http://example.com/jwks", "test-audience", "test-org")

	// Create a token with email claim
	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Email:   "test@example.com",
		OrgName: "test-org",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Valid = true

	email, err := verifier.ExtractEmailFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)
}

func TestExtractEmailFromToken_InvalidClaims(t *testing.T) {
	verifier := NewJWTVerifier("http://example.com/jwks", "test-audience", "test-org")

	// Create a token with standard claims (not CustomClaims)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Valid = true

	_, err := verifier.ExtractEmailFromToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token claims")
}

func TestExtractEmailFromToken_NoEmail(t *testing.T) {
	verifier := NewJWTVerifier("http://example.com/jwks", "test-audience", "test-org")

	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Email:   "", // Empty email
		OrgName: "test-org",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Valid = true

	_, err := verifier.ExtractEmailFromToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email claim not found")
}

func TestVerifyAndExtractEmail(t *testing.T) {
	// Create a mock JWKS server
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	// Encode public key
	nBytes := publicKey.N.Bytes()
	eBytes := []byte{byte(publicKey.E >> 24), byte(publicKey.E >> 16), byte(publicKey.E >> 8), byte(publicKey.E)}
	if publicKey.E < 256 {
		eBytes = []byte{byte(publicKey.E)}
	}

	jwk := JWK{
		Kty: "RSA",
		Kid: "test-kid",
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}

	jwks := JWKS{Keys: []JWK{jwk}}
	jwksJSON, _ := json.Marshal(jwks)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksJSON)
	}))
	defer server.Close()

	verifier := NewJWTVerifier(server.URL, "test-audience", "test-org")

	// Create a valid token
	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  []string{"test-audience"},
		},
		Email:   "test@example.com",
		OrgName: "test-org",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"
	tokenString, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	// Fetch JWKS first
	err = verifier.fetchJWKS()
	assert.NoError(t, err)

	email, err := verifier.VerifyAndExtractEmail(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)
}

func TestVerifyTokenAndMatchEmail(t *testing.T) {
	// Create a mock JWKS server
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey := &privateKey.PublicKey

	nBytes := publicKey.N.Bytes()
	eBytes := []byte{byte(publicKey.E >> 24), byte(publicKey.E >> 16), byte(publicKey.E >> 8), byte(publicKey.E)}
	if publicKey.E < 256 {
		eBytes = []byte{byte(publicKey.E)}
	}

	jwk := JWK{
		Kty: "RSA",
		Kid: "test-kid",
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}

	jwks := JWKS{Keys: []JWK{jwk}}
	jwksJSON, _ := json.Marshal(jwks)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksJSON)
	}))
	defer server.Close()

	verifier := NewJWTVerifier(server.URL, "test-audience", "test-org")

	// Create a valid token
	claims := &CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  []string{"test-audience"},
		},
		Email:   "test@example.com",
		OrgName: "test-org",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid"
	tokenString, err := token.SignedString(privateKey)
	assert.NoError(t, err)

	// Fetch JWKS first
	err = verifier.fetchJWKS()
	assert.NoError(t, err)

	// Test matching email
	match, err := verifier.VerifyTokenAndMatchEmail(tokenString, "test@example.com")
	assert.NoError(t, err)
	assert.True(t, match)

	// Test non-matching email
	match, err = verifier.VerifyTokenAndMatchEmail(tokenString, "other@example.com")
	assert.Error(t, err)
	assert.False(t, match)
	assert.Contains(t, err.Error(), "does not match")
}
