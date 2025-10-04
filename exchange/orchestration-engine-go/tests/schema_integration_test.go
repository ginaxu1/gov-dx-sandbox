package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

func TestSchemaIntegration_CompleteFlow(t *testing.T) {
	// This test simulates the complete flow from schema creation to GraphQL query execution

	// Step 1: Create initial schema version
	t.Run("Create initial schema version", func(t *testing.T) {
		// Mock database and services
		mockDB := new(MockDB)
		mockTx := new(MockTx)
		mockResult := new(MockResult)

		// Setup mocks for schema creation
		mockDB.On("Begin").Return(mockTx, nil)
		mockTx.On("QueryRow", "SELECT id FROM unified_schemas WHERE status = 'active' ORDER BY created_at DESC LIMIT 1").Return(&sql.Row{})
		mockTx.On("Exec", mock.AnythingOfType("string"), mock.AnythingOfType("[]interface{}")).Return(mockResult, nil)
		mockTx.On("Commit").Return(nil)
		mockTx.On("Rollback").Return(nil)

		// Create schema service
		schemaService := &services.SchemaService{
			DB: mockDB,
		}

		// Create initial schema
		request := &models.CreateSchemaRequest{
			Version:   "1.0.0",
			SDL:       "type Query { hello: String }",
			CreatedBy: "admin-123",
			Notes:     "Initial schema version",
		}

		response, err := schemaService.CreateSchemaVersion(request)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "1.0.0", response.Version)
		assert.False(t, response.IsActive)

		// Verify mocks
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	// Step 2: Activate the schema
	t.Run("Activate schema version", func(t *testing.T) {
		// Mock database and services
		mockDB := new(MockDB)
		mockTx := new(MockTx)
		mockResult := new(MockResult)

		// Setup mocks for schema activation
		mockDB.On("Begin").Return(mockTx, nil)
		mockTx.On("Exec", "UPDATE unified_schemas SET status = 'inactive' WHERE status = 'active'").Return(mockResult, nil)
		mockTx.On("Exec", "UPDATE unified_schemas SET status = 'active' WHERE version = $1", "1.0.0").Return(mockResult, nil)
		mockTx.On("Commit").Return(nil)
		mockTx.On("Rollback").Return(nil)

		// Create schema service
		schemaService := &services.SchemaService{
			DB: mockDB,
		}

		// Activate schema
		request := &models.UpdateSchemaStatusRequest{
			IsActive: true,
			Reason:   "Initial activation",
		}

		response, err := schemaService.UpdateSchemaStatus("1.0.0", request)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "1.0.0", response.Version)
		assert.True(t, response.IsActive)

		// Verify mocks
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	// Step 3: Create minor version update
	t.Run("Create minor version update", func(t *testing.T) {
		// Mock database and services
		mockDB := new(MockDB)
		mockTx := new(MockTx)
		mockResult := new(MockResult)

		// Setup mocks for minor version creation
		mockDB.On("Begin").Return(mockTx, nil)
		mockTx.On("QueryRow", "SELECT id FROM unified_schemas WHERE status = 'active' ORDER BY created_at DESC LIMIT 1").Return(&sql.Row{})
		mockTx.On("Exec", mock.AnythingOfType("string"), mock.AnythingOfType("[]interface{}")).Return(mockResult, nil)
		mockTx.On("Commit").Return(nil)
		mockTx.On("Rollback").Return(nil)

		// Create schema service
		schemaService := &services.SchemaService{
			DB: mockDB,
		}

		// Create minor version
		request := &models.CreateSchemaRequest{
			Version:   "1.1.0",
			SDL:       "type Query { hello: String, world: String }",
			CreatedBy: "admin-123",
			Notes:     "Added world field",
		}

		response, err := schemaService.CreateSchemaVersion(request)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "1.1.0", response.Version)
		assert.False(t, response.IsActive)

		// Verify mocks
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	// Step 4: Activate minor version
	t.Run("Activate minor version", func(t *testing.T) {
		// Mock database and services
		mockDB := new(MockDB)
		mockTx := new(MockTx)
		mockResult := new(MockResult)

		// Setup mocks for minor version activation
		mockDB.On("Begin").Return(mockTx, nil)
		mockTx.On("Exec", "UPDATE unified_schemas SET status = 'inactive' WHERE status = 'active'").Return(mockResult, nil)
		mockTx.On("Exec", "UPDATE unified_schemas SET status = 'active' WHERE version = $1", "1.1.0").Return(mockResult, nil)
		mockTx.On("Commit").Return(nil)
		mockTx.On("Rollback").Return(nil)

		// Create schema service
		schemaService := &services.SchemaService{
			DB: mockDB,
		}

		// Activate minor version
		request := &models.UpdateSchemaStatusRequest{
			IsActive: true,
			Reason:   "Minor version activation",
		}

		response, err := schemaService.UpdateSchemaStatus("1.1.0", request)

		// Assertions
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "1.1.0", response.Version)
		assert.True(t, response.IsActive)

		// Verify mocks
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
		mockResult.AssertExpectations(t)
	})

	// Step 5: Test GraphQL query with version routing
	t.Run("GraphQL query with version routing", func(t *testing.T) {
		// Create test schemas
		queryTypeV1 := &ast.Definition{
			Kind: ast.Object,
			Name: "Query",
			Fields: ast.FieldList{
				&ast.FieldDefinition{
					Name: "hello",
					Type: ast.NamedType("String", nil),
				},
			},
		}

		queryTypeV2 := &ast.Definition{
			Kind: ast.Object,
			Name: "Query",
			Fields: ast.FieldList{
				&ast.FieldDefinition{
					Name: "hello",
					Type: ast.NamedType("String", nil),
				},
				&ast.FieldDefinition{
					Name: "world",
					Type: ast.NamedType("String", nil),
				},
			},
		}

		schemaV1 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV1}}
		schemaV2 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV2}}

		// Mock GraphQL service
		mockGraphQLService := new(MockGraphQLService)
		mockGraphQLService.On("RouteQuery", "query { hello }", "1.0.0").Return(schemaV1, nil)
		mockGraphQLService.On("RouteQuery", "query { hello, world }", "1.1.0").Return(schemaV2, nil)
		mockGraphQLService.On("ProcessQuery", "query { hello }", schemaV1).Return(
			map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello",
				},
			}, nil)
		mockGraphQLService.On("ProcessQuery", "query { hello, world }", schemaV2).Return(
			map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello",
					"world": "World",
				},
			}, nil)

		// Create GraphQL handler
		handler := &services.GraphQLHandler{
			GraphQLService: mockGraphQLService,
		}

		// Test query with version 1.0.0
		t.Run("Query with version 1.0.0", func(t *testing.T) {
			request := models.GraphQLRequest{
				Query: "query { hello }",
			}

			jsonBody, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Schema-Version", "1.0.0")

			rr := httptest.NewRecorder()
			handler.HandleGraphQL(rr, req)

			// Assertions
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "1.0.0", rr.Header().Get("X-Schema-Version-Used"))

			var result map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, "Hello", result["data"].(map[string]interface{})["hello"])
		})

		// Test query with version 1.1.0
		t.Run("Query with version 1.1.0", func(t *testing.T) {
			request := models.GraphQLRequest{
				Query: "query { hello, world }",
			}

			jsonBody, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Schema-Version", "1.1.0")

			rr := httptest.NewRecorder()
			handler.HandleGraphQL(rr, req)

			// Assertions
			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "1.1.0", rr.Header().Get("X-Schema-Version-Used"))

			var result map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, "Hello", result["data"].(map[string]interface{})["hello"])
			assert.Equal(t, "World", result["data"].(map[string]interface{})["world"])
		})

		// Verify all expectations
		mockGraphQLService.AssertExpectations(t)
	})
}

