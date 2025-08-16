// database/database.go
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"policy-governance/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v5"
)

var conn *pgxpool.Pool

// GetConsumerPolicy fetches and parses the policy for a given consumer from the database.
// This is a variable so we can monkey-patch it during testing.
var GetConsumerPolicy = func(consumerId string) (*models.Policy, error) {
	if conn == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var policyJson []byte
	query := `SELECT policy FROM consumer_policies WHERE consumer_id = $1`
	err := conn.QueryRow(context.Background(), query, consumerId).Scan(&policyJson)
	if err != nil {
		if err == pgx.ErrNoRows {
			// It's not an error, but the consumer has no policy, so they have no permissions.
			// Return a nil policy and nil error to indicate this state.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query policy: %w", err)
	}

	var policy models.Policy
	err = json.Unmarshal(policyJson, &policy)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy JSON: %w", err)
	}

	return &policy, nil
}

// Init initializes the database connection using the DATABASE_URL environment variable.
func Init() error {
	var err error
	// Prevent initialization if in test mode (DATABASE_URL is not set)
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		fmt.Println("DATABASE_URL not set, skipping database initialization for test mode.")
		return nil
	}

	conn, err = pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	fmt.Println("Successfully connected to the database!")
	return nil
}

// Close gracefully terminates the database connection.
func Close() {
	if conn != nil {
		conn.Close()
		fmt.Println("Database connection closed.")
	}
}
