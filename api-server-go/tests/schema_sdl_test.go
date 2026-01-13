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
	ts := NewTestServer()
	defer ts.Close()
	apiServer := ts.APIServer
	mux := ts.Mux

	// First, create a provider profile so we can submit schemas
	providerID := createProviderProfile(t, mux, apiServer)

	t.Run("Create Provider Schema with SDL", func(t *testing.T) {
		reqBody := models.CreateProviderSchemaSubmissionRequest{
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
		req := httptest.NewRequest("POST", "/providers/"+providerID+"/schema-submissions", bytes.NewBuffer(jsonBody))
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

		if response["status"] != "draft" {
			t.Errorf("Expected status 'draft', got %v", response["status"])
		}
	})

	t.Run("Create Provider Schema with SDL - Provider Not Found", func(t *testing.T) {
		reqBody := models.CreateProviderSchemaSubmissionRequest{
			SDL: "type User { id: ID! }",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/providers/nonexistent-provider/schema-submissions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Create Provider Schema with SDL - Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/providers/test-provider/schema-submissions", bytes.NewBufferString("invalid json"))
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
	// Create a provider profile directly using the service
	profile, err := apiServer.GetProviderService().CreateProviderProfileForTesting(
		"Test Provider",
		"test@example.com",
		"1234567890",
		"government",
	)
	if err != nil {
		t.Fatalf("Failed to create provider profile: %v", err)
	}

	t.Logf("Created provider with ID: %s", profile.ProviderID)
	return profile.ProviderID
}

// Health and debug endpoint tests
func TestHealthAndDebugEndpoints(t *testing.T) {
	ts := NewTestServer()

	// Add health and debug endpoints
	ts.Mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"api-server"}`))
	}))
	ts.Mux.Handle("/debug", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"path":"` + r.URL.Path + `","method":"` + r.Method + `"}`))
	}))

	t.Run("Health and Debug", func(t *testing.T) {
		// Test health endpoint
		w := ts.MakeGETRequest("/health")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test debug endpoint
		w = ts.MakeGETRequest("/debug")
		AssertResponseStatus(t, w, http.StatusOK)
	})
}
