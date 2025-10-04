package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/models"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/services"
)

// MockSchemaService is a mock implementation of SchemaService for testing
type MockSchemaService struct {
	mock.Mock
}

func (m *MockSchemaService) CreateSchema(req *models.CreateSchemaRequest) (*models.UnifiedSchema, error) {
	args := m.Called(req)
	return args.Get(0).(*models.UnifiedSchema), args.Error(1)
}

func (m *MockSchemaService) UpdateSchemaStatus(version string, isActive bool, reason *string) error {
	args := m.Called(version, isActive, reason)
	return args.Error(0)
}

func (m *MockSchemaService) GetAllSchemaVersions(status *models.SchemaStatus, limit, offset int) ([]*models.UnifiedSchema, int, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]*models.UnifiedSchema), args.Get(1).(int), args.Error(2)
}

func (m *MockSchemaService) GetSchemaVersion(version string) (*models.UnifiedSchema, error) {
	args := m.Called(version)
	return args.Get(0).(*models.UnifiedSchema), args.Error(1)
}

func (m *MockSchemaService) ActivateVersion(version string) error {
	args := m.Called(version)
	return args.Error(0)
}

func TestSchemaHandler_CreateSchema(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockSchemaService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Successfully create schema version",
			requestBody: models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "type Query { hello: String, world: String }",
				CreatedBy: "admin-123",
				Notes:     "Added world field",
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("CreateSchemaVersion", mock.AnythingOfType("*models.CreateSchemaRequest")).Return(
					&models.CreateSchemaResponse{
						Success:   true,
						Message:   "Schema version created successfully",
						Version:   "1.1.0",
						IsActive:  false,
						CreatedAt: "2025-10-01T12:00:00Z",
						NextSteps: "Use PUT /sdl/versions/1.1.0/status to activate this version",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Invalid request body",
			requestBody: map[string]interface{}{
				"version": "1.1.0",
				// Missing required fields
			},
			setupMocks:     func(mockService *MockSchemaService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Service error",
			requestBody: models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "type Query { hello: String }",
				CreatedBy: "admin-123",
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("CreateSchemaVersion", mock.AnythingOfType("*models.CreateSchemaRequest")).Return(
					(*models.CreateSchemaResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name: "Invalid SDL syntax",
			requestBody: models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "invalid graphql syntax",
				CreatedBy: "admin-123",
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("CreateSchemaVersion", mock.AnythingOfType("*models.CreateSchemaRequest")).Return(
					(*models.CreateSchemaResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockSchemaService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.SchemaHandler{
				SchemaService: mockService,
			}

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/sdl", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler.CreateSchemaVersion(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response models.CreateSchemaResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestSchemaHandler_UpdateSchemaStatus(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		requestBody    interface{}
		setupMocks     func(*MockSchemaService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "Successfully activate schema",
			version: "1.1.0",
			requestBody: models.UpdateSchemaStatusRequest{
				IsActive: true,
				Reason:   "Testing completed successfully",
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("UpdateSchemaStatus", "1.1.0", mock.AnythingOfType("*models.UpdateSchemaStatusRequest")).Return(
					&models.UpdateSchemaStatusResponse{
						Success:   true,
						Message:   "Schema version 1.1.0 activated successfully",
						Version:   "1.1.0",
						IsActive:  true,
						UpdatedAt: "2025-10-01T12:00:00Z",
						Reloaded:  true,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "Successfully deactivate schema",
			version: "1.1.0",
			requestBody: models.UpdateSchemaStatusRequest{
				IsActive: false,
				Reason:   "Rollback due to issues",
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("UpdateSchemaStatus", "1.1.0", mock.AnythingOfType("*models.UpdateSchemaStatusRequest")).Return(
					&models.UpdateSchemaStatusResponse{
						Success:   true,
						Message:   "Schema version 1.1.0 deactivated successfully",
						Version:   "1.1.0",
						IsActive:  false,
						UpdatedAt: "2025-10-01T12:00:00Z",
						Reloaded:  true,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "Invalid request body",
			version: "1.1.0",
			requestBody: map[string]interface{}{
				// Missing is_active field
				"reason": "Test reason",
			},
			setupMocks:     func(mockService *MockSchemaService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:    "Service error",
			version: "1.1.0",
			requestBody: models.UpdateSchemaStatusRequest{
				IsActive: true,
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("UpdateSchemaStatus", "1.1.0", mock.AnythingOfType("*models.UpdateSchemaStatusRequest")).Return(
					(*models.UpdateSchemaStatusResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:    "Version not found",
			version: "2.0.0",
			requestBody: models.UpdateSchemaStatusRequest{
				IsActive: true,
			},
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("UpdateSchemaStatus", "2.0.0", mock.AnythingOfType("*models.UpdateSchemaStatusRequest")).Return(
					(*models.UpdateSchemaStatusResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockSchemaService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.SchemaHandler{
				SchemaService: mockService,
			}

			// Create request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/sdl/versions/"+tt.version+"/status", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler.UpdateSchemaStatus(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var response models.UpdateSchemaStatusResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestSchemaHandler_ListVersions(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockSchemaService)
		expectedStatus int
		expectedError  bool
		expectedCount  int
	}{
		{
			name: "Successfully list versions",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("GetAllVersions").Return([]models.SchemaVersion{
					{
						ID:         1,
						Version:    "1.0.0",
						SDL:        "type Query { hello: String }",
						CreatedAt:  time.Now(),
						CreatedBy:  "admin-123",
						Status:     "active",
						ChangeType: "major",
						Notes:      "Initial version",
					},
					{
						ID:         2,
						Version:    "1.1.0",
						SDL:        "type Query { hello: String, world: String }",
						CreatedAt:  time.Now(),
						CreatedBy:  "admin-123",
						Status:     "inactive",
						ChangeType: "minor",
						Notes:      "Added world field",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedCount:  2,
		},
		{
			name: "No versions found",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("GetAllVersions").Return([]models.SchemaVersion{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			expectedCount:  0,
		},
		{
			name: "Service error",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("GetAllVersions").Return(([]models.SchemaVersion)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockSchemaService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.SchemaHandler{
				SchemaService: mockService,
			}

			// Create request
			req := httptest.NewRequest("GET", "/sdl/versions", nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler.ListVersions(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var versions []models.SchemaVersion
				err := json.Unmarshal(rr.Body.Bytes(), &versions)
				assert.NoError(t, err)
				assert.Len(t, versions, tt.expectedCount)
			}

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestSchemaHandler_GetVersion(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		setupMocks     func(*MockSchemaService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "Successfully get version",
			version: "1.1.0",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("GetSchemaVersion", "1.1.0").Return(&models.SchemaVersion{
					ID:         2,
					Version:    "1.1.0",
					SDL:        "type Query { hello: String, world: String }",
					CreatedAt:  time.Now(),
					CreatedBy:  "admin-123",
					Status:     "inactive",
					ChangeType: "minor",
					Notes:      "Added world field",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "Version not found",
			version: "2.0.0",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("GetSchemaVersion", "2.0.0").Return((*models.SchemaVersion)(nil), assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockSchemaService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.SchemaHandler{
				SchemaService: mockService,
			}

			// Create request
			req := httptest.NewRequest("GET", "/sdl/versions/"+tt.version, nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler.GetVersion(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectedError {
				var version models.SchemaVersion
				err := json.Unmarshal(rr.Body.Bytes(), &version)
				assert.NoError(t, err)
				assert.Equal(t, tt.version, version.Version)
			}

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestSchemaHandler_ActivateVersion(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		setupMocks     func(*MockSchemaService)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "Successfully activate version",
			version: "1.1.0",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("ActivateVersion", "1.1.0").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "Activation failed",
			version: "2.0.0",
			setupMocks: func(mockService *MockSchemaService) {
				mockService.On("ActivateVersion", "2.0.0").Return(assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockSchemaService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			handler := &services.SchemaHandler{
				SchemaService: mockService,
			}

			// Create request
			req := httptest.NewRequest("POST", "/sdl/versions/"+tt.version+"/activate", nil)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute handler
			handler.ActivateVersion(rr, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify all expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestSchemaHandler_ValidateCreateRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.CreateSchemaRequest
		expectedError bool
	}{
		{
			name: "Valid request",
			request: &models.CreateSchemaRequest{
				Version:   "1.1.0",
				SDL:       "type Query { hello: String }",
				CreatedBy: "admin-123",
				Notes:     "Test notes",
			},
			expectedError: false,
		},
		{
			name: "Missing version",
			request: &models.CreateSchemaRequest{
				SDL:       "type Query { hello: String }",
				CreatedBy: "admin-123",
			},
			expectedError: true,
		},
		{
			name: "Missing SDL",
			request: &models.CreateSchemaRequest{
				Version:   "1.1.0",
				CreatedBy: "admin-123",
			},
			expectedError: true,
		},
		{
			name: "Missing created by",
			request: &models.CreateSchemaRequest{
				Version: "1.1.0",
				SDL:     "type Query { hello: String }",
			},
			expectedError: true,
		},
		{
			name: "Invalid version format",
			request: &models.CreateSchemaRequest{
				Version:   "invalid",
				SDL:       "type Query { hello: String }",
				CreatedBy: "admin-123",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := &services.SchemaHandler{}

			// Execute validation
			err := handler.ValidateCreateRequest(tt.request)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchemaHandler_CheckVersionCompatibility(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		expectedError bool
	}{
		{
			name:          "Valid semantic version",
			version:       "1.2.3",
			expectedError: false,
		},
		{
			name:          "Invalid version format",
			version:       "invalid",
			expectedError: true,
		},
		{
			name:          "Empty version",
			version:       "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := &services.SchemaHandler{}

			// Execute validation
			err := handler.CheckVersionCompatibility(tt.version)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
