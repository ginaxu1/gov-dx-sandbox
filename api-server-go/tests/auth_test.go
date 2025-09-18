package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestAuthService_AuthenticateConsumer_Success(t *testing.T) {
	// Create services
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer and approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create and approve an application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	app, err := consumerService.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Approve the application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}
	approvedApp, err := consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve application: %v", err)
	}

	// Test authentication
	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     approvedApp.Credentials.APISecret,
	}

	response, err := authService.AuthenticateConsumer(authReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.AccessToken == "" {
		t.Error("Expected access token to be generated")
	}

	if response.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got %s", response.TokenType)
	}

	if response.ConsumerID != consumer.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", consumer.ConsumerID, response.ConsumerID)
	}

	if response.ExpiresIn <= 0 {
		t.Error("Expected positive expiration time")
	}

	if response.ExpiresAt.Before(time.Now()) {
		t.Error("Expected expiration time to be in the future")
	}
}

func TestAuthService_AuthenticateConsumer_InvalidCredentials(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	authReq := models.AuthRequest{
		ConsumerID: "nonexistent",
		Secret:     "wrong-secret",
	}

	_, err := authService.AuthenticateConsumer(authReq)
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}

	expectedError := "consumer not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAuthService_AuthenticateConsumer_NoApprovedApplication(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer but no approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     "some-secret",
	}

	_, err = authService.AuthenticateConsumer(authReq)
	if err == nil {
		t.Error("Expected error for no approved application")
	}

	expectedError := "no approved application found for consumer"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAuthService_AuthenticateConsumer_WrongSecret(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer and approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create and approve an application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	app, err := consumerService.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Approve the application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}
	_, err = consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve application: %v", err)
	}

	// Test with wrong secret
	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     "wrong-secret",
	}

	_, err = authService.AuthenticateConsumer(authReq)
	if err == nil {
		t.Error("Expected error for wrong secret")
	}

	expectedError := "invalid credentials"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAuthService_ValidateToken_ValidToken(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer and approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create and approve an application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	app, err := consumerService.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Approve the application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}
	approvedApp, err := consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve application: %v", err)
	}

	// Generate a token
	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     approvedApp.Credentials.APISecret,
	}
	authResponse, err := authService.AuthenticateConsumer(authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Validate the token
	response, err := authService.ValidateToken(authResponse.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !response.Valid {
		t.Error("Expected token to be valid")
	}

	if response.ConsumerID != consumer.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", consumer.ConsumerID, response.ConsumerID)
	}

	if response.Error != "" {
		t.Errorf("Expected no error message, got %s", response.Error)
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Test with invalid token format
	response, err := authService.ValidateToken("invalid.token.format")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Valid {
		t.Error("Expected token to be invalid")
	}

	if response.Error == "" {
		t.Error("Expected error message for invalid token")
	}
}

func TestAuthService_ValidateToken_ConsumerNoLongerApproved(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer and approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create and approve an application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	app, err := consumerService.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Approve the application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}
	_, err = consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve application: %v", err)
	}

	// Generate a token
	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     "some-secret", // We'll use a mock secret for this test
	}

	// First create a valid token by getting the actual secret
	apps, err := consumerService.GetConsumerAppsByConsumerID(consumer.ConsumerID)
	if err != nil {
		t.Fatalf("Failed to get consumer apps: %v", err)
	}

	var actualSecret string
	for _, app := range apps {
		if app.Status == models.StatusApproved && app.Credentials != nil {
			actualSecret = app.Credentials.APISecret
			break
		}
	}

	authReq.Secret = actualSecret
	authResponse, err := authService.AuthenticateConsumer(authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Now reject the application
	rejectReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusDenied}[0],
	}
	_, err = consumerService.UpdateConsumerApp(app.SubmissionID, rejectReq)
	if err != nil {
		t.Fatalf("Failed to reject application: %v", err)
	}

	// Validate the token - should now be invalid
	response, err := authService.ValidateToken(authResponse.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Valid {
		t.Error("Expected token to be invalid after application rejection")
	}

	expectedError := "Consumer application no longer approved"
	if response.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, response.Error)
	}
}

func TestAuthService_GetConsumerIDFromToken_ValidToken(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer and approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create and approve an application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	app, err := consumerService.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Approve the application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}
	approvedApp, err := consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve application: %v", err)
	}

	// Generate a token
	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     approvedApp.Credentials.APISecret,
	}
	authResponse, err := authService.AuthenticateConsumer(authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Test GetConsumerIDFromToken
	consumerID, err := authService.GetConsumerIDFromToken(authResponse.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if consumerID != consumer.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", consumer.ConsumerID, consumerID)
	}
}

