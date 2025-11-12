package services

import (
	"context"
	"errors"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockIdentityProviderAPI is a mock implementation of idp.IdentityProviderAPI
type MockIdentityProviderAPI struct {
	mock.Mock
}

func (m *MockIdentityProviderAPI) CreateUser(ctx context.Context, user *idp.User) (*idp.UserInfo, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.UserInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) UpdateUser(ctx context.Context, userID string, user *idp.User) (*idp.UserInfo, error) {
	args := m.Called(ctx, userID, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.UserInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockIdentityProviderAPI) GetUser(ctx context.Context, userID string) (*idp.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.UserInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) AddMemberToGroupByGroupName(ctx context.Context, groupName string, member *idp.GroupMember) (*string, error) {
	args := m.Called(ctx, groupName, member)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if groupId, ok := args.Get(0).(string); ok {
		return &groupId, args.Error(1)
	}
	if groupIdPtr, ok := args.Get(0).(*string); ok {
		return groupIdPtr, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockIdentityProviderAPI) RemoveMemberFromGroup(ctx context.Context, groupID string, userID string) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

func (m *MockIdentityProviderAPI) GetGroup(ctx context.Context, groupID string) (*idp.GroupInfo, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.GroupInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) GetGroupByName(ctx context.Context, groupName string) (*string, error) {
	args := m.Called(ctx, groupName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	groupId := args.Get(0).(string)
	return &groupId, args.Error(1)
}

func (m *MockIdentityProviderAPI) CreateGroup(ctx context.Context, group *idp.Group) (*idp.GroupInfo, error) {
	args := m.Called(ctx, group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.GroupInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) UpdateGroup(ctx context.Context, groupID string, group *idp.Group) (*idp.GroupInfo, error) {
	args := m.Called(ctx, groupID, group)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.GroupInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) AddMemberToGroup(ctx context.Context, groupID string, memberInfo *idp.GroupMember) error {
	args := m.Called(ctx, groupID, memberInfo)
	return args.Error(0)
}

func (m *MockIdentityProviderAPI) CreateApplication(ctx context.Context, app *idp.Application) (*string, error) {
	args := m.Called(ctx, app)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	appId := args.Get(0).(string)
	return &appId, args.Error(1)
}

func (m *MockIdentityProviderAPI) DeleteApplication(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockIdentityProviderAPI) DeleteGroup(ctx context.Context, groupID string) error {
	args := m.Called(ctx, groupID)
	return args.Error(0)
}

func (m *MockIdentityProviderAPI) GetApplicationInfo(ctx context.Context, applicationID string) (*idp.ApplicationInfo, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.ApplicationInfo), args.Error(1)
}

func (m *MockIdentityProviderAPI) GetApplicationOIDC(ctx context.Context, applicationID string) (*idp.ApplicationOIDCInfo, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*idp.ApplicationOIDCInfo), args.Error(1)
}

func TestMemberService_CreateMember(t *testing.T) {
	t.Run("CreateMember_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		req := &models.CreateMemberRequest{
			Name:        "Test User",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}

		// Mock IDP responses
		createdUser := &idp.UserInfo{
			Id:    "idp-user-123",
			Email: "test@example.com",
		}
		groupId := "group-123"

		mockIDP.On("CreateUser", ctx, mock.AnythingOfType("*idp.User")).Return(createdUser, nil)
		mockIDP.On("AddMemberToGroupByGroupName", ctx, string(models.UserGroupMember), mock.AnythingOfType("*idp.GroupMember")).Return(&groupId, nil)

		result, err := service.CreateMember(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Email, result.Email)
		assert.Equal(t, req.PhoneNumber, result.PhoneNumber)
		assert.NotEmpty(t, result.MemberID)
		assert.Equal(t, createdUser.Id, result.IdpUserID)

		// Verify member was created in database
		var member models.Member
		err = db.Where("email = ?", req.Email).First(&member).Error
		assert.NoError(t, err)
		assert.Equal(t, req.Email, member.Email)

		mockIDP.AssertExpectations(t)
	})

	t.Run("CreateMember_IDPCreateUserFails", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		req := &models.CreateMemberRequest{
			Name:        "Test User",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}

		mockIDP.On("CreateUser", ctx, mock.AnythingOfType("*idp.User")).Return(nil, errors.New("IDP error"))

		result, err := service.CreateMember(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create user in IDP")

		mockIDP.AssertExpectations(t)
	})

	t.Run("CreateMember_EmailMismatch", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		req := &models.CreateMemberRequest{
			Name:        "Test User",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}

		// Mock IDP returns different email
		createdUser := &idp.UserInfo{
			Id:    "idp-user-123",
			Email: "different@example.com",
		}

		mockIDP.On("CreateUser", ctx, mock.AnythingOfType("*idp.User")).Return(createdUser, nil)
		mockIDP.On("DeleteUser", ctx, createdUser.Id).Return(nil)

		result, err := service.CreateMember(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "IDP user email mismatch")

		mockIDP.AssertExpectations(t)
	})

	t.Run("CreateMember_AddToGroupFails", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		req := &models.CreateMemberRequest{
			Name:        "Test User",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}

		createdUser := &idp.UserInfo{
			Id:    "idp-user-123",
			Email: "test@example.com",
		}

		mockIDP.On("CreateUser", ctx, mock.AnythingOfType("*idp.User")).Return(createdUser, nil)
		mockIDP.On("AddMemberToGroupByGroupName", ctx, string(models.UserGroupMember), mock.AnythingOfType("*idp.GroupMember")).Return(nil, errors.New("group error"))
		mockIDP.On("DeleteUser", ctx, createdUser.Id).Return(nil)

		result, err := service.CreateMember(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to add user to group")

		mockIDP.AssertExpectations(t)
	})

	t.Run("CreateMember_DatabaseCreateFails", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		req := &models.CreateMemberRequest{
			Name:        "Test User",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}

		createdUser := &idp.UserInfo{
			Id:    "idp-user-123",
			Email: "test@example.com",
		}
		groupId := "group-123"

		mockIDP.On("CreateUser", ctx, mock.AnythingOfType("*idp.User")).Return(createdUser, nil)
		mockIDP.On("AddMemberToGroupByGroupName", ctx, string(models.UserGroupMember), mock.AnythingOfType("*idp.GroupMember")).Return(&groupId, nil)
		mockIDP.On("RemoveMemberFromGroup", ctx, groupId, createdUser.Id).Return(nil)
		mockIDP.On("DeleteUser", ctx, createdUser.Id).Return(nil)

		// Create a duplicate member to cause database error
		duplicateMember := models.Member{
			MemberID:    "mem_duplicate",
			Email:       "test@example.com", // Same email
			Name:        "Duplicate",
			PhoneNumber: "1234567890",
			IdpUserID:   "other-idp-id",
		}
		db.Create(&duplicateMember)

		result, err := service.CreateMember(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create member in database")

		mockIDP.AssertExpectations(t)
	})
}

