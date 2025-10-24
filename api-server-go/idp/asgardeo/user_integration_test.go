package asgardeo

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/idp"
)

func TestGetUserIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	testUserID := os.Getenv("ASGARDEO_TEST_USER_ID")

	if clientID == "" || clientSecret == "" || baseURL == "" || testUserID == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update"},
	)

	user, err := client.GetUser(ctx, testUserID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.Id != testUserID {
		t.Errorf("Expected user ID %s, got %s", testUserID, user.Id)
	}
}

func TestCreateUserIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" || baseURL == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update"},
	)

	userInstance := &idp.User{
		Email:     "testuser1@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if createdUser.Email != userInstance.Email {
		t.Errorf("Expected user email %s, got %s", userInstance.Email, createdUser.Email)
	}

	// delete the created user
	err = client.DeleteUser(ctx, createdUser.Id)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

func TestDeleteUserIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	testUserID := os.Getenv("ASGARDEO_TEST_USER_ID_TO_DELETE")

	if clientID == "" || clientSecret == "" || baseURL == "" || testUserID == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update"},
	)

	err := client.DeleteUser(ctx, testUserID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

// a common function that creates a user, then checks if it exists, then deletes it
func TestUserLifecycleIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" || baseURL == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update"},
	)

	// Step 1: Create User
	userInstance := &idp.User{
		Email:     "testuser@example.com",
		FirstName: "Test",
		LastName:  "User",
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Step 2: Get User
	retrievedUser, err := client.GetUser(ctx, createdUser.Id)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrievedUser.Email != userInstance.Email {
		t.Errorf("Expected user email %s, got %s", userInstance.Email, retrievedUser.Email)
	}

	// Step 3: Delete User
	err = client.DeleteUser(ctx, createdUser.Id)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}
