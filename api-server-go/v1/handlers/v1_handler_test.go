package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/idp/idpfactory"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/gov-dx-sandbox/api-server-go/v1/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockPDPService is a mock implementation of PDPService
type MockPDPService struct {
	mock.Mock
}

func (m *MockPDPService) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPDPService) CreatePolicyMetadata(schemaId string, sdl string) (*models.PolicyMetadataCreateResponse, error) {
	args := m.Called(schemaId, sdl)
	return args.Get(0).(*models.PolicyMetadataCreateResponse), args.Error(1)
}

func (m *MockPDPService) UpdateAllowList(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error) {
	args := m.Called(request)
	return args.Get(0).(*models.AllowListUpdateResponse), args.Error(1)
}

// TestV1Handler tests the V1 API handler
type TestV1Handler struct {
	*testing.T
	db      *gorm.DB
	handler *V1Handler
}

// NewTestV1Handler creates a new test handler with PostgreSQL test database
func NewTestV1Handler(t *testing.T) *TestV1Handler {
	// Use test database configuration
	testDSN := "host=localhost port=5432 user=postgres password=postgres dbname=gov_dx_sandbox_test sslmode=disable"

	db, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Skipf("Skipping test: could not connect to test database: %v", err)
		return nil
	}

	// Auto-migrate the database in the correct order
	err = db.AutoMigrate(
		&models.Entity{},
		&models.Consumer{},
		&models.Provider{},
		&models.Application{},
		&models.ApplicationSubmission{},
		&models.Schema{},
		&models.SchemaSubmission{},
	)
	if err != nil {
		t.Skipf("Skipping test: could not migrate test database: %v", err)
		return nil
	}

	// Create handler with mock PDP service
	handler := NewTestV1HandlerWithMockPDP(t, db)

	return &TestV1Handler{
		T:       t,
		db:      db,
		handler: handler,
	}
}

// NewTestV1HandlerWithMockPDP creates a handler with mock PDP service for testing
func NewTestV1HandlerWithMockPDP(t *testing.T, db *gorm.DB) *V1Handler {
	// Create a test IDP provider (using dummy values for testing)
	idpProvider, err := idpfactory.NewIdpAPIProvider(idpfactory.FactoryConfig{
		ProviderType: idp.ProviderAsgardeo,
		BaseURL:      "http://localhost:9443",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{},
	})
	if err != nil {
		t.Fatalf("Failed to create test IDP provider: %v", err)
	}
	entityService := services.NewEntityService(db, &idpProvider)
	mockPDP := &MockPDPService{}

	// Set up mock expectations for successful operations
	mockPDP.On("UpdateAllowList", mock.AnythingOfType("models.AllowListUpdateRequest")).Return(
		&models.AllowListUpdateResponse{
			Records: []models.AllowListUpdateResponseRecord{
				{
					FieldName: "person.fullName",
					SchemaID:  "test-schema-1",
					ExpiresAt: "2024-12-31T23:59:59Z",
					UpdatedAt: "2024-01-01T00:00:00Z",
				},
			},
		}, nil)

	mockPDP.On("CreatePolicyMetadata", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		&models.PolicyMetadataCreateResponse{
			Records: []models.PolicyMetadataResponse{
				{
					ID:                "policy-1",
					SchemaID:          "test-schema-1",
					FieldName:         "person.fullName",
					DisplayName:       stringPtr("Full Name"),
					Description:       stringPtr("Person's full name"),
					Source:            models.SourcePrimary,
					IsOwner:           true,
					AccessControlType: models.AccessControlTypeRestricted,
					AllowList:         models.AllowList{},
					CreatedAt:         "2024-01-01T00:00:00Z",
					UpdatedAt:         "2024-01-01T00:00:00Z",
				},
			},
		}, nil)

	return &V1Handler{
		entityService:      entityService,
		providerService:    services.NewProviderService(db, entityService),
		consumerService:    services.NewConsumerService(db, entityService),
		schemaService:      services.NewSchemaService(db, mockPDP),
		applicationService: services.NewApplicationService(db, mockPDP),
	}
}

