package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/database"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/handlers"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

// TestSchemaAPIEndpoints tests the HTTP API endpoints
func TestSchemaAPIEndpoints(t *testing.T) {
	// Skip if no database connection
	if !hasDatabaseConnection() {
		t.Skip("Skipping integration tests - no database connection")
	}

	// Setup
	connectionString := "host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable"

	// Create schema mapping database
	schemaMappingDB, err := database.NewSchemaMappingDB(connectionString)
	if err != nil {
		t.Fatalf("Failed to connect to schema mapping database: %v", err)
	}
	defer schemaMappingDB.Close()

	schemaService := services.NewSchemaService(schemaMappingDB)
	schemaHandler := handlers.NewSchemaHandler(schemaService)

	// Test 1: Create schema
	t.Run("CreateSchema", func(t *testing.T) {
		reqBody := map[string]string{
			"version":    "1.0.0",
			"sdl":        "type Query { hello: String }",
			"created_by": "test-user",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/sdl", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.CreateSchema(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["version"] != "1.0.0" {
			t.Errorf("Expected version 1.0.0, got %v", response["version"])
		}
	})

	// Test 2: Get all schemas
	t.Run("GetAllSchemas", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sdl/versions", nil)
		w := httptest.NewRecorder()

		schemaHandler.GetSchemas(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if len(response) == 0 {
			t.Error("Expected at least one schema")
		}
	})

	// Test 3: Get active schema (should be empty initially)
	t.Run("GetActiveSchemaEmpty", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sdl", nil)
		w := httptest.NewRecorder()

		schemaHandler.GetActiveSchema(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 4: Activate schema
	t.Run("ActivateSchema", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/sdl/versions/1.0.0/activate", nil)
		w := httptest.NewRecorder()

		schemaHandler.ActivateSchema(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 5: Get active schema (should now have the schema)
	t.Run("GetActiveSchema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sdl", nil)
		w := httptest.NewRecorder()

		schemaHandler.GetActiveSchema(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["sdl"] != "type Query { hello: String }" {
			t.Errorf("Expected SDL 'type Query { hello: String }', got '%s'", response["sdl"])
		}
	})

	// Test 6: Validate SDL
	t.Run("ValidateSDL", func(t *testing.T) {
		reqBody := map[string]string{
			"sdl": "type Query { hello: String }",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/sdl/validate", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.ValidateSDL(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]bool
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if !response["valid"] {
			t.Error("Expected valid SDL to return true")
		}
	})

	// Test 7: Check compatibility
	t.Run("CheckCompatibility", func(t *testing.T) {
		reqBody := map[string]string{
			"sdl": "type Query { hello: String world: String }",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/sdl/check-compatibility", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.CheckCompatibility(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		if response["compatible"] != true {
			t.Errorf("Expected compatible to be true, got %v", response["compatible"])
		}
	})
}

// TestSchemaAPIErrorHandling tests error handling in API endpoints
func TestSchemaAPIErrorHandling(t *testing.T) {
	// Skip if no database connection
	if !hasDatabaseConnection() {
		t.Skip("Skipping integration tests - no database connection")
	}

	connectionString := "host=localhost port=5432 user=postgres password=password dbname=orchestration_engine sslmode=disable"

	// Create schema mapping database
	schemaMappingDB, err := database.NewSchemaMappingDB(connectionString)
	if err != nil {
		t.Fatalf("Failed to connect to schema mapping database: %v", err)
	}
	defer schemaMappingDB.Close()

	schemaService := services.NewSchemaService(schemaMappingDB)
	schemaHandler := handlers.NewSchemaHandler(schemaService)

	// Test 1: Create schema with invalid JSON
	t.Run("CreateSchemaInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/sdl", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.CreateSchema(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 2: Create schema with missing required fields
	t.Run("CreateSchemaMissingFields", func(t *testing.T) {
		reqBody := map[string]string{
			"version": "1.0.0",
			// Missing sdl and created_by
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/sdl", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.CreateSchema(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 3: Activate non-existent schema
	t.Run("ActivateNonExistentSchema", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/sdl/versions/non-existent/activate", nil)
		w := httptest.NewRecorder()

		schemaHandler.ActivateSchema(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 4: Validate SDL with invalid JSON
	t.Run("ValidateSDLInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/sdl/validate", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.ValidateSDL(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 5: Check compatibility with invalid JSON
	t.Run("CheckCompatibilityInvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/sdl/check-compatibility", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.CheckCompatibility(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// TestSchemaAPIWithoutDatabase tests API behavior when database is not available
func TestSchemaAPIWithoutDatabase(t *testing.T) {
	// Create handler without database connection
	schemaHandler := handlers.NewSchemaHandler(nil)

	// Test 1: Create schema without database
	t.Run("CreateSchemaWithoutDatabase", func(t *testing.T) {
		reqBody := map[string]string{
			"version":    "1.0.0",
			"sdl":        "type Query { hello: String }",
			"created_by": "test-user",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/sdl", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		schemaHandler.CreateSchema(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 2: Get schemas without database
	t.Run("GetSchemasWithoutDatabase", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/sdl/versions", nil)
		w := httptest.NewRecorder()

		schemaHandler.GetSchemas(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503, got %d: %s", w.Code, w.Body.String())
		}
	})
}
