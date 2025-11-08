package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		ApplicationUuid: "passport-app",
		Subscriber:      "passport-app",
		ApplicationId:   "passport-app",
	}

	if result.ApplicationUuid != expected.ApplicationUuid {
		t.Errorf("Expected ApplicationUuid %s, got %s", expected.ApplicationUuid, result.ApplicationUuid)
	}
	if result.Subscriber != expected.Subscriber {
		t.Errorf("Expected Subscriber %s, got %s", expected.Subscriber, result.Subscriber)
	}
	if result.ApplicationId != expected.ApplicationId {
		t.Errorf("Expected ApplicationId %s, got %s", expected.ApplicationId, result.ApplicationId)
	}
}

func TestGetConsumerJwtFromToken_XJWTAssertionHeader(t *testing.T) {

	claims := jwt.MapClaims{
		ClaimAppUUID:    "test-app-uuid",
		ClaimSubscriber: "test-subscriber",
		ClaimAppID:      "test-app-id",
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

	if result.ApplicationUuid != "test-app-uuid" {
		t.Errorf("Expected ApplicationUuid 'test-app-uuid', got '%s'", result.ApplicationUuid)
	}
	if result.Subscriber != "test-subscriber" {
		t.Errorf("Expected Subscriber 'test-subscriber', got '%s'", result.Subscriber)
	}
	if result.ApplicationId != "test-app-id" {
		t.Errorf("Expected ApplicationId 'test-app-id', got '%s'", result.ApplicationId)
	}
}

func TestGetConsumerJwtFromToken_AuthorizationHeaderWithBearer(t *testing.T) {

	claims := jwt.MapClaims{
		ClaimAppUUID:    "bearer-app-uuid",
		ClaimSubscriber: "bearer-subscriber",
		ClaimAppID:      "bearer-app-id",
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

	if result.ApplicationUuid != "bearer-app-uuid" {
		t.Errorf("Expected ApplicationUuid 'bearer-app-uuid', got '%s'", result.ApplicationUuid)
	}
	if result.Subscriber != "bearer-subscriber" {
		t.Errorf("Expected Subscriber 'bearer-subscriber', got '%s'", result.Subscriber)
	}
	if result.ApplicationId != "bearer-app-id" {
		t.Errorf("Expected ApplicationId 'bearer-app-id', got '%s'", result.ApplicationId)
	}
}

func TestGetConsumerJwtFromToken_AuthorizationHeaderWithoutBearer(t *testing.T) {

	claims := jwt.MapClaims{
		ClaimAppUUID:    "no-bearer-uuid",
		ClaimSubscriber: "no-bearer-subscriber",
		ClaimAppID:      "no-bearer-id",
	}

	tokenString := createUnsignedTestToken(claims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", tokenString)

	result, err := GetConsumerJwtFromToken("production", req)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.ApplicationUuid != "no-bearer-uuid" {
		t.Errorf("Expected ApplicationUuid 'no-bearer-uuid', got '%s'", result.ApplicationUuid)
	}
}

func TestGetConsumerJwtFromToken_MissingToken(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	result, err := GetConsumerJwtFromToken("production", req)

	if err == nil {
		t.Error("Expected error when token is missing")
	}

	if result != nil {
		t.Error("Expected nil result when token is missing")
	}

	expectedError := "missing token"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetConsumerJwtFromToken_InvalidToken(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", "invalid.token.string")

	result, err := GetConsumerJwtFromToken("production", req)

	if err == nil {
		t.Error("Expected error when token is invalid")
	}

	if result != nil {
		t.Error("Expected nil result when token is invalid")
	}
}

func TestGetConsumerJwtFromToken_XJWTAssertionTakesPrecedence(t *testing.T) {

	xJWTClaims := jwt.MapClaims{
		ClaimAppUUID:    "x-jwt-uuid",
		ClaimSubscriber: "x-jwt-subscriber",
		ClaimAppID:      "x-jwt-id",
	}

	authClaims := jwt.MapClaims{
		ClaimAppUUID:    "auth-uuid",
		ClaimSubscriber: "auth-subscriber",
		ClaimAppID:      "auth-id",
	}

	xJWTToken := createUnsignedTestToken(xJWTClaims)
	authToken := createUnsignedTestToken(authClaims)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-JWT-Assertion", xJWTToken)
	req.Header.Set("Authorization", "Bearer "+authToken)

	result, err := GetConsumerJwtFromToken("production", req)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Should use X-JWT-Assertion values, not Authorization
	if result.ApplicationUuid != "x-jwt-uuid" {
		t.Errorf("Expected X-JWT-Assertion to take precedence, got ApplicationUuid '%s'", result.ApplicationUuid)
	}
	if result.Subscriber != "x-jwt-subscriber" {
		t.Errorf("Expected X-JWT-Assertion to take precedence, got Subscriber '%s'", result.Subscriber)
	}
}

func TestGetConsumerJwtFromToken_MissingClaims(t *testing.T) {

	// Token with missing claims - they should be formatted as "<nil>"
	claims := jwt.MapClaims{
		"some-other-claim": "value",
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

	// Missing claims should be formatted as "<nil>"
	if result.ApplicationUuid != "<nil>" {
		t.Errorf("Expected ApplicationUuid '<nil>' for missing claim, got '%s'", result.ApplicationUuid)
	}
	if result.Subscriber != "<nil>" {
		t.Errorf("Expected Subscriber '<nil>' for missing claim, got '%s'", result.Subscriber)
	}
	if result.ApplicationId != "<nil>" {
		t.Errorf("Expected ApplicationId '<nil>' for missing claim, got '%s'", result.ApplicationId)
	}
}

func TestGetConsumerJwtFromToken_NumericClaims(t *testing.T) {

	// Test with numeric values in claims
	claims := jwt.MapClaims{
		ClaimAppUUID:    12345,
		ClaimSubscriber: 67890,
		ClaimAppID:      99999,
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

	// Numeric values should be converted to strings
	if result.ApplicationUuid != "12345" {
		t.Errorf("Expected ApplicationUuid '12345', got '%s'", result.ApplicationUuid)
	}
	if result.Subscriber != "67890" {
		t.Errorf("Expected Subscriber '67890', got '%s'", result.Subscriber)
	}
	if result.ApplicationId != "99999" {
		t.Errorf("Expected ApplicationId '99999', got '%s'", result.ApplicationId)
	}
}

func TestGetConsumerJwtFromToken_ShortBearerPrefix(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Authorization header too short to contain "Bearer "
	req.Header.Set("Authorization", "Bear")

	result, err := GetConsumerJwtFromToken("production", req)

	if err == nil {
		t.Error("Expected error with invalid short token")
	}

	if result != nil {
		t.Error("Expected nil result with invalid token")
	}
}

func TestConsumerAssertion_StructFields(t *testing.T) {
	// Test that ConsumerAssertion struct can be created and fields are accessible
	ca := &ConsumerAssertion{
		ApplicationUuid: "test-uuid",
		Subscriber:      "test-subscriber",
		ApplicationId:   "test-id",
	}

	if ca.ApplicationUuid != "test-uuid" {
		t.Errorf("Expected ApplicationUuid 'test-uuid', got '%s'", ca.ApplicationUuid)
	}
	if ca.Subscriber != "test-subscriber" {
		t.Errorf("Expected Subscriber 'test-subscriber', got '%s'", ca.Subscriber)
	}
	if ca.ApplicationId != "test-id" {
		t.Errorf("Expected ApplicationId 'test-id', got '%s'", ca.ApplicationId)
	}
}

func TestConstants(t *testing.T) {
	// Verify that constants are correctly defined
	expectedPrefix := "http://wso2.org/claims/"

	if WSO2ClaimPrefix != expectedPrefix {
		t.Errorf("Expected WSO2ClaimPrefix '%s', got '%s'", expectedPrefix, WSO2ClaimPrefix)
	}

	if ClaimSubscriber != expectedPrefix+"subscriber" {
		t.Errorf("Expected ClaimSubscriber '%s', got '%s'", expectedPrefix+"subscriber", ClaimSubscriber)
	}

	if ClaimAppUUID != expectedPrefix+"applicationUUId" {
		t.Errorf("Expected ClaimAppUUID '%s', got '%s'", expectedPrefix+"applicationUUId", ClaimAppUUID)
	}

	if ClaimAppID != expectedPrefix+"applicationid" {
		t.Errorf("Expected ClaimAppID '%s', got '%s'", expectedPrefix+"applicationid", ClaimAppID)
	}
}