// TestConsumerEndpoints tests all consumer-related endpoints
func TestConsumerEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)

	t.Run("CreateConsumer", func(t *testing.T) {
		req := models.CreateConsumerRequest{
			Name:        "Test Consumer",
			Email:       fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
			PhoneNumber: "1234567890",
			IdpUserID:   "test-user-123",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/consumers", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.ConsumerResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Email, response.Email)
		assert.Equal(t, req.PhoneNumber, response.PhoneNumber)
		assert.Equal(t, req.IdpUserID, response.IdpUserID)
		assert.NotEmpty(t, response.ConsumerID)
		assert.NotEmpty(t, response.EntityID)
	})

	t.Run("GetConsumer", func(t *testing.T) {
		// First create a consumer
		createReq := models.CreateConsumerRequest{
			Name:        "Test Consumer",
			Email:       fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
			PhoneNumber: "1234567890",
			IdpUserID:   "test-user-123",
		}

		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/consumers", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")

		createW := httptest.NewRecorder()
		testHandler.handler.handleConsumers(createW, createHttpReq)

		var createResponse models.ConsumerResponse
		json.Unmarshal(createW.Body.Bytes(), &createResponse)

		// Now get the consumer
		getHttpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/consumers/%s", createResponse.ConsumerID), nil)
		getW := httptest.NewRecorder()
		testHandler.handler.handleConsumers(getW, getHttpReq)

		assert.Equal(t, http.StatusOK, getW.Code)

		var getResponse models.ConsumerResponse
		err := json.Unmarshal(getW.Body.Bytes(), &getResponse)
		assert.NoError(t, err)
		assert.Equal(t, createResponse.ConsumerID, getResponse.ConsumerID)
		assert.Equal(t, createResponse.Name, getResponse.Name)
	})

	t.Run("UpdateConsumer", func(t *testing.T) {
		// First create a consumer
		createReq := models.CreateConsumerRequest{
			Name:        "Test Consumer",
			Email:       fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
			PhoneNumber: "1234567890",
			IdpUserID:   "test-user-123",
		}

		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/consumers", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")

		createW := httptest.NewRecorder()
		testHandler.handler.handleConsumers(createW, createHttpReq)

		var createResponse models.ConsumerResponse
		json.Unmarshal(createW.Body.Bytes(), &createResponse)

		// Now update the consumer
		updatedEmail := fmt.Sprintf("updated-%d@example.com", time.Now().UnixNano())
		updateReq := models.UpdateConsumerRequest{
			Name:        stringPtr("Updated Consumer"),
			Email:       stringPtr(updatedEmail),
			PhoneNumber: stringPtr("9876543210"),
		}

		updateBody, _ := json.Marshal(updateReq)
		updateHttpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/consumers/%s", createResponse.ConsumerID), bytes.NewBuffer(updateBody))
		updateHttpReq.Header.Set("Content-Type", "application/json")

		updateW := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(updateW, updateHttpReq)

		assert.Equal(t, http.StatusOK, updateW.Code)

		var updateResponse models.ConsumerResponse
		err := json.Unmarshal(updateW.Body.Bytes(), &updateResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Consumer", updateResponse.Name)
		assert.Equal(t, updatedEmail, updateResponse.Email)
		assert.Equal(t, "9876543210", updateResponse.PhoneNumber)
	})

	t.Run("GetAllConsumers", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/consumers", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.CollectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Items)
		assert.GreaterOrEqual(t, response.Count, 1)
	})
}

// TestProviderEndpoints tests all provider-related endpoints
func TestProviderEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)

	t.Run("CreateProvider", func(t *testing.T) {
		req := models.CreateProviderRequest{
			Name:        "Test Provider",
			Email:       fmt.Sprintf("provider-%d@example.com", time.Now().UnixNano()),
			PhoneNumber: "1234567890",
			IdpUserID:   "provider-user-123",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.ProviderResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Email, response.Email)
		assert.Equal(t, req.PhoneNumber, response.PhoneNumber)
		assert.Equal(t, req.IdpUserID, response.IdpUserID)
		assert.NotEmpty(t, response.ProviderID)
		assert.NotEmpty(t, response.EntityID)
	})

	t.Run("GetProvider", func(t *testing.T) {
		// First create a provider
		createReq := models.CreateProviderRequest{
			Name:        "Test Provider",
			Email:       fmt.Sprintf("provider-%d@example.com", time.Now().UnixNano()),
			PhoneNumber: "1234567890",
			IdpUserID:   "provider-user-123",
		}

		createBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewBuffer(createBody))
		createHttpReq.Header.Set("Content-Type", "application/json")

		createW := httptest.NewRecorder()
		testHandler.handler.handleProviders(createW, createHttpReq)

		var createResponse models.ProviderResponse
		json.Unmarshal(createW.Body.Bytes(), &createResponse)

		// Now get the provider
		getHttpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/providers/%s", createResponse.ProviderID), nil)
		getW := httptest.NewRecorder()
		testHandler.handler.handleProviders(getW, getHttpReq)

		assert.Equal(t, http.StatusOK, getW.Code)

		var getResponse models.ProviderResponse
		err := json.Unmarshal(getW.Body.Bytes(), &getResponse)
		assert.NoError(t, err)
		assert.Equal(t, createResponse.ProviderID, getResponse.ProviderID)
		assert.Equal(t, createResponse.Name, getResponse.Name)
	})

	t.Run("GetAllProviders", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/providers", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.CollectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Items)
		assert.GreaterOrEqual(t, response.Count, 1)
	})
}

