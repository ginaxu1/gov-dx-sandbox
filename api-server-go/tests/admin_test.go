package tests

import (
	"net/http"
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
		RequiredFields: map[string]bool{"person.fullName": true},
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
	metrics, err := service.GetMetrics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check metrics structure
	if len(metrics) == 0 {
		t.Error("Expected metrics to have data")
	}

	if metrics["total_applications"].(int) != 1 {
		t.Errorf("Expected 1 application, got %v", metrics["total_applications"])
	}

	if metrics["total_submissions"].(int) != 1 {
		t.Errorf("Expected 1 submission, got %v", metrics["total_submissions"])
	}

	// Check that metrics has expected fields
	if _, ok := metrics["total_applications"]; !ok {
		t.Error("Expected total_applications in metrics")
	}
	if _, ok := metrics["total_submissions"]; !ok {
		t.Error("Expected total_submissions in metrics")
	}
}

func TestAdminService_GetMetrics_Empty(t *testing.T) {
	service := services.NewAdminService()

	metrics, err := service.GetMetrics()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected metrics to have data")
	}

	// All counts should be 0 for empty metrics
	expectedCounts := []string{"total_applications", "total_submissions", "total_profiles", "total_schemas"}
	for _, countKey := range expectedCounts {
		if metrics[countKey].(int) != 0 {
			t.Errorf("Expected %s to be 0, got %v", countKey, metrics[countKey])
		}
	}
}

// HTTP endpoint tests for admin resources
func TestAdminEndpoints(t *testing.T) {
	ts := NewTestServer()

	t.Run("Admin Resources", func(t *testing.T) {
		// Test GET /admin/metrics
		w := ts.MakeGETRequest("/admin/metrics")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test GET /admin/recent-activity
		w = ts.MakeGETRequest("/admin/recent-activity")
		AssertResponseStatus(t, w, http.StatusOK)

		// Test GET /admin/statistics
		w = ts.MakeGETRequest("/admin/statistics")
		AssertResponseStatus(t, w, http.StatusOK)
	})
}
