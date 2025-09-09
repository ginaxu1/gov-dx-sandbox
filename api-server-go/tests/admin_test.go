package services_test

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
)

func TestAdminService_GetDashboard(t *testing.T) {
	service := services.NewAdminService()

	// Create some test data
	consumerService := service.GetConsumerService()
	providerService := service.GetProviderService()

	// Create applications
	appReq := models.CreateApplicationRequest{
		RequiredFields: map[string]interface{}{"person.fullName": true},
	}
	_, err := consumerService.CreateApplication(appReq)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}

	// Create provider submission
	subReq := models.CreateProviderSubmissionRequest{
		ProviderName: "Test Department",
		ContactEmail: "test@example.com",
		PhoneNumber:  "1234567890",
		ProviderType: models.ProviderTypeGovernment,
	}
	_, err = providerService.CreateProviderSubmission(subReq)
	if err != nil {
		t.Fatalf("Failed to create provider submission: %v", err)
	}

	// Get dashboard
	dashboard, err := service.GetDashboard()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check dashboard structure
	overview, ok := dashboard["overview"].(map[string]interface{})
	if !ok {
		t.Error("Expected overview section in dashboard")
	}

	if overview["total_applications"].(int) != 1 {
		t.Errorf("Expected 1 application, got %v", overview["total_applications"])
	}

	if overview["total_submissions"].(int) != 1 {
		t.Errorf("Expected 1 submission, got %v", overview["total_submissions"])
	}

	// Check that recent activity exists
	activity, ok := dashboard["recent_activity"].([]map[string]interface{})
	if !ok {
		t.Error("Expected recent_activity section in dashboard")
	}

	if len(activity) == 0 {
		t.Error("Expected recent activity to have entries")
	}
}

func TestAdminService_GetDashboard_Empty(t *testing.T) {
	service := services.NewAdminService()

	dashboard, err := service.GetDashboard()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	overview, ok := dashboard["overview"].(map[string]interface{})
	if !ok {
		t.Error("Expected overview section in dashboard")
	}

	// All counts should be 0 for empty dashboard
	expectedCounts := []string{"total_applications", "total_submissions", "total_profiles", "total_schemas"}
	for _, countKey := range expectedCounts {
		if overview[countKey].(int) != 0 {
			t.Errorf("Expected %s to be 0, got %v", countKey, overview[countKey])
		}
	}
}
