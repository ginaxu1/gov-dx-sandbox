package tests

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

func TestProviderService_CreateProviderSubmission(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetProviderService()

	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Provider",
		ContactEmail: "test@provider.com",
		PhoneNumber:  "1234567890",
		ProviderType: "government",
	}

	submission, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if submission.SubmissionID == "" {
		t.Error("Expected SubmissionID to be generated")
	}
	if submission.ProviderName != req.ProviderName {
		t.Errorf("Expected ProviderName %s, got %s", req.ProviderName, submission.ProviderName)
	}
	if submission.ContactEmail != req.ContactEmail {
		t.Errorf("Expected ContactEmail %s, got %s", req.ContactEmail, submission.ContactEmail)
	}
}

func TestProviderService_GetProviderSubmission(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetProviderService()

	// First create a submission
	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Provider",
		ContactEmail: "test@provider.com",
		PhoneNumber:  "1234567890",
		ProviderType: "government",
	}

	createdSubmission, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Failed to create provider submission: %v", err)
	}

	// Now get the submission
	submission, err := service.GetProviderSubmission(createdSubmission.SubmissionID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if submission.SubmissionID != createdSubmission.SubmissionID {
		t.Errorf("Expected SubmissionID %s, got %s", createdSubmission.SubmissionID, submission.SubmissionID)
	}
}

func TestProviderService_GetProviderSubmission_NotFound(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetProviderService()

	_, err := service.GetProviderSubmission("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent submission")
	}
}

func TestProviderService_GetAllProviderSubmissions(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetProviderService()

	// Create multiple submissions
	submissions := []models.CreateProviderSubmissionRequest{
		{ProviderName: "Provider 1", ContactEmail: "provider1@example.com", PhoneNumber: "1111111111", ProviderType: "government"},
		{ProviderName: "Provider 2", ContactEmail: "provider2@example.com", PhoneNumber: "2222222222", ProviderType: "private"},
	}

	for _, req := range submissions {
		_, err := service.CreateProviderSubmission(req)
		if err != nil {
			t.Fatalf("Failed to create provider submission: %v", err)
		}
	}

	// Get all submissions
	allSubmissions, err := service.GetAllProviderSubmissions()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allSubmissions) != len(submissions) {
		t.Errorf("Expected %d submissions, got %d", len(submissions), len(allSubmissions))
	}
}

func TestProviderService_UpdateProviderSubmission(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetProviderService()

	// Create a submission
	req := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Provider",
		ContactEmail: "test@provider.com",
		PhoneNumber:  "1234567890",
		ProviderType: "government",
	}

	createdSubmission, err := service.CreateProviderSubmission(req)
	if err != nil {
		t.Fatalf("Failed to create provider submission: %v", err)
	}

	// Update the submission
	status := models.SubmissionStatusApproved
	updateReq := models.UpdateProviderSubmissionRequest{
		Status: &status,
	}

	updatedSubmission, err := service.UpdateProviderSubmission(createdSubmission.SubmissionID, updateReq)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updatedSubmission.Status != "approved" {
		t.Errorf("Expected Status 'approved', got %s", updatedSubmission.Status)
	}
}

func TestProviderService_GetAllProviderProfilesWithEntity(t *testing.T) {
	ts := NewTestServer()
	defer ts.Close()
	service := ts.APIServer.GetProviderService()

	// Create provider profiles directly for testing
	_, err := service.CreateProviderProfileForTesting("Provider 1", "provider1@example.com", "1111111111", "government")
	if err != nil {
		t.Fatalf("Failed to create provider profile: %v", err)
	}

	_, err = service.CreateProviderProfileForTesting("Provider 2", "provider2@example.com", "2222222222", "private")
	if err != nil {
		t.Fatalf("Failed to create provider profile: %v", err)
	}

	// Get all profiles
	profiles, err := service.GetAllProviderProfilesWithEntity()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if profiles == nil {
		t.Error("Expected profiles data to be returned")
	}

	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}
}
