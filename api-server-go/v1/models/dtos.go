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

type UpdateProviderSchemaSubmissionRequest struct {
	Status *string `json:"status,omitempty"`
}

// CreateConsumerApplicationSubmissionRequest Consumer Application Submission DTOs
type CreateConsumerApplicationSubmissionRequest struct {
	ApplicationName        string   `json:"applicationName" validate:"required"`
	ApplicationDescription *string  `json:"applicationDescription,omitempty"`
	SelectedFields         []string `json:"selectedFields" validate:"required,min=1"`
	PreviousApplicationID  *string  `json:"previousApplicationId,omitempty"`
}

type UpdateConsumerApplicationSubmissionRequest struct {
	Status *string `json:"status,omitempty"`
}

// EntityResponse Response DTOs
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
