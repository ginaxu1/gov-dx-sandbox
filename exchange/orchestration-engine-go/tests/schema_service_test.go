package tests

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

// MockDB is a mock implementation of sql.DB for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	args2 := m.Called(query, args)
	return args2.Get(0).(*sql.Rows), args2.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	args2 := m.Called(query, args)
	return args2.Get(0).(*sql.Row)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	args2 := m.Called(query, args)
	return args2.Get(0).(sql.Result), args2.Error(1)
}

func (m *MockDB) Begin() (*sql.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sql.Tx), args.Error(1)
}

// MockTx is a mock implementation of sql.Tx for testing
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	args2 := m.Called(query, args)
	return args2.Get(0).(*sql.Rows), args2.Error(1)
}

func (m *MockTx) QueryRow(query string, args ...interface{}) *sql.Row {
	args2 := m.Called(query, args)
	return args2.Get(0).(*sql.Row)
}

func (m *MockTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	args2 := m.Called(query, args)
	return args2.Get(0).(sql.Result), args2.Error(1)
}

func (m *MockTx) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTx) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// MockResult is a mock implementation of sql.Result for testing
type MockResult struct {
	mock.Mock
}

func (m *MockResult) LastInsertId() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResult) RowsAffected() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestSchemaService_CreateSchemaVersion(t *testing.T) {
	tests := []struct {
		name           string
		request        *models.CreateSchemaRequest
		setupMocks     func(*MockDB, *MockTx, *MockResult)
		expectedError  bool
		expectedResult *models.CreateSchemaResponse
	}{
		{
			name: "Successfully create minor version",
			request: &models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "type Query { hello: String, world: String }",
				CreatedBy: "admin-123",
				Notes:     "Added world field",
			},
			setupMocks: func(mockDB *MockDB, mockTx *MockTx, mockResult *MockResult) {
				// Mock transaction begin
				mockDB.On("Begin").Return(mockTx, nil)

				// Mock getPreviousVersionID
				mockTx.On("QueryRow", "SELECT id FROM unified_schemas WHERE status = 'active' ORDER BY created_at DESC LIMIT 1").Return(&sql.Row{})

				// Mock saveSchemaVersion
				mockTx.On("Exec", mock.AnythingOfType("string"), mock.AnythingOfType("[]interface{}")).Return(mockResult, nil)

				// Mock commit
				mockTx.On("Commit").Return(nil)

				// Mock rollback (defer)
				mockTx.On("Rollback").Return(nil)
			},
			expectedError: false,
			expectedResult: &models.CreateSchemaResponse{
				Success:   true,
				Message:   "Schema version created successfully",
				Version:   "1.1.0",
				IsActive:  false,
				NextSteps: "Use PUT /sdl/versions/1.1.0/status to activate this version",
			},
		},
		{
			name: "Invalid SDL syntax",
			request: &models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "invalid graphql syntax",
				CreatedBy: "admin-123",
			},
			setupMocks: func(mockDB *MockDB, mockTx *MockTx, mockResult *MockResult) {
				// Mock transaction begin
				mockDB.On("Begin").Return(mockTx, nil)
				// Mock rollback (defer)
				mockTx.On("Rollback").Return(nil)
			},
			expectedError: true,
		},
		{
			name: "Database transaction failure",
			request: &models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "type Query { hello: String }",
				CreatedBy: "admin-123",
			},
			setupMocks: func(mockDB *MockDB, mockTx *MockTx, mockResult *MockResult) {
				// Mock transaction begin failure
				mockDB.On("Begin").Return((*sql.Tx)(nil), assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDB := new(MockDB)
			mockTx := new(MockTx)
			mockResult := new(MockResult)

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB, mockTx, mockResult)
			}

			// Create service with mock DB
			service := &services.SchemaService{
				DB: mockDB,
			}

			// Execute test
			result, err := service.CreateSchemaVersion(tt.request)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Success, result.Success)
				assert.Equal(t, tt.expectedResult.Version, result.Version)
				assert.Equal(t, tt.expectedResult.IsActive, result.IsActive)
			}

			// Verify all expectations
			mockDB.AssertExpectations(t)
			mockTx.AssertExpectations(t)
			mockResult.AssertExpectations(t)
		})
	}
}

