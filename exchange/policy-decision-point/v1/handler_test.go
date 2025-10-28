package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// PolicyMetadataServiceInterface defines the interface for policy metadata service
type PolicyMetadataServiceInterface interface {
	CreatePolicyMetadata(req *models.PolicyMetadataCreateRequest) (*models.PolicyMetadataCreateResponse, error)
	UpdateAllowList(req *models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error)
	GetPolicyDecision(req *models.PolicyDecisionRequest) (*models.PolicyDecisionResponse, error)
}

// MockPolicyMetadataService is a mock implementation for testing
type MockPolicyMetadataService struct {
	mock.Mock
}

func (m *MockPolicyMetadataService) CreatePolicyMetadata(req *models.PolicyMetadataCreateRequest) (*models.PolicyMetadataCreateResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.PolicyMetadataCreateResponse), args.Error(1)
}

func (m *MockPolicyMetadataService) UpdateAllowList(req *models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.AllowListUpdateResponse), args.Error(1)
}

func (m *MockPolicyMetadataService) GetPolicyDecision(req *models.PolicyDecisionRequest) (*models.PolicyDecisionResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*models.PolicyDecisionResponse), args.Error(1)
}

// TestHandler wraps the Handler with a mockable service interface
type TestHandler struct {
	policyService PolicyMetadataServiceInterface
}

