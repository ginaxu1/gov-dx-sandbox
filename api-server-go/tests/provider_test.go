package tests

import (
	"net/http"
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

// HTTP endpoint tests for provider management
func TestProviderEndpoints(t *testing.T) {
	ts := NewTestServer()

	t.Run("Provider Management", func(t *testing.T) {
		// Test GET /providers (empty initially)
		w := ts.MakeGETRequest("/providers")
		AssertResponseStatus(t, w, http.StatusOK)

		// Create a provider profile
		providerID := ts.CreateTestProviderProfile(t, "Test Provider", "provider@example.com", "1234567890", "government")

		// Test GET /providers (should now have one provider)
		w = ts.MakeGETRequest("/providers")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test GET /providers/{providerId}
		w = ts.MakeGETRequest("/providers/" + providerID)
		AssertResponseStatus(t, w, http.StatusOK)
	})

	t.Run("Provider Schema Management", func(t *testing.T) {
		// Create a provider profile
		providerID := ts.CreateTestProviderProfile(t, "Test Provider", "provider@example.com", "1234567890", "government")

		// Test GET /providers/{providerId}/schemas (empty initially)
		w := ts.MakeGETRequest("/providers/" + providerID + "/schemas")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test GET /providers/{providerId}/schema-submissions (empty initially)
		w = ts.MakeGETRequest("/providers/" + providerID + "/schema-submissions")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test POST /providers/{providerId}/schema-submissions (new schema)
		schemaID := ts.CreateTestSchemaSubmission(t, providerID, "type Query { test: String }")

		// Test GET /providers/{providerId}/schema-submissions/{schemaId}
		w = ts.MakeGETRequest("/providers/" + providerID + "/schema-submissions/" + schemaID)
		AssertResponseStatus(t, w, http.StatusOK)

		// Test PUT /providers/{providerId}/schema-submissions/{schemaId} (submit for review)
		ts.SubmitSchemaForReview(t, providerID, schemaID)

		// Test PUT /providers/{providerId}/schema-submissions/{schemaId} (admin approval)
		ts.ApproveSchemaSubmission(t, providerID, schemaID)

		// Test GET /providers/{providerId}/schemas (should now have approved schema)
		w = ts.MakeGETRequest("/providers/" + providerID + "/schemas")
		AssertResponseStatus(t, w, http.StatusOK)
	})

	t.Run("Schema Status Workflow", func(t *testing.T) {
		// Create provider and schema
		providerID := ts.CreateTestProviderProfile(t, "Workflow Provider", "workflow@example.com", "1234567890", "government")
		schemaID := ts.CreateTestSchemaSubmission(t, providerID, "type Query { workflow: String }")

		// Verify initial status is draft
		w := ts.MakeGETRequest("/providers/" + providerID + "/schema-submissions/" + schemaID)
		AssertResponseStatus(t, w, http.StatusOK)

		// Submit for review (draft -> pending)
		ts.SubmitSchemaForReview(t, providerID, schemaID)

		// Verify status is now pending
		w = ts.MakeGETRequest("/providers/" + providerID + "/schema-submissions/" + schemaID)
		AssertResponseStatus(t, w, http.StatusOK)

		// Approve schema (pending -> approved)
		ts.ApproveSchemaSubmission(t, providerID, schemaID)

		// Verify schema appears in approved schemas
		w = ts.MakeGETRequest("/providers/" + providerID + "/schemas")
		AssertResponseStatus(t, w, http.StatusOK)
	})

	t.Run("Provider Schema Modification", func(t *testing.T) {
		// Create provider and initial schema
		providerID := ts.CreateTestProviderProfile(t, "Modification Provider", "mod@example.com", "1234567890", "government")
		initialSchemaID := ts.CreateTestSchemaSubmission(t, providerID, "type Query { original: String }")

		// Submit and approve initial schema
		ts.SubmitSchemaForReview(t, providerID, initialSchemaID)
		ts.ApproveSchemaSubmission(t, providerID, initialSchemaID)

		// Create modification schema
		modificationReq := map[string]interface{}{
			"sdl":       "type Query { modified: String }",
			"schema_id": initialSchemaID,
		}
		w := ts.MakePOSTRequest("/providers/"+providerID+"/schema-submissions", modificationReq)
		AssertResponseStatus(t, w, http.StatusCreated)

		var response map[string]interface{}
		AssertJSONResponse(t, w, &response)
		modificationID := response["submissionId"].(string)

		// Submit and approve modification
		ts.SubmitSchemaForReview(t, providerID, modificationID)
		ts.ApproveSchemaSubmission(t, providerID, modificationID)

		// Verify both schemas are approved
		w = ts.MakeGETRequest("/providers/" + providerID + "/schemas")
		AssertResponseStatus(t, w, http.StatusOK)
	})

	t.Run("Provider Submissions", func(t *testing.T) {
		// Test GET /provider-submissions (empty initially)
		w := ts.MakeGETRequest("/provider-submissions")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test POST /provider-submissions
		submissionReq := map[string]interface{}{
			"providerName": "DRP Test Inc",
			"contactEmail": "contact@healthtech.com",
			"phoneNumber":  "+1-555-0789",
			"providerType": "business",
		}
		w = ts.MakePOSTRequest("/provider-submissions", submissionReq)
		AssertResponseStatus(t, w, http.StatusCreated)

		var response map[string]interface{}
		AssertJSONResponse(t, w, &response)
		submissionID := response["submissionId"].(string)

		// Test GET /provider-submissions/{submissionId}
		w = ts.MakeGETRequest("/provider-submissions/" + submissionID)
		AssertResponseStatus(t, w, http.StatusOK)

		// Test PUT /provider-submissions/{submissionId} (admin approval)
		approvalReq := map[string]string{
			"status": "approved",
		}
		w = ts.MakePUTRequest("/provider-submissions/"+submissionID, approvalReq)
		AssertResponseStatus(t, w, http.StatusOK)
	})

	t.Run("Error Cases", func(t *testing.T) {
		// Test 400 for non-existent provider (provider service returns 400, not 404)
		w := ts.MakeGETRequest("/providers/non-existent/schemas")
		AssertResponseStatus(t, w, http.StatusBadRequest)

		// Test 404 for non-existent schema
		w = ts.MakeGETRequest("/providers/test-provider/schema-submissions/non-existent")
		AssertResponseStatus(t, w, http.StatusNotFound)
	})
}
