package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
)

// TestServer represents a test HTTP server with common setup
type TestServer struct {
	APIServer *handlers.APIServer
	Mux       *http.ServeMux
}

// NewTestServer creates a new test server instance
func NewTestServer() *TestServer {
	apiServer := handlers.NewAPIServer()
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	return &TestServer{
		APIServer: apiServer,
		Mux:       mux,
	}
}

// MakeRequest makes an HTTP request and returns the response
func (ts *TestServer) MakeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			panic("Failed to marshal request body: " + err.Error())
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.Mux.ServeHTTP(w, req)

	return w
}

// MakeGETRequest makes a GET request
func (ts *TestServer) MakeGETRequest(url string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", url, nil)
}

// MakePOSTRequest makes a POST request
func (ts *TestServer) MakePOSTRequest(url string, body interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", url, body)
}

// MakePUTRequest makes a PUT request
func (ts *TestServer) MakePUTRequest(url string, body interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", url, body)
}

// MakeDELETERequest makes a DELETE request
func (ts *TestServer) MakeDELETERequest(url string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", url, nil)
}

// AssertResponseStatus checks if the response has the expected status code
func AssertResponseStatus(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	if w.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Response: %s", expectedStatus, w.Code, w.Body.String())
	}
}

// AssertJSONResponse checks if the response can be unmarshaled as JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
		t.Errorf("Failed to unmarshal response: %v. Response: %s", err, w.Body.String())
	}
}

// AssertErrorResponse checks if the response contains an error
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	AssertResponseStatus(t, w, expectedStatus)

	var errorResp map[string]string
	AssertJSONResponse(t, w, &errorResp)

	if _, hasError := errorResp["error"]; !hasError {
		t.Error("Expected error field in response")
	}
}

// AssertSuccessResponse checks if the response is successful
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	AssertResponseStatus(t, w, expectedStatus)

	// Try to unmarshal as JSON to ensure it's valid
	var response interface{}
	AssertJSONResponse(t, w, &response)
}

// CreateTestConsumer creates a consumer for testing and returns the consumer ID
func (ts *TestServer) CreateTestConsumer(t *testing.T, name, email, phone string) string {
	consumerReq := map[string]string{
		"consumerName": name,
		"contactEmail": email,
		"phoneNumber":  phone,
	}

	w := ts.MakePOSTRequest("/consumers", consumerReq)
	AssertResponseStatus(t, w, http.StatusCreated)

	var consumer map[string]interface{}
	AssertJSONResponse(t, w, &consumer)

	consumerID, ok := consumer["consumerId"].(string)
	if !ok {
		t.Fatal("Expected consumerId in response")
	}

	return consumerID
}

// CreateTestConsumerApp creates a consumer application for testing and returns the submission ID
func (ts *TestServer) CreateTestConsumerApp(t *testing.T, consumerID string, requiredFields map[string]bool) string {
	appReq := map[string]interface{}{
		"required_fields": requiredFields,
	}

	w := ts.MakePOSTRequest("/consumer-applications/"+consumerID, appReq)
	AssertResponseStatus(t, w, http.StatusCreated)

	var app map[string]interface{}
	AssertJSONResponse(t, w, &app)

	submissionID, ok := app["submissionId"].(string)
	if !ok {
		t.Fatal("Expected submissionId in response")
	}

	return submissionID
}

// CreateTestProviderProfile creates a provider profile directly for testing and returns the provider ID
func (ts *TestServer) CreateTestProviderProfile(t *testing.T, name, email, phone, providerType string) string {
	profile, err := ts.APIServer.GetProviderService().CreateProviderProfileForTesting(
		name, email, phone, providerType,
	)
	if err != nil {
		t.Fatalf("Failed to create provider profile: %v", err)
	}
	return profile.ProviderID
}

// CreateTestSchemaSubmission creates a schema submission for testing and returns the schema ID
func (ts *TestServer) CreateTestSchemaSubmission(t *testing.T, providerID, sdl string) string {
	schemaReq := map[string]interface{}{
		"sdl":       sdl,
		"schema_id": nil,
	}

	w := ts.MakePOSTRequest("/providers/"+providerID+"/schema-submissions", schemaReq)
	AssertResponseStatus(t, w, http.StatusCreated)

	var schema map[string]interface{}
	AssertJSONResponse(t, w, &schema)

	schemaID, ok := schema["submissionId"].(string)
	if !ok {
		t.Fatal("Expected submissionId in response")
	}

	return schemaID
}

// SubmitSchemaForReview submits a draft schema for admin review
func (ts *TestServer) SubmitSchemaForReview(t *testing.T, providerID, schemaID string) {
	updateReq := map[string]string{
		"status": "pending",
	}
	w := ts.MakePUTRequest("/providers/"+providerID+"/schema-submissions/"+schemaID, updateReq)
	AssertResponseStatus(t, w, http.StatusOK)
}

// ApproveSchemaSubmission approves a schema submission for testing
func (ts *TestServer) ApproveSchemaSubmission(t *testing.T, providerID, schemaID string) {
	updateReq := map[string]string{
		"status": "approved",
	}

	w := ts.MakePUTRequest("/providers/"+providerID+"/schema-submissions/"+schemaID, updateReq)
	AssertResponseStatus(t, w, http.StatusOK)
}
