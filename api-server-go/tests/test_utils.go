package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	_ "github.com/lib/pq"
)

// TestServer represents a test HTTP server with common setup
type TestServer struct {
	APIServer *handlers.APIServer
	Mux       *http.ServeMux
	DB        *sql.DB
}

// NewTestServer creates a new test server instance with PostgreSQL test database
func NewTestServer() *TestServer {
	// Get test database configuration from environment variables
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5434")
	user := getEnvOrDefault("TEST_DB_USER", "test_user")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "test_password")
	database := getEnvOrDefault("TEST_DB_NAME", "api_server_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	// Create PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslmode)

	// Connect to test database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		panic("Failed to ping test database: " + err.Error())
	}

	// Create tables for testing
	if err := createTestTables(db); err != nil {
		panic("Failed to create test tables: " + err.Error())
	}

	// Create API server with database
	apiServer := handlers.NewAPIServerWithDB(db)
	mux := http.NewServeMux()
	apiServer.SetupRoutes(mux)

	return &TestServer{
		APIServer: apiServer,
		Mux:       mux,
		DB:        db,
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createTestTables creates the necessary tables for testing using PostgreSQL syntax
func createTestTables(db *sql.DB) error {
	// Drop tables if they exist (for clean test runs)
	dropTables := []string{
		"DROP TABLE IF EXISTS consumer_grants CASCADE",
		"DROP TABLE IF EXISTS consumer_apps CASCADE",
		"DROP TABLE IF EXISTS consumers CASCADE",
		"DROP TABLE IF EXISTS provider_schemas CASCADE",
		"DROP TABLE IF EXISTS provider_profiles CASCADE",
		"DROP TABLE IF EXISTS provider_submissions CASCADE",
		"DROP TABLE IF EXISTS provider_metadata CASCADE",
	}

	for _, drop := range dropTables {
		db.Exec(drop) // Ignore errors for cleanup
	}

	// Create tables with PostgreSQL syntax
	tables := []string{
		`CREATE TABLE consumers (
			consumer_id VARCHAR(255) PRIMARY KEY,
			consumer_name VARCHAR(255) NOT NULL,
			contact_email VARCHAR(255) NOT NULL,
			phone_number VARCHAR(50),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE consumer_apps (
			submission_id VARCHAR(255) PRIMARY KEY,
			consumer_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			required_fields JSONB,
			credentials JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			FOREIGN KEY (consumer_id) REFERENCES consumers(consumer_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE provider_submissions (
			submission_id VARCHAR(255) PRIMARY KEY,
			provider_name VARCHAR(255) NOT NULL,
			contact_email VARCHAR(255) NOT NULL,
			phone_number VARCHAR(50),
			provider_type VARCHAR(100) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE provider_profiles (
			provider_id VARCHAR(255) PRIMARY KEY,
			provider_name VARCHAR(255) NOT NULL,
			contact_email VARCHAR(255) NOT NULL,
			phone_number VARCHAR(50),
			provider_type VARCHAR(100) NOT NULL,
			approved_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE provider_schemas (
			schema_id VARCHAR(255) PRIMARY KEY,
			provider_id VARCHAR(255) NOT NULL,
			submission_id VARCHAR(255),
			status VARCHAR(50) NOT NULL,
			schema_data JSONB,
			schema_input TEXT,
			sdl TEXT,
			field_configurations JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			FOREIGN KEY (provider_id) REFERENCES provider_profiles(provider_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE consumer_grants (
			consumer_id VARCHAR(255) PRIMARY KEY,
			approved_fields JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			FOREIGN KEY (consumer_id) REFERENCES consumers(consumer_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE provider_metadata (
			field_name VARCHAR(255) PRIMARY KEY,
			owner VARCHAR(255) NOT NULL,
			provider VARCHAR(255) NOT NULL,
			consent_required BOOLEAN NOT NULL DEFAULT FALSE,
			access_control_type VARCHAR(50) NOT NULL,
			allow_list JSONB,
			description TEXT,
			expiry_time VARCHAR(50),
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// Close cleans up the test server
func (ts *TestServer) Close() {
	if ts.DB != nil {
		ts.DB.Close()
	}
}

// MakeRequest makes an HTTP request and returns the response
func (ts *TestServer) MakeRequest(method, url string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			panic("Failed to marshal request body: " + err.Error())
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	ts.Mux.ServeHTTP(w, req)

	return w
}

// MakeGETRequest makes a GET request
func (ts *TestServer) MakeGETRequest(url string) *httptest.ResponseRecorder {
	return ts.MakeRequest("GET", url, nil)
}

// MakePOSTRequest makes a POST request
func (ts *TestServer) MakePOSTRequest(url string, body interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("POST", url, body)
}

// MakePUTRequest makes a PUT request
func (ts *TestServer) MakePUTRequest(url string, body interface{}) *httptest.ResponseRecorder {
	return ts.MakeRequest("PUT", url, body)
}

// MakeDELETERequest makes a DELETE request
func (ts *TestServer) MakeDELETERequest(url string) *httptest.ResponseRecorder {
	return ts.MakeRequest("DELETE", url, nil)
}

// AssertResponseStatus checks if the response has the expected status code
func AssertResponseStatus(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	if w.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Response: %s", expectedStatus, w.Code, w.Body.String())
	}
}

// AssertJSONResponse checks if the response can be unmarshaled as JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
		t.Errorf("Failed to unmarshal response: %v. Response: %s", err, w.Body.String())
	}
}

// AssertErrorResponse checks if the response contains an error
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	AssertResponseStatus(t, w, expectedStatus)

	var errorResp map[string]string
	AssertJSONResponse(t, w, &errorResp)

	if _, hasError := errorResp["error"]; !hasError {
		t.Error("Expected error field in response")
	}
}

// AssertSuccessResponse checks if the response is successful
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	AssertResponseStatus(t, w, expectedStatus)

	// Try to unmarshal as JSON to ensure it's valid
	var response interface{}
	AssertJSONResponse(t, w, &response)
}

// CreateTestConsumer creates a consumer for testing and returns the consumer ID
func (ts *TestServer) CreateTestConsumer(t *testing.T, name, email, phone string) string {
	consumerReq := map[string]string{
		"consumerName": name,
		"contactEmail": email,
		"phoneNumber":  phone,
	}

	w := ts.MakePOSTRequest("/consumers", consumerReq)
	AssertResponseStatus(t, w, http.StatusCreated)

	var consumer map[string]interface{}
	AssertJSONResponse(t, w, &consumer)

	consumerID, ok := consumer["consumerId"].(string)
	if !ok {
		t.Fatal("Expected consumerId in response")
	}

	return consumerID
}

// CreateTestConsumerApp creates a consumer application for testing and returns the submission ID
func (ts *TestServer) CreateTestConsumerApp(t *testing.T, consumerID string, requiredFields map[string]bool) string {
	appReq := map[string]interface{}{
		"required_fields": requiredFields,
	}

	w := ts.MakePOSTRequest("/consumer-applications/"+consumerID, appReq)
	AssertResponseStatus(t, w, http.StatusCreated)

	var app map[string]interface{}
	AssertJSONResponse(t, w, &app)

	submissionID, ok := app["submissionId"].(string)
	if !ok {
		t.Fatal("Expected submissionId in response")
	}

	return submissionID
}

// CreateTestProviderProfile creates a provider profile directly for testing and returns the provider ID
func (ts *TestServer) CreateTestProviderProfile(t *testing.T, name, email, phone, providerType string) string {
	profile, err := ts.APIServer.GetProviderService().CreateProviderProfileForTesting(
		name, email, phone, providerType,
	)
	if err != nil {
		t.Fatalf("Failed to create provider profile: %v", err)
	}
	return profile.ProviderID
}

// CreateTestSchemaSubmission creates a schema submission for testing and returns the schema ID
func (ts *TestServer) CreateTestSchemaSubmission(t *testing.T, providerID, sdl string) string {
	schemaReq := map[string]interface{}{
		"sdl":       sdl,
		"schema_id": nil,
	}

	w := ts.MakePOSTRequest("/providers/"+providerID+"/schema-submissions", schemaReq)
	AssertResponseStatus(t, w, http.StatusCreated)

	var schema map[string]interface{}
	AssertJSONResponse(t, w, &schema)

	schemaID, ok := schema["submissionId"].(string)
	if !ok {
		t.Fatal("Expected submissionId in response")
	}

	return schemaID
}

// SubmitSchemaForReview submits a draft schema for admin review
func (ts *TestServer) SubmitSchemaForReview(t *testing.T, providerID, schemaID string) {
	updateReq := map[string]string{
		"status": "pending",
	}
	w := ts.MakePUTRequest("/providers/"+providerID+"/schema-submissions/"+schemaID, updateReq)
	AssertResponseStatus(t, w, http.StatusOK)
}

// ApproveSchemaSubmission approves a schema submission for testing
func (ts *TestServer) ApproveSchemaSubmission(t *testing.T, providerID, schemaID string) {
	updateReq := map[string]string{
		"status": "approved",
	}

	w := ts.MakePUTRequest("/providers/"+providerID+"/schema-submissions/"+schemaID, updateReq)
	AssertResponseStatus(t, w, http.StatusOK)
}
