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
	// Create user in the IDP using the factory-created client
	ctx := context.Background()

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

	// Create Member in the database
	Member := models.Member{
		MemberID:    "ent_" + uuid.New().String(),
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		IdpUserID:   createdUser.Id,
	}

	if err := s.db.Create(&Member).Error; err != nil {
		// Rollback IDP user creation if DB operation fails (not implemented here)
		err := (*s.idp).DeleteUser(ctx, createdUser.Id)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to create Member: %w", err)
	}

	response := &models.MemberResponse{
		MemberID:    Member.MemberID,
		IdpUserID:   Member.IdpUserID,
		Name:        Member.Name,
		Email:       Member.Email,
		PhoneNumber: Member.PhoneNumber,
		CreatedAt:   Member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   Member.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// UpdateMember updates an existing Member
func (s *MemberService) UpdateMember(MemberID string, req *models.UpdateMemberRequest) (*models.MemberResponse, error) {
	var Member models.Member
	err := s.db.First(&Member, "Member_id = ?", MemberID).Error
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// Check if we need to update the IDP user
	needsIdpUpdate := false
	beforeUpdateName := Member.Name
	beforeUpdatePhoneNumber := Member.PhoneNumber

	// Update fields if provided
	if req.Name != nil {
		Member.Name = *req.Name
		needsIdpUpdate = true
	}
	if req.PhoneNumber != nil {
		Member.PhoneNumber = *req.PhoneNumber
		needsIdpUpdate = true
	}

	// Update user in IDP if necessary
	if needsIdpUpdate {
		ctx := context.Background()
		userInstance := &idp.User{
			Email:       Member.Email,
			FirstName:   Member.Name,
			LastName:    "",
			PhoneNumber: Member.PhoneNumber,
		}

		_, err := (*s.idp).UpdateUser(ctx, Member.IdpUserID, userInstance)
		if err != nil {
			return nil, fmt.Errorf("failed to update user in IDP: %w", err)
		}

		slog.Info("Updated user in IDP", "userID", Member.IdpUserID)
	}

	if err := s.db.Save(&Member).Error; err != nil {
		// Rollback IDP user update if DB operation fails (not implemented here)
		_, err := (*s.idp).UpdateUser(context.Background(), Member.IdpUserID, &idp.User{
			Email:       Member.Email,
			FirstName:   beforeUpdateName,
			LastName:    "",
			PhoneNumber: beforeUpdatePhoneNumber,
		})
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to update Member: %w", err)
	}

	response := &models.MemberResponse{
		MemberID:    Member.MemberID,
		IdpUserID:   Member.IdpUserID,
		Name:        Member.Name,
		Email:       Member.Email,
		PhoneNumber: Member.PhoneNumber,
		CreatedAt:   Member.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   Member.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetMember retrieves a Member by ID with associated provider/consumer information
func (s *MemberService) GetMember(MemberID string) (*models.MemberResponse, error) {
	var result struct {
		models.Member
	}
	err := s.db.Table("entities").
		Where("entities.Member_id = ?", MemberID).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch Member: %w", err)
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

// GetAllMembers retrieves all entities with their associated provider/consumer information
func (s *MemberService) GetAllMembers() ([]models.MemberResponse, error) {
	var results []struct {
		models.Member
	}

	err := s.db.Table("members").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %w", err)
	}

	response := make([]models.MemberResponse, len(results))
	for i, result := range results {
		response[i] = models.MemberResponse{
			MemberID:    result.Member.MemberID,
			IdpUserID:   result.Member.IdpUserID,
			Name:        result.Member.Name,
			Email:       result.Member.Email,
			PhoneNumber: result.Member.PhoneNumber,
			CreatedAt:   result.Member.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   result.Member.UpdatedAt.Format(time.RFC3339),
		}
	}

	return response, nil
}
