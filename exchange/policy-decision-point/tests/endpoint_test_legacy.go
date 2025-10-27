// NOTE: These tests are for the legacy API endpoints that were removed during V1 refactoring.
// These tests need to be updated to work with the new V1 API endpoints:
// - POST /api/v1/policy/metadata (instead of /policy-metadata)
// - POST /api/v1/policy/update-allowlist (instead of /allow-list)
// - POST /api/v1/policy/decide (new endpoint)
// The models and handler logic have also changed to use GORM and V1 architecture.
// 
// ⚠️  THESE TESTS WILL NOT COMPILE until updated for V1 architecture.
// See tests/README.md for migration guide.

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	// TODO: Update to V1 models and fix struct compatibility
	// "github.com/gov-dx-sandbox/exchange/policy-decision-point/v1/models"
)

// Temporary mock models to prevent compilation errors
type PolicyMetadataCreateRequest struct {
	FieldName         string           `json:"field_name"`
	DisplayName       string           `json:"display_name"`
	Description       string           `json:"description"`
	Source            string           `json:"source"`
	IsOwner           bool             `json:"is_owner"`
	AccessControlType string           `json:"access_control_type"`
	AllowList         []AllowListEntry `json:"allow_list"`
}

type AllowListEntry struct {
	ApplicationID string `json:"application_id"`
	ExpiresAt     int64  `json:"expires_at"`
}

type PolicyMetadataCreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id"`
}

type AllowListUpdateRequest struct {
	FieldName     string `json:"field_name"`
	ApplicationID string `json:"application_id"`
	ExpiresAt     string `json:"expires_at"`
}

type AllowListUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestCreatePolicyMetadata(t *testing.T) {
	// Create mock database service
	dbService := NewMockDatabaseService()
	handler := NewMetadataHandler(dbService)

	// Create test request
	req := models.PolicyMetadataCreateRequest{
		FieldName:         "person.fullName",
		DisplayName:       "Full Name",
		Description:       "Complete name of the person",
		Source:            "passport_system",
		IsOwner:           true,
		AccessControlType: "restricted",
		AllowList:         []models.AllowListEntry{},
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/policy-metadata", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreatePolicyMetadata(w, httpReq)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response models.PolicyMetadataCreateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}

	if response.ID == "" {
		t.Errorf("Expected ID to be set, got empty string")
	}
}

func TestUpdateAllowList(t *testing.T) {
	// Create mock database service
	dbService := NewMockDatabaseService()
	handler := NewMetadataHandler(dbService)

	// Create test request
	req := models.AllowListUpdateRequest{
		FieldName:     "person.fullName",
		ApplicationID: "test-app",
		ExpiresAt:     "2024-12-31T23:59:59Z",
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/allow-list", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.UpdateAllowList(w, httpReq)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.AllowListUpdateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success to be true, got false")
	}
}

func TestCreatePolicyMetadataValidation(t *testing.T) {
	// Create mock database service
	dbService := NewMockDatabaseService()
	handler := NewMetadataHandler(dbService)

	// Create invalid request (missing required fields)
	req := models.PolicyMetadataCreateRequest{
		FieldName: "", // Missing required field
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/policy-metadata", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.CreatePolicyMetadata(w, httpReq)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateAllowListValidation(t *testing.T) {
	// Create mock database service
	dbService := NewMockDatabaseService()
	handler := NewMetadataHandler(dbService)

	// Create invalid request (missing required fields)
	req := models.AllowListUpdateRequest{
		FieldName: "", // Missing required field
	}

	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/allow-list", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler.UpdateAllowList(w, httpReq)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