func TestAuthService_GetConsumerIDFromToken_WithBearerPrefix(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Create a consumer and approved application
	consumerReq := models.CreateConsumerRequest{
		ConsumerName: "Test Consumer",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
	}
	consumer, err := consumerService.CreateConsumer(consumerReq)
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Create and approve an application
	appReq := models.CreateConsumerAppRequest{
		ConsumerID:     consumer.ConsumerID,
		RequiredFields: map[string]bool{"person.fullName": true},
	}
	app, err := consumerService.CreateConsumerApp(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Approve the application
	updateReq := models.UpdateConsumerAppRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}
	approvedApp, err := consumerService.UpdateConsumerApp(app.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve application: %v", err)
	}

	// Generate a token
	authReq := models.AuthRequest{
		ConsumerID: consumer.ConsumerID,
		Secret:     approvedApp.Credentials.APISecret,
	}
	authResponse, err := authService.AuthenticateConsumer(authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Test GetConsumerIDFromToken with Bearer prefix
	consumerID, err := authService.GetConsumerIDFromToken("Bearer " + authResponse.AccessToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if consumerID != consumer.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", consumer.ConsumerID, consumerID)
	}
}

func TestAuthService_GetConsumerIDFromToken_InvalidToken(t *testing.T) {
	consumerService := services.NewConsumerService()
	authService := services.NewAuthService(consumerService)

	// Test with invalid token
	_, err := authService.GetConsumerIDFromToken("invalid.token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

// HTTP endpoint tests for authentication
func TestAuthEndpoints(t *testing.T) {
	ts := NewTestServer()

	t.Run("POST /auth/token", func(t *testing.T) {
		// Create a consumer and approved application first
		consumerID := ts.CreateTestConsumer(t, "Test Consumer", "test@example.com", "1234567890")

		// Create and approve an application
		requiredFields := map[string]bool{"person.fullName": true}
		submissionID := ts.CreateTestConsumerApp(t, consumerID, requiredFields)

		// Approve the application
		updateReq := map[string]string{
			"status": "approved",
		}
		w := ts.MakePUTRequest("/consumer-applications/"+submissionID, updateReq)
		AssertResponseStatus(t, w, http.StatusOK)

		// Get the approved application to get credentials
		w = ts.MakeGETRequest("/consumer-applications/" + submissionID)
		AssertResponseStatus(t, w, http.StatusOK)

		var app map[string]interface{}
		AssertJSONResponse(t, w, &app)

		credentials, ok := app["credentials"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected credentials in response")
		}

		apiSecret, ok := credentials["apiSecret"].(string)
		if !ok {
			t.Fatal("Expected apiSecret in credentials")
		}

		// Test authentication
		authReq := map[string]string{
			"consumerId": consumerID,
			"secret":     apiSecret,
		}

		w = ts.MakePOSTRequest("/auth/token", authReq)
		AssertResponseStatus(t, w, http.StatusOK)

		var authResp map[string]interface{}
		AssertJSONResponse(t, w, &authResp)

		if _, ok := authResp["accessToken"]; !ok {
			t.Error("Expected accessToken in response")
		}

		if authResp["tokenType"] != "Bearer" {
			t.Errorf("Expected tokenType 'Bearer', got %v", authResp["tokenType"])
		}

		if authResp["consumerId"] != consumerID {
			t.Errorf("Expected consumerId %s, got %v", consumerID, authResp["consumerId"])
		}
	})

	t.Run("POST /auth/token - Invalid Credentials", func(t *testing.T) {
		authReq := map[string]string{
			"consumerId": "nonexistent",
			"secret":     "wrong-secret",
		}

		w := ts.MakePOSTRequest("/auth/token", authReq)
		AssertResponseStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("POST /auth/token - Missing Fields", func(t *testing.T) {
		authReq := map[string]string{
			"consumerId": "test",
			// Missing secret
		}

		w := ts.MakePOSTRequest("/auth/token", authReq)
		AssertResponseStatus(t, w, http.StatusBadRequest)
	})

	t.Run("POST /auth/validate (Asgardeo)", func(t *testing.T) {
		// This test requires Asgardeo service to be configured
		// For now, we'll test the endpoint structure without actual Asgardeo validation
		// In a real test environment, you would need to mock the Asgardeo service

		// Test with a mock Asgardeo token
		validateReq := map[string]string{
			"token": "mock.asgardeo.token",
		}

		w := ts.MakePOSTRequest("/auth/validate", validateReq)
		// The response will depend on whether Asgardeo service is configured
		// We expect either 200 with validation result or 500 if service not configured
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 200 or 500, got %d", w.Code)
		}
	})

	t.Run("POST /auth/validate - Invalid Token", func(t *testing.T) {
		validateReq := map[string]string{
			"token": "invalid.token",
		}

		w := ts.MakePOSTRequest("/auth/validate", validateReq)
		// Asgardeo validation will return 500 if service not configured, or 200 with validation result
		if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 200 or 500, got %d", w.Code)
		}
	})

	t.Run("POST /auth/validate - Missing Token", func(t *testing.T) {
		validateReq := map[string]string{
			// Missing token
		}

		w := ts.MakePOSTRequest("/auth/validate", validateReq)
		AssertResponseStatus(t, w, http.StatusBadRequest)
	})

	t.Run("Unsupported Methods", func(t *testing.T) {
		// Test GET on auth endpoints
		w := ts.MakeGETRequest("/auth/token")
		AssertResponseStatus(t, w, http.StatusMethodNotAllowed)

		w = ts.MakeGETRequest("/auth/validate")
		AssertResponseStatus(t, w, http.StatusMethodNotAllowed)
	})
}
