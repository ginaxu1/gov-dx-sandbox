package models

// Request/Response DTOs for V1 API endpoints

// CreateProviderSchemaSubmissionRequest Provider Schema Submission DTOs
type CreateProviderSchemaSubmissionRequest struct {
	SchemaName        string  `json:"schemaName" validate:"required"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               string  `json:"sdl" validate:"required"`
	SchemaEndpoint    string  `json:"schemaEndpoint" validate:"required"`
	PreviousSchemaID  *string `json:"previousSchemaId,omitempty"`
}

// UpdateProviderSchemaSubmissionRequest updates the status of a provider schema submission
type UpdateProviderSchemaSubmissionRequest struct {
	SchemaName        *string `json:"schemaName,omitempty"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               *string `json:"sdl,omitempty"`
	SchemaEndpoint    *string `json:"schemaEndpoint,omitempty"`
	Status            *string `json:"status,omitempty"`
	PreviousSchemaID  *string `json:"previousSchemaId,omitempty"`
}

// CreateProviderSchemaRequest creates a new provider schema
type CreateProviderSchemaRequest struct {
	SchemaName        string  `json:"schemaName" validate:"required"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               string  `json:"sdl" validate:"required"`
	Endpoint          string  `json:"endpoint" validate:"required"`
}

// UpdateProviderSchemaRequest updates an existing provider schema
type UpdateProviderSchemaRequest struct {
	SchemaName        *string `json:"schemaName,omitempty"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               *string `json:"sdl,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"`
	Version           *string `json:"version,omitempty"`
}

// CreateConsumerApplicationSubmissionRequest Consumer Application Submission DTOs
type CreateConsumerApplicationSubmissionRequest struct {
	ApplicationName        string   `json:"applicationName" validate:"required"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields" validate:"required,min=1"`
	PreviousApplicationID  *string  `json:"previousApplicationId,omitempty"`
}

// UpdateConsumerApplicationSubmissionRequest updates the status of a consumer application submission
type UpdateConsumerApplicationSubmissionRequest struct {
	ApplicationName        *string  `json:"applicationName,omitempty"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields,omitempty"`
	Status                 *string  `json:"status,omitempty"`
	PreviousApplicationID  *string  `json:"previousApplicationId,omitempty"`
}

// CreateConsumerApplicationRequest creates a new consumer application
type CreateConsumerApplicationRequest struct {
	ApplicationName        string   `json:"applicationName" validate:"required"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields" validate:"required,min=1"`
}

// UpdateConsumerApplicationRequest updates an existing consumer application
type UpdateConsumerApplicationRequest struct {
	ApplicationName        *string  `json:"applicationName,omitempty"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields,omitempty"`
	Version                *string  `json:"version,omitempty"`
}

type CreateEntityRequest struct {
	Name        string `json:"name" validate:"required"`
	EntityType  string `json:"entityType" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type UpdateEntityRequest struct {
	Name        *string `json:"name,omitempty"`
	EntityType  *string `json:"entityType,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type CreateConsumerRequest struct {
	Name        string  `json:"name" validate:"required"`
	EntityType  string  `json:"entityType" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"phoneNumber" validate:"required"`
	EntityID    *string `json:"entityId,omitempty"`
}

type UpdateConsumerRequest struct {
	Name        *string `json:"name,omitempty"`
	EntityType  *string `json:"entityType,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type CreateProviderRequest struct {
	Name        string  `json:"name" validate:"required"`
	EntityType  string  `json:"entityType" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"phoneNumber" validate:"required"`
	EntityID    *string `json:"entityId,omitempty"`
}

type UpdateProviderRequest struct {
	Name        *string `json:"name,omitempty"`
	EntityType  *string `json:"entityType,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type EntityResponse struct {
	EntityID    string  `json:"entityId"`
	Name        string  `json:"name"`
	EntityType  string  `json:"entityType"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phoneNumber"`
	ProviderID  *string `json:"providerId,omitempty"`
	ConsumerID  *string `json:"consumerId,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// ToEntity converts an EntityResponse to an Entity model (for internal use)
func (e *EntityResponse) ToEntity() Entity {
	return Entity{
		EntityID:    e.EntityID,
		Name:        e.Name,
		EntityType:  e.EntityType,
		Email:       e.Email,
		PhoneNumber: e.PhoneNumber,
	}
}

type ProviderResponse struct {
	ProviderID  string `json:"providerId"`
	EntityID    string `json:"entityId"`
	Name        string `json:"name"`
	EntityType  string `json:"entityType"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ConsumerResponse struct {
	ConsumerID  string `json:"consumerId"`
	EntityID    string `json:"entityId"`
	Name        string `json:"name"`
	EntityType  string `json:"entityType"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ProviderSchemaResponse struct {
	SchemaID          string  `json:"schemaId"`
	ProviderID        string  `json:"providerId"`
	SchemaName        string  `json:"schemaName"`
	SDL               string  `json:"sdl"`
	Endpoint          string  `json:"endpoint"`
	Version           string  `json:"version"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type ProviderSchemaSubmissionResponse struct {
	SubmissionID      string  `json:"submissionId"`
	PreviousSchemaID  *string `json:"previousSchemaId,omitempty"`
	SchemaName        string  `json:"schemaName"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               string  `json:"sdl"`
	SchemaEndpoint    string  `json:"schemaEndpoint"`
	Status            string  `json:"status"`
	ProviderID        string  `json:"providerId"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type ConsumerApplicationResponse struct {
	ApplicationID          string   `json:"applicationId"`
	ApplicationName        string   `json:"applicationName"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields"`
	ConsumerID             string   `json:"consumerId"`
	Version                string   `json:"version"`
	CreatedAt              string   `json:"createdAt"`
	UpdatedAt              string   `json:"updatedAt"`
}

type ConsumerApplicationSubmissionResponse struct {
	SubmissionID           string   `json:"submissionId"`
	PreviousApplicationID  *string  `json:"previousApplicationId,omitempty"`
	ApplicationName        string   `json:"applicationName"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields"`
	ConsumerID             string   `json:"consumerId"`
	Status                 string   `json:"status"`
	CreatedAt              string   `json:"createdAt"`
	UpdatedAt              string   `json:"updatedAt"`
}

// CollectionResponse Generic collection response
type CollectionResponse struct {
	Items interface{} `json:"items"`
	Count int         `json:"count"`
}
