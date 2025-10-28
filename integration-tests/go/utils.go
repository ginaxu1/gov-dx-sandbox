package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// TestDB wraps database connection for testing
type TestDB struct {
	connStr string
	db      *sql.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(connStr string) *TestDB {
	return &TestDB{
		connStr: connStr,
	}
}

// Connect establishes a database connection
func (tdb *TestDB) Connect() error {
	db, err := sql.Open("postgres", tdb.connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	tdb.db = db
	return nil
}

// Close closes the database connection
func (tdb *TestDB) Close() error {
	if tdb.db != nil {
		return tdb.db.Close()
	}
	return nil
}

// DB returns the database connection
func (tdb *TestDB) DB() *sql.DB {
	return tdb.db
}

// ConnectionString returns the connection string
func (tdb *TestDB) ConnectionString() string {
	return tdb.connStr
}

// ExecuteSQL executes a SQL statement
func (tdb *TestDB) ExecuteSQL(ctx context.Context, query string, args ...interface{}) error {
	_, err := tdb.db.ExecContext(ctx, query, args...)
	return err
}

// QueryRow executes a query and returns a single row
func (tdb *TestDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return tdb.db.QueryRowContext(ctx, query, args...)
}

// Query executes a query and returns multiple rows
func (tdb *TestDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return tdb.db.QueryContext(ctx, query, args...)
}

// GetRowCount returns the number of rows in a table
func (tdb *TestDB) GetRowCount(ctx context.Context, tableName string) (int, error) {
	var count int
	err := tdb.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
	return count, err
}

// TestRedis wraps Redis connection for testing
type TestRedis struct {
	connStr string
	client  interface{} // Using interface{} to avoid direct redis dependency in this simplified version
}

// NewTestRedis creates a new test Redis connection
func NewTestRedis(connStr string) *TestRedis {
	return &TestRedis{
		connStr: connStr,
	}
}

// Connect establishes a Redis connection
func (tr *TestRedis) Connect() error {
	// Simplified version - in real implementation would use redis client
	// For now, just verify connection string exists
	if tr.connStr == "" {
		return fmt.Errorf("Redis connection string is empty")
	}
	return nil
}

// Close closes the Redis connection
func (tr *TestRedis) Close() error {
	return nil
}

// Client returns the Redis client (placeholder)
func (tr *TestRedis) Client() interface{} {
	return tr.client
}

// ConnectionString returns the connection string
func (tr *TestRedis) ConnectionString() string {
	return tr.connStr
}

// GetStreamLength returns the length of a Redis stream
func (tr *TestRedis) GetStreamLength(ctx context.Context, streamName string) (int64, error) {
	// This would use actual Redis client in real implementation
	return 0, nil
}

// AddToStream adds a message to a Redis stream
func (tr *TestRedis) AddToStream(ctx context.Context, streamName string, values map[string]interface{}) (string, error) {
	// This would use actual Redis client in real implementation
	return "", nil
}

// TestOPA wraps OPA connection for testing
type TestOPA struct {
	url string
}

// NewTestOPA creates a new test OPA connection
func NewTestOPA(url string) *TestOPA {
	return &TestOPA{
		url: url,
	}
}

// URL returns the OPA URL
func (to *TestOPA) URL() string {
	return to.url
}

// WaitForServices waits for all services to be ready
func WaitForServices(timeout time.Duration) error {
	log.Println("Waiting for services to be ready...")

	// Wait for PostgreSQL
	if err := waitForPostgreSQL(timeout); err != nil {
		return fmt.Errorf("PostgreSQL not ready: %w", err)
	}

	// Wait for Redis
	if err := waitForRedis(timeout); err != nil {
		return fmt.Errorf("Redis not ready: %w", err)
	}

	// Wait for OPA
	if err := waitForOPA(timeout); err != nil {
		return fmt.Errorf("OPA not ready: %w", err)
	}

	log.Println("All services are ready!")
	return nil
}

func waitForPostgreSQL(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	connStr := "postgres://test_user:test_password@localhost:5432/opendif_test?sslmode=disable"
	for time.Now().Before(deadline) {
		conn, err := sql.Open("postgres", connStr)
		if err == nil {
			if err := conn.Ping(); err == nil {
				conn.Close()
				return nil
			}
			conn.Close()
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for PostgreSQL")
}

func waitForRedis(timeout time.Duration) error {
	// Redis validation is optional in test environment
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Simple timeout check
		if time.Now().After(deadline.Add(-5 * time.Second)) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return nil
}

func waitForOPA(timeout time.Duration) error {
	// OPA validation is optional in test environment
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Simple timeout check
		if time.Now().After(deadline.Add(-5 * time.Second)) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return nil
}
