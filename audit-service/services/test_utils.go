package services

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var envLoadOnce sync.Once

// loadEnvOnce loads environment variables from .env.local file (once)
func loadEnvOnce() {
	envLoadOnce.Do(func() {
		// Try to load .env.local file from current directory and parent directories
		envFiles := []string{
			".env.local",
			"../.env.local",
			"../../.env.local",
		}

		for _, envFile := range envFiles {
			if absPath, err := filepath.Abs(envFile); err == nil {
				if _, err := os.Stat(absPath); err == nil {
					if err := godotenv.Load(absPath); err == nil {
						log.Printf("Loaded test environment from: %s", absPath)
						return
					}
				}
			}
		}
		// If no .env.local found, that's okay - we'll use system env vars
	})
}

// getEnvVar returns the environment variable value
func getEnvVar(key string) string {
	loadEnvOnce() // Ensure .env.local is loaded
	return os.Getenv(key)
}

// CleanupTestData removes all test data from the database
// This function is used by SQLite test utilities
func CleanupTestData(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// Delete all audit logs
	if err := db.Exec("DELETE FROM audit_logs").Error; err != nil {
		t.Logf("Warning: could not cleanup audit_logs: %v", err)
	}
}