func TestSchemaIntegration_ContractTesting(t *testing.T) {
	// This test simulates contract testing for schema compatibility

	// Create test schemas
	queryTypeV1 := &ast.Definition{
		Kind: ast.Object,
		Name: "Query",
		Fields: ast.FieldList{
			&ast.FieldDefinition{
				Name: "hello",
				Type: ast.NamedType("String", nil),
			},
		},
	}

	queryTypeV2 := &ast.Definition{
		Kind: ast.Object,
		Name: "Query",
		Fields: ast.FieldList{
			&ast.FieldDefinition{
				Name: "hello",
				Type: ast.NamedType("String", nil),
			},
			&ast.FieldDefinition{
				Name: "world",
				Type: ast.NamedType("String", nil),
			},
		},
	}

	schemaV1 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV1}}
	schemaV2 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV2}}

	// Create contract tests
	contractTests := []models.ContractTest{
		{
			Name:        "Hello query test",
			Query:       "query { hello }",
			Variables:   map[string]interface{}{},
			Expected:    map[string]interface{}{"data": map[string]interface{}{"hello": "Hello World"}},
			Description: "Test basic hello query",
		},
		{
			Name:        "World query test",
			Query:       "query { world }",
			Variables:   map[string]interface{}{},
			Expected:    map[string]interface{}{"data": map[string]interface{}{"world": "Hello World"}},
			Description: "Test world query",
		},
	}

	// Mock database
	mockDB := new(MockDB)
	mockRows := &sql.Rows{}
	mockDB.On("Query", mock.AnythingOfType("string")).Return(mockRows, nil)

	// Create contract test suite
	suite := &services.ContractTestSuite{
		DB: mockDB,
	}

	// Mock LoadContractTests
	suite.LoadContractTests = func() ([]models.ContractTest, error) {
		return contractTests, nil
	}

	// Test contract tests against schema V1
	t.Run("Contract tests against schema V1", func(t *testing.T) {
		results, err := suite.ExecuteContractTests(schemaV1)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, 2, results.TotalTests)
		assert.Equal(t, 1, results.Passed) // Only hello query should pass
		assert.Equal(t, 1, results.Failed) // world query should fail
	})

	// Test contract tests against schema V2
	t.Run("Contract tests against schema V2", func(t *testing.T) {
		results, err := suite.ExecuteContractTests(schemaV2)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, 2, results.TotalTests)
		assert.Equal(t, 2, results.Passed) // Both queries should pass
		assert.Equal(t, 0, results.Failed)
	})

	// Verify mocks
	mockDB.AssertExpectations(t)
}