// TestApplicationEndpoints tests all application-related endpoints
func TestApplicationEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)

	// First create a consumer for the application
	consumerReq := models.CreateConsumerRequest{
		Name:        "Test Consumer",
		Email:       fmt.Sprintf("test-consumer-%d@example.com", time.Now().UnixNano()),
		PhoneNumber: "1234567890",
		IdpUserID:   "test-user-123",
	}

	consumerBody, _ := json.Marshal(consumerReq)
	consumerHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/consumers", bytes.NewBuffer(consumerBody))
	consumerHttpReq.Header.Set("Content-Type", "application/json")

	consumerW := httptest.NewRecorder()
	testHandler.handler.handleConsumers(consumerW, consumerHttpReq)

	var consumerResponse models.ConsumerResponse
	json.Unmarshal(consumerW.Body.Bytes(), &consumerResponse)

	t.Run("CreateApplication", func(t *testing.T) {
		req := models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: stringPtr("Test Description"),
			SelectedFields: []models.SelectedFieldRecord{
				{
					FieldName: "person.fullName",
					SchemaID:  "test-schema-1",
				},
			},
			ConsumerID: consumerResponse.ConsumerID,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.ApplicationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, req.ApplicationName, response.ApplicationName)
		assert.Equal(t, req.ApplicationDescription, response.ApplicationDescription)
		assert.Equal(t, req.ConsumerID, response.ConsumerID)
		assert.NotEmpty(t, response.ApplicationID)
		assert.Len(t, response.SelectedFields, 1)
	})

	t.Run("GetApplications", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/applications", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.CollectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Items)
		assert.GreaterOrEqual(t, response.Count, 0)
	})
}

// TestSchemaEndpoints tests all schema-related endpoints
func TestSchemaEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)

	// First create a provider for the schema
	providerReq := models.CreateProviderRequest{
		Name:        "Test Provider",
		Email:       fmt.Sprintf("test-provider-%d@example.com", time.Now().UnixNano()),
		PhoneNumber: "1234567890",
		IdpUserID:   "provider-user-123",
	}

	providerBody, _ := json.Marshal(providerReq)
	providerHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/providers", bytes.NewBuffer(providerBody))
	providerHttpReq.Header.Set("Content-Type", "application/json")

	providerW := httptest.NewRecorder()
	testHandler.handler.handleProviders(providerW, providerHttpReq)

	var providerResponse models.ProviderResponse
	json.Unmarshal(providerW.Body.Bytes(), &providerResponse)

	t.Run("CreateSchema", func(t *testing.T) {
		req := models.CreateSchemaRequest{
			SchemaName:        "Test Schema",
			SchemaDescription: stringPtr("Test Schema Description"),
			SDL:               "type Person { fullName: String }",
			Endpoint:          "http://example.com/graphql",
			ProviderID:        providerResponse.ProviderID,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.SchemaResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, req.SchemaName, response.SchemaName)
		assert.Equal(t, req.SchemaDescription, response.SchemaDescription)
		assert.Equal(t, req.SDL, response.SDL)
		assert.Equal(t, req.Endpoint, response.Endpoint)
		assert.Equal(t, req.ProviderID, response.ProviderID)
		assert.NotEmpty(t, response.SchemaID)
	})

	t.Run("GetSchemas", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schemas", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.CollectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Items)
		assert.GreaterOrEqual(t, response.Count, 0)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
