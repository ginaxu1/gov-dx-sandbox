package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestBaseModel_BeforeCreate(t *testing.T) {
	t.Run("BeforeCreate_SetsTimestamps", func(t *testing.T) {
		// Use a test database connection
		dsn := "host=localhost port=5432 user=postgres password=password dbname=api_server_test sslmode=disable"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Skipf("Skipping test: could not connect to test database: %v", err)
			return
		}

		// Create a test model that embeds BaseModel
		type TestModel struct {
			ID string `gorm:"primarykey"`
			BaseModel
			Name string
		}

		// Auto-migrate
		db.AutoMigrate(&TestModel{})

		// Create a record
		model := TestModel{
			ID:   "test-123",
			Name: "Test",
		}

		err = db.Create(&model).Error
		assert.NoError(t, err)

		// Verify timestamps were set
		assert.False(t, model.CreatedAt.IsZero())
		assert.False(t, model.UpdatedAt.IsZero())
		assert.WithinDuration(t, time.Now(), model.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, time.Now(), model.UpdatedAt, 5*time.Second)

		// Cleanup
		db.Exec("DELETE FROM test_models")
	})
}

func TestBaseModel_BeforeUpdate(t *testing.T) {
	t.Run("BeforeUpdate_UpdatesTimestamp", func(t *testing.T) {
		dsn := "host=localhost port=5432 user=postgres password=password dbname=api_server_test sslmode=disable"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			t.Skipf("Skipping test: could not connect to test database: %v", err)
			return
		}

		type TestModel struct {
			ID string `gorm:"primarykey"`
			BaseModel
			Name string
		}

		db.AutoMigrate(&TestModel{})

		// Create a record
		model := TestModel{
			ID:   "test-123",
			Name: "Original",
		}
		db.Create(&model)
		originalUpdatedAt := model.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(100 * time.Millisecond)

		// Update the record
		model.Name = "Updated"
		err = db.Save(&model).Error
		assert.NoError(t, err)

		// Verify UpdatedAt was changed
		assert.True(t, model.UpdatedAt.After(originalUpdatedAt))

		// Cleanup
		db.Exec("DELETE FROM test_models")
	})
}
