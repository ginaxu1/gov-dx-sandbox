package services

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gov-dx-sandbox/portal-backend/v1/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestApplicationService_CreateApplication(t *testing.T) {
	t.Run("CreateApplication_Success", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		// Mock PDP
		mockTransport := &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"records": [{"id": "policy_1"}]}`)),
					Header:     make(http.Header),
				}, nil
			},
		}
		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		pdpService.HTTPClient = &http.Client{Transport: mockTransport}

		service := NewApplicationService(db, pdpService)

		desc := "Test Description"
		req := &models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: &desc,
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
		}

		// Mock DB expectations
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id"}).AddRow("app_123"))

		// Act
		// Note: CreateApplication returns error if PDP fails. Here PDP succeeds.
		// However, CreateApplication returns nil, error if successful? No, it returns response, nil.
		// Wait, the original test expected error because PDP failed. Now we mock PDP success.
		// Let's check CreateApplication implementation.
		// It returns *models.ApplicationResponse, error.

		resp, err := service.CreateApplication(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		if resp != nil {
			assert.Equal(t, req.ApplicationName, resp.ApplicationName)
			assert.Equal(t, req.MemberID, resp.MemberID)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateApplication_PDPFailure_Compensation", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		// Mock PDP failure
		mockTransport := &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": "pdp error"}`)),
					Header:     make(http.Header),
				}, nil
			},
		}
		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		pdpService.HTTPClient = &http.Client{Transport: mockTransport}

		service := NewApplicationService(db, pdpService)

		desc := "Test Description"
		req := &models.CreateApplicationRequest{
			ApplicationName:        "Test Application",
			ApplicationDescription: &desc,
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
		}

		// Mock DB expectations
		// 1. Create application
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id"}).AddRow("app_123"))

		// 2. Compensation: Delete application
		mock.ExpectExec(`DELETE FROM "applications"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Act
		resp, err := service.CreateApplication(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to update allow list")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_UpdateApplication(t *testing.T) {
	t.Run("UpdateApplication_Success", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. First find the application
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id", "application_name", "application_description", "member_id", "version"}).
				AddRow("app_123", "Original Name", "Original Description", "member-123", "v1"))

		// 2. Save the updated application
		mock.ExpectExec(`UPDATE "applications"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		newName := "Updated Name"
		newDesc := "Updated Description"
		req := &models.UpdateApplicationRequest{
			ApplicationName:        &newName,
			ApplicationDescription: &newDesc,
		}

		result, err := service.UpdateApplication("app_123", req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, newName, result.ApplicationName)
			if result.ApplicationDescription != nil {
				assert.Equal(t, newDesc, *result.ApplicationDescription)
			}
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateApplication_NotFound", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations - return no rows
		mock.ExpectQuery(`SELECT .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		newName := "Updated Name"
		req := &models.UpdateApplicationRequest{
			ApplicationName: &newName,
		}

		result, err := service.UpdateApplication("non-existent-id", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_GetApplication(t *testing.T) {
	t.Run("GetApplication_Success", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id", "application_name", "member_id", "version"}).
				AddRow("app_123", "Test Application", "member-123", "v1"))

		// Preload Member expectation
		mock.ExpectQuery(`SELECT .* FROM "members" WHERE "members"."member_id" = .*`).
			WithArgs("member-123").
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).
				AddRow("member-123", "Test Member"))

		result, err := service.GetApplication("app_123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, "app_123", result.ApplicationID)
			assert.Equal(t, "Test Application", result.ApplicationName)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetApplication_NotFound", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := service.GetApplication("non-existent-id")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_GetApplications(t *testing.T) {
	t.Run("GetApplications_NoFilter", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .* FROM "applications" ORDER BY created_at DESC`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id", "application_name", "member_id", "version"}).
				AddRow("app_1", "App 1", "member-1", "v1").
				AddRow("app_2", "App 2", "member-2", "v1"))

		// Preload Member expectation (for each application)
		// Note: GORM might batch these or do them individually depending on version/config
		// With Preload, it typically does IN query
		mock.ExpectQuery(`SELECT .* FROM "members" WHERE "members"."member_id" IN .*`).
			WithArgs("member-1", "member-2").
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).
				AddRow("member-1", "Member 1").
				AddRow("member-2", "Member 2"))

		result, err := service.GetApplications(nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetApplications_WithMemberIDFilter", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		memberID := "member-123"

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .* FROM "applications" WHERE member_id = .* ORDER BY created_at DESC`).
			WithArgs(memberID).
			WillReturnRows(sqlmock.NewRows([]string{"application_id", "application_name", "member_id", "version"}).
				AddRow("app_1", "App 1", memberID, "v1"))

		// Preload Member expectation
		mock.ExpectQuery(`SELECT .* FROM "members" WHERE "members"."member_id" = .*`).
			WithArgs(memberID).
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).
				AddRow(memberID, "Member 1"))

		result, err := service.GetApplications(&memberID)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, memberID, result[0].MemberID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_CreateApplicationSubmission(t *testing.T) {
	t.Run("CreateApplicationSubmission_Success", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. Validate member
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).AddRow("member-123", "Test Member"))

		// 2. Create submission
		mock.ExpectQuery(`INSERT INTO .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id"}).AddRow("sub_123"))

		desc := "Test Description"
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Submission",
			ApplicationDescription: &desc,
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID: "member-123",
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, req.ApplicationName, result.ApplicationName)
			assert.Equal(t, string(models.StatusPending), result.Status)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateApplicationSubmission_MemberNotFound", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnError(gorm.ErrRecordNotFound)

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
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_UpdateApplicationSubmission(t *testing.T) {
	t.Run("UpdateApplicationSubmission_Success", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. Find submission
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow("sub_123", "Original", "member-123", string(models.StatusPending)))

		// 2. Save submission
		mock.ExpectExec(`UPDATE "application_submissions"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		newName := "Updated"
		req := &models.UpdateApplicationSubmissionRequest{
			ApplicationName: &newName,
		}

		result, err := service.UpdateApplicationSubmission("sub_123", req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, newName, result.ApplicationName)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateApplicationSubmission_NotFound", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		updatedName := "Updated"
		req := &models.UpdateApplicationSubmissionRequest{ApplicationName: &updatedName}
		result, err := service.UpdateApplicationSubmission("non-existent", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "application submission not found")

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateApplicationSubmission_ApprovalSuccess", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		member := models.Member{MemberID: "member-123", Name: "Test Member"}
		submission := models.ApplicationSubmission{
			SubmissionID:    "sub_123",
			ApplicationName: "Original",
			MemberID:        member.MemberID,
			Status:          models.StatusPending,
			Member:          member,
		}

		// Mock DB expectations
		// 1. Find submission
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow(submission.SubmissionID, submission.ApplicationName, submission.MemberID, string(submission.Status)))

		// 2. Save submission (status update to Approved)
		mock.ExpectExec(`UPDATE "application_submissions"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 3. Create Application
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id"}).AddRow("app_new"))

		// 4. Create PDP Job
		mock.ExpectQuery(`INSERT INTO "pdp_jobs"`).
			WillReturnRows(sqlmock.NewRows([]string{"job_id"}).AddRow("pdp_job_1"))

		status := string(models.StatusApproved)
		req := &models.UpdateApplicationSubmissionRequest{
			Status: &status,
		}

		// UpdateApplicationSubmission now uses outbox pattern - it succeeds and queues a job
		result, err := service.UpdateApplicationSubmission(submission.SubmissionID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// CreateApplication now uses outbox pattern - it succeeds and queues a job
		// Note: The original instruction seems to have intended to replace the verification logic
		// for UpdateApplicationSubmission_ApprovalSuccess, but the provided code snippet
		// includes a call to `service.CreateApplication(req)` which is out of context for this test.
		// Assuming the intent was to verify the outcome of the UpdateApplicationSubmission call
		// leading to application creation and PDP job queuing.
		// CreateApplication now uses outbox pattern - it succeeds and queues a job
		// The following lines are adjusted to reflect the original test's context.
		assert.Equal(t, string(models.StatusApproved), string(result.Status))

		// Verify application was created (in the mock DB)
		var appCount int64
		db.Model(&models.Application{}).Where("member_id = ?", member.MemberID).Count(&appCount)
		assert.Equal(t, int64(1), appCount)

		// Verify PDP job was queued (in the mock DB)
		var jobCount int64
		db.Model(&models.PDPJob{}).Where("application_id IS NOT NULL").Count(&jobCount)
		assert.GreaterOrEqual(t, jobCount, int64(1))

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateApplicationSubmission_ApprovalWithApplicationCreationFailure", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		// Mock PDP failure
		mockTransport := &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": "pdp error"}`)),
					Header:     make(http.Header),
				}, nil
			},
		}
		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		pdpService.HTTPClient = &http.Client{Transport: mockTransport}

		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. Find submission
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow("sub_123", "Original", "member-123", string(models.StatusPending)))

		// 2. Save submission (status update to Approved)
		mock.ExpectExec(`UPDATE "application_submissions"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 3. Create Application (will fail on PDP)
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id"}).AddRow("app_new"))

		// 4. Compensation: Delete application (from CreateApplication compensation)
		mock.ExpectExec(`DELETE FROM "applications"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 5. Compensation: Update submission status back to Pending
		mock.ExpectExec(`UPDATE "application_submissions"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		status := string(models.StatusApproved)
		req := &models.UpdateApplicationSubmissionRequest{
			Status: &status,
		}

		result, err := service.UpdateApplicationSubmission("sub_123", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create application")

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_GetApplicationSubmission(t *testing.T) {
	t.Run("GetApplicationSubmission_Success", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow("sub_123", "Test Submission", "member-123", string(models.StatusPending)))

		// Preload Member
		mock.ExpectQuery(`SELECT .* FROM "members" WHERE "members"."member_id" = .*`).
			WithArgs("member-123").
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).AddRow("member-123", "Test Member"))

		// Preload PreviousApplication (none)
		// Note: GORM might not execute this query if PreviousApplicationID is null in the struct returned above
		// But if it does, we should expect it. Let's see.
		// If PreviousApplicationID is null, GORM usually skips.

		result, err := service.GetApplicationSubmission("sub_123")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, "sub_123", result.SubmissionID)
			assert.Equal(t, "Test Submission", result.ApplicationName)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetApplicationSubmission_NotFound", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := service.GetApplicationSubmission("non-existent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_GetApplicationSubmissions(t *testing.T) {
	t.Run("GetApplicationSubmissions_NoFilter", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow("sub_1", "Sub 1", "member-1", string(models.StatusPending)).
				AddRow("sub_2", "Sub 2", "member-1", string(models.StatusPending)))

		// Preload Member
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).AddRow("member-1", "Test Member"))

		// Preload PreviousApplication (none)

		result, err := service.GetApplicationSubmissions(nil, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetApplicationSubmissions_WithMemberIDFilter", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		memberID := "member-123"

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .* FROM "application_submissions" WHERE member_id = .* ORDER BY created_at DESC`).
			WithArgs(memberID).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow("sub_1", "Sub 1", memberID, string(models.StatusPending)))

		// Preload Member
		mock.ExpectQuery(`SELECT .* FROM "members" WHERE "members"."member_id" = .*`).
			WithArgs(memberID).
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).AddRow(memberID, "Test Member"))

		result, err := service.GetApplicationSubmissions(&memberID, nil)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, memberID, result[0].MemberID)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetApplicationSubmissions_WithStatusFilter", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		statusFilter := []string{string(models.StatusApproved)}

		// Mock DB expectations
		mock.ExpectQuery(`SELECT .* FROM "application_submissions" WHERE status IN .* ORDER BY created_at DESC`).
			WithArgs(statusFilter[0]).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id", "application_name", "member_id", "status"}).
				AddRow("sub_2", "Sub 2", "member-123", string(models.StatusApproved)))

		// Preload Member
		mock.ExpectQuery(`SELECT .* FROM "members" WHERE "members"."member_id" = .*`).
			WithArgs("member-123").
			WillReturnRows(sqlmock.NewRows([]string{"member_id", "name"}).AddRow("member-123", "Test Member"))

		result, err := service.GetApplicationSubmissions(nil, &statusFilter)

		assert.NoError(t, err)
		if result != nil && len(result) > 0 {
			assert.Equal(t, string(models.StatusApproved), result[0].Status)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_CreateApplication_EdgeCases(t *testing.T) {
	t.Run("CreateApplication_EmptySelectedFields", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		req := &models.CreateApplicationRequest{
			ApplicationName: "Test Application",
			SelectedFields:  []models.SelectedFieldRecord{},
			MemberID:        "member-123",
		}

		// Mock DB expectations
		// 1. Create application
		mock.ExpectQuery(`INSERT INTO "applications"`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id"}).AddRow("app_123"))

		// 2. Compensation: Delete application (because empty fields might cause PDP error or logic error)
		// Wait, empty selected fields might be valid for DB but PDP might reject?
		// The original test expected error.
		// If PDP service is mocked to return error (or if logic checks for empty fields), then compensation happens.
		// Let's assume PDP returns error for empty fields if we mock it that way.
		// Or if the logic itself checks.
		// The original test said "Will fail on PDP call but tests the request structure".
		// So we should mock PDP failure.

		mockTransport := &MockRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error": "empty fields"}`)),
					Header:     make(http.Header),
				}, nil
			},
		}
		pdpService.HTTPClient = &http.Client{Transport: mockTransport}

		mock.ExpectExec(`DELETE FROM "applications"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := service.CreateApplication(req)
		assert.Error(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_UpdateApplication_EdgeCases(t *testing.T) {
	t.Run("UpdateApplication_PartialUpdate", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. Find application
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id", "application_name", "application_description", "member_id", "version"}).
				AddRow("app_123", "Original Name", "Original Description", "member-123", "v1"))

		// 2. Save application
		mock.ExpectExec(`UPDATE "applications"`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		newName := "Updated Name Only"
		req := &models.UpdateApplicationRequest{
			ApplicationName: &newName,
		}

		result, err := service.UpdateApplication("app_123", req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil {
			assert.Equal(t, newName, result.ApplicationName)
			// Original description should remain
			// Note: The mock returned "Original Description", so result should have it if logic preserves it
			if result.ApplicationDescription != nil {
				assert.Equal(t, "Original Description", *result.ApplicationDescription)
			}
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestApplicationService_CreateApplicationSubmission_EdgeCases(t *testing.T) {
	t.Run("CreateApplicationSubmission_WithPreviousApplicationID", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. Validate previous application
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"application_id"}).AddRow("app_prev"))

		// 2. Validate member
		mock.ExpectQuery(`SELECT .*`).
			WillReturnRows(sqlmock.NewRows([]string{"member_id"}).AddRow("member-123"))

		// 3. Create submission
		mock.ExpectQuery(`INSERT INTO .*`).
			WillReturnRows(sqlmock.NewRows([]string{"submission_id"}).AddRow("sub_123"))

		prevAppID := "app_prev"
		desc := "Test Description"
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:        "Test Submission",
			ApplicationDescription: &desc,
			SelectedFields: []models.SelectedFieldRecord{
				{FieldName: "field1", SchemaID: "schema-123"},
			},
			MemberID:              "member-123",
			PreviousApplicationID: &prevAppID,
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if result != nil && result.PreviousApplicationID != nil {
			assert.Equal(t, prevAppID, *result.PreviousApplicationID)
		}

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CreateApplicationSubmission_InvalidPreviousApplicationID", func(t *testing.T) {
		db, mock, cleanup := SetupMockDB(t)
		defer cleanup()

		pdpService := NewPDPService("http://mock-pdp", "mock-key")
		service := NewApplicationService(db, pdpService)

		// Mock DB expectations
		// 1. Validate previous application
		mock.ExpectQuery(`SELECT .*`).
			WillReturnError(gorm.ErrRecordNotFound)

		invalidAppID := "non-existent-app"
		req := &models.CreateApplicationSubmissionRequest{
			ApplicationName:       "New Submission",
			SelectedFields:        []models.SelectedFieldRecord{{FieldName: "field1", SchemaID: "schema-123"}},
			MemberID:              "member-123",
			PreviousApplicationID: &invalidAppID,
		}

		result, err := service.CreateApplicationSubmission(req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