func TestSchemaService_UpdateSchemaStatus(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		request       *models.UpdateSchemaStatusRequest
		setupMocks    func(*MockDB, *MockTx, *MockResult)
		expectedError bool
	}{
		{
			name:    "Successfully activate schema",
			version: "1.1.0",
			request: &models.UpdateSchemaStatusRequest{
				IsActive: true,
				Reason:   "Testing completed",
			},
			setupMocks: func(mockDB *MockDB, mockTx *MockTx, mockResult *MockResult) {
				// Mock transaction begin
				mockDB.On("Begin").Return(mockTx, nil)

				// Mock deactivate all schemas
				mockTx.On("Exec", "UPDATE unified_schemas SET status = 'inactive' WHERE status = 'active'").Return(mockResult, nil)

				// Mock activate specific version
				mockTx.On("Exec", "UPDATE unified_schemas SET status = 'active' WHERE version = $1", "1.1.0").Return(mockResult, nil)

				// Mock commit
				mockTx.On("Commit").Return(nil)

				// Mock rollback (defer)
				mockTx.On("Rollback").Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "Successfully deactivate schema",
			version: "1.1.0",
			request: &models.UpdateSchemaStatusRequest{
				IsActive: false,
				Reason:   "Rollback due to issues",
			},
			setupMocks: func(mockDB *MockDB, mockTx *MockTx, mockResult *MockResult) {
				// Mock transaction begin
				mockDB.On("Begin").Return(mockTx, nil)

				// Mock deactivate all schemas
				mockTx.On("Exec", "UPDATE unified_schemas SET status = 'inactive' WHERE status = 'active'").Return(mockResult, nil)

				// Mock commit
				mockTx.On("Commit").Return(nil)

				// Mock rollback (defer)
				mockTx.On("Rollback").Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "Database transaction failure",
			version: "1.1.0",
			request: &models.UpdateSchemaStatusRequest{
				IsActive: true,
			},
			setupMocks: func(mockDB *MockDB, mockTx *MockTx, mockResult *MockResult) {
				// Mock transaction begin failure
				mockDB.On("Begin").Return((*sql.Tx)(nil), assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDB := new(MockDB)
			mockTx := new(MockTx)
			mockResult := new(MockResult)

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB, mockTx, mockResult)
			}

			// Create service with mock DB
			service := &services.SchemaService{
				DB: mockDB,
			}

			// Execute test
			result, err := service.UpdateSchemaStatus(tt.version, tt.request)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.Success)
				assert.Equal(t, tt.version, result.Version)
				assert.Equal(t, tt.request.IsActive, result.IsActive)
			}

			// Verify all expectations
			mockDB.AssertExpectations(t)
			mockTx.AssertExpectations(t)
			mockResult.AssertExpectations(t)
		})
	}
}

