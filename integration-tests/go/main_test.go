package integration

import (
	"log"
	"os"
	"testing"
	"time"
)

// Test environment configuration
var (
	testDB    *TestDB
	testRedis *TestRedis
	testOPA   *TestOPA
)

// TestMain manages the lifecycle of the integration test environment
func TestMain(m *testing.M) {
	// Check if we should use Docker Compose
	useCompose := os.Getenv("USE_COMPOSE") == "true"

	var exitCode int

	if useCompose {
		// Use Docker Compose for full environment
		log.Println("Using Docker Compose for test environment...")
		log.Println("Please start Docker Compose manually: docker-compose up -d")
		log.Println("Waiting for services to be ready...")
		time.Sleep(5 * time.Second)

		log.Println("Test environment is up and running!")
		exitCode = m.Run()
	} else {
		// Use direct database connections (assuming services are running)
		log.Println("Using existing services for test environment...")
		log.Println("Assuming services are running on localhost")

		// Initialize test utilities
		testDB = NewTestDB(getPostgresURL())
		testRedis = NewTestRedis(getRedisURL())

		log.Println("Test environment is ready!")
		exitCode = m.Run()
	}

	os.Exit(exitCode)
}

// getPostgresURL returns the PostgreSQL connection string
func getPostgresURL() string {
	// Always use localhost since tests run on host, not in container
	// Docker Compose ports are mapped to localhost
	if os.Getenv("USE_COMPOSE") == "true" {
		// When using compose, connect via localhost with fixed port mapping
		// Using port 5433 to avoid conflicts with local PostgreSQL
		return "postgres://test_user:test_password@localhost:5433/opendif_test?sslmode=disable"
	}
	if testDB != nil {
		return testDB.ConnectionString()
	}
	return "postgres://test_user:test_password@localhost:5432/opendif_test?sslmode=disable"
}

// getRedisURL returns the Redis connection string
func getRedisURL() string {
	// Always use localhost since tests run on host, not in container
	if os.Getenv("USE_COMPOSE") == "true" {
		return "localhost:6379"
	}
	if testRedis != nil {
		return testRedis.ConnectionString()
	}
	return "localhost:6379"
}

// getOPAURL returns the OPA service URL
func getOPAURL() string {
	// Always use localhost since tests run on host, not in container
	return "http://localhost:8181"
}
