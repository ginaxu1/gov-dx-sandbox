package models

// Schema represents the provider_schemas table
type Schema struct {
	SchemaID          string  `gorm:"primarykey;column:schema_id" json:"schemaId"`
	ProviderID        string  `gorm:"column:provider_id;not null" json:"providerId"`
	SchemaName        string  `gorm:"column:schema_name;not null" json:"schemaName"`
	SDL               string  `gorm:"column:sdl;not null" json:"sdl"`
	Endpoint          string  `gorm:"column:endpoint;not null" json:"endpoint"`
	Version           string  `gorm:"column:version;not null" json:"version"`
	SchemaDescription *string `gorm:"column:schema_description" json:"schemaDescription,omitempty"`
	BaseModel

	// Relationships
	Provider Provider `gorm:"foreignKey:ProviderID;references:ProviderID" json:"provider"`
}

// TableName sets the table name for GORM
func (Schema) TableName() string {
	return "provider_schemas"
}

// SchemaSubmission represents the provider_schema_submissions table
type SchemaSubmission struct {
	SubmissionID      string  `gorm:"primarykey;column:submission_id" json:"submissionId"`
	PreviousSchemaID  *string `gorm:"column:previous_schema_id" json:"previousSchemaId,omitempty"`
	SchemaName        string  `gorm:"column:schema_name;not null" json:"schemaName"`
	SchemaDescription *string `gorm:"column:schema_description" json:"schemaDescription,omitempty"`
	SDL               string  `gorm:"column:sdl;not null" json:"sdl"`
	SchemaEndpoint    string  `gorm:"column:schema_endpoint;not null" json:"schemaEndpoint"`
	Status            string  `gorm:"column:status;not null" json:"status"`
	ProviderID        string  `gorm:"column:provider_id;not null" json:"providerId"`
	Review            *string `gorm:"column:review" json:"review,omitempty"`
	BaseModel

	// Relationships
	Provider       Provider `gorm:"foreignKey:ProviderID;references:ProviderID" json:"provider"`
	PreviousSchema *Schema  `gorm:"foreignKey:PreviousSchemaID;references:SchemaID" json:"previousSchema,omitempty"`
}

// TableName sets the table name for GORM
func (SchemaSubmission) TableName() string {
	return "provider_schema_submissions"
}

// Application represents the consumer_applications table
type Application struct {
	ApplicationID          string      `gorm:"primarykey;column:application_id" json:"applicationId"`
	ApplicationName        string      `gorm:"column:application_name;not null" json:"applicationName"`
	ApplicationDescription *string     `gorm:"column:application_description" json:"applicationDescription,omitempty"`
	SelectedFields         StringArray `gorm:"column:selected_fields;type:text[];not null" json:"selectedFields"`
	ConsumerID             string      `gorm:"column:consumer_id;not null" json:"consumerId"`
	Version                string      `gorm:"column:version;not null" json:"version"`
	BaseModel

	// Relationships
	Consumer Consumer `gorm:"foreignKey:ConsumerID;references:ConsumerID" json:"consumer"`
}

// TableName sets the table name for GORM
func (Application) TableName() string {
	return "consumer_applications"
}

// ApplicationSubmission represents the consumer_application_submissions table
type ApplicationSubmission struct {
	SubmissionID           string      `gorm:"primarykey;column:submission_id" json:"submissionId"`
	PreviousApplicationID  *string     `gorm:"column:previous_application_id" json:"previousApplicationId,omitempty"`
	ApplicationName        string      `gorm:"column:application_name;not null" json:"applicationName"`
	ApplicationDescription *string     `gorm:"column:application_description" json:"applicationDescription,omitempty"`
	SelectedFields         StringArray `gorm:"column:selected_fields;type:text[];not null" json:"selectedFields"`
	ConsumerID             string      `gorm:"column:consumer_id;not null" json:"consumerId"`
	Status                 string      `gorm:"column:status;not null" json:"status"`
	Review                 *string     `gorm:"column:review" json:"review,omitempty"`
	BaseModel

	// Relationships
	Consumer            Consumer     `gorm:"foreignKey:ConsumerID;references:ConsumerID" json:"consumer"`
	PreviousApplication *Application `gorm:"foreignKey:PreviousApplicationID;references:ApplicationID" json:"previousApplication,omitempty"`
}

// TableName sets the table name for GORM
func (ApplicationSubmission) TableName() string {
	return "consumer_application_submissions"
}