func TestSchemaService_LoadSchema(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockDB)
		expectedError bool
		expectedCount int
	}{
		{
			name: "Successfully load schemas",
			setupMocks: func(mockDB *MockDB) {
				// Mock getAllActiveSchemas query
				mockRows := &sql.Rows{}
				mockDB.On("Query", "SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id FROM unified_schemas WHERE status IN ('active', 'inactive') ORDER BY created_at DESC").Return(mockRows, nil)
			},
			expectedError: false,
			expectedCount: 0, // No schemas in mock
		},
		{
			name: "Database query failure",
			setupMocks: func(mockDB *MockDB) {
				// Mock query failure
				mockDB.On("Query", mock.AnythingOfType("string")).Return((*sql.Rows)(nil), assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDB := new(MockDB)

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB)
			}

			// Create service with mock DB
			service := &services.SchemaService{
				DB:             mockDB,
				SchemaVersions: make(map[string]*ast.Document),
			}

			// Execute test
			err := service.LoadSchema()

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, service.SchemaVersions, tt.expectedCount)
			}

			// Verify all expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestSchemaService_RouteQuery(t *testing.T) {
	// Create test schemas
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
		version        string
		schemaVersions map[string]*ast.Document
		currentSchema  *ast.Document
		expectedError  bool
		expectedSchema *ast.Document
	}{
		{
			name:           "Route to specific version",
			query:          "query { hello }",
			version:        "1.1.0",
			schemaVersions: map[string]*ast.Document{"1.1.0": testSchema},
			currentSchema:  nil,
			expectedError:  false,
			expectedSchema: testSchema,
		},
		{
			name:           "Route to current schema when no version specified",
			query:          "query { hello }",
			version:        "",
			schemaVersions: map[string]*ast.Document{},
			currentSchema:  testSchema,
			expectedError:  false,
			expectedSchema: testSchema,
		},
		{
			name:           "Version not found",
			query:          "query { hello }",
			version:        "2.0.0",
			schemaVersions: map[string]*ast.Document{"1.1.0": testSchema},
			currentSchema:  testSchema,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			service := &services.SchemaService{
				SchemaVersions: tt.schemaVersions,
				CurrentSchema:  tt.currentSchema,
			}

			// Execute test
			schema, err := service.RouteQuery(tt.query, tt.version)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSchema, schema)
			}
		})
	}
}

func TestSchemaService_GetAllActiveSchemas(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockDB)
		expectedError bool
		expectedCount int
	}{
		{
			name: "Successfully retrieve schemas",
			setupMocks: func(mockDB *MockDB) {
				// Mock successful query
				mockRows := &sql.Rows{}
				mockDB.On("Query", "SELECT id, version, sdl, created_at, created_by, status, change_type, notes, previous_version_id FROM unified_schemas WHERE status IN ('active', 'inactive') ORDER BY created_at DESC").Return(mockRows, nil)
			},
			expectedError: false,
			expectedCount: 0, // No rows in mock
		},
		{
			name: "Database query failure",
			setupMocks: func(mockDB *MockDB) {
				// Mock query failure
				mockDB.On("Query", mock.AnythingOfType("string")).Return((*sql.Rows)(nil), assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDB := new(MockDB)

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB)
			}

			// Create service with mock DB
			service := &services.SchemaService{
				DB: mockDB,
			}

			// Execute test
			schemas, err := service.GetAllActiveSchemas()

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, schemas, tt.expectedCount)
			}

			// Verify all expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestSchemaService_GetPreviousVersionID(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		setupMocks    func(*MockTx)
		expectedError bool
		expectedID    *int
	}{
		{
			name:    "Previous version exists",
			version: "1.1.0",
			setupMocks: func(mockTx *MockTx) {
				// Mock successful query with result
				mockRow := &sql.Row{}
				mockTx.On("QueryRow", "SELECT id FROM unified_schemas WHERE status = 'active' ORDER BY created_at DESC LIMIT 1").Return(mockRow)
			},
			expectedError: false,
			expectedID:    nil, // Mock returns nil
		},
		{
			name:    "No previous version",
			version: "1.0.0",
			setupMocks: func(mockTx *MockTx) {
				// Mock no rows found
				mockRow := &sql.Row{}
				mockTx.On("QueryRow", "SELECT id FROM unified_schemas WHERE status = 'active' ORDER BY created_at DESC LIMIT 1").Return(mockRow)
			},
			expectedError: false,
			expectedID:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockTx := new(MockTx)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTx)
			}

			// Create service
			service := &services.SchemaService{}

			// Execute test
			id, err := service.GetPreviousVersionID(mockTx, tt.version)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			// Verify all expectations
			mockTx.AssertExpectations(t)
		})
	}
}
