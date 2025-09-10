package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestProviderSchemaSDLAPI(t *testing.T) {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	// First, create a provider profile so we can submit schemas
	providerID := createProviderProfile(t, mux, apiServer)

	t.Run("Create Provider Schema with SDL", func(t *testing.T) {
		reqBody := models.CreateProviderSchemaSDLRequest{
			SDL: `directive @accessControl(type: String!) on FIELD_DEFINITION

directive @source(value: String!) on FIELD_DEFINITION

directive @isOwner(value: Boolean!) on FIELD_DEFINITION

directive @description(value: String!) on FIELD_DEFINITION

type BirthInfo {
  birthCertificateID: ID! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
  birthPlace: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
  birthDate: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
}

type User {
  id: ID! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
  name: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
  email: String! @accessControl(type: "public") @source(value: "authoritative") @isOwner(value: false)
  birthInfo: BirthInfo @accessControl(type: "public") @source(value: "authoritative") @description(value: "Default Description")
}

type Query {
  getUser(id: ID!): User @description(value: "Default Description")
  listUsers: [User!]! @description(value: "Default Description")
  getBirthInfo(userId: ID!): BirthInfo @description(value: "Default Description")
  listUsersByBirthPlace(birthPlace: String!): [User!]! @description(value: "Default Description")
  searchUsersByName(name: String!): [User!]! @description(value: "Default Description")
}`,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/provider-schemas/"+providerID, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response["providerId"] != providerID {
			t.Errorf("Expected providerId '%s', got %v", providerID, response["providerId"])
		}

		if response["sdl"] == nil {
			t.Error("Expected SDL to be present in response")
		}

		if response["status"] != "pending" {
			t.Errorf("Expected status 'pending', got %v", response["status"])
		}
	})

	t.Run("Create Provider Schema with SDL - Provider Not Found", func(t *testing.T) {
		reqBody := models.CreateProviderSchemaSDLRequest{
			SDL: "type User { id: ID! }",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/provider-schemas/nonexistent-provider", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Create Provider Schema with SDL - Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/provider-schemas/test-provider", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// Helper function to create a provider profile for testing
func createProviderProfile(t *testing.T, mux *http.ServeMux, apiServer *handlers.APIServer) string {
	// Create a provider submission first
	submissionReq := map[string]interface{}{
		"providerName": "Test Provider",
		"contactEmail": "test@example.com",
		"phoneNumber":  "1234567890",
		"providerType": "government",
	}

	jsonBody, _ := json.Marshal(submissionReq)
	req := httptest.NewRequest("POST", "/provider-submissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create provider submission: %d", w.Code)
	}

	var submissionResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &submissionResponse); err != nil {
		t.Fatalf("Failed to unmarshal submission response: %v", err)
	}

	submissionID := submissionResponse["submissionId"].(string)

	// Approve the submission to create a provider profile
	approveReq := map[string]interface{}{
		"status": "approved",
	}

	jsonBody, _ = json.Marshal(approveReq)
	req = httptest.NewRequest("PUT", "/provider-submissions/"+submissionID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to approve provider submission: %d", w.Code)
	}

	// Extract provider ID from the response
	var approveResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &approveResponse); err != nil {
		t.Fatalf("Failed to unmarshal approval response: %v", err)
	}

	providerID, ok := approveResponse["providerId"].(string)
	if !ok {
		t.Fatalf("Provider ID not found in approval response: %v", approveResponse)
	}

	t.Logf("Created provider with ID: %s", providerID)
	return providerID
}
