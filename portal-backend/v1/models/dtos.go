package models

// Request/Response DTOs for V1 API endpoints

// CreateSchemaSubmissionRequest Provider Schema Submission DTOs
type CreateSchemaSubmissionRequest struct {
	SchemaName        string  `json:"schemaName" validate:"required"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               string  `json:"sdl" validate:"required"`
	SchemaEndpoint    string  `json:"schemaEndpoint" validate:"required"`
	PreviousSchemaID  *string `json:"previousSchemaId,omitempty"`
	MemberID          string  `json:"memberId" validate:"required"`
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
	MemberID          string  `json:"memberId" validate:"required"`
}

// UpdateSchemaRequest updates an existing provider schema
type UpdateSchemaRequest struct {
	SchemaName        *string `json:"schemaName,omitempty"`
	SchemaDescription *string `json:"schemaDescription,omitempty"`
	SDL               *string `json:"sdl,omitempty"`
	Endpoint          *string `json:"endpoint,omitempty"`
	Version           *string `json:"version,omitempty"`
}

// CreateApplicationSubmissionRequest Consumer Application Submission DTOs
type CreateApplicationSubmissionRequest struct {
	ApplicationName        string                `json:"applicationName" validate:"required"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields" validate:"required,min=1"`
	PreviousApplicationID  *string               `json:"previousApplicationId,omitempty"`
	MemberID               string                `json:"memberId" validate:"required"`
}

// UpdateApplicationSubmissionRequest updates the status of a consumer application submission
type UpdateApplicationSubmissionRequest struct {
	ApplicationName        *string                `json:"applicationName,omitempty"`
	ApplicationDescription *string                `json:"applicationDescription,omitempty"`
	SelectedFields         *[]SelectedFieldRecord `json:"selectedFields,omitempty"`
	Status                 *string                `json:"status,omitempty"`
	PreviousApplicationID  *string                `json:"previousApplicationId,omitempty"`
	Review                 *string                `json:"review,omitempty"`
}

// CreateApplicationRequest creates a new consumer application
type CreateApplicationRequest struct {
	ApplicationName        string                `json:"applicationName" validate:"required"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields" validate:"required,min=1"`
	MemberID               string                `json:"memberId" validate:"required"`
}

// UpdateApplicationRequest updates an existing consumer application
type UpdateApplicationRequest struct {
	ApplicationName        *string `json:"applicationName,omitempty"`
	ApplicationDescription *string `json:"applicationDescription,omitempty"`
	Version                *string `json:"version,omitempty"`
	// Note: SelectedFields is intentionally omitted from UpdateApplicationRequest.
	// Field updates should be handled through a separate endpoint or process. That is not implemented yet.
}

type CreateMemberRequest struct {
	Name        string `json:"name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type UpdateMemberRequest struct {
	Name        *string `json:"name,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type MemberResponse struct {
	MemberID    string `json:"memberId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	IdpUserID   string `json:"idpUserId"`
}

// ToMember converts a MemberResponse to a Member model (for internal use)
func (e *MemberResponse) ToMember() Member {
	return Member{
		MemberID:    e.MemberID,
		Name:        e.Name,
		Email:       e.Email,
		PhoneNumber: e.PhoneNumber,
		IdpUserID:   e.IdpUserID,
	}
}

type SchemaResponse struct {
	SchemaID          string  `json:"schemaId"`
	MemberID          string  `json:"memberId"`
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
	MemberID          string  `json:"memberId"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
	Review            *string `json:"review,omitempty"`
}

type ApplicationResponse struct {
	ApplicationID          string                `json:"applicationId"`
	ApplicationName        string                `json:"applicationName"`
	ApplicationDescription *string               `json:"applicationDescription,omitempty"`
	SelectedFields         []SelectedFieldRecord `json:"selectedFields"`
	MemberID               string                `json:"memberId"`
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
	MemberID               string                `json:"memberId"`
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
