package models

// Request/Response DTOs for V1 API endpoints

// CreateSchemaSubmissionRequest Provider Schema Submission DTOs
type CreateSchemaSubmissionRequest struct {
	SchemaName        string  `json:"schemaName" validate:"required"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               string  `json:"sdl" validate:"required"`
	SchemaEndpoint    string  `json:"schemaEndpoint" validate:"required"`
	PreviousSchemaID  *string `json:"previousSchemaId,omitempty"`
	ProviderID        string  `json:"providerId" validate:"required"`
}

// UpdateSchemaSubmissionRequest updates the status of a provider schema submission
type UpdateSchemaSubmissionRequest struct {
	SchemaName        *string `json:"schemaName,omitempty"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               *string `json:"sdl,omitempty"`
	SchemaEndpoint    *string `json:"schemaEndpoint,omitempty"`
	Status            *string `json:"status,omitempty"`
	PreviousSchemaID  *string `json:"previousSchemaId,omitempty"`
	Review            *string `json:"review,omitempty"`
}

// CreateSchemaRequest creates a new provider schema
type CreateSchemaRequest struct {
	SchemaName        string  `json:"schemaName" validate:"required"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               string  `json:"sdl" validate:"required"`
	Endpoint          string  `json:"endpoint" validate:"required"`
	ProviderID        string  `json:"providerId" validate:"required"`
}

// UpdateSchemaRequest updates an existing provider schema
type UpdateSchemaRequest struct {
	SchemaName        *string `json:"schemaName,omitempty"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               *string `json:"sdl,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"`
	Version           *string `json:"version,omitempty"`
}

// // SelectedFieldRecord represents a record of selected fields
// type SelectedFieldRecord struct {
// 	FieldName string `json:"fieldName" validate:"required"`
// 	SchemaID  string `json:"schemaId" validate:"required"`
// }

// CreateApplicationSubmissionRequest Consumer Application Submission DTOs
type CreateApplicationSubmissionRequest struct {
	ApplicationName        string                `json:"applicationName" validate:"required"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields" validate:"required,min=1"`
	PreviousApplicationID  *string               `json:"previousApplicationId,omitempty"`
	ConsumerID             string                `json:"consumerId" validate:"required"`
}

// UpdateApplicationSubmissionRequest updates the status of a consumer application submission
type UpdateApplicationSubmissionRequest struct {
	ApplicationName        *string                `json:"applicationName,omitempty"`
	ApplicationDescription *string                `json:"applicationDescription,omitempty"`
	SelectedFields         *[]SelectedFieldRecord `json:"selectedFields,omitempty" validate:"required,min=1"`
	Status                 *string                `json:"status,omitempty"`
	PreviousApplicationID  *string                `json:"previousApplicationId,omitempty"`
	Review                 *string                `json:"review,omitempty"`
}

// CreateApplicationRequest creates a new consumer application
type CreateApplicationRequest struct {
	ApplicationName        string                `json:"applicationName" validate:"required"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields" validate:"required,min=1"`
	ConsumerID             string                `json:"consumerId" validate:"required"`
}

// UpdateApplicationRequest updates an existing consumer application
type UpdateApplicationRequest struct {
	ApplicationName        *string                `json:"applicationName,omitempty"`
	ApplicationDescription *string                `json:"applicationDescription,omitempty"`
	SelectedFields         *[]SelectedFieldRecord `json:"selectedFields,omitempty" validate:"required,min=1"`
	Version                *string                `json:"version,omitempty"`
}

type CreateEntityRequest struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
	IdpUserID   string `json:"idpUserId" validate:"required"`
}

type UpdateEntityRequest struct {
	Name        *string `json:"name,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	IdpUserID   *string `json:"idpUserId,omitempty"`
}

type CreateConsumerRequest struct {
	Name        string  `json:"name" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"phoneNumber" validate:"required"`
	IdpUserID   string  `json:"idpUserId" validate:"required"`
	EntityID    *string `json:"entityId,omitempty"`
}

type UpdateConsumerRequest struct {
	Name        *string `json:"name,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	IdpUserID   *string `json:"idpUserId,omitempty"`
}

type CreateProviderRequest struct {
	Name        string  `json:"name" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber string  `json:"phoneNumber" validate:"required"`
	EntityID    *string `json:"entityId,omitempty"`
	IdpUserID   string  `json:"idpUserId" validate:"required"`
}

type UpdateProviderRequest struct {
	Name        *string `json:"name,omitempty"`
	IdpUserID   *string `json:"idpUserId,omitempty"`
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type EntityResponse struct {
	EntityID    string `json:"entityId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	IdpUserID   string `json:"idpUserId"`
}

// ToEntity converts an EntityResponse to an Entity model (for internal use)
func (e *EntityResponse) ToEntity() Entity {
	return Entity{
		EntityID:    e.EntityID,
		Name:        e.Name,
		Email:       e.Email,
		PhoneNumber: e.PhoneNumber,
		IdpUserID:   e.IdpUserID,
	}
}

type ProviderResponse struct {
	ProviderID  string `json:"providerId"`
	EntityID    string `json:"entityId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	IdpUserID   string `json:"idpUserId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ConsumerResponse struct {
	ConsumerID  string `json:"consumerId"`
	EntityID    string `json:"entityId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	IdpUserID   string `json:"idpUserId"`
}

type SchemaResponse struct {
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

type SchemaSubmissionResponse struct {
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
	Review            *string `json:"review,omitempty"`
}

type ApplicationResponse struct {
	ApplicationID          string                `json:"applicationId"`
	ApplicationName        string                `json:"applicationName"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields"`
	ConsumerID             string                `json:"consumerId"`
	Version                string                `json:"version"`
	CreatedAt              string                `json:"createdAt"`
	UpdatedAt              string                `json:"updatedAt"`
}

type ApplicationSubmissionResponse struct {
	SubmissionID           string                `json:"submissionId"`
	PreviousApplicationID  *string               `json:"previousApplicationId,omitempty"`
	ApplicationName        string                `json:"applicationName"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields"`
	ConsumerID             string                `json:"consumerId"`
	Status                 string                `json:"status"`
	CreatedAt              string                `json:"createdAt"`
	UpdatedAt              string                `json:"updatedAt"`
	Review                 *string               `json:"review,omitempty"`
}

// CollectionResponse Generic collection response
type CollectionResponse struct {
	Items interface{} `json:"items"`
	Count int         `json:"count"`
}
