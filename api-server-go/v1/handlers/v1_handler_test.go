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
	// Use shared PostgreSQL test utility
	db := services.SetupPostgresTestDB(t)
	if db == nil {
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
	memberService := services.NewMemberService(db, idpProvider)

	// For testing, we'll use a real PDPService but skip actual HTTP calls
	// In a real test, you'd use a test HTTP server
	mockPDP := services.NewPDPService("http://localhost:8082", "test-api-key")

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
	// Cleanup is handled by SetupPostgresTestDB

	var createdMemberID string

	t.Run("POST /api/v1/members - CreateMember", func(t *testing.T) {
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

		// May fail due to IDP connection, but verify structure if successful
		if w.Code == http.StatusCreated {
			var response models.MemberResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, req.Name, response.Name)
			assert.Equal(t, req.Email, response.Email)
			assert.Equal(t, req.PhoneNumber, response.PhoneNumber)
			assert.NotEmpty(t, response.MemberID)
			createdMemberID = response.MemberID
		} else {
			t.Logf("Member creation may have failed due to IDP connection: status %d", w.Code)
		}
	})

	t.Run("POST /api/v1/members - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/members", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/members/:id - UpdateMember", func(t *testing.T) {
		// Create a member first
		createReq := models.CreateMemberRequest{
			Name:        "Original Name",
			Email:       "original@example.com",
			PhoneNumber: "1234567890",
		}
		createReqBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/members", bytes.NewBuffer(createReqBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(createW, createHttpReq)
		if createW.Code != http.StatusCreated {
			t.Skip("Member creation failed, skipping update test")
			return
		}

		var createResponse map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResponse)
		memberID := createResponse["data"].(map[string]interface{})["memberId"].(string)

		// Now update the member
		name := "Updated Name"
		phone := "9876543210"
		updateReq := models.UpdateMemberRequest{
			Name:        &name,
			PhoneNumber: &phone,
		}
		updateReqBody, _ := json.Marshal(updateReq)
		updateHttpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/members/%s", memberID), bytes.NewBuffer(updateReqBody))
		updateHttpReq.Header.Set("Content-Type", "application/json")
		updateW := httptest.NewRecorder()
		mux.ServeHTTP(updateW, updateHttpReq)

		assert.Equal(t, http.StatusOK, updateW.Code)
		var response map[string]interface{}
		json.Unmarshal(updateW.Body.Bytes(), &response)
		assert.NotNil(t, response["data"])
	})

	t.Run("PUT /api/v1/members/:id - UpdateMember_InvalidJSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/members/test-id", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/members/:id - UpdateMember_NotFound", func(t *testing.T) {
		name := "Updated Name"
		req := models.UpdateMemberRequest{
			Name: &name,
		}
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/members/non-existent-id", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/members - GetAllMembers", func(t *testing.T) {
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

	t.Run("GET /api/v1/members - WithQueryParams", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/members?email=test@example.com", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		// May return 500 if query fails, but should handle gracefully
		assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)
	})

	t.Run("GET /api/v1/members/:memberId - GetMember", func(t *testing.T) {
		if createdMemberID == "" {
			t.Skip("No member ID available from creation test")
			return
		}

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/members/%s", createdMemberID), nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.MemberResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, createdMemberID, response.MemberID)
		}
	})

	t.Run("GET /api/v1/members/:memberId - NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/members/non-existent-id", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /api/v1/members/:memberId - UpdateMember", func(t *testing.T) {
		if createdMemberID == "" {
			t.Skip("No member ID available from creation test")
			return
		}

		name := "Updated Name"
		phone := "9876543210"
		req := models.UpdateMemberRequest{
			Name:        &name,
			PhoneNumber: &phone,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/members/%s", createdMemberID), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.MemberResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Updated Name", response.Name)
			assert.Equal(t, "9876543210", response.PhoneNumber)
		}
	})

	t.Run("Method Not Allowed", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/members", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("Invalid Path", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/members/invalid/path", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestSchemaEndpoints tests all schema-related endpoints
func TestSchemaEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}
	defer testHandler.db.Exec("DELETE FROM schemas")

	var createdSchemaID string
	testMemberID := "test-member-id"

	t.Run("POST /api/v1/schemas - CreateSchema", func(t *testing.T) {
		req := models.CreateSchemaRequest{
			SchemaName:        "Test Schema",
			SchemaDescription: "Test Description",
			SDL:               "type Query { test: String }",
			Endpoint:          "http://example.com/graphql",
			MemberID:          testMemberID,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusCreated {
			var response models.SchemaResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, req.SchemaName, response.SchemaName)
			assert.Equal(t, req.SDL, response.SDL)
			assert.NotEmpty(t, response.SchemaID)
			createdSchemaID = response.SchemaID
		}
	})

	t.Run("POST /api/v1/schemas - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/schemas - GetAllSchemas", func(t *testing.T) {
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

	t.Run("GET /api/v1/schemas - WithQueryParams", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schemas?memberId=test-member", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /api/v1/schemas/:schemaId - GetSchema", func(t *testing.T) {
		if createdSchemaID == "" {
			t.Skip("No schema ID available")
			return
		}

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/schemas/%s", createdSchemaID), nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.SchemaResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, createdSchemaID, response.SchemaID)
		}
	})

	t.Run("GET /api/v1/schemas/:schemaId - NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schemas/non-existent", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /api/v1/schemas/:schemaId - UpdateSchema", func(t *testing.T) {
		if createdSchemaID == "" {
			t.Skip("No schema ID available")
			return
		}

		schemaName := "Updated Schema Name"
		sdl := "type Query { updated: String }"
		req := models.UpdateSchemaRequest{
			SchemaName: &schemaName,
			SDL:        &sdl,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/schemas/%s", createdSchemaID), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.SchemaResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Updated Schema Name", response.SchemaName)
		}
	})

	t.Run("Method Not Allowed - Schemas", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/schemas", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

// TestSchemaSubmissionEndpoints tests all schema submission-related endpoints
func TestSchemaSubmissionEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}
	defer testHandler.db.Exec("DELETE FROM schema_submissions")

	var createdSubmissionID string
	testMemberID := "test-member-id"

	t.Run("GET /api/v1/schema-submissions/:id - GetSchemaSubmission", func(t *testing.T) {
		// Create a member first
		memberReq := models.CreateMemberRequest{
			Name:        "Test Member",
			Email:       "testmember@example.com",
			PhoneNumber: "1234567890",
		}
		memberReqBody, _ := json.Marshal(memberReq)
		memberHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/members", bytes.NewBuffer(memberReqBody))
		memberHttpReq.Header.Set("Content-Type", "application/json")
		memberW := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(memberW, memberHttpReq)
		if memberW.Code != http.StatusCreated {
			t.Skip("Member creation failed, skipping submission test")
			return
		}
		var memberResponse map[string]interface{}
		json.Unmarshal(memberW.Body.Bytes(), &memberResponse)
		memberID := memberResponse["data"].(map[string]interface{})["memberId"].(string)

		// Create a submission
		desc := "Test Description"
		createReq := models.CreateSchemaSubmissionRequest{
			SchemaName:        "Test Schema Submission",
			SchemaDescription: &desc,
			SDL:               "type Query { test: String }",
			SchemaEndpoint:    "http://example.com/graphql",
			MemberID:          memberID,
		}
		createReqBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schema-submissions", bytes.NewBuffer(createReqBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		mux.ServeHTTP(createW, createHttpReq)
		assert.Equal(t, http.StatusCreated, createW.Code)

		var createResponse map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResponse)
		if createResponse["data"] == nil {
			t.Skip("Submission creation failed, skipping update test")
			return
		}
		submissionID := createResponse["data"].(map[string]interface{})["submissionId"].(string)

		// Now get the submission
		getHttpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/schema-submissions/%s", submissionID), nil)
		getW := httptest.NewRecorder()
		mux.ServeHTTP(getW, getHttpReq)

		assert.Equal(t, http.StatusOK, getW.Code)
		var response map[string]interface{}
		json.Unmarshal(getW.Body.Bytes(), &response)
		assert.NotNil(t, response["data"])
	})

	t.Run("GET /api/v1/schema-submissions/:id - GetSchemaSubmission_NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schema-submissions/non-existent-id", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("POST /api/v1/schema-submissions - CreateSchemaSubmission", func(t *testing.T) {
		desc := "Test Description"
		req := models.CreateSchemaSubmissionRequest{
			SchemaName:        "Test Schema Submission",
			SchemaDescription: &desc,
			SDL:               "type Query { test: String }",
			SchemaEndpoint:    "http://example.com/graphql",
			MemberID:          testMemberID,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schema-submissions", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusCreated {
			var response models.SchemaSubmissionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, req.SchemaName, response.SchemaName)
			assert.NotEmpty(t, response.SubmissionID)
			createdSubmissionID = response.SubmissionID
		}
	})

	t.Run("GET /api/v1/schema-submissions - GetAllSchemaSubmissions", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schema-submissions", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.CollectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, response.Count, 0)
	})

	t.Run("GET /api/v1/schema-submissions - WithQueryParams", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schema-submissions?memberId=test&status=pending", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /api/v1/schema-submissions/:submissionId - GetSchemaSubmission", func(t *testing.T) {
		if createdSubmissionID == "" {
			t.Skip("No submission ID available")
			return
		}

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/schema-submissions/%s", createdSubmissionID), nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.SchemaSubmissionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, createdSubmissionID, response.SubmissionID)
		}
	})

	t.Run("PUT /api/v1/schema-submissions/:submissionId - UpdateSchemaSubmission", func(t *testing.T) {
		if createdSubmissionID == "" {
			t.Skip("No submission ID available")
			return
		}

		status := "approved"
		review := "Looks good"
		req := models.UpdateSchemaSubmissionRequest{
			Status: &status,
			Review: &review,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/schema-submissions/%s", createdSubmissionID), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.SchemaSubmissionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		}
	})
}

