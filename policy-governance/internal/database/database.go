package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"policy-governance/internal/models"

	"github.com/jackc/pgx/v4"
)

var conn *pgx.Conn

var GetConsumerPolicy = func(consumerId string) (*models.Policy, error) {
	if conn == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var policyJson []byte
	query := `SELECT policy FROM consumer_policies WHERE consumer_id = $1`
	err := conn.QueryRow(context.Background(), query, consumerId).Scan(&policyJson)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no policy found for consumer: %s", consumerId)
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

func Init() error {
	var err error
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	conn, err = pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	fmt.Println("Successfully connected to the database!")
	return nil
}

func Close() {
	if conn != nil {
		conn.Close(context.Background())
		fmt.Println("Database connection closed.")
	}
}
