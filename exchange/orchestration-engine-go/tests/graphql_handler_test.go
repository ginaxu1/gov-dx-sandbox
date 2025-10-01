package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

// MockGraphQLService is a mock implementation of GraphQLService for testing
type MockGraphQLService struct {
	mock.Mock
}

func (m *MockGraphQLService) ProcessQuery(query string, schema *ast.Document) (interface{}, error) {
	args := m.Called(query, schema)
	return args.Get(0), args.Error(1)
}

func (m *MockGraphQLService) RouteQuery(query string, version string) (*ast.Document, error) {
	args := m.Called(query, version)
	return args.Get(0).(*ast.Document), args.Error(1)
}

func TestGraphQLHandler_HandleGraphQL(t *testing.T) {
	// Create test schema
	queryType := &ast.Definition{
		Kind: ast.Object,
		Name: "Query",
		Fields: ast.FieldList{
			&ast.FieldDefinition{
				Name: "hello",
				Type: ast.NamedType("String", nil),
			},
		},
	}

	testSchema := &ast.Document{
		Definitions: ast.DefinitionList{queryType},
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		headers        map[string]string
		queryParams    map[string]string
		setupMocks     func(*MockGraphQLService)
		expectedStatus int
		expectedError  bool
		expectedResult interface{}
	}{
		{
			name: "Successfully process query with X-Schema-Version header",
			requestBody: models.GraphQLRequest{
				Query: "query { hello }",
			},
			headers: map[string]string{
				"X-Schema-Version": "1.1.0",
			},
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("RouteQuery", "query { hello }", "1.1.0").Return(testSchema, nil)
				mockService.On("ProcessQuery", "query { hello }", testSchema).Return(
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "Hello World",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedResult: map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello World",
				},
			},
		},
		{
			name: "Successfully process query with version query parameter",
			requestBody: models.GraphQLRequest{
				Query: "query { hello }",
			},
			queryParams: map[string]string{
				"version": "1.1.0",
			},
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("RouteQuery", "query { hello }", "1.1.0").Return(testSchema, nil)
				mockService.On("ProcessQuery", "query { hello }", testSchema).Return(
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "Hello World",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedResult: map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello World",
				},
			},
		},
		{
			name: "Successfully process query with default schema (no version)",
			requestBody: models.GraphQLRequest{
				Query: "query { hello }",
			},
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("RouteQuery", "query { hello }", "").Return(testSchema, nil)
				mockService.On("ProcessQuery", "query { hello }", testSchema).Return(
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "Hello World",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedResult: map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello World",
				},
			},
		},
		{
			name: "Invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "body",
			},
			setupMocks:     func(mockService *MockGraphQLService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Schema routing error",
			requestBody: models.GraphQLRequest{
				Query: "query { hello }",
			},
			headers: map[string]string{
				"X-Schema-Version": "2.0.0",
			},
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("RouteQuery", "query { hello }", "2.0.0").Return((*ast.Document)(nil), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Query processing error",
			requestBody: models.GraphQLRequest{
				Query: "query { hello }",
			},
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("RouteQuery", "query { hello }", "").Return(testSchema, nil)
				mockService.On("ProcessQuery", "query { hello }", testSchema).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name: "Query with variables",
			requestBody: models.GraphQLRequest{
				Query:     "query GetHello($name: String) { hello(name: $name) }",
				Variables: map[string]interface{}{"name": "World"},
			},
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("RouteQuery", "query GetHello($name: String) { hello(name: $name) }", "").Return(testSchema, nil)
				mockService.On("ProcessQuery", "query GetHello($name: String) { hello(name: $name) }", testSchema).Return(
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "Hello World",
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedResult: map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello World",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockGraphQLService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.GraphQLHandler{
				GraphQLService: mockService,
			}

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Set query parameters
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler.HandleGraphQL(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var result interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)

				// Check response headers
				if tt.headers["X-Schema-Version"] != "" {
					assert.Equal(t, tt.headers["X-Schema-Version"], rr.Header().Get("X-Schema-Version-Used"))
				} else if tt.queryParams["version"] != "" {
					assert.Equal(t, tt.queryParams["version"], rr.Header().Get("X-Schema-Version-Used"))
				}
			}

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGraphQLHandler_ExtractVersionFromRequest(t *testing.T) {
	tests := []struct {
		name            string
		headers         map[string]string
		queryParams     map[string]string
		expectedVersion string
	}{
		{
			name: "Version from X-Schema-Version header",
			headers: map[string]string{
				"X-Schema-Version": "1.1.0",
			},
			expectedVersion: "1.1.0",
		},
		{
			name: "Version from query parameter",
			queryParams: map[string]string{
				"version": "1.2.0",
			},
			expectedVersion: "1.2.0",
		},
		{
			name: "Header takes precedence over query parameter",
			headers: map[string]string{
				"X-Schema-Version": "1.1.0",
			},
			queryParams: map[string]string{
				"version": "1.2.0",
			},
			expectedVersion: "1.1.0",
		},
		{
			name:            "No version specified",
			expectedVersion: "",
		},
		{
			name: "Empty header and parameter",
			headers: map[string]string{
				"X-Schema-Version": "",
			},
			queryParams: map[string]string{
				"version": "",
			},
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("POST", "/graphql", nil)

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Set query parameters
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, value := range tt.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			// Create handler
			handler := &services.GraphQLHandler{}

			// Execute test
			version := handler.ExtractVersionFromRequest(req)

			// Assertions
			assert.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestGraphQLHandler_ValidateGraphQLRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.GraphQLRequest
		expectedError bool
	}{
		{
			name: "Valid request with query",
			request: &models.GraphQLRequest{
				Query: "query { hello }",
			},
			expectedError: false,
		},
		{
			name: "Valid request with query and variables",
			request: &models.GraphQLRequest{
				Query:     "query GetHello($name: String) { hello(name: $name) }",
				Variables: map[string]interface{}{"name": "World"},
			},
			expectedError: false,
		},
		{
			name: "Missing query",
			request: &models.GraphQLRequest{
				Variables: map[string]interface{}{"name": "World"},
			},
			expectedError: true,
		},
		{
			name: "Empty query",
			request: &models.GraphQLRequest{
				Query: "",
			},
			expectedError: true,
		},
		{
			name: "Invalid variables type",
			request: &models.GraphQLRequest{
				Query:     "query { hello }",
				Variables: "invalid", // Should be map[string]interface{}
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := &services.GraphQLHandler{}

			// Execute validation
			err := handler.ValidateGraphQLRequest(tt.request)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGraphQLHandler_ProcessGraphQLQuery(t *testing.T) {
	// Create test schema
	queryType := &ast.Definition{
		Kind: ast.Object,
		Name: "Query",
		Fields: ast.FieldList{
			&ast.FieldDefinition{
				Name: "hello",
				Type: ast.NamedType("String", nil),
			},
		},
	}

	testSchema := &ast.Document{
		Definitions: ast.DefinitionList{queryType},
	}

	tests := []struct {
		name           string
		query          string
		schema         *ast.Document
		setupMocks     func(*MockGraphQLService)
		expectedError  bool
		expectedResult interface{}
	}{
		{
			name:   "Successfully process query",
			query:  "query { hello }",
			schema: testSchema,
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("ProcessQuery", "query { hello }", testSchema).Return(
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "Hello World",
						},
					}, nil)
			},
			expectedError: false,
			expectedResult: map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello World",
				},
			},
		},
		{
			name:   "Query processing error",
			query:  "query { hello }",
			schema: testSchema,
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("ProcessQuery", "query { hello }", testSchema).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
		{
			name:   "Complex query with multiple fields",
			query:  "query { hello, world }",
			schema: testSchema,
			setupMocks: func(mockService *MockGraphQLService) {
				mockService.On("ProcessQuery", "query { hello, world }", testSchema).Return(
					map[string]interface{}{
						"data": map[string]interface{}{
							"hello": "Hello",
							"world": "World",
						},
					}, nil)
			},
			expectedError: false,
			expectedResult: map[string]interface{}{
				"data": map[string]interface{}{
					"hello": "Hello",
					"world": "World",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockGraphQLService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.GraphQLHandler{
				GraphQLService: mockService,
			}

			// Execute test
			result, err := handler.ProcessGraphQLQuery(tt.query, tt.schema)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestGraphQLHandler_SetResponseHeaders(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		expectedHeader string
	}{
		{
			name:           "Set X-Schema-Version-Used header",
			version:        "1.1.0",
			expectedHeader: "1.1.0",
		},
		{
			name:           "Empty version",
			version:        "",
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create response recorder
			rr := httptest.NewRecorder()

			// Create handler
			handler := &services.GraphQLHandler{}

			// Execute test
			handler.SetResponseHeaders(rr, tt.version)

			// Assertions
			assert.Equal(t, tt.expectedHeader, rr.Header().Get("X-Schema-Version-Used"))
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		})
	}
}