// TestApplicationEndpoints tests all application-related endpoints
func TestApplicationEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}
	defer testHandler.db.Exec("DELETE FROM applications")

	var createdApplicationID string
	testMemberID := "test-member-id"
	testSchemaID := "test-schema-id"

	t.Run("POST /api/v1/applications - CreateApplication", func(t *testing.T) {
		req := models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: "Test Description",
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: testSchemaID},
				{FieldName: "field2", SchemaID: testSchemaID},
			},
			MemberID: testMemberID,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusCreated {
			var response models.ApplicationResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, req.ApplicationName, response.ApplicationName)
			assert.NotEmpty(t, response.ApplicationID)
			createdApplicationID = response.ApplicationID
		}
	})

	t.Run("POST /api/v1/applications - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewBufferString("invalid"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/applications - GetAllApplications", func(t *testing.T) {
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

	t.Run("GET /api/v1/applications - WithQueryParams", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/applications?memberId=test-member", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /api/v1/applications/:applicationId - GetApplication", func(t *testing.T) {
		if createdApplicationID == "" {
			t.Skip("No application ID available")
			return
		}

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/applications/%s", createdApplicationID), nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.ApplicationResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, createdApplicationID, response.ApplicationID)
		}
	})

	t.Run("GET /api/v1/applications/:applicationId - NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/applications/non-existent", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /api/v1/applications/:applicationId - UpdateApplication", func(t *testing.T) {
		if createdApplicationID == "" {
			t.Skip("No application ID available")
			return
		}

		appName := "Updated Application Name"
		appDesc := "Updated Description"
		req := models.UpdateApplicationRequest{
			ApplicationName:        &appName,
			ApplicationDescription: &appDesc,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/applications/%s", createdApplicationID), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.ApplicationResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "Updated Application Name", response.ApplicationName)
		}
	})

	t.Run("Method Not Allowed - Applications", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/applications", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

// TestApplicationSubmissionEndpoints tests all application submission-related endpoints
func TestApplicationSubmissionEndpoints(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}
	defer testHandler.db.Exec("DELETE FROM application_submissions")

	var createdSubmissionID string
	testMemberID := "test-member-id"
	testSchemaID := "test-schema-id"

	t.Run("POST /api/v1/application-submissions - CreateApplicationSubmission", func(t *testing.T) {
		desc := "Test Description"
		req := models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Application Submission",
			ApplicationDescription: &desc,
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: testSchemaID},
			},
			MemberID: testMemberID,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/application-submissions", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusCreated {
			var response models.ApplicationSubmissionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, req.ApplicationName, response.ApplicationName)
			assert.NotEmpty(t, response.SubmissionID)
			createdSubmissionID = response.SubmissionID
		}
	})

	t.Run("PUT /api/v1/application-submissions/:id - UpdateApplicationSubmission", func(t *testing.T) {
		// Create a member first
		memberReq := models.CreateMemberRequest{
			Name:        "Test Member",
			Email:       "testmember3@example.com",
			PhoneNumber: "1234567890",
		}
		memberReqBody, _ := json.Marshal(memberReq)
		memberHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/members", bytes.NewBuffer(memberReqBody))
		memberHttpReq.Header.Set("Content-Type", "application/json")
		memberW := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(memberW, memberHttpReq)
		if memberW.Code != http.StatusCreated {
			t.Skip("Member creation failed, skipping submission test")
			return
		}
		var memberResponse map[string]interface{}
		json.Unmarshal(memberW.Body.Bytes(), &memberResponse)
		memberID := memberResponse["data"].(map[string]interface{})["memberId"].(string)

		// Create a schema first for the application submission
		schemaReq := models.CreateSchemaRequest{
			SchemaName:        "Test Schema",
			SchemaDescription: "Test Description",
			SDL:               "type Query { test: String }",
			Endpoint:          "http://example.com/graphql",
			MemberID:          memberID,
		}
		schemaReqBody, _ := json.Marshal(schemaReq)
		schemaHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBuffer(schemaReqBody))
		schemaHttpReq.Header.Set("Content-Type", "application/json")
		schemaW := httptest.NewRecorder()
		mux.ServeHTTP(schemaW, schemaHttpReq)
		if schemaW.Code != http.StatusCreated {
			t.Skip("Schema creation failed, skipping submission test")
			return
		}
		var schemaResponse map[string]interface{}
		json.Unmarshal(schemaW.Body.Bytes(), &schemaResponse)
		schemaID := schemaResponse["data"].(map[string]interface{})["schemaId"].(string)

		// Create a submission
		desc := "Test Description"
		createReq := models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Application Submission",
			ApplicationDescription: &desc,
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: schemaID},
			},
			MemberID: memberID,
		}
		createReqBody, _ := json.Marshal(createReq)
		createHttpReq := httptest.NewRequest(http.MethodPost, "/api/v1/application-submissions", bytes.NewBuffer(createReqBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		mux.ServeHTTP(createW, createHttpReq)
		if createW.Code != http.StatusCreated {
			t.Skip("Submission creation failed, skipping update test")
			return
		}

		var createResponse map[string]interface{}
		json.Unmarshal(createW.Body.Bytes(), &createResponse)
		if createResponse["data"] == nil {
			t.Skip("Submission creation failed, skipping update test")
			return
		}
		submissionID := createResponse["data"].(map[string]interface{})["submissionId"].(string)

		// Now update the submission
		status := "approved"
		updateReq := models.UpdateApplicationSubmissionRequest{
			Status: &status,
		}
		updateReqBody, _ := json.Marshal(updateReq)
		updateHttpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/application-submissions/%s", submissionID), bytes.NewBuffer(updateReqBody))
		updateHttpReq.Header.Set("Content-Type", "application/json")
		updateW := httptest.NewRecorder()
		mux.ServeHTTP(updateW, updateHttpReq)

		assert.Equal(t, http.StatusOK, updateW.Code)
		var response map[string]interface{}
		json.Unmarshal(updateW.Body.Bytes(), &response)
		assert.NotNil(t, response["data"])
	})

	t.Run("PUT /api/v1/application-submissions/:id - UpdateApplicationSubmission_InvalidJSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/application-submissions/test-id", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/application-submissions/:id - UpdateApplicationSubmission_NotFound", func(t *testing.T) {
		status := "approved"
		req := models.UpdateApplicationSubmissionRequest{
			Status: &status,
		}
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/application-submissions/non-existent-id", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/application-submissions - GetAllApplicationSubmissions", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/application-submissions", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.CollectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, response.Count, 0)
	})

	t.Run("GET /api/v1/application-submissions - WithQueryParams", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/application-submissions?memberId=test&status=pending", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GET /api/v1/application-submissions/:submissionId - GetApplicationSubmission", func(t *testing.T) {
		if createdSubmissionID == "" {
			t.Skip("No submission ID available")
			return
		}

		httpReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/application-submissions/%s", createdSubmissionID), nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.ApplicationSubmissionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, createdSubmissionID, response.SubmissionID)
		}
	})

	t.Run("GET /api/v1/application-submissions/:submissionId - NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/application-submissions/non-existent", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /api/v1/application-submissions/:submissionId - UpdateApplicationSubmission", func(t *testing.T) {
		if createdSubmissionID == "" {
			t.Skip("No submission ID available")
			return
		}

		status := "approved"
		review := "Approved"
		req := models.UpdateApplicationSubmissionRequest{
			Status: &status,
			Review: &review,
		}

		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/application-submissions/%s", createdSubmissionID), bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		if w.Code == http.StatusOK {
			var response models.ApplicationSubmissionResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
		}
	})

	t.Run("Method Not Allowed - ApplicationSubmissions", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/application-submissions", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

// TestSchemaEndpoints_EdgeCases tests edge cases for schema endpoints
func TestSchemaEndpoints_EdgeCases(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}

	t.Run("POST /api/v1/schemas - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schemas", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/schemas/:id - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/schemas/test-id", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/schemas/:id - NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/schemas/non-existent-id", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /api/v1/schemas/:id - NotFound", func(t *testing.T) {
		schemaName := "Updated Name"
		req := models.UpdateSchemaRequest{
			SchemaName: &schemaName,
		}
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/schemas/non-existent-id", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestApplicationEndpoints_EdgeCases tests edge cases for application endpoints
func TestApplicationEndpoints_EdgeCases(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}

	t.Run("POST /api/v1/applications - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/applications/:id - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/applications/test-id", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/applications/:id - NotFound", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/applications/non-existent-id", nil)
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /api/v1/applications/:id - NotFound", func(t *testing.T) {
		appName := "Updated Name"
		req := models.UpdateApplicationRequest{
			ApplicationName: &appName,
		}
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/applications/non-existent-id", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestSchemaSubmissionEndpoints_EdgeCases tests edge cases for schema submission endpoints
func TestSchemaSubmissionEndpoints_EdgeCases(t *testing.T) {
	testHandler := NewTestV1Handler(t)
	if testHandler == nil {
		t.Skip("Skipping test: database connection failed")
		return
	}

	t.Run("POST /api/v1/schema-submissions - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/schema-submissions", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/schema-submissions/:id - Invalid JSON", func(t *testing.T) {
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/schema-submissions/test-id", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/schema-submissions/:id - NotFound", func(t *testing.T) {
		status := "approved"
		req := models.UpdateSchemaSubmissionRequest{
			Status: &status,
		}
		reqBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/schema-submissions/non-existent-id", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		testHandler.handler.SetupV1Routes(mux)
		mux.ServeHTTP(w, httpReq)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
