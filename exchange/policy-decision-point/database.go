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
	CreatePolicyMetadata(req *models.PolicyMetadataCreateRequest) (string, error)
	UpdateAllowList(req *models.AllowListUpdateRequest) error
	GetAllPolicyMetadata() (map[string]interface{}, error)
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

// CreatePolicyMetadata creates a new policy metadata record
func (ds *DatabaseService) CreatePolicyMetadata(req *models.PolicyMetadataCreateRequest) (string, error) {
	// Generate UUID for the new record
	var id string
	err := ds.db.QueryRow("SELECT gen_random_uuid()").Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}

	// Insert new policy metadata record
	query := `INSERT INTO policy_metadata 
		(id, field_name, display_name, description, source, is_owner, owner, access_control_type, allow_list, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	now := time.Now().UTC()
	owner := "CITIZEN"    // Default owner as per schema
	allowListJSON := "[]" // Default empty allow list

	_, err = ds.db.Exec(query, id, req.FieldName, req.DisplayName, req.Description,
		req.Source, req.IsOwner, owner, req.AccessControlType, allowListJSON, now, now)
	if err != nil {
		return "", fmt.Errorf("failed to create policy metadata: %w", err)
	}

	slog.Info("Created policy metadata", "id", id, "field_name", req.FieldName)
	return id, nil
}

// UpdateAllowList updates the allow list for a specific field
func (ds *DatabaseService) UpdateAllowList(req *models.AllowListUpdateRequest) error {
	// Parse expires_at timestamp
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		return fmt.Errorf("invalid expires_at format, expected RFC3339: %w", err)
	}

	// Get current allow list for the field
	var currentAllowListJSON string
	query := `SELECT allow_list FROM policy_metadata WHERE field_name = $1`
	err = ds.db.QueryRow(query, req.FieldName).Scan(&currentAllowListJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("field %s not found", req.FieldName)
		}
		return fmt.Errorf("failed to get current allow list: %w", err)
	}

	// Parse current allow list
	var currentAllowList []models.AllowListEntry
	if err := json.Unmarshal([]byte(currentAllowListJSON), &currentAllowList); err != nil {
		return fmt.Errorf("failed to parse current allow list: %w", err)
	}

	// Check if application already exists in allow list
	found := false
	for i, entry := range currentAllowList {
		if entry.ApplicationID == req.ApplicationID {
			// Update existing entry
			currentAllowList[i] = models.AllowListEntry{
				ApplicationID: req.ApplicationID,
				ExpiresAt:     expiresAt.Unix(),
			}
			found = true
			break
		}
	}

	// Add new entry if not found
	if !found {
		newEntry := models.AllowListEntry{
			ApplicationID: req.ApplicationID,
			ExpiresAt:     expiresAt.Unix(),
		}
		currentAllowList = append(currentAllowList, newEntry)
	}

	// Serialize updated allow list
	updatedAllowListJSON, err := json.Marshal(currentAllowList)
	if err != nil {
		return fmt.Errorf("failed to marshal updated allow list: %w", err)
	}

	// Update the allow list in database
	updateQuery := `UPDATE policy_metadata SET allow_list = $1, updated_at = $2 WHERE field_name = $3`
	now := time.Now().UTC()
	_, err = ds.db.Exec(updateQuery, updatedAllowListJSON, now, req.FieldName)
	if err != nil {
		return fmt.Errorf("failed to update allow list: %w", err)
	}

	slog.Info("Updated allow list", "field_name", req.FieldName, "application_id", req.ApplicationID)
	return nil
}

// GetAllPolicyMetadata retrieves all policy metadata from the database
func (ds *DatabaseService) GetAllPolicyMetadata() (map[string]interface{}, error) {
	query := `SELECT field_name, display_name, description, source, is_owner, owner, access_control_type, allow_list 
			  FROM policy_metadata ORDER BY created_at DESC`

	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy metadata: %w", err)
	}
	defer rows.Close()

	fields := make(map[string]interface{})
	for rows.Next() {
		var fieldName, displayName, description, source, owner, accessControlType string
		var isOwner bool
		var allowListJSON string

		err := rows.Scan(&fieldName, &displayName, &description, &source, &isOwner, &owner, &accessControlType, &allowListJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy metadata field: %w", err)
		}

		// Parse allow list JSON
		var allowList []interface{}
		if allowListJSON != "" {
			if err := json.Unmarshal([]byte(allowListJSON), &allowList); err != nil {
				slog.Warn("Failed to parse allow list", "field", fieldName, "error", err)
				allowList = []interface{}{}
			}
		}

		// Create field metadata structure
		fieldMetadata := map[string]interface{}{
			"owner":               owner,
			"provider":            source, // Using source as provider for now
			"is_owner":            isOwner,
			"access_control_type": accessControlType,
			"allow_list":          allowList,
		}

		// Add optional fields if they exist
		if displayName != "" {
			fieldMetadata["display_name"] = displayName
		}
		if description != "" {
			fieldMetadata["description"] = description
		}

		fields[fieldName] = fieldMetadata
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return map[string]interface{}{
		"fields": fields,
	}, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
