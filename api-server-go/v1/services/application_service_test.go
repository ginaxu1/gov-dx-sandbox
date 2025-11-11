package services

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDBForApplication creates an in-memory SQLite database for application testing
func setupTestDBForApplication(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto-migrate tables (GormValue now handles SQLite compatibility)
	err = db.AutoMigrate(&models.Member{}, &models.Application{}, &models.ApplicationSubmission{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestApplicationService_CreateApplication(t *testing.T) {
	t.Run("CreateApplication_Success", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		// Use a real PDPService - it will fail on HTTP calls but we're testing DB operations
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		req := &models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: stringPtr("Test Description"),
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
		}

		// This will fail on PDP call, but we can verify DB operations
		_, err := service.CreateApplication(req)

		// Expect error due to PDP failure, but verify compensation worked
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update allow list")
		// Verify application was rolled back
		var count int64
		db.Model(&models.Application{}).Where("application_name = ?", req.ApplicationName).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestApplicationService_UpdateApplication(t *testing.T) {
	t.Run("UpdateApplication_Success", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		// Create an application first using GORM (now works with SQLite)
		application := models.Application{
			ApplicationID:          "app_123",
			ApplicationName:        "Original Name",
			ApplicationDescription: stringPtr("Original Description"),
			SelectedFields: models.SelectedFieldRecords{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
			Version:  string(models.ActiveVersion),
		}
		err := db.Create(&application).Error
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		newName := "Updated Name"
		newDesc := "Updated Description"
		req := &models.UpdateApplicationRequest{
			ApplicationName:        &newName,
			ApplicationDescription: &newDesc,
		}

		result, err := service.UpdateApplication(application.ApplicationID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.ApplicationName)
		if result.ApplicationDescription != nil {
			assert.Equal(t, newDesc, *result.ApplicationDescription)
		}

		// Verify database was updated
		var updatedApp models.Application
		err = db.Where("application_id = ?", application.ApplicationID).First(&updatedApp).Error
		assert.NoError(t, err)
		assert.Equal(t, newName, updatedApp.ApplicationName)
	})

	t.Run("UpdateApplication_NotFound", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		newName := "Updated Name"
		req := &models.UpdateApplicationRequest{
			ApplicationName: &newName,
		}

		result, err := service.UpdateApplication("non-existent-id", req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestApplicationService_GetApplication(t *testing.T) {
	t.Run("GetApplication_Success", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		// Create an application using GORM
		application := models.Application{
			ApplicationID:          "app_123",
			ApplicationName:        "Test Application",
			ApplicationDescription: stringPtr("Test Description"),
			SelectedFields: models.SelectedFieldRecords{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
			Version:  string(models.ActiveVersion),
		}
		err := db.Create(&application).Error
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		result, err := service.GetApplication("app_123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "app_123", result.ApplicationID)
		assert.Equal(t, "Test Application", result.ApplicationName)
	})

	t.Run("GetApplication_NotFound", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		result, err := service.GetApplication("non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestApplicationService_GetApplications(t *testing.T) {
	t.Run("GetApplications_NoFilter", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		// Create multiple applications using GORM (now works with SQLite)
		applications := []models.Application{
			{
				ApplicationID:   "app_1",
				ApplicationName: "App 1",
				MemberID:        "member-1",
				SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-1"}},
				Version:         string(models.ActiveVersion),
			},
			{
				ApplicationID:   "app_2",
				ApplicationName: "App 2",
				MemberID:        "member-2",
				SelectedFields:  models.SelectedFieldRecords{{FieldName: "field2", SchemaID: "schema-2"}},
				Version:         string(models.ActiveVersion),
			},
		}
		for _, app := range applications {
			err := db.Create(&app).Error
			assert.NoError(t, err)
		}

		result, err := service.GetApplications(nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("GetApplications_WithMemberIDFilter", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		memberID := "member-123"
		application := models.Application{
			ApplicationID:   "app_1",
			ApplicationName: "App 1",
			MemberID:        memberID,
			SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			Version:         string(models.ActiveVersion),
		}
		err := db.Create(&application).Error
		assert.NoError(t, err)

		result, err := service.GetApplications(&memberID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, memberID, result[0].MemberID)
	})
}

func TestApplicationService_CreateApplicationSubmission(t *testing.T) {
	t.Run("CreateApplicationSubmission_Success", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		// Create a member first
		member := models.Member{
			MemberID:    "member-123",
			Name:        "Test Member",
			Email:       "test@example.com",
			PhoneNumber: "1234567890",
		}
		err := db.Create(&member).Error
		assert.NoError(t, err)

		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Submission",
			ApplicationDescription: stringPtr("Test Description"),
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: member.MemberID,
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.ApplicationName, result.ApplicationName)
		assert.NotEmpty(t, result.SubmissionID)
		assert.Equal(t, string(models.StatusPending), result.Status)
	})

	t.Run("CreateApplicationSubmission_MemberNotFound", func(t *testing.T) {
		db := setupTestDBForApplication(t)
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Submission",
			ApplicationDescription: stringPtr("Test Description"),
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "non-existent-member",
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