func TestSchemaIntegration_VersionCompatibility(t *testing.T) {
	// This test simulates version compatibility checking

	// Create test schemas
	queryTypeV1 := &ast.Definition{
		Kind: ast.Object,
		Name: "Query",
		Fields: ast.FieldList{
			&ast.FieldDefinition{
				Name: "hello",
				Type: ast.NamedType("String", nil),
			},
		},
	}

	queryTypeV2 := &ast.Definition{
		Kind: ast.Object,
		Name: "Query",
		Fields: ast.FieldList{
			&ast.FieldDefinition{
				Name: "hello",
				Type: ast.NamedType("String", nil),
			},
			&ast.FieldDefinition{
				Name: "world",
				Type: ast.NamedType("String", nil),
			},
		},
	}

	schemaV1 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV1}}
	schemaV2 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV2}}

	// Create schema service
	service := &services.SchemaService{}

	// Mock getCurrentActiveSchema
	service.GetCurrentActiveSchema = func() (string, *ast.Document, error) {
		return "1.0.0", schemaV1, nil
	}

	// Test minor version compatibility
	t.Run("Minor version compatibility", func(t *testing.T) {
		compatibility, err := service.CheckCompatibility("1.1.0", schemaV2)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "minor", compatibility.ChangeType)
		assert.Empty(t, compatibility.BreakingChanges)
		assert.Contains(t, compatibility.NewFields, "New field 'Query.world' added")
	})

	// Test major version compatibility
	t.Run("Major version compatibility", func(t *testing.T) {
		// Create breaking change schema
		queryTypeV3 := &ast.Definition{
			Kind: ast.Object,
			Name: "Query",
			Fields: ast.FieldList{
				&ast.FieldDefinition{
					Name: "greeting",
					Type: ast.NamedType("String", nil),
				},
			},
		}

		schemaV3 := &ast.Document{Definitions: ast.DefinitionList{queryTypeV3}}

		compatibility, err := service.CheckCompatibility("2.0.0", schemaV3)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, "major", compatibility.ChangeType)
		assert.Contains(t, compatibility.BreakingChanges, "Field 'Query.hello' was removed")
		assert.Contains(t, compatibility.NewFields, "New field 'Query.greeting' added")
	})
}

func TestSchemaIntegration_ErrorHandling(t *testing.T) {
	// This test simulates error handling scenarios

	// Test invalid SDL
	t.Run("Invalid SDL handling", func(t *testing.T) {
		// Mock database
		mockDB := new(MockDB)
		mockTx := new(MockTx)

		// Setup mocks
		mockDB.On("Begin").Return(mockTx, nil)
		mockTx.On("Rollback").Return(nil)

		// Create schema service
		service := &services.SchemaService{
			DB: mockDB,
		}

		// Create request with invalid SDL
		request := &models.CreateSchemaRequest{
			Version:   "1.1.0",
			SDL:       "invalid graphql syntax",
			CreatedBy: "admin-123",
		}

		response, err := service.CreateSchemaVersion(request)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)

		// Verify mocks
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	// Test database transaction failure
	t.Run("Database transaction failure", func(t *testing.T) {
		// Mock database
		mockDB := new(MockDB)

		// Setup mocks for transaction failure
		mockDB.On("Begin").Return((*sql.Tx)(nil), assert.AnError)

		// Create schema service
		service := &services.SchemaService{
			DB: mockDB,
		}

		// Create request
		request := &models.CreateSchemaRequest{
			Version:   "1.1.0",
			SDL:       "type Query { hello: String }",
			CreatedBy: "admin-123",
		}

		response, err := service.CreateSchemaVersion(request)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)

		// Verify mocks
		mockDB.AssertExpectations(t)
	})

	// Test version not found
	t.Run("Version not found", func(t *testing.T) {
		// Mock GraphQL service
		mockService := new(MockGraphQLService)
		mockService.On("RouteQuery", "query { hello }", "2.0.0").Return((*ast.Document)(nil), assert.AnError)

		// Create GraphQL handler
		handler := &services.GraphQLHandler{
			GraphQLService: mockService,
		}

		// Create request
		request := models.GraphQLRequest{
			Query: "query { hello }",
		}

		jsonBody, _ := json.Marshal(request)
		req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Schema-Version", "2.0.0")

		rr := httptest.NewRecorder()
		handler.HandleGraphQL(rr, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		// Verify mocks
		mockService.AssertExpectations(t)
	})
}
