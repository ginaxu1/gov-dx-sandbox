package services

import (
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/v1/models"
	"github.com/stretchr/testify/assert"
)

func TestApplicationService_CreateApplication(t *testing.T) {
	t.Run("CreateApplication_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		// Use a real PDPService - it will fail on HTTP calls but we're testing DB operations
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		req := &models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: "Test Description",
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		// Create an application first using GORM (now works with SQLite)
		application := models.Application{
			ApplicationID:          "app_123",
			ApplicationName:        "Original Name",
			ApplicationDescription: "Original Description",
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		// Create an application using GORM
		desc := "Test Description"
		application := models.Application{
			ApplicationID:          "app_123",
			ApplicationName:        "Test Application",
			ApplicationDescription: desc,
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		result, err := service.GetApplication("non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestApplicationService_GetApplications(t *testing.T) {
	t.Run("GetApplications_NoFilter", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
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

		desc := "Test Description"
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Submission",
			ApplicationDescription: &desc,
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
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		desc := "Test Description"
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Submission",
			ApplicationDescription: &desc,
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

func TestApplicationService_UpdateApplicationSubmission(t *testing.T) {
	t.Run("UpdateApplicationSubmission_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.ApplicationSubmission{
			SubmissionID:    "sub_123",
			ApplicationName: "Original",
			SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:        member.MemberID,
			Status:          string(models.StatusPending),
		}
		db.Create(&submission)

		newName := "Updated"
		req := &models.UpdateApplicationSubmissionRequest{
			ApplicationName: &newName,
		}

		result, err := service.UpdateApplicationSubmission(submission.SubmissionID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.ApplicationName)
	})

	t.Run("UpdateApplicationSubmission_NotFound", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		updatedName := "Updated"
		req := &models.UpdateApplicationSubmissionRequest{ApplicationName: &updatedName}
		result, err := service.UpdateApplicationSubmission("non-existent", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "application submission not found")
	})

	t.Run("UpdateApplicationSubmission_ApprovalWithApplicationCreationFailure", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.ApplicationSubmission{
			SubmissionID:    "sub_123",
			ApplicationName: "Original",
			SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:        member.MemberID,
			Status:          string(models.StatusPending),
		}
		db.Create(&submission)

		status := string(models.StatusApproved)
		req := &models.UpdateApplicationSubmissionRequest{
			Status: &status,
		}

		// This will fail because PDP service is not available, but tests compensation logic
		result, err := service.UpdateApplicationSubmission(submission.SubmissionID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		// Verify submission status was rolled back to pending
		var updatedSubmission models.ApplicationSubmission
		db.Where("submission_id = ?", submission.SubmissionID).First(&updatedSubmission)
		assert.Equal(t, string(models.StatusPending), updatedSubmission.Status)
	})
}

func TestApplicationService_GetApplicationSubmission(t *testing.T) {
	t.Run("GetApplicationSubmission_Success", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.ApplicationSubmission{
			SubmissionID:    "sub_123",
			ApplicationName: "Test Submission",
			SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:        member.MemberID,
			Status:          string(models.StatusPending),
		}
		db.Create(&submission)

		result, err := service.GetApplicationSubmission(submission.SubmissionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, submission.SubmissionID, result.SubmissionID)
		assert.Equal(t, submission.ApplicationName, result.ApplicationName)
	})

	t.Run("GetApplicationSubmission_NotFound", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		result, err := service.GetApplicationSubmission("non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestApplicationService_GetApplicationSubmissions(t *testing.T) {
	t.Run("GetApplicationSubmissions_NoFilter", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submissions := []models.ApplicationSubmission{
			{SubmissionID: "sub_1", ApplicationName: "Sub 1", SelectedFields: models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-1"}}, MemberID: member.MemberID, Status: string(models.StatusPending)},
			{SubmissionID: "sub_2", ApplicationName: "Sub 2", SelectedFields: models.SelectedFieldRecords{{FieldName: "field2", SchemaID: "schema-2"}}, MemberID: member.MemberID, Status: string(models.StatusPending)},
		}
		for _, s := range submissions {
			db.Create(&s)
		}

		result, err := service.GetApplicationSubmissions(nil, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("GetApplicationSubmissions_WithMemberIDFilter", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		memberID := "member-123"
		member := models.Member{MemberID: memberID, Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submission := models.ApplicationSubmission{
			SubmissionID:    "sub_1",
			ApplicationName: "Sub 1",
			SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:        memberID,
			Status:          string(models.StatusPending),
		}
		db.Create(&submission)

		result, err := service.GetApplicationSubmissions(&memberID, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, memberID, result[0].MemberID)
	})

	t.Run("GetApplicationSubmissions_WithStatusFilter", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)
		submissions := []models.ApplicationSubmission{
			{SubmissionID: "sub_1", ApplicationName: "Sub 1", SelectedFields: models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-1"}}, MemberID: member.MemberID, Status: string(models.StatusPending)},
			{SubmissionID: "sub_2", ApplicationName: "Sub 2", SelectedFields: models.SelectedFieldRecords{{FieldName: "field2", SchemaID: "schema-2"}}, MemberID: member.MemberID, Status: string(models.StatusApproved)},
		}
		for _, s := range submissions {
			db.Create(&s)
		}

		statusFilter := []string{string(models.StatusApproved)}
		result, err := service.GetApplicationSubmissions(nil, &statusFilter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, string(models.StatusApproved), result[0].Status)
	})
}

func TestApplicationService_CreateApplication_EdgeCases(t *testing.T) {
	t.Run("CreateApplication_EmptySelectedFields", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		req := &models.CreateApplicationRequest{
			ApplicationName: "Test Application",
			SelectedFields:  []models.SelectedFieldRecord{},
			MemberID:        "member-123",
		}

		// Will fail on PDP call but tests the request structure
		_, err := service.CreateApplication(req)
		assert.Error(t, err)
		// Verify application was rolled back
		var count int64
		db.Model(&models.Application{}).Where("application_name = ?", req.ApplicationName).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CreateApplication_WithDescription", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		req := &models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: "Test Description",
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
		}

		// Will fail on PDP call but tests the request structure with description
		_, err := service.CreateApplication(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update allow list")
		// Verify application was rolled back
		var count int64
		db.Model(&models.Application{}).Where("application_name = ?", req.ApplicationName).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestApplicationService_UpdateApplication_EdgeCases(t *testing.T) {
	t.Run("UpdateApplication_PartialUpdate", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		application := models.Application{
			ApplicationID:          "app_123",
			ApplicationName:        "Original Name",
			ApplicationDescription: "Original Description",
			SelectedFields:         models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:               "member-123",
			Version:                string(models.ActiveVersion),
		}
		db.Create(&application)

		// Only update name
		newName := "Updated Name Only"
		req := &models.UpdateApplicationRequest{
			ApplicationName: &newName,
		}

		result, err := service.UpdateApplication(application.ApplicationID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.ApplicationName)
		// Original description should remain
		if application.ApplicationDescription != "" {
			assert.NotNil(t, result.ApplicationDescription)
			assert.Equal(t, application.ApplicationDescription, *result.ApplicationDescription)
		}
	})
}

func TestApplicationService_CreateApplicationSubmission_EdgeCases(t *testing.T) {
	t.Run("CreateApplicationSubmission_WithPreviousApplicationID", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)

		// Create a previous application
		previousApp := models.Application{
			ApplicationID:   "app_prev",
			ApplicationName: "Previous App",
			SelectedFields:  models.SelectedFieldRecords{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:        member.MemberID,
			Version:         string(models.ActiveVersion),
		}
		db.Create(&previousApp)

		previousAppID := previousApp.ApplicationID
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:       "New Submission",
			SelectedFields:        []models.SelectedFieldRecord{{FieldName: "field2", SchemaID: "schema-456"}},
			MemberID:              member.MemberID,
			PreviousApplicationID: &previousAppID,
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, previousAppID, *result.PreviousApplicationID)
	})

	t.Run("CreateApplicationSubmission_InvalidPreviousApplicationID", func(t *testing.T) {
		db := SetupPostgresTestDB(t)
		if db == nil {
			return
		}
		pdpService := NewPDPService("http://localhost:9999", "test-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test", Email: "test@example.com", PhoneNumber: "123"}
		db.Create(&member)

		invalidAppID := "non-existent-app"
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:       "New Submission",
			SelectedFields:        []models.SelectedFieldRecord{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:              member.MemberID,
			PreviousApplicationID: &invalidAppID,
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
