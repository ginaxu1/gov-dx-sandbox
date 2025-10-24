package asgardeo

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/idp"
)

func TestGetApplicationInfoIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	testApplicationID := os.Getenv("ASGARDEO_TEST_APPLICATION_ID")

	if clientID == "" || clientSecret == "" || baseURL == "" || testApplicationID == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_application_mgt_create internal_application_mgt_delete internal_application_mgt_update internal_application_mgt_view"},
	)

	applicationInfo, err := client.GetApplicationInfo(ctx, testApplicationID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if applicationInfo.Id != testApplicationID {
		t.Errorf("Expected user ID %s, got %s", testApplicationID, applicationInfo.Id)
	}
}

//func TestCreateApplicationIntegration(t *testing.T) {
//	ctx := context.Background()
//
//	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
//	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
//	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
//
//	if clientID == "" || clientSecret == "" || baseURL == "" {
//		t.Skip("Skipping integration test: missing Asgardeo environment variables")
//	}
//
//	client := NewClient(
//		baseURL,
//		clientID,
//		clientSecret,
//		[]string{"internal_application_mgt_create internal_application_mgt_delete internal_application_mgt_update internal_application_mgt_view"},
//	)
//
//	appInstance := &idp.Application{
//		Name:        "TestApp",
//		Description: "This is a test application",
//		TemplateId:  "m2m-application",
//	}
//
//	applicationId, err := client.CreateApplication(ctx, appInstance)
//
//	if err != nil {
//		t.Fatalf("CreateApplication failed: %v", err)
//	}
//
//	if applicationId == nil || *applicationId == "" {
//		t.Errorf("Expected non-empty application ID, got %v", applicationId)
//	}
//}

func TestGetApplicationOIDCIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL")
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	testApplicationID := os.Getenv("ASGARDEO_TEST_APPLICATION_ID")

	if clientID == "" || clientSecret == "" || baseURL == "" || testApplicationID == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_application_mgt_create internal_application_mgt_delete internal_application_mgt_update internal_application_mgt_view"},
	)

	oidcInfo, err := client.GetApplicationOIDC(ctx, testApplicationID)
	if err != nil {
		t.Fatalf("GetApplicationOIDC failed: %v", err)
	}

	if oidcInfo.ClientId == "" || oidcInfo.ClientSecret == "" {
		t.Errorf("Expected non-empty ClientId and ClientSecret, got ClientId: %s, ClientSecret: %s", oidcInfo.ClientId, oidcInfo.ClientSecret)
	}
}

func TestApplicationLifecycleIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL")
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" || baseURL == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_application_mgt_create internal_application_mgt_delete internal_application_mgt_update internal_application_mgt_view"},
	)

	// Create Application
	appInstance := &idp.Application{
		Name:        "LifecycleTestApp",
		Description: "This is a test application for lifecycle",
		TemplateId:  "m2m-application",
	}

	appId, err := client.CreateApplication(ctx, appInstance)

	if err != nil {
		t.Fatalf("CreateApplication failed: %v", err)
	}

	// Delete Application
	err = client.DeleteApplication(ctx, *appId)
	if err != nil {
		t.Fatalf("DeleteApplication failed: %v", err)
	}
}