func TestMemberService_UpdateMember(t *testing.T) {
	t.Run("UpdateMember_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		// Create a member first
		member := models.Member{
			MemberID:    "mem_123",
			Name:        "Original Name",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
			IdpUserID:   "idp-user-123",
		}
		db.Create(&member)

		ctx := context.Background()
		newName := "Updated Name"
		newPhone := "9876543210"
		req := &models.UpdateMemberRequest{
			Name:        &newName,
			PhoneNumber: &newPhone,
		}

		updatedUser := &idp.UserInfo{
			Id:    "idp-user-123",
			Email: "test@example.com",
		}

		mockIDP.On("UpdateUser", ctx, member.IdpUserID, mock.AnythingOfType("*idp.User")).Return(updatedUser, nil)

		result, err := service.UpdateMember(ctx, member.MemberID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.Name)
		assert.Equal(t, newPhone, result.PhoneNumber)

		// Verify database was updated
		var updatedMember models.Member
		db.Where("member_id = ?", member.MemberID).First(&updatedMember)
		assert.Equal(t, newName, updatedMember.Name)
		assert.Equal(t, newPhone, updatedMember.PhoneNumber)

		mockIDP.AssertExpectations(t)
	})

	t.Run("UpdateMember_NotFound", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		newName := "Updated Name"
		req := &models.UpdateMemberRequest{
			Name: &newName,
		}

		result, err := service.UpdateMember(ctx, "non-existent-id", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "member not found")
	})

	t.Run("UpdateMember_IDPUpdateFails", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		// Create a member first
		member := models.Member{
			MemberID:    "mem_123",
			Name:        "Original Name",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
			IdpUserID:   "idp-user-123",
		}
		db.Create(&member)

		ctx := context.Background()
		newName := "Updated Name"
		req := &models.UpdateMemberRequest{
			Name: &newName,
		}

		mockIDP.On("UpdateUser", ctx, member.IdpUserID, mock.AnythingOfType("*idp.User")).Return(nil, errors.New("IDP error"))

		result, err := service.UpdateMember(ctx, member.MemberID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to update user in IDP")
	})
}

func TestMemberService_GetMember(t *testing.T) {
	t.Run("GetMember_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		// Create a member
		member := models.Member{
			MemberID:    "mem_123",
			Name:        "Test User",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
			IdpUserID:   "idp-user-123",
		}
		db.Create(&member)

		ctx := context.Background()
		result, err := service.GetMember(ctx, member.MemberID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, member.MemberID, result.MemberID)
		assert.Equal(t, member.Name, result.Name)
		assert.Equal(t, member.Email, result.Email)
	})

	t.Run("GetMember_NotFound", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		ctx := context.Background()
		result, err := service.GetMember(ctx, "non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to fetch member")
	})
}

func TestMemberService_GetAllMembers(t *testing.T) {
	t.Run("GetAllMembers_NoFilters", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		// Create multiple members
		members := []models.Member{
			{MemberID: "mem_1", Name: "User 1", Email: "user1@example.com", PhoneNumber: "111", IdpUserID: "idp1"},
			{MemberID: "mem_2", Name: "User 2", Email: "user2@example.com", PhoneNumber: "222", IdpUserID: "idp2"},
		}
		for _, m := range members {
			db.Create(&m)
		}

		ctx := context.Background()
		result, err := service.GetAllMembers(ctx, nil, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("GetAllMembers_WithEmailFilter", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		// Create members
		member := models.Member{
			MemberID:    "mem_1",
			Name:        "User 1",
			Email:       "user1@example.com",
			PhoneNumber: "111",
			IdpUserID:   "idp1",
		}
		db.Create(&member)

		ctx := context.Background()
		email := "user1@example.com"
		result, err := service.GetAllMembers(ctx, nil, &email)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, email, result[0].Email)
	})

	t.Run("GetAllMembers_WithIdpUserIDFilter", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		mockIDP := new(MockIdentityProviderAPI)
		service := NewMemberService(db, mockIDP)

		// Create members
		member := models.Member{
			MemberID:    "mem_1",
			Name:        "User 1",
			Email:       "user1@example.com",
			PhoneNumber: "111",
			IdpUserID:   "idp-user-123",
		}
		db.Create(&member)

		ctx := context.Background()
		idpUserID := "idp-user-123"
		result, err := service.GetAllMembers(ctx, &idpUserID, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, idpUserID, result[0].IdpUserID)
	})
}
