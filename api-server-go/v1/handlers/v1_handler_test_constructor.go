package handlers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewV1Handler(t *testing.T) {
	t.Run("NewV1Handler_MissingPDPURL", func(t *testing.T) {
		// Save original value
		originalURL := os.Getenv("CHOREO_PDP_CONNECTION_SERVICEURL")
		originalKey := os.Getenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")
		defer func() {
			if originalURL != "" {
				os.Setenv("CHOREO_PDP_CONNECTION_SERVICEURL", originalURL)
			} else {
				os.Unsetenv("CHOREO_PDP_CONNECTION_SERVICEURL")
			}
			if originalKey != "" {
				os.Setenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY", originalKey)
			} else {
				os.Unsetenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")
			}
		}()

		// Unset required environment variables
		os.Unsetenv("CHOREO_PDP_CONNECTION_SERVICEURL")
		os.Unsetenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")

		dsn := "host=localhost port=5432 user=postgres password=password dbname=api_server_test sslmode=disable"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Skip("Skipping test: could not connect to test database")
			return
		}

		handler, err := NewV1Handler(db)
		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "CHOREO_PDP_CONNECTION_SERVICEURL")
	})

	t.Run("NewV1Handler_MissingPDPKey", func(t *testing.T) {
		// Save original value
		originalURL := os.Getenv("CHOREO_PDP_CONNECTION_SERVICEURL")
		originalKey := os.Getenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")
		defer func() {
			if originalURL != "" {
				os.Setenv("CHOREO_PDP_CONNECTION_SERVICEURL", originalURL)
			} else {
				os.Unsetenv("CHOREO_PDP_CONNECTION_SERVICEURL")
			}
			if originalKey != "" {
				os.Setenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY", originalKey)
			} else {
				os.Unsetenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")
			}
		}()

		// Set URL but not key
		os.Setenv("CHOREO_PDP_CONNECTION_SERVICEURL", "http://localhost:9999")
		os.Unsetenv("CHOREO_PDP_CONNECTION_CHOREOAPIKEY")

		dsn := "host=localhost port=5432 user=postgres password=password dbname=api_server_test sslmode=disable"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Skip("Skipping test: could not connect to test database")
			return
		}

		handler, err := NewV1Handler(db)
		assert.Error(t, err)
		assert.Nil(t, handler)
		assert.Contains(t, err.Error(), "CHOREO_PDP_CONNECTION_CHOREOAPIKEY")
	})
}
