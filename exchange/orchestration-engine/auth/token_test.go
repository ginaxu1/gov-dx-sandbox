package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/configs"
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

	result, err := GetConsumerJwtFromToken("local", nil, req)
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

	result, err := GetConsumerJwtFromToken("production", nil, req)
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

	result, err := GetConsumerJwtFromToken("production", nil, req)
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

	result, err := GetConsumerJwtFromToken("production", nil, req)

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
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
		ClaimNbf:      float64(time.Now().Add(time.Hour).Unix()), // Valid in future
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", nil, req)

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
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
		ClaimNbf:      float64(time.Now().Add(time.Hour).Unix()), // Valid in future
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", nil, req)

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

	result, err := GetConsumerJwtFromToken("production", nil, req)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Subscriber != "azp-subscriber" {
		t.Errorf("Expected Subscriber to fall back to azp 'azp-subscriber', got '%s'", result.Subscriber)
	}
}

func TestGetConsumerJwtFromToken_MissingExp(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimSub:      "some-subscriber",
		// missing exp
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	result, err := GetConsumerJwtFromToken("production", nil, req)

	if err == nil {
		t.Error("Expected error when exp claim is missing")
	}
	if result != nil {
		t.Error("Expected nil result when exp claim is missing")
	}
	if err != nil && err.Error() != "missing or invalid exp claim" {
		t.Errorf("Expected error message 'missing or invalid exp claim', got: %v", err)
	}
}

func TestGetConsumerJwtFromToken_InvalidIssuer(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimSub:      "subscriber",
		ClaimIss:      "https://wrong-issuer.com",
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	jwtConfig := &configs.JWTConfig{
		ExpectedIssuer: "https://expected-issuer.com",
	}

	result, err := GetConsumerJwtFromToken("production", jwtConfig, req)

	if err == nil {
		t.Error("Expected error when issuer doesn't match")
	}
	if result != nil {
		t.Error("Expected nil result when issuer doesn't match")
	}
}

func TestGetConsumerJwtFromToken_ValidIssuer(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimSub:      "subscriber",
		ClaimIss:      "https://expected-issuer.com",
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	jwtConfig := &configs.JWTConfig{
		ExpectedIssuer: "https://expected-issuer.com",
	}

	result, err := GetConsumerJwtFromToken("production", jwtConfig, req)
	if err != nil {
		t.Errorf("Expected no error when issuer matches, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result when issuer matches")
	}
	if result.Iss != "https://expected-issuer.com" {
		t.Errorf("Expected issuer 'https://expected-issuer.com', got '%s'", result.Iss)
	}
}

func TestGetConsumerJwtFromToken_InvalidAudience(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimSub:      "subscriber",
		ClaimAud:      []string{"https://wrong-api.com"},
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	jwtConfig := &configs.JWTConfig{
		ValidAudiences: []string{"https://api1.com", "https://api2.com"},
	}

	result, err := GetConsumerJwtFromToken("production", jwtConfig, req)

	if err == nil {
		t.Error("Expected error when audience doesn't match any valid audience")
	}
	if result != nil {
		t.Error("Expected nil result when audience doesn't match")
	}
}

func TestGetConsumerJwtFromToken_ValidAudience(t *testing.T) {
	claims := jwt.MapClaims{
		ClaimClientId: "client-id",
		ClaimSub:      "subscriber",
		ClaimAud:      []string{"https://api2.com"},
		ClaimExp:      float64(time.Now().Add(time.Hour).Unix()),
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", tokenString)

	jwtConfig := &configs.JWTConfig{
		ValidAudiences: []string{"https://api1.com", "https://api2.com"},
	}

	result, err := GetConsumerJwtFromToken("production", jwtConfig, req)
	if err != nil {
		t.Errorf("Expected no error when audience matches, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result when audience matches")
	}
	if len(result.Aud) == 0 || result.Aud[0] != "https://api2.com" {
		t.Errorf("Expected audience 'https://api2.com', got '%v'", result.Aud)
	}
}
