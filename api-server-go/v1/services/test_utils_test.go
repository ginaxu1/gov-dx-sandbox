package services

import (
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
)

func TestGetEnvOrDefault(t *testing.T) {
	t.Run("GetEnvOrDefault_WithValue", func(t *testing.T) {
		key := "TEST_ENV_VAR_12345"
		os.Setenv(key, "test-value")
		defer os.Unsetenv(key)

		result := getEnvOrDefault(key, "default")
		assert.Equal(t, "test-value", result)
	})

	t.Run("GetEnvOrDefault_WithoutValue", func(t *testing.T) {
		key := "TEST_ENV_VAR_NONEXISTENT_12345"
		os.Unsetenv(key)

		result := getEnvOrDefault(key, "default-value")
		assert.Equal(t, "default-value", result)
	})

	t.Run("GetEnvOrDefault_EmptyString", func(t *testing.T) {
		key := "TEST_ENV_VAR_EMPTY_12345"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		result := getEnvOrDefault(key, "default")
		assert.Equal(t, "default", result)
	})
}

func TestCleanupTestData(t *testing.T) {
	t.Run("CleanupTestData_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}

		// Create some test data
		member := models.Member{
			MemberID:    "test-member-cleanup",
			Name:        "Test Member",
			Email:       "cleanup@example.com",
			PhoneNumber: "1234567890",
		}
		db.Create(&member)

		// Verify data exists
		var count int64
		db.Model(&models.Member{}).Where("member_id = ?", member.MemberID).Count(&count)
		assert.Greater(t, count, int64(0))

		// Cleanup
		CleanupTestData(t, db)

		// Verify data was deleted
		db.Model(&models.Member{}).Where("member_id = ?", member.MemberID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
