package asgardeo

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
)

// Test constants and helper functions
const (
	testUserEmail1   = "hello@example.com"
	testUserEmail2   = "addmember@example.com"
	testUserEmail3   = "removemember@example.com"
	testUserEmail4   = "lifecyclemember1@example.com"
	testUserEmail5   = "lifecyclemember2@example.com"
	testPhoneNumber1 = "+1234567890"
	testPhoneNumber2 = "+9876543210"
)

var (
	groupScopes = []string{"internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"}
	userScopes  = []string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"}
)

// setupTestClient creates a test client with proper environment validation
func setupTestClient(t *testing.T, scopes []string) *Client {
	baseURL := os.Getenv("ASGARDEO_BASE_URL")
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" || baseURL == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	return NewClient(baseURL, clientID, clientSecret, scopes)
}

// cleanupResources handles cleanup of test resources with proper error handling
func cleanupResources(t *testing.T, client *Client, ctx context.Context, groupIDs []string, userIDs []string) {
	// Clean up groups first
	for _, groupID := range groupIDs {
		if err := client.DeleteGroup(ctx, groupID); err != nil {
			t.Errorf("Failed to cleanup group %s: %v", groupID, err)
		}
	}

	// Clean up users
	for _, userID := range userIDs {
		if err := client.DeleteUser(ctx, userID); err != nil {
			t.Errorf("Failed to cleanup user %s: %v", userID, err)
		}
	}
}

func TestGetGroupIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, groupScopes)

	testGroupID := os.Getenv("ASGARDEO_TEST_GROUP_ID")
	if testGroupID == "" {
		t.Skip("Skipping test: ASGARDEO_TEST_GROUP_ID not set")
	}

	group, err := client.GetGroup(ctx, testGroupID)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if group.Id != testGroupID {
		t.Errorf("Expected group ID %s, got %s", testGroupID, group.Id)
	}
}

func TestGetGroupByNameIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, groupScopes)

	testGroupName := os.Getenv("ASGARDEO_TEST_GROUP_NAME")
	if testGroupName == "" {
		t.Skip("Skipping test: ASGARDEO_TEST_GROUP_NAME not set")
	}

	groupID, err := client.GetGroupByName(ctx, testGroupName)
	if err != nil {
		t.Fatalf("GetGroupByName failed: %v", err)
	}

	if groupID == nil {
		t.Fatalf("GetGroupByName returned nil group ID")
	}

	groupObj, err := client.GetGroup(ctx, *groupID)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if groupObj.DisplayName != testGroupName {
		t.Errorf("Expected group name %s, got %s", testGroupName, groupObj.DisplayName)
	}
}

func TestCreateGroupIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, groupScopes)

	groupInstance := &idp.Group{
		DisplayName: "TestGroup1",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	if createdGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, createdGroup.DisplayName)
	}
}

func TestCreateGroupWithMembersIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, userScopes)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       testUserEmail1,
		FirstName:   "Example",
		LastName:    "User (Group Addition Test)",
		PhoneNumber: testPhoneNumber1,
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, nil, []string{createdUser.Id})

	// Step 2: Create group with the user as a member
	groupInstance := &idp.Group{
		DisplayName: "TestGroupWithMembers",
		Members: []*idp.GroupMember{
			{
				Value:   createdUser.Id,
				Display: createdUser.Email,
			},
		},
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	if createdGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, createdGroup.DisplayName)
	}

	// Verify the created group has the expected member
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(retrievedGroup.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(retrievedGroup.Members))
	}

	if len(retrievedGroup.Members) > 0 && retrievedGroup.Members[0].Value != createdUser.Id {
		t.Errorf("Expected member ID %s, got %s", createdUser.Id, retrievedGroup.Members[0].Value)
	}
}

func TestUpdateGroupIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, groupScopes)

	// Step 1: Create a group
	groupInstance := &idp.Group{
		DisplayName: "UpdateTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	// Step 2: Update the group
	updatedGroupInstance := &idp.Group{
		DisplayName: "UpdatedTestGroup",
	}

	updatedGroup, err := client.UpdateGroup(ctx, createdGroup.Id, updatedGroupInstance)
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	if updatedGroup.DisplayName != updatedGroupInstance.DisplayName {
		t.Errorf("Expected updated display name %s, got %s", updatedGroupInstance.DisplayName, updatedGroup.DisplayName)
	}
}

func TestDeleteGroupIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, groupScopes)

	groupInstance := &idp.Group{
		DisplayName: "DeleteTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	// Test: Delete Group
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}

func TestGroupLifecycleIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, groupScopes)

	// Step 1: Create Group
	groupInstance := &idp.Group{
		DisplayName: "LifecycleTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	if createdGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, createdGroup.DisplayName)
	}

	// Step 2: Get Group
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if retrievedGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, retrievedGroup.DisplayName)
	}

	// Step 3: Update Group
	updatedGroupInstance := &idp.Group{
		DisplayName: "UpdatedLifecycleTestGroup",
	}

	updatedGroup, err := client.UpdateGroup(ctx, createdGroup.Id, updatedGroupInstance)
	if err != nil {
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	if updatedGroup.DisplayName != updatedGroupInstance.DisplayName {
		t.Errorf("Expected updated display name %s, got %s", updatedGroupInstance.DisplayName, updatedGroup.DisplayName)
	}

	// Step 4: Delete Group (handled by defer cleanup)
}

func TestAddMemberToGroupIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, userScopes)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       testUserEmail2,
		FirstName:   "Add",
		LastName:    "Member",
		PhoneNumber: testPhoneNumber1,
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, nil, []string{createdUser.Id})

	// Step 2: Create a group
	groupInstance := &idp.Group{
		DisplayName: "AddMemberTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	patchGroupMember := &idp.GroupMember{
		Value:   createdUser.Id,
		Display: createdUser.Email,
	}

	// Step 3: Add user to group
	err = client.AddMemberToGroup(ctx, &createdGroup.Id, patchGroupMember)
	if err != nil {
		t.Fatalf("AddMemberToGroup failed: %v", err)
	}

	// Step 4: Verify member was added
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	found := false
	for _, member := range retrievedGroup.Members {
		if member.Value == createdUser.Id {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("User %s was not found in group members", createdUser.Id)
	}
}

func TestAddMemberToGroupByGroupNameIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, userScopes)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       testUserEmail2,
		FirstName:   "Add",
		LastName:    "Member",
		PhoneNumber: testPhoneNumber1,
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, nil, []string{createdUser.Id})

	patchGroupMember := &idp.GroupMember{
		Value:   createdUser.Id,
		Display: createdUser.Email,
	}

	// Step 2: Add user to group by group name
	groupName := string(models.UserGroupMember)
	groupId, err := client.AddMemberToGroupByGroupName(ctx, groupName, patchGroupMember)
	if err != nil {
		t.Fatalf("AddMemberToGroupByGroupName failed: %v", err)
	}

	// Step 3: Verify member was added
	retrievedGroup, err := client.GetGroup(ctx, *groupId)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	found := false
	for _, member := range retrievedGroup.Members {
		if member.Value == createdUser.Id {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("User %s was not found in group members", createdUser.Id)
	}
}

func TestRemoveMemberFromGroupIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, userScopes)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       testUserEmail3,
		FirstName:   "Remove",
		LastName:    "Member",
		PhoneNumber: testPhoneNumber1,
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, nil, []string{createdUser.Id})

	// Step 2: Create a group with the user as a member
	groupInstance := &idp.Group{
		DisplayName: "RemoveMemberTestGroup",
		Members: []*idp.GroupMember{{
			Value:   createdUser.Id,
			Display: createdUser.Email,
		}},
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	// Step 3: Remove user from group
	err = client.RemoveMemberFromGroup(ctx, createdGroup.Id, createdUser.Id)
	if err != nil {
		t.Fatalf("RemoveMemberFromGroup failed: %v", err)
	}

	// Step 4: Verify member was removed
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	for _, member := range retrievedGroup.Members {
		if member.Value == createdUser.Id {
			t.Errorf("User %s should have been removed from group but was still found", createdUser.Id)
		}
	}
}

func TestGroupWithMembersLifecycleIntegration(t *testing.T) {
	ctx := context.Background()
	client := setupTestClient(t, userScopes)

	// Step 1: Create two test users
	user1Instance := &idp.User{
		Email:       testUserEmail4,
		FirstName:   "Lifecycle",
		LastName:    "Member1",
		PhoneNumber: testPhoneNumber1,
	}

	createdUser1, err := client.CreateUser(ctx, user1Instance)
	if err != nil {
		t.Fatalf("CreateUser1 failed: %v", err)
	}

	user2Instance := &idp.User{
		Email:       testUserEmail5,
		FirstName:   "Lifecycle",
		LastName:    "Member2",
		PhoneNumber: testPhoneNumber2,
	}

	createdUser2, err := client.CreateUser(ctx, user2Instance)
	if err != nil {
		t.Fatalf("CreateUser2 failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, nil, []string{createdUser1.Id, createdUser2.Id})

	// Step 2: Create group with first user
	groupInstance := &idp.Group{
		DisplayName: "LifecycleGroupWithMembers",
		Members: []*idp.GroupMember{{
			Value:   createdUser1.Id,
			Display: createdUser1.Email,
		}},
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	defer cleanupResources(t, client, ctx, []string{createdGroup.Id}, nil)

	// Step 3: Verify group has one member
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(retrievedGroup.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(retrievedGroup.Members))
	}

	// Step 4: Add second user to group
	patchMember2 := &idp.GroupMember{
		Value:   createdUser2.Id,
		Display: createdUser2.Email,
	}

	err = client.AddMemberToGroup(ctx, &createdGroup.Id, patchMember2)
	if err != nil {
		t.Fatalf("AddMemberToGroup failed: %v", err)
	}

	// Step 5: Verify group has two members
	retrievedGroup, err = client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(retrievedGroup.Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(retrievedGroup.Members))
	}

	// Step 6: Remove first user from group
	err = client.RemoveMemberFromGroup(ctx, createdGroup.Id, createdUser1.Id)
	if err != nil {
		t.Fatalf("RemoveMemberFromGroup failed: %v", err)
	}

	// Step 7: Verify group has one member
	retrievedGroup, err = client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(retrievedGroup.Members) != 1 {
		t.Errorf("Expected 1 member after removal, got %d", len(retrievedGroup.Members))
	}

	if len(retrievedGroup.Members) > 0 && retrievedGroup.Members[0].Value != createdUser2.Id {
		t.Errorf("Expected remaining member to be %s, got %s", createdUser2.Id, retrievedGroup.Members[0].Value)
	}
}