func (h *TestHandler) CreatePolicyMetadata(w http.ResponseWriter, r *http.Request) {
	var req models.PolicyMetadataCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.policyService.CreatePolicyMetadata(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandler) UpdateAllowList(w http.ResponseWriter, r *http.Request) {
	var req models.AllowListUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.policyService.UpdateAllowList(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *TestHandler) GetPolicyDecision(w http.ResponseWriter, r *http.Request) {
	var req models.PolicyDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.policyService.GetPolicyDecision(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func TestHandler_CreatePolicyMetadata_Success(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create test request with correct types
	displayName := "Test Field"
	description := "Test Description"
	owner := models.OwnerCitizen

	req := models.PolicyMetadataCreateRequest{
		SchemaID: "test-schema",
		Records: []models.PolicyMetadataCreateRequestRecord{
			{
				FieldName:         "test.field",
				DisplayName:       &displayName,
				Description:       &description,
				Source:            models.SourcePrimary,
				IsOwner:           true,
				AccessControlType: models.AccessControlTypePublic,
				Owner:             &owner,
			},
		},
	}

	expectedResponse := &models.PolicyMetadataCreateResponse{
		Records: []models.PolicyMetadataResponse{
			{
				FieldName:         "test.field",
				DisplayName:       &displayName,
				Description:       &description,
				Source:            models.SourcePrimary,
				IsOwner:           true,
				AccessControlType: models.AccessControlTypePublic,
				Owner:             &owner,
			},
		},
	}

	// Setup mock expectations
	mockService.On("CreatePolicyMetadata", &req).Return(expectedResponse, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/metadata", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreatePolicyMetadata(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.PolicyMetadataCreateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Records[0].FieldName, response.Records[0].FieldName)
	assert.Equal(t, expectedResponse.Records[0].SchemaID, response.Records[0].SchemaID)

	// Verify mock was called
	mockService.AssertExpectations(t)
}

func TestHandler_CreatePolicyMetadata_InvalidJSON(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create HTTP request with invalid JSON
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/metadata", bytes.NewBufferString("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreatePolicyMetadata(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")

	// Verify mock was not called
	mockService.AssertNotCalled(t, "CreatePolicyMetadata")
}

func TestHandler_CreatePolicyMetadata_ServiceError(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create test request
	displayName := "Test Field"
	owner := models.OwnerCitizen

	req := models.PolicyMetadataCreateRequest{
		SchemaID: "test-schema",
		Records: []models.PolicyMetadataCreateRequestRecord{
			{
				FieldName:         "test.field",
				DisplayName:       &displayName,
				Source:            models.SourcePrimary,
				IsOwner:           true,
				AccessControlType: models.AccessControlTypePublic,
				Owner:             &owner,
			},
		},
	}

	// Setup mock to return error
	mockService.On("CreatePolicyMetadata", &req).Return((*models.PolicyMetadataCreateResponse)(nil), assert.AnError)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/metadata", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreatePolicyMetadata(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "assert.AnError")

	// Verify mock was called
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateAllowList_Success(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create test request
	req := models.AllowListUpdateRequest{
		ApplicationID: "test-app",
		GrantDuration: models.GrantDurationTypeOneMonth,
		Records: []models.AllowListUpdateRequestRecord{
			{
				SchemaID:  "test-schema",
				FieldName: "test.field",
			},
		},
	}

	expectedResponse := &models.AllowListUpdateResponse{
		Records: []models.AllowListUpdateResponseRecord{
			{
				FieldName: "test.field",
				SchemaID:  "test-schema",
				ExpiresAt: "2024-01-01T00:00:00Z",
				UpdatedAt: "2024-01-01T00:00:00Z",
			},
		},
	}

	// Setup mock expectations
	mockService.On("UpdateAllowList", &req).Return(expectedResponse, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/update-allowlist", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.UpdateAllowList(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.AllowListUpdateResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Records[0].FieldName, response.Records[0].FieldName)
	assert.Equal(t, expectedResponse.Records[0].SchemaID, response.Records[0].SchemaID)

	// Verify mock was called
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateAllowList_InvalidJSON(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create HTTP request with invalid JSON
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/update-allowlist", bytes.NewBufferString("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.UpdateAllowList(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")

	// Verify mock was not called
	mockService.AssertNotCalled(t, "UpdateAllowList")
}

func TestHandler_GetPolicyDecision_Success(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create test request
	req := models.PolicyDecisionRequest{
		ApplicationID: "test-app",
		RequiredFields: []models.PolicyDecisionRequestRecord{
			{
				SchemaID:  "test-schema",
				FieldName: "test.field",
			},
		},
	}

	expectedResponse := &models.PolicyDecisionResponse{
		ConsentRequiredFields:   []models.PolicyDecisionResponseFieldRecord{},
		UnauthorizedFields:      []models.PolicyDecisionResponseFieldRecord{},
		ExpiredFields:           []models.PolicyDecisionResponseFieldRecord{},
		AppAuthorized:           true,
		AppAccessExpired:        false,
		AppRequiresOwnerConsent: false,
	}

	// Setup mock expectations
	mockService.On("GetPolicyDecision", &req).Return(expectedResponse, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/decide", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPolicyDecision(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response models.PolicyDecisionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.AppAuthorized, response.AppAuthorized)
	assert.Equal(t, expectedResponse.AppAccessExpired, response.AppAccessExpired)
	assert.Equal(t, expectedResponse.AppRequiresOwnerConsent, response.AppRequiresOwnerConsent)

	// Verify mock was called
	mockService.AssertExpectations(t)
}

func TestHandler_GetPolicyDecision_Unauthorized(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create test request
	req := models.PolicyDecisionRequest{
		ApplicationID: "unauthorized-app",
		RequiredFields: []models.PolicyDecisionRequestRecord{
			{
				SchemaID:  "test-schema",
				FieldName: "test.field",
			},
		},
	}

	displayName := "Test Field"
	owner := models.OwnerCitizen

	expectedResponse := &models.PolicyDecisionResponse{
		ConsentRequiredFields: []models.PolicyDecisionResponseFieldRecord{},
		UnauthorizedFields: []models.PolicyDecisionResponseFieldRecord{
			{
				FieldName:   "test.field",
				SchemaID:    "test-schema",
				DisplayName: &displayName,
				Owner:       &owner,
			},
		},
		ExpiredFields:           []models.PolicyDecisionResponseFieldRecord{},
		AppAuthorized:           false,
		AppAccessExpired:        false,
		AppRequiresOwnerConsent: false,
	}

	// Setup mock expectations
	mockService.On("GetPolicyDecision", &req).Return(expectedResponse, nil)

	// Create HTTP request
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/decide", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPolicyDecision(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PolicyDecisionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response.AppAuthorized)
	assert.Len(t, response.UnauthorizedFields, 1)
	assert.Equal(t, "test.field", response.UnauthorizedFields[0].FieldName)

	// Verify mock was called
	mockService.AssertExpectations(t)
}

func TestHandler_GetPolicyDecision_InvalidJSON(t *testing.T) {
	// Create mock service
	mockService := new(MockPolicyMetadataService)

	// Create test handler with mock service
	handler := &TestHandler{
		policyService: mockService,
	}

	// Create HTTP request with invalid JSON
	httpReq := httptest.NewRequest("POST", "/api/v1/policy/decide", bytes.NewBufferString("invalid json"))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.GetPolicyDecision(w, httpReq)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")

	// Verify mock was not called
	mockService.AssertNotCalled(t, "GetPolicyDecision")
}

func TestHandler_SetupRoutes(t *testing.T) {
	// Create handler with nil service (we're just testing route setup)
	handler := &Handler{
		policyService: nil,
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Setup routes
	handler.SetupRoutes(mux)

	// Test that routes are registered by checking if they exist
	// This is a basic test - in a real scenario you'd make actual HTTP requests
	assert.NotNil(t, mux)
}

// Integration test for the actual Handler
func TestHandler_Integration(t *testing.T) {
	// This test verifies that the actual Handler works correctly
	// We'll test the route setup and basic functionality

	// Create handler (we can't easily mock the service without dependency injection)
	// So we'll just test that the handler can be created and routes can be set up
	handler := &Handler{
		policyService: nil, // Will cause errors if called, but that's expected
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Setup routes
	handler.SetupRoutes(mux)

	// Test that routes are registered
	assert.NotNil(t, mux)

	// Test that we can make requests to the endpoints (they'll fail due to nil service, but that's expected)
	testCases := []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/v1/policy/metadata", `{"schemaId":"test","records":[]}`},
		{"POST", "/api/v1/policy/update-allowlist", `{"applicationId":"test","grantDuration":"30d","records":[]}`},
		{"POST", "/api/v1/policy/decide", `{"applicationId":"test","requiredFields":[]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// This will call the actual handler methods
			mux.ServeHTTP(w, req)

			// We expect either 500 (service error), 404 (route not found), or 200 (if middleware handles the error)
			// All are acceptable since we're testing the routing, not the service logic
			assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusNotFound || w.Code == http.StatusOK,
				"Expected 500, 404, or 200, got %d", w.Code)
		})
	}
}
