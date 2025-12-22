package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Helper function to create a token without signing (matches ParseUnverified usage)
func createUnsignedTestToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	// For unsigned tokens, use empty string as key
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	return tokenString
}

func TestGetConsumerJwtFromToken_LocalEnvironment(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	result, err := GetConsumerJwtFromToken("local", req)
	if err != nil {
		t.Errorf("Expected no error in local environment, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result in local environment")
	}

	expected := &ConsumerAssertion{
		ClientId:   "passport-app",
		Subscriber: "passport-app",
		Iss:        "https://idp.example.com",
	}

	if result.ClientId != expected.ClientId {
		t.Errorf("Expected ClientId %s, got %s", expected.ClientId, result.ClientId)
	}
	if result.Subscriber != expected.Subscriber {
		t.Errorf("Expected Subscriber %s, got %s", expected.Subscriber, result.Subscriber)
	}
}

func TestGetConsumerJwtFromToken_XJWTAssertionHeader(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimIss:      "https://idp.test.com",
		ClaimClientId: "test-client-id",
		ClaimSub:      "test-subscriber",
		ClaimAud:      []string{"https://api.test.com"},
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
		ClaimIat:      float64(time.Now().Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", req)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.ClientId != "test-client-id" {
		t.Errorf("Expected ClientId 'test-client-id', got '%s'", result.ClientId)
	}
	if result.Subscriber != "test-subscriber" {
		t.Errorf("Expected Subscriber 'test-subscriber', got '%s'", result.Subscriber)
	}
	if result.Iss != "https://idp.test.com" {
		t.Errorf("Expected Iss 'https://idp.test.com', got '%s'", result.Iss)
	}
}

func TestGetConsumerJwtFromToken_AuthorizationHeaderWithBearer(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "bearer-client-id",
		ClaimSub:      "bearer-subscriber",
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	result, err := GetConsumerJwtFromToken("production", req)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.ClientId != "bearer-client-id" {
		t.Errorf("Expected ClientId 'bearer-client-id', got '%s'", result.ClientId)
	}
}

func TestGetConsumerJwtFromToken_MissingClientId(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimSub: "some-subscriber",
		ClaimExp: float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", req)

	if err == nil {
		t.Error("Expected error when client_id is missing")
	}
	if result != nil {
		t.Error("Expected nil result when client_id is missing")
	}
}

func TestGetConsumerJwtFromToken_ExpiredToken(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimExp:      float64(time.Now().Add(-time.Hour).Unix()), // Expired 1 hour ago
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", req)

	if err == nil {
		t.Error("Expected error when token is expired")
	}
	if result != nil {
		t.Error("Expected nil result when token is expired")
	}
}

func TestGetConsumerJwtFromToken_NbfFuture(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimNbf:      float64(time.Now().Add(time.Hour).Unix()), // Valid in future
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", req)

	if err == nil {
		t.Error("Expected error when token is not yet valid (nbf)")
	}
	if result != nil {
		t.Error("Expected nil result when token is not yet valid")
	}
}

func TestGetConsumerJwtFromToken_AzpFallback(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimAzp:      "azp-subscriber",
		// missing sub
		ClaimExp: float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", req)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Subscriber != "azp-subscriber" {
		t.Errorf("Expected Subscriber to fall back to azp 'azp-subscriber', got '%s'", result.Subscriber)
	}
}
