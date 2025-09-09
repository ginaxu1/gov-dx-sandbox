package services_test

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestConsumerService_CreateApplication(t *testing.T) {
	service := services.NewConsumerService()

	req := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{
			"person.fullName": true,
			"person.nic":      true,
		},
	}

	app, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if app.AppID == "" {
		t.Error("Expected AppID to be generated")
	}

	if app.Status != models.StatusPending {
		t.Errorf("Expected status %s, got %s", models.StatusPending, app.Status)
	}

	if len(app.RequiredFields) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(app.RequiredFields))
	}
}

func TestConsumerService_GetApplication(t *testing.T) {
	service := services.NewConsumerService()

	// Create an application first
	req := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{"person.fullName": true},
	}
	createdApp, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Retrieve the application
	app, err := service.GetApplication(createdApp.AppID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if app.AppID != createdApp.AppID {
		t.Errorf("Expected AppID %s, got %s", createdApp.AppID, app.AppID)
	}
}

func TestConsumerService_GetApplication_NotFound(t *testing.T) {
	service := services.NewConsumerService()

	_, err := service.GetApplication("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent application")
	}
}

func TestConsumerService_UpdateApplication(t *testing.T) {
	service := services.NewConsumerService()

	// Create an application
	req := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{"person.fullName": true},
	}
	createdApp, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Update the application
	updateReq := models.UpdateApplicationRequest{
		Status: &[]models.ApplicationStatus{models.StatusApproved}[0],
	}

	updatedApp, err := service.UpdateApplication(createdApp.AppID, updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedApp.Status != models.StatusApproved {
		t.Errorf("Expected status %s, got %s", models.StatusApproved, updatedApp.Status)
	}

	// Check that credentials were generated for approved application
	if updatedApp.Credentials == nil {
		t.Error("Expected credentials to be generated for approved application")
	}

	if updatedApp.Credentials.APIKey == "" || updatedApp.Credentials.APISecret == "" {
		t.Error("Expected credentials to have API key and secret")
	}
}

func TestConsumerService_DeleteApplication(t *testing.T) {
	service := services.NewConsumerService()

	// Create an application
	req := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{"person.fullName": true},
	}
	createdApp, err := service.CreateApplication(req)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Delete the application
	err = service.DeleteApplication(createdApp.AppID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it's deleted
	_, err = service.GetApplication(createdApp.AppID)
	if err == nil {
		t.Error("Expected error for deleted application")
	}
}

func TestConsumerService_GetAllApplications(t *testing.T) {
	service := services.NewConsumerService()

	// Create multiple applications
	req1 := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{"person.fullName": true},
	}
	req2 := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{"person.nic": true},
	}

	_, err := service.CreateApplication(req1)
	if err != nil {
		t.Fatalf("Failed to create first application: %v", err)
	}

	_, err = service.CreateApplication(req2)
	if err != nil {
		t.Fatalf("Failed to create second application: %v", err)
	}

	// Get all applications
	apps, err := service.GetAllApplications()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("Expected 2 applications, got %d", len(apps))
	}
}
