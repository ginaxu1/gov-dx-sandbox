package services_test

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestProviderService_CreateProviderSubmission(t *testing.T) {
	service := services.NewProviderService()

	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Department",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
		ProviderType: models.ProviderTypeGovernment,
	}

	submission, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if submission.SubmissionID == "" {
		t.Error("Expected SubmissionID to be generated")
	}

	if submission.Status != models.SubmissionStatusPending {
		t.Errorf("Expected status %s, got %s", models.SubmissionStatusPending, submission.Status)
	}

	if submission.ProviderName != req.ProviderName {
		t.Errorf("Expected provider name %s, got %s", req.ProviderName, submission.ProviderName)
	}
}

func TestProviderService_CreateProviderSubmission_Duplicate(t *testing.T) {
	service := services.NewProviderService()

	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Department",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
		ProviderType: models.ProviderTypeGovernment,
	}

	// Create first submission
	_, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Failed to create first submission: %v", err)
	}

	// Try to create duplicate
	_, err = service.CreateProviderSubmission(req)
	if err == nil {
		t.Error("Expected error for duplicate submission")
	}
}

func TestProviderService_UpdateProviderSubmission_Approval(t *testing.T) {
	service := services.NewProviderService()

	// Create a submission
	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Department",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
		ProviderType: models.ProviderTypeGovernment,
	}

	submission, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	// Approve the submission
	updateReq := models.UpdateProviderSubmissionRequest{
		Status: &[]models.ProviderSubmissionStatus{models.SubmissionStatusApproved}[0],
	}

	updatedSubmission, err := service.UpdateProviderSubmission(submission.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedSubmission.Status != models.SubmissionStatusApproved {
		t.Errorf("Expected status %s, got %s", models.SubmissionStatusApproved, updatedSubmission.Status)
	}

	// Check that provider profile was created
	profiles, err := service.GetAllProviderProfiles()
	if err != nil {
		t.Fatalf("Failed to get provider profiles: %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("Expected 1 provider profile, got %d", len(profiles))
	}

	profile := profiles[0]
	if profile.ProviderName != submission.ProviderName {
		t.Errorf("Expected provider name %s, got %s", submission.ProviderName, profile.ProviderName)
	}
}

func TestProviderService_CreateProviderSchema(t *testing.T) {
	service := services.NewProviderService()

	// First create a provider profile
	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Department",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
		ProviderType: models.ProviderTypeGovernment,
	}

	submission, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	// Approve to create profile
	updateReq := models.UpdateProviderSubmissionRequest{
		Status: &[]models.ProviderSubmissionStatus{models.SubmissionStatusApproved}[0],
	}

	_, err = service.UpdateProviderSubmission(submission.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Failed to approve submission: %v", err)
	}

	// Get the created profile
	profiles, err := service.GetAllProviderProfiles()
	if err != nil {
		t.Fatalf("Failed to get profiles: %v", err)
	}

	providerID := profiles[0].ProviderID

	// Create schema for the provider
	schemaReq := models.CreateProviderSchemaRequest{
		ProviderID: providerID,
		FieldConfigurations: models.FieldConfigurations{
			"PersonData": {
				"fullName": {
					Source:      "authoritative",
					IsOwner:     true,
					Description: "Full name of the person",
				},
			},
		},
	}

	schema, err := service.CreateProviderSchema(schemaReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if schema.SubmissionID == "" {
		t.Error("Expected SubmissionID to be generated")
	}

	if schema.ProviderID != providerID {
		t.Errorf("Expected provider ID %s, got %s", providerID, schema.ProviderID)
	}

	if schema.Status != models.SchemaStatusPending {
		t.Errorf("Expected status %s, got %s", models.SchemaStatusPending, schema.Status)
	}
}

func TestProviderService_GetProviderSubmission_NotFound(t *testing.T) {
	service := services.NewProviderService()

	_, err := service.GetProviderSubmission("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent submission")
	}
}

func TestProviderService_GetProviderProfile_NotFound(t *testing.T) {
	service := services.NewProviderService()

	_, err := service.GetProviderProfile("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent profile")
	}
}

func TestProviderService_GetProviderSchema_NotFound(t *testing.T) {
	service := services.NewProviderService()

	_, err := service.GetProviderSchema("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent schema")
	}
}
