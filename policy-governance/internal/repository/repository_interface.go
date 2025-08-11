// internal/repository/repository_interface.go
package repository

import (
	"context"
	"policy-governance/internal/models"
)

// PolicyRepositoryInterface defines the methods that the policy repository must implement.
type PolicyRepositoryInterface interface {
	GetPolicy(ctx context.Context, consumerID, providerID string) (*models.PolicyMapping, error)
}