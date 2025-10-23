package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
)

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
