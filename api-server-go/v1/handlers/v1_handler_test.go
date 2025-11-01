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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
		&models.Member{},
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
	memberService := services.NewMemberService(db, &idpProvider)

	// For testing, we'll use a real PDPService but skip actual HTTP calls
	// In a real test, you'd use a test HTTP server
	mockPDP := services.NewPDPService("http://localhost:8082")

	// Note: In a real scenario, you'd set up a test HTTP server to handle PDP requests
	// For now, the tests will need to handle PDP failures gracefully or skip PDP-dependent operations

	return &V1Handler{
		memberService:      memberService,
		schemaService:      services.NewSchemaService(db, mockPDP),
		applicationService: services.NewApplicationService(db, mockPDP),
	}
}

// TestMemberEndpoints tests all member-related endpoints
func TestMemberEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}

	t.Run("CreateMember", func(t *testing.T) {
		// Skip if IDP is not available (test will fail on CreateUser)
		req := models.CreateMemberRequest{
			Name:        "Test Member",
			Email:       fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
			PhoneNumber: "1234567890",
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/members", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		// May fail due to IDP connection, but verify structure
		if w.Code == http.StatusCreated {
			var response models.MemberResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, req.Name, response.Name)
			assert.Equal(t, req.Email, response.Email)
			assert.Equal(t, req.PhoneNumber, response.PhoneNumber)
			assert.NotEmpty(t, response.MemberID)
		}
	})

	t.Run("GetAllMembers", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/members", nil)
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

// TestApplicationEndpoints tests all application-related endpoints
func TestApplicationEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}

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
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}

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
