package services

import (
	"database/sql"

	"github.com/gov-dx-sandbox/api-server-go/models"
)

// Service interfaces for dependency injection

type ConsumerServiceInterface interface {
	CreateConsumer(req models.CreateConsumerRequest) (*models.Consumer, error)
	GetConsumer(id string) (*models.Consumer, error)
	GetAllConsumers() ([]*models.Consumer, error)
	UpdateConsumer(id string, req models.UpdateConsumerRequest) (*models.Consumer, error)
	DeleteConsumer(id string) error
	CreateConsumerApp(req models.CreateConsumerAppRequest) (*models.ConsumerApp, error)
	GetConsumerApp(id string) (*models.ConsumerApp, error)
	UpdateConsumerApp(id string, req models.UpdateConsumerAppRequest) (*models.UpdateConsumerAppResponse, error)
	GetAllConsumerApps() ([]*models.ConsumerApp, error)
	GetConsumerAppsByConsumerID(consumerID string) ([]*models.ConsumerApp, error)
	// Legacy methods
	GetAllApplications() ([]*models.Application, error)
	CreateApplication(req models.CreateApplicationRequest) (*models.Application, error)
	GetApplication(id string) (*models.Application, error)
	UpdateApplication(id string, req models.UpdateApplicationRequest) (*models.UpdateApplicationResponse, error)
	DeleteApplication(id string) error
}

type ProviderServiceInterface interface {
	GetAllProviderSubmissions() ([]*models.ProviderSubmission, error)
	GetProviderSubmissionsByStatus(status string) ([]*models.ProviderSubmission, error)
	CreateProviderSubmission(req models.CreateProviderSubmissionRequest) (*models.ProviderSubmission, error)
	GetProviderSubmission(id string) (*models.ProviderSubmission, error)
	UpdateProviderSubmission(id string, req models.UpdateProviderSubmissionRequest) (*models.UpdateProviderSubmissionResponse, error)
	GetAllProviderProfiles() ([]*models.ProviderProfile, error)
	GetProviderProfile(id string) (*models.ProviderProfile, error)
	GetAllProviderSchemas() ([]*models.ProviderSchema, error)
	CreateProviderSchema(req models.CreateProviderSchemaRequest) (*models.ProviderSchema, error)
	CreateProviderSchemaSDL(providerID string, req models.CreateProviderSchemaSDLRequest) (*models.ProviderSchema, error)
	CreateProviderSchemaSubmission(providerID string, req models.CreateProviderSchemaSubmissionRequest) (*models.ProviderSchema, error)
	SubmitSchemaForReview(schemaID string) (*models.ProviderSchema, error)
	GetApprovedSchemasByProviderID(providerID string) ([]*models.ProviderSchema, error)
	GetProviderSchemasByProviderID(providerID string) ([]*models.ProviderSchema, error)
	GetProviderSchema(id string) (*models.ProviderSchema, error)
	UpdateProviderSchema(id string, req models.UpdateProviderSchemaRequest) (*models.ProviderSchema, error)
	CreateProviderProfileForTesting(providerName, contactEmail, phoneNumber, providerType string) (*models.ProviderProfile, error)
}

type GrantsServiceInterface interface {
	GetAllConsumerGrants() (*models.ConsumerGrantsData, error)
	GetConsumerGrant(consumerID string) (*models.ConsumerGrant, error)
	CreateConsumerGrant(req models.CreateConsumerGrantRequest) (*models.ConsumerGrant, error)
	UpdateConsumerGrant(consumerID string, req models.UpdateConsumerGrantRequest) (*models.ConsumerGrant, error)
	DeleteConsumerGrant(consumerID string) error
	GetAllProviderFields() (*models.ProviderMetadataData, error)
	GetProviderField(fieldName string) (*models.ProviderField, error)
	CreateProviderField(req models.CreateProviderFieldRequest) (*models.ProviderField, error)
	UpdateProviderField(fieldName string, req models.UpdateProviderFieldRequest) (*models.ProviderField, error)
	DeleteProviderField(fieldName string) error
	ConvertSchemaToProviderMetadata(req models.SchemaConversionRequest) (*models.SchemaConversionResponse, error)
	ExportConsumerGrants() ([]byte, error)
	ExportProviderMetadata() ([]byte, error)
	ImportConsumerGrants(data []byte) error
	ImportProviderMetadata(data []byte) error
	AddConsumerToAllowList(fieldName string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error)
	RemoveConsumerFromAllowList(fieldName, consumerID string) (*models.AllowListManagementResponse, error)
	GetAllowListForField(fieldName string) (*models.AllowListListResponse, error)
	UpdateConsumerInAllowList(fieldName, consumerID string, req models.AllowListManagementRequest) (*models.AllowListManagementResponse, error)
}

// Service factory functions

func NewConsumerServiceWithDB(db *sql.DB) ConsumerServiceInterface {
	return NewConsumerService(db)
}

func NewProviderServiceWithDB(db *sql.DB) ProviderServiceInterface {
	return NewProviderService(db)
}

func NewGrantsServiceWithDB(db *sql.DB) GrantsServiceInterface {
	return NewGrantsService(db)
}
