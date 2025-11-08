package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"gorm.io/gorm"
)

// MemberService handles Member-related operations
type MemberService struct {
	db  *gorm.DB
	idp *idp.IdentityProviderAPI
}

// NewMemberService creates a new Member service
func NewMemberService(db *gorm.DB, idp *idp.IdentityProviderAPI) *MemberService {
	return &MemberService{db: db, idp: idp}
}

// CreateMember creates a new Member
func (s *MemberService) CreateMember(req *models.CreateMemberRequest) (*models.MemberResponse, error) {
	ctx := context.Background()

	// Create user in the IDP using the factory-created client
	userInstance := &idp.User{
		Email:       req.Email,
		FirstName:   req.Name,
		LastName:    "",
		PhoneNumber: req.PhoneNumber,
	}
	createdUser, err := (*s.idp).CreateUser(ctx, userInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in IDP: %w", err)
	}
	if createdUser.Email != userInstance.Email {
		return nil, fmt.Errorf("IDP user email mismatch")
	}
	slog.Info("Created user in IDP", "userID", createdUser.Id)

	// Add user to UserGroupMember in the IDP
	patchGroupMember := &idp.GroupMember{
		Value:   createdUser.Id,
		Display: createdUser.Email,
	}
	groupId, err := (*s.idp).AddMemberToGroupByGroupName(ctx, string(models.UserGroupMember), patchGroupMember)
	if err != nil {
		deleteErr := (*s.idp).DeleteUser(ctx, createdUser.Id)
		if deleteErr != nil {
			return nil, fmt.Errorf("failed to rollback user creation in IDP: %w", deleteErr)
		}
		return nil, fmt.Errorf("failed to add user to group in IDP: %w", err)
	}

	// Create Member in the database
	member := models.Member{
		MemberID:    "mem_" + uuid.New().String(),
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IdpUserID:   createdUser.Id,
	}
	if err := s.db.Create(&member).Error; err != nil {
		// Delete user from IDP group if adding to DB fails
		removeErr := (*s.idp).RemoveMemberFromGroup(ctx, *groupId, createdUser.Id)
		if removeErr != nil {
			return nil, fmt.Errorf("failed to rollback group membership in IDP: %w", removeErr)
		}
		// Rollback IDP user creation if DB operation fails (not implemented here)
		err := (*s.idp).DeleteUser(ctx, createdUser.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to rollback user creation in IDP: %w", err)
		}
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	response := &models.MemberResponse{
		MemberID:    member.MemberID,
		IdpUserID:   member.IdpUserID,
		Name:        member.Name,
		Email:       member.Email,
		PhoneNumber: member.PhoneNumber,
		CreatedAt:   member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   member.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateMember updates an existing Member
func (s *MemberService) UpdateMember(memberID string, req *models.UpdateMemberRequest) (*models.MemberResponse, error) {
	var member models.Member
	err := s.db.First(&member, "member_id = ?", memberID).Error
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// Check if we need to update the IDP user
	needsIdpUpdate := false
	beforeUpdateName := member.Name
	beforeUpdatePhoneNumber := member.PhoneNumber

	// Update fields if provided
	if req.Name != nil {
		member.Name = *req.Name
		needsIdpUpdate = true
	}
	if req.PhoneNumber != nil {
		member.PhoneNumber = *req.PhoneNumber
		needsIdpUpdate = true
	}

	// Update user in IDP if necessary
	if needsIdpUpdate {
		ctx := context.Background()
		userInstance := &idp.User{
			Email:       member.Email,
			FirstName:   member.Name,
			LastName:    "",
			PhoneNumber: member.PhoneNumber,
		}

		_, err := (*s.idp).UpdateUser(ctx, member.IdpUserID, userInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to update user in IDP: %w", err)
		}

		slog.Info("Updated user in IDP", "userID", member.IdpUserID)
	}

	if err := s.db.Save(&member).Error; err != nil {
		// Rollback IDP user update if DB operation fails (not implemented here)
		_, err := (*s.idp).UpdateUser(context.Background(), member.IdpUserID, &idp.User{
			Email:       member.Email,
			FirstName:   beforeUpdateName,
			LastName:    "",
			PhoneNumber: beforeUpdatePhoneNumber,
		})
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	response := &models.MemberResponse{
		MemberID:    member.MemberID,
		IdpUserID:   member.IdpUserID,
		Name:        member.Name,
		Email:       member.Email,
		PhoneNumber: member.PhoneNumber,
		CreatedAt:   member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   member.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetMember retrieves a Member by ID
func (s *MemberService) GetMember(memberID string) (*models.MemberResponse, error) {
	var result struct {
		models.Member
	}
	err := s.db.Table("members").
		Where("members.member_id = ?", memberID).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch member: %w", err)
	}

	response := &models.MemberResponse{
		MemberID:    result.Member.MemberID,
		IdpUserID:   result.Member.IdpUserID,
		Name:        result.Member.Name,
		Email:       result.Member.Email,
		PhoneNumber: result.Member.PhoneNumber,
		CreatedAt:   result.Member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   result.Member.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetAllMembers retrieves all members
func (s *MemberService) GetAllMembers(idpUserId *string, email *string) ([]models.MemberResponse, error) {
	if (idpUserId != nil && *idpUserId != "") || (email != nil && *email != "") {
		var member models.Member
		query := s.db.Table("members")
		if idpUserId != nil && *idpUserId != "" {
			query = query.Where("idp_user_id = ?", *idpUserId)
		}
		if email != nil && *email != "" {
			query = query.Where("email = ?", *email)
		}
		err := query.First(&member).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch member: %w", err)
		}

		response := []models.MemberResponse{{
			MemberID:    member.MemberID,
			IdpUserID:   member.IdpUserID,
			Name:        member.Name,
			Email:       member.Email,
			PhoneNumber: member.PhoneNumber,
			CreatedAt:   member.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   member.UpdatedAt.Format(time.RFC3339),
		}}

		return response, nil
	}
	var results []models.Member

	err := s.db.Table("members").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members: %w", err)
	}

	response := make([]models.MemberResponse, len(results))
	for i, result := range results {
		response[i] = models.MemberResponse{
			MemberID:    result.MemberID,
			IdpUserID:   result.IdpUserID,
			Name:        result.Name,
			Email:       result.Email,
			PhoneNumber: result.PhoneNumber,
			CreatedAt:   result.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   result.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}
