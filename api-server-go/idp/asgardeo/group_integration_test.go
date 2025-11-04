package asgardeo

import (
	"context"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/idp"
)

func TestGetGroupIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL") // e.g. https://api.asgardeo.io/t/yourorg
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	testGroupID := os.Getenv("ASGARDEO_TEST_GROUP_ID")

	if clientID == "" || clientSecret == "" || baseURL == "" || testGroupID == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	group, err := client.GetGroup(ctx, testGroupID)
	if err != nil {
		t.Fatalf("GetGroup failed: %v", err)
	}

	if group.Id != testGroupID {
		t.Errorf("Expected group ID %s, got %s", testGroupID, group.Id)
	}
}

func TestCreateGroupIntegration(t *testing.T) {
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
		[]string{"internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	groupInstance := &idp.Group{
		DisplayName: "TestGroup1",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if createdGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, createdGroup.DisplayName)
	}

	// Clean up: delete the created group
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}

func TestCreateGroupWithMembersIntegration(t *testing.T) {
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
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       "hello@example.com",
		FirstName:   "Example",
		LastName:    "User (Group Addition Test)",
		PhoneNumber: "+1234567890",
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

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
		// Clean up user
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if createdGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, createdGroup.DisplayName)
	}

	// get the created group to verify members
	getCreatedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up user and group
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(getCreatedGroup.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(createdGroup.Members))
	}

	if len(createdGroup.Members) > 0 && createdGroup.Members[0].Value != createdUser.Id {
		t.Errorf("Expected member ID %s, got %s", createdUser.Id, createdGroup.Members[0].Value)
	}

	// Clean up: delete group and user
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	err = client.DeleteUser(ctx, createdUser.Id)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

