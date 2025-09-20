package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gov-dx-sandbox/exchange/consent-engine"
)

// TestData contains common test data structures
type TestData struct {
	ConsentPortalURL string
	JWKSURL          string
}

// DefaultTestData returns default test configuration
func DefaultTestData() TestData {
	return TestData{
		ConsentPortalURL: getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173"),
		JWKSURL:          getEnvOrDefault("TEST_JWKS_URL", "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/jwks"),
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConsentRequestBuilder helps build test consent requests
type ConsentRequestBuilder struct {
	request *consent.ConsentRequest
}

// NewConsentRequestBuilder creates a new builder with default values
func NewConsentRequestBuilder() *ConsentRequestBuilder {
	return &ConsentRequestBuilder{
		request: &consent.ConsentRequest{
			AppID:     "test-app",
			Purpose:   "test_purpose",
			SessionID: "test-session-123",
			DataFields: []consent.DataField{
				{
					OwnerType:  "citizen",
					OwnerID:    "test-owner-123",
					OwnerEmail: "test@example.com",
					Fields:     []string{"person.name", "person.email"},
				},
			},
		},
	}
}

// WithAppID sets the app ID
func (b *ConsentRequestBuilder) WithAppID(appID string) *ConsentRequestBuilder {
	b.request.AppID = appID
	return b
}

// WithPurpose sets the purpose
func (b *ConsentRequestBuilder) WithPurpose(purpose string) *ConsentRequestBuilder {
	b.request.Purpose = purpose
	return b
}

// WithSessionID sets the session ID
func (b *ConsentRequestBuilder) WithSessionID(sessionID string) *ConsentRequestBuilder {
	b.request.SessionID = sessionID
	return b
}

// WithDataField adds a data field
func (b *ConsentRequestBuilder) WithDataField(dataField consent.DataField) *ConsentRequestBuilder {
	b.request.DataFields = append(b.request.DataFields, dataField)
	return b
}

// WithOwner sets the owner information
func (b *ConsentRequestBuilder) WithOwner(ownerType, ownerID, ownerEmail string, fields []string) *ConsentRequestBuilder {
	b.request.DataFields = []consent.DataField{
		{
			OwnerType:  ownerType,
			OwnerID:    ownerID,
			OwnerEmail: ownerEmail,
			Fields:     fields,
		},
	}
	return b
}

// Build returns the built consent request
func (b *ConsentRequestBuilder) Build() *consent.ConsentRequest {
	return b.request
}

// DataFieldBuilder helps build test data fields
type DataFieldBuilder struct {
	field *consent.DataField
}

// NewDataFieldBuilder creates a new data field builder
func NewDataFieldBuilder() *DataFieldBuilder {
	return &DataFieldBuilder{
		field: &consent.DataField{
			OwnerType:  "citizen",
			OwnerID:    "test-owner-123",
			OwnerEmail: "test@example.com",
			Fields:     []string{"person.name"},
		},
	}
}

// WithOwnerType sets the owner type
func (b *DataFieldBuilder) WithOwnerType(ownerType string) *DataFieldBuilder {
	b.field.OwnerType = ownerType
	return b
}

// WithOwnerID sets the owner ID
func (b *DataFieldBuilder) WithOwnerID(ownerID string) *DataFieldBuilder {
	b.field.OwnerID = ownerID
	return b
}

// WithOwnerEmail sets the owner email
func (b *DataFieldBuilder) WithOwnerEmail(ownerEmail string) *DataFieldBuilder {
	b.field.OwnerEmail = ownerEmail
	return b
}

// WithFields sets the fields
func (b *DataFieldBuilder) WithFields(fields []string) *DataFieldBuilder {
	b.field.Fields = fields
	return b
}

// Build returns the built data field
func (b *DataFieldBuilder) Build() consent.DataField {
	return *b.field
}

// HTTPTestHelper provides common HTTP testing utilities
type HTTPTestHelper struct {
	t *testing.T
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(t *testing.T) *HTTPTestHelper {
	return &HTTPTestHelper{t: t}
}

// CreateRequest creates an HTTP request with JSON body
func (h *HTTPTestHelper) CreateRequest(method, url string, body interface{}) *http.Request {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			h.t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateRequestWithHeaders creates an HTTP request with custom headers
func (h *HTTPTestHelper) CreateRequestWithHeaders(method, url string, body interface{}, headers map[string]string) *http.Request {
	req := h.CreateRequest(method, url, body)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

// CreateRecorder creates a new response recorder
func (h *HTTPTestHelper) CreateRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// AssertStatusCode asserts the response status code
func (h *HTTPTestHelper) AssertStatusCode(recorder *httptest.ResponseRecorder, expectedCode int) {
	if recorder.Code != expectedCode {
		h.t.Errorf("Expected status %d, got %d. Body: %s", expectedCode, recorder.Code, recorder.Body.String())
	}
}

// AssertJSONResponse asserts and unmarshals a JSON response
func (h *HTTPTestHelper) AssertJSONResponse(recorder *httptest.ResponseRecorder, expectedCode int, target interface{}) {
	h.AssertStatusCode(recorder, expectedCode)

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		h.t.Fatalf("Failed to unmarshal response: %v. Body: %s", err, recorder.Body.String())
	}
}

// AssertErrorResponse asserts an error response
func (h *HTTPTestHelper) AssertErrorResponse(recorder *httptest.ResponseRecorder, expectedCode int) {
	h.AssertStatusCode(recorder, expectedCode)

	var errorResp struct {
		Error string `json:"error"`
	}
	h.AssertJSONResponse(recorder, expectedCode, &errorResp)

	if errorResp.Error == "" {
		h.t.Error("Expected error message in response")
	}
}

// JWTTestHelper provides JWT testing utilities
type JWTTestHelper struct {
	t *testing.T
}

// NewJWTTestHelper creates a new JWT test helper
func NewJWTTestHelper(t *testing.T) *JWTTestHelper {
	return &JWTTestHelper{t: t}
}

// CreateJWTVerifier creates a JWT verifier for testing
func (h *JWTTestHelper) CreateJWTVerifier(jwksURL string) *consent.JWTVerifier {
	verifier := consent.NewJWTVerifier(jwksURL)
	if verifier == nil {
		h.t.Fatal("JWT verifier should not be nil")
	}
	return verifier
}

// CreateMockUserToken creates a mock user JWT token for testing
func (h *JWTTestHelper) CreateMockUserToken(email string) string {
	claims := jwt.MapClaims{
		"sub":   "test-user-id",
		"email": email,
		"iss":   getEnvOrDefault("TEST_ASGARDEO_ISSUER", "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/token"),
		"aud":   getEnvOrDefault("TEST_ASGARDEO_AUDIENCE", "test-audience"),
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
		"aut":   "user",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		h.t.Fatalf("Failed to create mock token: %v", err)
	}

	return tokenString
}

// CreateMockM2MToken creates a mock M2M JWT token for testing
func (h *JWTTestHelper) CreateMockM2MToken() string {
	claims := jwt.MapClaims{
		"sub":       "test-client-id",
		"client_id": "test-client-id",
		"iss":       getEnvOrDefault("TEST_ASGARDEO_ISSUER", "https://api.asgardeo.io/t/lankasoftwarefoundation/oauth2/token"),
		"aud":       getEnvOrDefault("TEST_ASGARDEO_AUDIENCE", "test-audience"),
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(time.Hour).Unix(),
		"aut":       "m2m",
		"scope":     "consent:read consent:write",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		h.t.Fatalf("Failed to create mock M2M token: %v", err)
	}

	return tokenString
}

// EngineTestHelper provides engine testing utilities
type EngineTestHelper struct {
	t *testing.T
}

// NewEngineTestHelper creates a new engine test helper
func NewEngineTestHelper(t *testing.T) *EngineTestHelper {
	return &EngineTestHelper{t: t}
}

// CreateTestEngine creates a test consent engine
func (h *EngineTestHelper) CreateTestEngine() *consent.ConsentEngine {
	consentPortalURL := getEnvOrDefault("TEST_CONSENT_PORTAL_URL", "http://localhost:5173")
	engine := consent.NewConsentEngine(consentPortalURL)
	if engine == nil {
		h.t.Fatal("Consent engine should not be nil")
	}
	return engine
}

// CreateTestConsent creates a test consent and returns the consent ID
func (h *EngineTestHelper) CreateTestConsent() string {
	engine := h.CreateTestEngine()

	req := NewConsentRequestBuilder().
		WithAppID("test-app").
		WithPurpose("test_purpose").
		WithSessionID("test-session-123").
		WithOwner("citizen", "test-owner-123", "test@example.com", []string{"person.name", "person.email"}).
		Build()

	record, err := engine.CreateConsent(*req)
	if err != nil {
		h.t.Fatalf("Failed to create test consent: %v", err)
	}

	return record.ConsentID
}

// AssertConsentRecord asserts common consent record fields
func (h *EngineTestHelper) AssertConsentRecord(record *consent.ConsentRecord, expectedAppID, expectedOwnerID, expectedOwnerEmail string) {
	if record.ConsentID == "" {
		h.t.Error("Expected non-empty consent ID")
	}

	if record.AppID != expectedAppID {
		h.t.Errorf("Expected AppID=%s, got %s", expectedAppID, record.AppID)
	}

	if record.OwnerID != expectedOwnerID {
		h.t.Errorf("Expected OwnerID=%s, got %s", expectedOwnerID, record.OwnerID)
	}

	if record.OwnerEmail != expectedOwnerEmail {
		h.t.Errorf("Expected OwnerEmail=%s, got %s", expectedOwnerEmail, record.OwnerEmail)
	}

	if record.Status != string(consent.StatusPending) {
		h.t.Errorf("Expected status=%s, got %s", string(consent.StatusPending), record.Status)
	}
}

// Common test data constants
const (
	TestAppID      = "test-app"
	TestPurpose    = "test_purpose"
	TestSessionID  = "test-session-123"
	TestOwnerType  = "citizen"
	TestOwnerID    = "test-owner-123"
	TestOwnerEmail = "test@example.com"
	TestFieldName  = "person.name"
	TestFieldEmail = "person.email"
)

// Common test field arrays
var (
	TestFieldsBasic    = []string{TestFieldName}
	TestFieldsExtended = []string{TestFieldName, TestFieldEmail}
)
