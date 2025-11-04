package services

import (
	"github.com/gov-dx-sandbox/api-server-go/v1/models"
)

// PDPClient defines the interface for PDP service operations
// This allows us to use mock implementations in tests
type PDPClient interface {
	CreatePolicyMetadata(schemaID, sdl string) (*models.PolicyMetadataCreateResponse, error)
	UpdateAllowList(request models.AllowListUpdateRequest) (*models.AllowListUpdateResponse, error)
	HealthCheck() error
}

// Ensure PDPService implements PDPClient
var _ PDPClient = (*PDPService)(nil)