func TestUpdateGroupIntegration(t *testing.T) {
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
		[]string{"internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// Step 1: Create a group
	groupInstance := &idp.Group{
		DisplayName: "UpdateTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	// Step 2: Update the group
	updatedGroupInstance := &idp.Group{
		DisplayName: "UpdatedTestGroup",
	}

	updatedGroup, err := client.UpdateGroup(ctx, createdGroup.Id, updatedGroupInstance)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	if updatedGroup.DisplayName != updatedGroupInstance.DisplayName {
		t.Errorf("Expected updated display name %s, got %s", updatedGroupInstance.DisplayName, updatedGroup.DisplayName)
	}

	// Step 3: Clean up
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}

func TestDeleteGroupIntegration(t *testing.T) {
	ctx := context.Background()

	baseURL := os.Getenv("ASGARDEO_BASE_URL")
	clientID := os.Getenv("ASGARDEO_CLIENT_ID")
	clientSecret := os.Getenv("ASGARDEO_CLIENT_SECRET")
	testGroupID := os.Getenv("ASGARDEO_TEST_GROUP_ID_TO_DELETE")

	if clientID == "" || clientSecret == "" || baseURL == "" || testGroupID == "" {
		t.Skip("Skipping integration test: missing Asgardeo environment variables")
	}

	client := NewClient(
		baseURL,
		clientID,
		clientSecret,
		[]string{"internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// create a group to delete
	groupInstance := &idp.Group{
		DisplayName: "DeleteTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	// Step: Delete Group
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}

func TestGroupLifecycleIntegration(t *testing.T) {
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
		[]string{"internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// Step 1: Create Group
	groupInstance := &idp.Group{
		DisplayName: "LifecycleTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}

	if createdGroup.DisplayName != groupInstance.DisplayName {
		t.Errorf("Expected group display name %s, got %s", groupInstance.DisplayName, createdGroup.DisplayName)
	}

	// Step 2: Get Group
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
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
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		t.Fatalf("UpdateGroup failed: %v", err)
	}

	if updatedGroup.DisplayName != updatedGroupInstance.DisplayName {
		t.Errorf("Expected updated display name %s, got %s", updatedGroupInstance.DisplayName, updatedGroup.DisplayName)
	}

	// Step 4: Delete Group
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}
}

func TestAddMemberToGroupIntegration(t *testing.T) {
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
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       "addmember@example.com",
		FirstName:   "Add",
		LastName:    "Member",
		PhoneNumber: "+1234567890",
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Step 2: Create a group
	groupInstance := &idp.Group{
		DisplayName: "AddMemberTestGroup",
	}

	createdGroup, err := client.CreateGroup(ctx, groupInstance)
	if err != nil {
		// Clean up user
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("CreateGroup failed: %v", err)
	}

	patchGroupMember := &idp.GroupMember{
		Value:   createdUser.Id,
		Display: createdUser.Email,
	}

	// Step 3: Add user to group
	err = client.AddMemberToGroup(ctx, createdGroup.Id, patchGroupMember)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("AddMemberToGroup failed: %v", err)
	}

	// Step 4: Verify member was added
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser.Id)
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

	// Clean up: delete group and user
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	err = client.DeleteUser(ctx, createdUser.Id)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

func TestRemoveMemberFromGroupIntegration(t *testing.T) {
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
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// Step 1: Create a test user
	userInstance := &idp.User{
		Email:       "removemember@example.com",
		FirstName:   "Remove",
		LastName:    "Member",
		PhoneNumber: "+1234567890",
	}

	createdUser, err := client.CreateUser(ctx, userInstance)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

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
		// Clean up user
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("CreateGroup failed: %v", err)
	}

	// Step 3: Remove user from group
	err = client.RemoveMemberFromGroup(ctx, createdGroup.Id, createdUser.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("RemoveMemberFromGroup failed: %v", err)
	}

	// Step 4: Verify member was removed
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser.Id)
		t.Fatalf("GetGroup failed: %v", err)
	}

	for _, member := range retrievedGroup.Members {
		if member.Value == createdUser.Id {
			t.Errorf("User %s should have been removed from group but was still found", createdUser.Id)
		}
	}

	// Clean up: delete group and user
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	err = client.DeleteUser(ctx, createdUser.Id)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}

func TestGroupWithMembersLifecycleIntegration(t *testing.T) {
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
		[]string{"internal_user_mgt_create internal_user_mgt_list internal_user_mgt_view internal_user_mgt_delete internal_user_mgt_update internal_group_mgt_create internal_group_mgt_list internal_group_mgt_view internal_group_mgt_delete internal_group_mgt_update"},
	)

	// Step 1: Create two test users
	user1Instance := &idp.User{
		Email:       "lifecyclemember1@example.com",
		FirstName:   "Lifecycle",
		LastName:    "Member1",
		PhoneNumber: "+1234567890",
	}

	createdUser1, err := client.CreateUser(ctx, user1Instance)
	if err != nil {
		t.Fatalf("CreateUser1 failed: %v", err)
	}

	user2Instance := &idp.User{
		Email:       "lifecyclemember2@example.com",
		FirstName:   "Lifecycle",
		LastName:    "Member2",
		PhoneNumber: "+9876543210",
	}

	createdUser2, err := client.CreateUser(ctx, user2Instance)
	if err != nil {
		// Clean up user1
		client.DeleteUser(ctx, createdUser1.Id)
		t.Fatalf("CreateUser2 failed: %v", err)
	}

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
		// Clean up users
		client.DeleteUser(ctx, createdUser1.Id)
		client.DeleteUser(ctx, createdUser2.Id)
		t.Fatalf("CreateGroup failed: %v", err)
	}

	getCreatedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up users and group
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser1.Id)
		client.DeleteUser(ctx, createdUser2.Id)
		t.Fatalf("GetGroup failed: %v", err)
	}

	// Step 3: Verify group has one member
	if len(getCreatedGroup.Members) != 1 {
		t.Errorf("Expected 1 member, got %d", len(getCreatedGroup.Members))
	}

	patchMember2 := &idp.GroupMember{
		Value:   createdUser2.Id,
		Display: createdUser2.Email,
	}

	// Step 4: Add second user to group
	err = client.AddMemberToGroup(ctx, createdGroup.Id, patchMember2)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser1.Id)
		client.DeleteUser(ctx, createdUser2.Id)
		t.Fatalf("AddMemberToGroup failed: %v", err)
	}

	// Step 5: Verify group has two members
	retrievedGroup, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser1.Id)
		client.DeleteUser(ctx, createdUser2.Id)
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(retrievedGroup.Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(retrievedGroup.Members))
	}

	// Step 6: Remove first user from group
	err = client.RemoveMemberFromGroup(ctx, createdGroup.Id, createdUser1.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser1.Id)
		client.DeleteUser(ctx, createdUser2.Id)
		t.Fatalf("RemoveMemberFromGroup failed: %v", err)
	}

	// Step 7: Verify group has one member
	retrievedGroup2, err := client.GetGroup(ctx, createdGroup.Id)
	if err != nil {
		// Clean up
		client.DeleteGroup(ctx, createdGroup.Id)
		client.DeleteUser(ctx, createdUser1.Id)
		client.DeleteUser(ctx, createdUser2.Id)
		t.Fatalf("GetGroup failed: %v", err)
	}

	if len(retrievedGroup2.Members) != 1 {
		t.Errorf("Expected 1 member after removal, got %d", len(retrievedGroup2.Members))
	}

	if len(retrievedGroup2.Members) > 0 && retrievedGroup2.Members[0].Value != createdUser2.Id {
		t.Errorf("Expected remaining member to be %s, got %s", createdUser2.Id, retrievedGroup2.Members[0].Value)
	}

	// Step 8: Clean up - delete group and users
	err = client.DeleteGroup(ctx, createdGroup.Id)
	if err != nil {
		t.Fatalf("DeleteGroup failed: %v", err)
	}

	err = client.DeleteUser(ctx, createdUser1.Id)
	if err != nil {
		t.Fatalf("DeleteUser1 failed: %v", err)
	}

	err = client.DeleteUser(ctx, createdUser2.Id)
	if err != nil {
		t.Fatalf("DeleteUser2 failed: %v", err)
	}
}
