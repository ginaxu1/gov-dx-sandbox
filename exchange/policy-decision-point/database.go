package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gov-dx-sandbox/exchange/policy-decision-point/models"
	_ "github.com/lib/pq"
)

// DatabaseServiceInterface defines the interface for database operations
type DatabaseServiceInterface interface {
	GetAllProviderMetadata() (*models.ProviderMetadata, error)
	UpdateProviderField(fieldName string, field models.ProviderMetadataField) error
	UpdateProviderMetadata(metadata *models.ProviderMetadata) error
	Close() error
}

// DatabaseService handles database operations for the PDP service
type DatabaseService struct {
	db *sql.DB
}

// NewDatabaseService creates a new database service
func NewDatabaseService() (*DatabaseService, error) {
	// Get database connection string from Choreo environment variables
	// Choreo-Defined Environment Variable Names:
	// HostName: CHOREO_DB_PDP_HOSTNAME
	// Port: CHOREO_DB_PDP_PORT
	// Username: CHOREO_DB_PDP_USERNAME
	// Password: CHOREO_DB_PDP_PASSWORD
	// DatabaseName: CHOREO_DB_PDP_DATABASENAME

	dbHost := getEnv("CHOREO_DB_PDP_HOSTNAME", getEnv("DB_HOST", "localhost"))
	dbPort := getEnv("CHOREO_DB_PDP_PORT", getEnv("DB_PORT", "5432"))
	dbUser := getEnv("CHOREO_DB_PDP_USERNAME", getEnv("DB_USER", "postgres"))
	dbPassword := getEnv("CHOREO_DB_PDP_PASSWORD", getEnv("DB_PASSWORD", "postgres"))
	// For Choreo, the database name might be different from the environment variable
	dbName := getEnv("CHOREO_DB_PDP_DATABASENAME", getEnv("DB_NAME", "defaultdb"))

	// Require SSL by default for better security
	dbSSLMode := getEnv("DB_SSLMODE", "require")

	// Build connection string with SSL configuration
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	slog.Info("Database connection established",
		"host", dbHost,
		"port", dbPort,
		"database", dbName,
		"user", dbUser,
		"ssl_mode", dbSSLMode,
		"is_choreo", os.Getenv("CHOREO_DB_PDP_HOSTNAME") != "")

	return &DatabaseService{db: db}, nil
}

// Close closes the database connection
func (ds *DatabaseService) Close() error {
	if ds.db != nil {
		return ds.db.Close()
	}
	return nil
}

// GetAllProviderMetadata retrieves all provider metadata from the database
func (ds *DatabaseService) GetAllProviderMetadata() (*models.ProviderMetadata, error) {
	query := `SELECT field_name, owner, provider, consent_required, access_control_type, allow_list, created_at, updated_at 
			  FROM provider_metadata ORDER BY created_at DESC`

	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider metadata: %w", err)
	}
	defer rows.Close()

	fields := make(map[string]models.ProviderMetadataField)
	for rows.Next() {
		field := models.ProviderMetadataField{}
		var fieldName string
		var allowListJSON sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(&fieldName, &field.Owner, &field.Provider, &field.ConsentRequired, &field.AccessControlType, &allowListJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider field: %w", err)
		}

		// Parse JSON fields
		if allowListJSON.Valid {
			err = json.Unmarshal([]byte(allowListJSON.String), &field.AllowList)
			if err != nil {
				slog.Warn("Failed to parse allow list", "field", fieldName, "error", err)
				field.AllowList = []models.PDPAllowListEntry{}
			}
		}

		fields[fieldName] = field
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return &models.ProviderMetadata{
		Fields: fields,
	}, nil
}

// UpdateProviderField updates a specific provider field in the database
func (ds *DatabaseService) UpdateProviderField(fieldName string, field models.ProviderMetadataField) error {
	// Serialize JSON fields
	allowListJSON, err := json.Marshal(field.AllowList)
	if err != nil {
		return fmt.Errorf("failed to marshal allow list: %w", err)
	}

	now := time.Now().UTC()

	// Check if field exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM provider_metadata WHERE field_name = $1)`
	err = ds.db.QueryRow(checkQuery, fieldName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if field exists: %w", err)
	}

	if exists {
		// Update existing field
		updateQuery := `UPDATE provider_metadata SET 
			owner = $1, provider = $2, consent_required = $3, access_control_type = $4, 
			allow_list = $5, updated_at = $6 
			WHERE field_name = $7`

		_, err = ds.db.Exec(updateQuery, field.Owner, field.Provider, field.ConsentRequired,
			field.AccessControlType, allowListJSON, now, fieldName)
		if err != nil {
			return fmt.Errorf("failed to update provider field: %w", err)
		}
		slog.Info("Updated provider field", "fieldName", fieldName)
	} else {
		// Insert new field
		insertQuery := `INSERT INTO provider_metadata 
			(field_name, owner, provider, consent_required, access_control_type, allow_list, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = ds.db.Exec(insertQuery, fieldName, field.Owner, field.Provider, field.ConsentRequired,
			field.AccessControlType, allowListJSON, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert provider field: %w", err)
		}
		slog.Info("Created provider field", "fieldName", fieldName)
	}

	return nil
}

// UpdateProviderMetadata updates multiple provider fields in the database
func (ds *DatabaseService) UpdateProviderMetadata(metadata *models.ProviderMetadata) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for fieldName, field := range metadata.Fields {
		// Serialize JSON fields
		allowListJSON, err := json.Marshal(field.AllowList)
		if err != nil {
			return fmt.Errorf("failed to marshal allow list for field %s: %w", fieldName, err)
		}

		now := time.Now().UTC()

		// Use UPSERT (INSERT ... ON CONFLICT ... DO UPDATE)
		upsertQuery := `INSERT INTO provider_metadata 
			(field_name, owner, provider, consent_required, access_control_type, allow_list, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (field_name) DO UPDATE SET
			owner = EXCLUDED.owner,
			provider = EXCLUDED.provider,
			consent_required = EXCLUDED.consent_required,
			access_control_type = EXCLUDED.access_control_type,
			allow_list = EXCLUDED.allow_list,
			updated_at = EXCLUDED.updated_at`

		_, err = tx.Exec(upsertQuery, fieldName, field.Owner, field.Provider, field.ConsentRequired,
			field.AccessControlType, allowListJSON, now, now)
		if err != nil {
			return fmt.Errorf("failed to upsert provider field %s: %w", fieldName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Updated provider metadata", "fields", len(metadata.Fields))
	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
