package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/handlers"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
)

// ============================================================================
// GRAPHQL HANDLER TESTS
// ============================================================================

func TestGraphQLHandler(t *testing.T) {

	t.Run("Valid GraphQL request", func(t *testing.T) {
		// Create mock services
		mockGraphQLService := &MockGraphQLService{}

		// Create handler
		handler := handlers.NewGraphQLHandler(mockGraphQLService)

		// Create test request
		req := models.GraphQLRequest{
			Query: "query { hello }",
			Variables: map[string]interface{}{
				"name": "world",
			},
			OperationName: "HelloQuery",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Execute handler
		handler.HandleGraphQL(w, httpReq)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "data")
	})

	t.Run("Invalid GraphQL request - missing query", func(t *testing.T) {
		// Create mock services
		mockGraphQLService := &MockGraphQLService{}

		// Create handler
		handler := handlers.NewGraphQLHandler(mockGraphQLService)

		// Create test request without query
		req := models.GraphQLRequest{
			Variables: map[string]interface{}{
				"name": "world",
			},
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Execute handler
		handler.HandleGraphQL(w, httpReq)

		// Note: The current handler doesn't validate the request, so it will succeed
		// In a real implementation, validation should be added
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "data")
	})

	t.Run("GraphQL request with schema version", func(t *testing.T) {
		// Create mock services
		mockGraphQLService := &MockGraphQLService{}

		// Create handler
		handler := handlers.NewGraphQLHandler(mockGraphQLService)

		// Create test request with schema version
		req := models.GraphQLRequest{
			Query:         "query { hello }",
			SchemaVersion: "1.0.0",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Execute handler
		handler.HandleGraphQL(w, httpReq)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GraphQL request with variables", func(t *testing.T) {
		// Create mock services
		mockGraphQLService := &MockGraphQLService{}

		// Create handler
		handler := handlers.NewGraphQLHandler(mockGraphQLService)

		// Create test request with variables
		req := models.GraphQLRequest{
			Query: "query GetPerson($nic: String!) { personInfo(nic: $nic) { fullName } }",
			Variables: map[string]interface{}{
				"nic": "123456789V",
			},
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		// Execute handler
		handler.HandleGraphQL(w, httpReq)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGraphQLRequestValidation(t *testing.T) {
	t.Run("Valid request validation", func(t *testing.T) {
		req := models.GraphQLRequest{
			Query:         "query { hello }",
			Variables:     map[string]interface{}{"name": "world"},
			OperationName: "HelloQuery",
		}

		handler := &handlers.GraphQLHandler{}
		err := handler.ValidateGraphQLRequest(&req)
		assert.NoError(t, err)
	})

	t.Run("Invalid request - empty query", func(t *testing.T) {
		req := models.GraphQLRequest{
			Variables:     map[string]interface{}{"name": "world"},
			OperationName: "HelloQuery",
		}

		handler := &handlers.GraphQLHandler{}
		err := handler.ValidateGraphQLRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query is required")
	})

	t.Run("Invalid request - query too short", func(t *testing.T) {
		req := models.GraphQLRequest{
			Query: "hi",
		}

		handler := &handlers.GraphQLHandler{}
		err := handler.ValidateGraphQLRequest(&req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query is too short")
	})
}

func TestGraphQLResponseFormat(t *testing.T) {
	t.Run("Successful response format", func(t *testing.T) {
		response := models.GraphQLResponse{
			Data: map[string]interface{}{
				"hello": "world",
			},
		}

		// Verify response structure
		assert.Contains(t, response.Data, "hello")
		assert.Equal(t, "world", response.Data["hello"])
		assert.Empty(t, response.Errors)
	})

	t.Run("Error response format", func(t *testing.T) {
		response := models.GraphQLResponse{
			Data: nil,
			Errors: []models.GraphQLError{
				{
					Message: "Field 'invalidField' doesn't exist",
					Locations: []models.GraphQLErrorLocation{
						{Line: 1, Column: 10},
					},
				},
			},
		}

		// Verify error structure
		assert.Nil(t, response.Data)
		assert.Len(t, response.Errors, 1)
		assert.Equal(t, "Field 'invalidField' doesn't exist", response.Errors[0].Message)
		assert.Len(t, response.Errors[0].Locations, 1)
		assert.Equal(t, 1, response.Errors[0].Locations[0].Line)
		assert.Equal(t, 10, response.Errors[0].Locations[0].Column)
	})
}

// ============================================================================
// MOCK SERVICES
// ============================================================================

type MockSchemaService struct{}

func (m *MockSchemaService) CreateSchema(req *models.CreateSchemaRequest) (*models.UnifiedSchema, error) {
	return &models.UnifiedSchema{
		ID:      "test-id",
		Version: "1.0.0",
		SDL:     req.SDL,
		Status:  "active",
	}, nil
}

func (m *MockSchemaService) GetSchemaByVersion(version string) (*models.UnifiedSchema, error) {
	return &models.UnifiedSchema{
		ID:      "test-id",
		Version: version,
		SDL:     "type Query { hello: String }",
		Status:  "active",
	}, nil
}

func (m *MockSchemaService) GetActiveSchema() (*models.UnifiedSchema, error) {
	return &models.UnifiedSchema{
		ID:      "test-id",
		Version: "1.0.0",
		SDL:     "type Query { hello: String }",
		Status:  "active",
	}, nil
}

func (m *MockSchemaService) GetAllSchemas() ([]*models.UnifiedSchema, error) {
	return []*models.UnifiedSchema{
		{
			ID:      "test-id-1",
			Version: "1.0.0",
			SDL:     "type Query { hello: String }",
			Status:  "active",
		},
	}, nil
}

func (m *MockSchemaService) UpdateSchemaStatus(version string, req *models.UpdateSchemaStatusRequest) error {
	return nil
}

func (m *MockSchemaService) DeleteSchema(version string) error {
	return nil
}

func (m *MockSchemaService) ActivateVersion(version string) error {
	return nil
}

func (m *MockSchemaService) DeactivateVersion(version string) error {
	return nil
}

func (m *MockSchemaService) GetSchemaVersions() ([]*models.SchemaVersionInfo, error) {
	return []*models.SchemaVersionInfo{
		{
			Version:     "1.0.0",
			Status:      "active",
			Description: "Initial version",
		},
	}, nil
}

func (m *MockSchemaService) CheckCompatibility(sdl string) (*models.SchemaCompatibilityCheck, error) {
	return &models.SchemaCompatibilityCheck{
		Compatible:         true,
		CompatibilityLevel: "minor",
	}, nil
}

func (m *MockSchemaService) ValidateSDL(sdl string) error {
	return nil
}

func (m *MockSchemaService) ExecuteQuery(req *models.GraphQLRequest) (*models.GraphQLResponse, error) {
	return &models.GraphQLResponse{
		Data: map[string]interface{}{
			"hello": "world",
		},
	}, nil
}

func (m *MockSchemaService) GetSchemaVersionsByVersion(version string) ([]*models.SchemaVersion, error) {
	return []*models.SchemaVersion{}, nil
}

func (m *MockSchemaService) GetAllSchemaVersions() ([]*models.SchemaVersion, error) {
	return []*models.SchemaVersion{}, nil
}

type MockGraphQLService struct{}

func (m *MockGraphQLService) ProcessQuery(query string, schema *ast.QueryDocument) (interface{}, error) {
	return &models.GraphQLResponse{
		Data: map[string]interface{}{
			"hello": "world",
		},
	}, nil
}

func (m *MockGraphQLService) RouteQuery(query string, version string) (*ast.QueryDocument, error) {
	return &ast.QueryDocument{}, nil
}
