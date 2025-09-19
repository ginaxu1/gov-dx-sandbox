package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJWTVerifierInitialization(t *testing.T) {
	jwksURL := "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks"
	jwtVerifier := NewJWTVerifier(jwksURL)

	if jwtVerifier == nil {
		t.Fatal("JWT verifier should not be nil")
	}

	if jwtVerifier.jwksURL != jwksURL {
		t.Errorf("Expected JWKS URL %s, got %s", jwksURL, jwtVerifier.jwksURL)
	}

	if jwtVerifier.httpClient == nil {
		t.Fatal("HTTP client should not be nil")
	}

	if jwtVerifier.keys == nil {
		t.Fatal("Keys map should not be nil")
	}
}

func TestFetchJWKS(t *testing.T) {
	jwksURL := "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks"
	jwtVerifier := NewJWTVerifier(jwksURL)

	err := jwtVerifier.fetchJWKS()
	if err != nil {
		t.Fatalf("Failed to fetch JWKS: %v", err)
	}

	if len(jwtVerifier.keys) == 0 {
		t.Fatal("Should have fetched at least one key")
	}

	t.Logf("Successfully fetched %d keys from JWKS", len(jwtVerifier.keys))
}

func TestVerifyInvalidToken(t *testing.T) {
	jwksURL := "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks"
	jwtVerifier := NewJWTVerifier(jwksURL)

	// First fetch the JWKS
	err := jwtVerifier.fetchJWKS()
	if err != nil {
		t.Fatalf("Failed to fetch JWKS: %v", err)
	}

	// Test with invalid token
	_, err = jwtVerifier.VerifyAndExtractEmail("invalid.token.here")
	if err == nil {
		t.Fatal("Should have rejected invalid token")
	}

	t.Logf("Correctly rejected invalid token: %v", err)
}

func TestDataInfoEndpoint(t *testing.T) {
	// Create a test consent engine
	engine := NewConsentEngine("http://localhost:5173")
	server := &apiServer{engine: engine}

	// Create a test consent record
	consentReq := ConsentRequest{
		AppID:     "test-app",
		Purpose:   "test-purpose",
		SessionID: "test-session",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "test-owner-123",
				OwnerEmail: "test@example.com",
				Fields:     []string{"person.name", "person.email"},
			},
		},
	}

	// Process the consent request
	consentRecord, err := engine.ProcessConsentRequest(consentReq)
	if err != nil {
		t.Fatalf("Failed to create consent record: %v", err)
	}

	// Test the data-info endpoint
	req := httptest.NewRequest("GET", "/data-info/"+consentRecord.ConsentID, nil)
	w := httptest.NewRecorder()

	server.dataInfoHandler(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check the response fields
	expectedOwnerID := "test-owner-123"
	expectedOwnerEmail := "test@example.com"

	if response["owner_id"] != expectedOwnerID {
		t.Errorf("Expected owner_id %s, got %s", expectedOwnerID, response["owner_id"])
	}

	if response["owner_email"] != expectedOwnerEmail {
		t.Errorf("Expected owner_email %s, got %s", expectedOwnerEmail, response["owner_email"])
	}

	t.Logf("✅ Data info endpoint test passed")
	t.Logf("   Owner ID: %s", response["owner_id"])
	t.Logf("   Owner Email: %s", response["owner_email"])
}

func TestDataInfoEndpointNotFound(t *testing.T) {
	engine := NewConsentEngine("http://localhost:5173")
	server := &apiServer{engine: engine}

	// Test with non-existent consent ID
	req := httptest.NewRequest("GET", "/data-info/non-existent-id", nil)
	w := httptest.NewRecorder()

	server.dataInfoHandler(w, req)

	// Check the response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	t.Logf("✅ Data info endpoint 404 test passed")
}

func TestJWTMiddlewareEmailMatching(t *testing.T) {
	// Create a test consent engine
	engine := NewConsentEngine("http://localhost:5173")

	// Create a test consent record
	consentReq := ConsentRequest{
		AppID:     "test-app",
		Purpose:   "test-purpose",
		SessionID: "test-session",
		DataFields: []DataField{
			{
				OwnerType:  "citizen",
				OwnerID:    "test-owner-123",
				OwnerEmail: "test@example.com",
				Fields:     []string{"person.name", "person.email"},
			},
		},
	}

	// Process the consent request
	consentRecord, err := engine.ProcessConsentRequest(consentReq)
	if err != nil {
		t.Fatalf("Failed to create consent record: %v", err)
	}

	// Test JWT middleware with matching email
	jwtVerifier := NewJWTVerifier("https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks")

	// Create a mock JWT token (this will fail verification, but we can test the flow)
	req := httptest.NewRequest("GET", "/consents/"+consentRecord.ConsentID, nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	// Create middleware
	middleware := jwtAuthMiddleware(jwtVerifier, engine)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Test the middleware
	handler.ServeHTTP(w, req)

	// Should return 403 due to invalid token
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for invalid token, got %d", w.Code)
	}

	t.Logf("✅ JWT middleware email matching test passed")
}
