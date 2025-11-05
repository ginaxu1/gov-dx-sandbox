package policy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
)

func TestNewPdpClient(t *testing.T) {
	baseUrl := "http://localhost:8080"
	client := NewPdpClient(baseUrl)

	if client == nil {
		t.Fatal("Expected non-nil PdpClient")
	}

	if client.baseUrl != baseUrl {
		t.Errorf("Expected baseUrl %s, got %s", baseUrl, client.baseUrl)
	}

	if client.httpClient == nil {
		t.Error("Expected non-nil httpClient")
	}

	if client.httpClient.Timeout.Seconds() != 10 {
		t.Errorf("Expected timeout of 10 seconds, got %v", client.httpClient.Timeout)
	}
}

func TestMakePdpRequest_Success(t *testing.T) {
	// Create a mock server

	logger.Init()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/policy/decide" {
			t.Errorf("Expected path /api/v1/policy/decide, got %s", r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify request body
		var req PdpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Send mock response
		response := PdpResponse{
			AppAuthorized:         true,
			ConsentRequired:       false,
			ConsentRequiredFields: []ConsentRequiredField{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := NewPdpClient(server.URL)

	// Create test request
	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "field1",
			},
		},
	}

	// Make request
	response, err := client.MakePdpRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	if !response.AppAuthorized {
		t.Error("Expected AppAuthorized to be true")
	}

	if response.ConsentRequired {
		t.Error("Expected ConsentRequired to be false")
	}
}

func TestMakePdpRequest_ConsentRequired(t *testing.T) {
	// Create a mock server that returns consent required
	displayName := "Test Field"
	description := "Test Description"
	owner := "Test Owner"

	logger.Init()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PdpResponse{
			AppAuthorized:   true,
			ConsentRequired: true,
			ConsentRequiredFields: []ConsentRequiredField{
				{
					FieldName:   "sensitiveField",
					SchemaID:    "schema1",
					DisplayName: &displayName,
					Description: &description,
					Owner:       &owner,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewPdpClient(server.URL)

	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "sensitiveField",
			},
		},
	}

	response, err := client.MakePdpRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !response.AppAuthorized {
		t.Error("Expected AppAuthorized to be true")
	}

	if !response.ConsentRequired {
		t.Error("Expected ConsentRequired to be true")
	}

	if len(response.ConsentRequiredFields) != 1 {
		t.Fatalf("Expected 1 consent required field, got %d", len(response.ConsentRequiredFields))
	}

	field := response.ConsentRequiredFields[0]
	if field.FieldName != "sensitiveField" {
		t.Errorf("Expected FieldName sensitiveField, got %s", field.FieldName)
	}

	if field.SchemaID != "schema1" {
		t.Errorf("Expected SchemaID schema1, got %s", field.SchemaID)
	}

	if field.DisplayName == nil || *field.DisplayName != displayName {
		t.Errorf("Expected DisplayName %s, got %v", displayName, field.DisplayName)
	}

	if field.Description == nil || *field.Description != description {
		t.Errorf("Expected Description %s, got %v", description, field.Description)
	}

	if field.Owner == nil || *field.Owner != owner {
		t.Errorf("Expected Owner %s, got %v", owner, field.Owner)
	}
}

func TestMakePdpRequest_ServerError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	logger.Init()

	client := NewPdpClient(server.URL)

	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "field1",
			},
		},
	}

	response, err := client.MakePdpRequest(request)

	if err == nil {
		t.Error("Expected error when server returns non-JSON response")
	}

	if response != nil {
		t.Errorf("Expected nil response on error, got %v", response)
	}
}

func TestMakePdpRequest_InvalidJSON(t *testing.T) {
	// Create a mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewPdpClient(server.URL)

	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "field1",
			},
		},
	}

	response, err := client.MakePdpRequest(request)

	if err == nil {
		t.Error("Expected error when server returns invalid JSON")
	}

	if response != nil {
		t.Errorf("Expected nil response on error, got %v", response)
	}
}

func TestMakePdpRequest_NetworkError(t *testing.T) {
	// Use an invalid URL to simulate network error
	client := NewPdpClient("http://invalid-url-that-does-not-exist:9999")

	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "field1",
			},
		},
	}

	response, err := client.MakePdpRequest(request)

	if err == nil {
		t.Error("Expected error when network request fails")
	}

	if response != nil {
		t.Errorf("Expected nil response on error, got %v", response)
	}
}

func TestMakePdpRequest_MultipleRequiredFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify multiple required fields in request
		var req PdpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.RequiredFields) != 3 {
			t.Errorf("Expected 3 required fields, got %d", len(req.RequiredFields))
		}

		response := PdpResponse{
			AppAuthorized:         true,
			ConsentRequired:       false,
			ConsentRequiredFields: []ConsentRequiredField{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewPdpClient(server.URL)

	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "field1",
			},
			{
				ProviderKey: "provider2",
				SchemaId:    "schema2",
				FieldName:   "field2",
			},
			{
				ProviderKey: "provider3",
				SchemaId:    "schema3",
				FieldName:   "field3",
			},
		},
	}

	response, err := client.MakePdpRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}
}

func TestMakePdpRequest_AppNotAuthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := PdpResponse{
			AppAuthorized:         false,
			ConsentRequired:       false,
			ConsentRequiredFields: []ConsentRequiredField{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewPdpClient(server.URL)

	request := &PdpRequest{
		ConsumerId: "consumer123",
		AppId:      "app456",
		RequestId:  "req789",
		RequiredFields: []RequiredField{
			{
				ProviderKey: "provider1",
				SchemaId:    "schema1",
				FieldName:   "field1",
			},
		},
	}

	response, err := client.MakePdpRequest(request)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.AppAuthorized {
		t.Error("Expected AppAuthorized to be false")
	}

	if response.ConsentRequired {
		t.Error("Expected ConsentRequired to be false")
	}
}
