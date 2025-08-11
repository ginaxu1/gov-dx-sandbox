// internal/repository/repository.go
package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"policy-governance/internal/database"
	"policy-governance/internal/models"
)

type PolicyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository() *PolicyRepository {
	return &PolicyRepository{db: database.DB}
}

// GetPolicy retrieves a policy mapping from the database.
func (r *PolicyRepository) GetPolicy(ctx context.Context, consumerID, providerID string) (*models.PolicyMapping, error) {
	var policy models.PolicyMapping
	result := r.db.WithContext(ctx).Where("consumer_id = ? AND provider_id = ?", consumerID, providerID).First(&policy)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No policy found, not an error
		}
		return nil, result.Error // Other database error
	}
	return &policy, nil
}