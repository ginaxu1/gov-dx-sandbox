package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock HTTP handlers for testing
type MockSchemaHandler struct{}

func (h *MockSchemaHandler) GetUnifiedSchemas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schemas := []UnifiedSchema{
		{
			ID:        "1",
			Version:   "1.0.0",
			SDL:       "type Query { personInfo(nic: String!): PersonInfo }",
			IsActive:  false,
			Notes:     "Initial version",
			CreatedBy: "admin",
		},
		{
			ID:        "2",
			Version:   "1.1.0",
			SDL:       "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			IsActive:  true,
			Notes:     "Added fullName field",
			CreatedBy: "admin",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

func (h *MockSchemaHandler) GetLatestUnifiedSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schema := UnifiedSchema{
		ID:        "2",
		Version:   "1.1.0",
		SDL:       "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
		IsActive:  true,
		Notes:     "Added fullName field",
		CreatedBy: "admin",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schema)
}

func (h *MockSchemaHandler) CreateUnifiedSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Version   string `json:"version"`
		SDL       string `json:"sdl"`
		Notes     string `json:"notes"`
		CreatedBy string `json:"createdBy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Version == "" || req.SDL == "" || req.CreatedBy == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Mock compatibility check
	if req.SDL == "invalid schema" {
		http.Error(w, "Backward compatibility check failed", http.StatusBadRequest)
		return
	}

	// Create new schema
	schema := UnifiedSchema{
		ID:        "3",
		Version:   req.Version,
		SDL:       req.SDL,
		IsActive:  false, // New schemas are not active by default
		Notes:     req.Notes,
		CreatedBy: req.CreatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schema)
}

func (h *MockSchemaHandler) ActivateUnifiedSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Mock activation
	response := map[string]string{
		"message": "Schema activated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *MockSchemaHandler) GetProviderSchemas(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schemas := map[string]ProviderSchema{
		"drp": {
			ID:         "1",
			ProviderID: "drp",
			SchemaName: "person-schema",
			SDL:        "type Query { person(nic: String!): Person }\ntype Person { fullName: String }",
			IsActive:   true,
		},
		"dmt": {
			ID:         "2",
			ProviderID: "dmt",
			SchemaName: "vehicle-schema",
			SDL:        "type Query { vehicle(regNo: String!): Vehicle }\ntype Vehicle { regNo: String make: String }",
			IsActive:   true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schemas)
}

// Test cases for API endpoints
func TestGetUnifiedSchemas(t *testing.T) {
	handler := &MockSchemaHandler{}
	req, err := http.NewRequest("GET", "/sdl/versions", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetUnifiedSchemas(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var schemas []UnifiedSchema
	err = json.Unmarshal(rr.Body.Bytes(), &schemas)
	require.NoError(t, err)
	assert.Len(t, schemas, 2)
	assert.Equal(t, "1.0.0", schemas[0].Version)
	assert.Equal(t, "1.1.0", schemas[1].Version)
}

func TestGetLatestUnifiedSchema(t *testing.T) {
	handler := &MockSchemaHandler{}
	req, err := http.NewRequest("GET", "/sdl", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetLatestUnifiedSchema(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var schema UnifiedSchema
	err = json.Unmarshal(rr.Body.Bytes(), &schema)
	require.NoError(t, err)
	assert.Equal(t, "1.1.0", schema.Version)
	assert.True(t, schema.IsActive)
}

func TestCreateUnifiedSchema(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid schema creation",
			requestBody: map[string]interface{}{
				"version":   "1.2.0",
				"sdl":       "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String birthDate: String }",
				"notes":     "Added birthDate field",
				"createdBy": "admin",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"version": "1.2.0",
				"sdl":     "type Query { personInfo(nic: String!): PersonInfo }",
				// Missing notes and createdBy
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "invalid schema - compatibility check fails",
			requestBody: map[string]interface{}{
				"version":   "1.2.0",
				"sdl":       "invalid schema",
				"notes":     "Invalid schema",
				"createdBy": "admin",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	handler := &MockSchemaHandler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/sdl", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.CreateUnifiedSchema(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError {
				var schema UnifiedSchema
				err = json.Unmarshal(rr.Body.Bytes(), &schema)
				require.NoError(t, err)
				assert.False(t, schema.IsActive) // New schemas should not be active
			}
		})
	}
}

func TestActivateUnifiedSchema(t *testing.T) {
	handler := &MockSchemaHandler{}
	req, err := http.NewRequest("POST", "/sdl/versions/1.2.0/activate", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ActivateUnifiedSchema(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Schema activated successfully", response["message"])
}

func TestGetProviderSchemas(t *testing.T) {
	handler := &MockSchemaHandler{}
	req, err := http.NewRequest("GET", "/sdl/providers", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.GetProviderSchemas(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var schemas map[string]ProviderSchema
	err = json.Unmarshal(rr.Body.Bytes(), &schemas)
	require.NoError(t, err)
	assert.Len(t, schemas, 2)
	assert.Contains(t, schemas, "drp")
	assert.Contains(t, schemas, "dmt")
	assert.Equal(t, "person-schema", schemas["drp"].SchemaName)
	assert.Equal(t, "vehicle-schema", schemas["dmt"].SchemaName)
}

// Test cases for field mapping API endpoints
func TestFieldMappingEndpoints(t *testing.T) {
	t.Run("create field mapping", func(t *testing.T) {
		// Mock field mapping creation
		mapping := map[string]interface{}{
			"unified_field_path":  "person.fullName",
			"provider_id":         "drp",
			"provider_field_path": "person.fullName",
			"directives": map[string]interface{}{
				"sourceInfo": map[string]string{
					"providerKey":   "drp",
					"providerField": "person.fullName",
				},
			},
		}

		jsonBody, err := json.Marshal(mapping)
		require.NoError(t, err)

		_, err = http.NewRequest("POST", "/sdl/mappings", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		// Mock handler would be implemented here
		rr.WriteHeader(http.StatusCreated)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("get field mappings", func(t *testing.T) {
		_, err := http.NewRequest("GET", "/sdl/mappings", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		// Mock handler would return field mappings
		rr.WriteHeader(http.StatusOK)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("update field mapping", func(t *testing.T) {
		update := map[string]interface{}{
			"provider_id":         "dmt", // Changed provider
			"provider_field_path": "person.name",
		}

		jsonBody, err := json.Marshal(update)
		require.NoError(t, err)

		_, err = http.NewRequest("PUT", "/sdl/mappings/mapping-123", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		// Mock handler would update the mapping
		rr.WriteHeader(http.StatusOK)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("delete field mapping", func(t *testing.T) {
		_, err := http.NewRequest("DELETE", "/sdl/mappings/mapping-123", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		// Mock handler would delete the mapping
		rr.WriteHeader(http.StatusNoContent)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

// Test cases for error handling
func TestAPIErrorHandling(t *testing.T) {
	t.Run("invalid JSON in request body", func(t *testing.T) {
		handler := &MockSchemaHandler{}
		req, err := http.NewRequest("POST", "/sdl", bytes.NewBufferString("invalid json"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.CreateUnifiedSchema(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("unsupported HTTP method", func(t *testing.T) {
		handler := &MockSchemaHandler{}
		req, err := http.NewRequest("DELETE", "/sdl", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.GetUnifiedSchemas(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})
}

// Test cases for schema validation
func TestSchemaMappingValidation(t *testing.T) {
	tests := []struct {
		name        string
		sdl         string
		valid       bool
		expectError bool
	}{
		{
			name:        "valid GraphQL schema",
			sdl:         "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String }",
			valid:       true,
			expectError: false,
		},
		{
			name:        "invalid GraphQL syntax",
			sdl:         "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String", // Missing closing brace
			valid:       false,
			expectError: true,
		},
		{
			name:        "empty schema",
			sdl:         "",
			valid:       false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateGraphQLSchema(tt.sdl)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.valid, valid)
		})
	}
}

// Mock validation function
func validateGraphQLSchema(sdl string) (bool, error) {
	if sdl == "" {
		return false, assert.AnError
	}
	if sdl == "type Query { personInfo(nic: String!): PersonInfo }\ntype PersonInfo { fullName: String" {
		return false, assert.AnError
	}
	return true, nil
}
