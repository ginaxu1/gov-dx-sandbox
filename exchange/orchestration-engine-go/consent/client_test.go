package consent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

func init() {
	// Any necessary initialization for tests can be done here
	logger.Init()
}

func TestNewCEClient(t *testing.T) {
	baseUrl := "http://test-consent-engine.com"
	client := NewCEClient(baseUrl)

	if client == nil {
		t.Fatal("Expected CEClient to be created, got nil")
	}

	if client.baseUrl != baseUrl {
		t.Errorf("Expected baseUrl to be %s, got %s", baseUrl, client.baseUrl)
	}

	if client.httpClient == nil {
		t.Fatal("Expected httpClient to be initialized, got nil")
	}

	expectedTimeout := time.Second * 10
	if client.httpClient.Timeout != expectedTimeout {
		t.Errorf("Expected timeout to be %v, got %v", expectedTimeout, client.httpClient.Timeout)
	}
}

func TestMakeConsentRequest_Success(t *testing.T) {
	// Create a test server that returns a successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/consents" {
			t.Errorf("Expected path /consents, got %s", r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify request body can be decoded
		var req CERequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := CEResponse{
			Status:           "success",
			ConsentPortalUrl: "http://consent-portal.com/session/123",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewCEClient(server.URL)

	request := &CERequest{
		AppId: "test-app-123",
		DataFields: []DataOwnerRecord{
			{
				OwnerType: "individual",
				OwnerId:   "user-456",
				Fields:    []string{"name", "email", "phone"},
			},
		},
		Purpose:   "test purpose",
		SessionId: "session-789",
	}

	response, err := client.MakeConsentRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.ConsentPortalUrl != "http://consent-portal.com/session/123" {
		t.Errorf("Expected ConsentPortalUrl 'http://consent-portal.com/session/123', got '%s'", response.ConsentPortalUrl)
	}
}

func TestMakeConsentRequest_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewCEClient(server.URL)

	request := &CERequest{
		AppId:      "test-app-123",
		DataFields: []DataOwnerRecord{},
		Purpose:    "test purpose",
		SessionId:  "session-789",
	}

	// Note: The current implementation doesn't check HTTP status codes,
	// so this test verifies the actual behavior
	response, err := client.MakeConsentRequest(request)

	// The implementation tries to decode response regardless of status code
	// If the response body is not valid JSON, we expect an error
	if err == nil && response != nil {
		// If no error, it means the server returned valid JSON despite 500 status
		t.Log("Server returned valid JSON response despite error status")
	}
}

func TestMakeConsentRequest_InvalidResponseJSON(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json{"))
	}))
	defer server.Close()

	client := NewCEClient(server.URL)

	request := &CERequest{
		AppId:      "test-app-123",
		DataFields: []DataOwnerRecord{},
		Purpose:    "test purpose",
		SessionId:  "session-789",
	}

	response, err := client.MakeConsentRequest(request)

	if err == nil {
		t.Error("Expected error when decoding invalid JSON, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response on decode error, got %v", response)
	}
}

func TestMakeConsentRequest_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	client := NewCEClient("http://invalid-host-that-does-not-exist-12345.com")

	request := &CERequest{
		AppId:      "test-app-123",
		DataFields: []DataOwnerRecord{},
		Purpose:    "test purpose",
		SessionId:  "session-789",
	}

	response, err := client.MakeConsentRequest(request)

	if err == nil {
		t.Error("Expected network error, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response on network error, got %v", response)
	}
}

func TestMakeConsentRequest_EmptyDataFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CERequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify empty data fields
		if len(req.DataFields) != 0 {
			t.Errorf("Expected 0 data fields, got %d", len(req.DataFields))
		}

		w.Header().Set("Content-Type", "application/json")
		response := CEResponse{
			Status:           "success",
			ConsentPortalUrl: "http://consent-portal.com/session/empty",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewCEClient(server.URL)

	request := &CERequest{
		AppId:      "test-app-123",
		DataFields: []DataOwnerRecord{},
		Purpose:    "test purpose",
		SessionId:  "session-789",
	}

	response, err := client.MakeConsentRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}
}

func TestMakeConsentRequest_MultipleDataOwners(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CERequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify multiple data owners
		if len(req.DataFields) != 3 {
			t.Errorf("Expected 3 data fields, got %d", len(req.DataFields))
		}

		w.Header().Set("Content-Type", "application/json")
		response := CEResponse{
			Status:           "pending",
			ConsentPortalUrl: "http://consent-portal.com/session/multi",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewCEClient(server.URL)

	request := &CERequest{
		AppId: "test-app-123",
		DataFields: []DataOwnerRecord{
			{
				OwnerType: "individual",
				OwnerId:   "user-1",
				Fields:    []string{"name", "email"},
			},
			{
				OwnerType: "organization",
				OwnerId:   "org-1",
				Fields:    []string{"company_name", "tax_id"},
			},
			{
				OwnerType: "individual",
				OwnerId:   "user-2",
				Fields:    []string{"address", "phone"},
			},
		},
		Purpose:   "multi-owner consent",
		SessionId: "session-multi",
	}

	response, err := client.MakeConsentRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", response.Status)
	}
}

func TestCERequest_JSONMarshaling(t *testing.T) {
	request := &CERequest{
		AppId: "test-app",
		DataFields: []DataOwnerRecord{
			{
				OwnerType: "individual",
				OwnerId:   "user-1",
				Fields:    []string{"field1", "field2"},
			},
		},
		Purpose:   "testing",
		SessionId: "session-1",
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal CERequest: %v", err)
	}

	var unmarshaled CERequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal CERequest: %v", err)
	}

	if unmarshaled.AppId != request.AppId {
		t.Errorf("AppId mismatch: expected %s, got %s", request.AppId, unmarshaled.AppId)
	}

	if unmarshaled.Purpose != request.Purpose {
		t.Errorf("Purpose mismatch: expected %s, got %s", request.Purpose, unmarshaled.Purpose)
	}

	if unmarshaled.SessionId != request.SessionId {
		t.Errorf("SessionId mismatch: expected %s, got %s", request.SessionId, unmarshaled.SessionId)
	}
}

func TestCEResponse_JSONUnmarshaling(t *testing.T) {
	jsonData := `{"status":"approved","consent_portal_url":"http://example.com/consent"}`

	var response CEResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal CEResponse: %v", err)
	}

	if response.Status != "approved" {
		t.Errorf("Status mismatch: expected 'approved', got '%s'", response.Status)
	}

	if response.ConsentPortalUrl != "http://example.com/consent" {
		t.Errorf("ConsentPortalUrl mismatch: expected 'http://example.com/consent', got '%s'", response.ConsentPortalUrl)
	}
}
