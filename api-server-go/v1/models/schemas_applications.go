package models

// ProviderSchema represents the provider_schemas table
type ProviderSchema struct {
	SchemaID          string  `gorm:"primarykey;column:schema_id" json:"schemaId"`
	ProviderID        string  `gorm:"column:provider_id;not null" json:"providerId"`
	SchemaName        string  `gorm:"column:schema_name;not null" json:"schemaName"`
	SDL               string  `gorm:"column:sdl;not null" json:"sdl"`
	Endpoint          string  `gorm:"column:endpoint;not null" json:"endpoint"`
	Version           string  `gorm:"column:version;not null" json:"version"`
	SchemaDescription *string `gorm:"column:schema_description" json:"schemaDescription,omitempty"`
	BaseModel

	// Relationships (disabled for auto-migration)
	Provider Provider `gorm:"-" json:"provider,omitempty"`
}

// TableName sets the table name for GORM
func (ProviderSchema) TableName() string {
	return "provider_schemas"
}

// ProviderSchemaSubmission represents the provider_schema_submissions table
type ProviderSchemaSubmission struct {
	SubmissionID      string  `gorm:"primarykey;column:submission_id" json:"submissionId"`
	PreviousSchemaID  *string `gorm:"column:previous_schema_id" json:"previousSchemaId,omitempty"`
	SchemaName        string  `gorm:"column:schema_name;not null" json:"schemaName"`
	SchemaDescription *string `gorm:"column:schema_description" json:"schemaDescription,omitempty"`
	SDL               string  `gorm:"column:sdl;not null" json:"sdl"`
	SchemaEndpoint    string  `gorm:"column:schema_endpoint;not null" json:"schemaEndpoint"`
	Status            string  `gorm:"column:status;not null" json:"status"`
	ProviderID        string  `gorm:"column:provider_id;not null" json:"providerId"`
	BaseModel

	// Relationships (disabled for auto-migration)
	Provider       Provider        `gorm:"-" json:"provider,omitempty"`
	PreviousSchema *ProviderSchema `gorm:"-" json:"previousSchema,omitempty"`
}

// TableName sets the table name for GORM
func (ProviderSchemaSubmission) TableName() string {
	return "provider_schema_submissions"
}

// ConsumerApplication represents the consumer_applications table
type ConsumerApplication struct {
	ApplicationID          string      `gorm:"primarykey;column:application_id" json:"applicationId"`
	ApplicationName        string      `gorm:"column:application_name;not null" json:"applicationName"`
	ApplicationDescription *string     `gorm:"column:application_description" json:"applicationDescription,omitempty"`
	SelectedFields         StringArray `gorm:"column:selected_fields;type:text[];not null" json:"selectedFields"`
	ConsumerID             string      `gorm:"column:consumer_id;not null" json:"consumerId"`
	Version                string      `gorm:"column:version;not null" json:"version"`
	BaseModel

	// Relationships (disabled for auto-migration)
	Consumer Consumer `gorm:"-" json:"consumer,omitempty"`
}

// TableName sets the table name for GORM
func (ConsumerApplication) TableName() string {
	return "consumer_applications"
}

// ConsumerApplicationSubmission represents the consumer_application_submissions table
type ConsumerApplicationSubmission struct {
	SubmissionID           string      `gorm:"primarykey;column:submission_id" json:"submissionId"`
	PreviousApplicationID  *string     `gorm:"column:previous_application_id" json:"previousApplicationId,omitempty"`
	ApplicationName        string      `gorm:"column:application_name;not null" json:"applicationName"`
	ApplicationDescription *string     `gorm:"column:application_description" json:"applicationDescription,omitempty"`
	SelectedFields         StringArray `gorm:"column:selected_fields;type:text[];not null" json:"selectedFields"`
	ConsumerID             string      `gorm:"column:consumer_id;not null" json:"consumerId"`
	Status                 string      `gorm:"column:status;not null" json:"status"`
	BaseModel

	// Relationships (disabled for auto-migration)
	Consumer            Consumer             `gorm:"-" json:"consumer,omitempty"`
	PreviousApplication *ConsumerApplication `gorm:"-" json:"previousApplication,omitempty"`
}

// TableName sets the table name for GORM
func (ConsumerApplicationSubmission) TableName() string {
	return "consumer_application_submissions"
}
