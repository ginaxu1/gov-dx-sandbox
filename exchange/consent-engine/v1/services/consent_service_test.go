package services

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return gormDB, mock
}

func TestNewConsentService(t *testing.T) {
	db, _ := setupMockDB(t)

	tests := []struct {
		name                 string
		consentPortalBaseURL string
		expectError          bool
	}{
		{
			name:                 "Valid URL",
			consentPortalBaseURL: "http://localhost:5173",
			expectError:          false,
		},
		{
			name:                 "Invalid URL - Empty",
			consentPortalBaseURL: "",
			expectError:          true,
		},
		{
			name:                 "Invalid URL - No Scheme",
			consentPortalBaseURL: "localhost:5173",
			expectError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewConsentService(db, tt.consentPortalBaseURL)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestGetConsentInternalView_ByID(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	idStr := id.String()

	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "status", "created_at", "updated_at", "grant_duration"}).
		AddRow(id, "user-1", "approved", time.Now(), time.Now(), "P30D")

	mock.ExpectQuery(`SELECT \* FROM "consent_records" WHERE consent_id = \$1 ORDER BY created_at DESC.* LIMIT \$2`).
		WithArgs(id, 1). // GORM uses numeric args for postgres
		WillReturnRows(rows)

	resp, err := service.GetConsentInternalView(ctx, &idStr, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, idStr, resp.ConsentID)
}

func TestUpdateConsentStatusByPortalAction(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	idStr := id.String()
	updatedBy := "user-action"

	req := models.ConsentPortalActionRequest{
		ConsentID: idStr,
		Action:    models.ActionApprove,
		UpdatedBy: updatedBy,
	}

	// Mock finding the record
	rows := sqlmock.NewRows([]string{"consent_id", "status", "grant_duration"}).
		AddRow(id, "pending", "P30D")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1 ORDER BY "consent_records"."consent_id" LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	// Mock updating the record
	// GORM updates can be complex, often inside transaction or save
	// Expect UPDATE
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consent_records"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.UpdateConsentStatusByPortalAction(ctx, req)
	require.NoError(t, err)

	// Verify expectations
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateConsentRecord_New(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	consentType := models.TypeRealtime
	grantDuration := string(models.DurationOneDay)

	req := models.CreateConsentRequest{
		AppID:   "app-1",
		AppName: nil,
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user-1@example.com",
			Fields: []models.ConsentField{
				{FieldName: "email", SchemaID: "schema-1", Owner: "citizen"},
			},
		},
		ConsentType:   &consentType,
		GrantDuration: &grantDuration,
	}

	// Mock GetConsentInternalView returning RecordNotFound
	// Query: SELECT * FROM "consent_records" WHERE owner_id = ? AND app_id = ? ORDER BY created_at DESC LIMIT 1
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock Create
	// mock.ExpectBegin() // SkipDefaultTransaction is true
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
		WillReturnRows(sqlmock.NewRows([]string{"consent_id"}).AddRow(uuid.New()))
	// mock.ExpectCommit()

	resp, err := service.CreateConsentRecord(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, string(models.StatusPending), resp.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateConsentRecord_ExistingMatch(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	fields := []models.ConsentField{{FieldName: "email", SchemaID: "schema-1", Owner: "citizen"}}
	req := models.CreateConsentRequest{
		AppID: "app-1",
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user-1@example.com",
			Fields:     fields,
		},
	}

	// Mock GetConsentInternalView returning existing Pending consent
	rows := sqlmock.NewRows([]string{"consent_id", "status", "fields"}).
		AddRow(uuid.New(), "pending", `[{"fieldName":"email","schemaId":"schema-1","owner":"citizen"}]`) // JSONB match

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnRows(rows)

	resp, err := service.CreateConsentRecord(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, string(models.StatusPending), resp.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateConsentRecord_RevokeAndCreate(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	// New fields differ from stored fields
	req := models.CreateConsentRequest{
		AppID: "app-1",
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user-1@example.com",
			Fields:     []models.ConsentField{{FieldName: "new_field", SchemaID: "schema-1", Owner: "citizen"}},
		},
	}

	existID := uuid.New()

	// Mock GetConsentInternalView returning existing Approved consent with old fields
	rows := sqlmock.NewRows([]string{"consent_id", "status", "fields"}).
		AddRow(existID, "approved", `[{"fieldName":"old_field","schemaId":"schema-1","owner":"citizen"}]`)

	// Note: GORM select
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnRows(rows)

	// Expect Transaction Begin (RevokeAndCreate uses transaction)
	mock.ExpectBegin()

	// Revoke: Find Existing to Revoke
	rowsExist := sqlmock.NewRows([]string{"consent_id", "status"}).
		AddRow(existID, "approved")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)).
		WithArgs(existID, 1). // LIMIT 1
		WillReturnRows(rowsExist)

	// Revoke: Update Status
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consent_records"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create: Insert New
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
		WillReturnRows(sqlmock.NewRows([]string{"consent_id"}).AddRow(uuid.New()))

	mock.ExpectCommit()

	resp, err := service.CreateConsentRecord(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConsentInternalView_ByOwnerApp(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	ownerID := "user-1"
	appID := "app-1"
	id := uuid.New()

	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "status", "created_at", "updated_at", "grant_duration"}).
		AddRow(id, "user-1", "approved", time.Now(), time.Now(), "P30D")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)).
		WithArgs(ownerID, appID, 1). // LIMIT 1
		WillReturnRows(rows)

	resp, err := service.GetConsentInternalView(ctx, nil, &ownerID, nil, &appID)
	require.NoError(t, err)
	assert.Equal(t, id.String(), resp.ConsentID)
}

func TestGetConsentInternalView_ExpiredPending(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	idStr := id.String()

	// Pending expired
	expiredTime := time.Now().Add(-2 * time.Hour)
	rows := sqlmock.NewRows([]string{"consent_id", "status", "pending_expires_at"}).
		AddRow(id, "pending", expiredTime)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)).
		WithArgs(id, 1). // LIMIT 1
		WillReturnRows(rows)

	// Expect Update to Expired
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consent_records"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := service.GetConsentInternalView(ctx, &idStr, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, string(models.StatusExpired), resp.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateConsentRecord_InvalidInput(t *testing.T) {
	db, _ := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	// Missing AppID
	req := models.CreateConsentRequest{
		AppID: "",
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user-1@example.com",
			Fields:     []models.ConsentField{{FieldName: "email"}},
		},
	}

	_, err := service.CreateConsentRecord(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "appId is required")
}

func TestGetConsentInternalView_ExpiredApproved(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	idStr := id.String()

	// Approved expired
	expiredTime := time.Now().Add(-2 * time.Hour)
	rows := sqlmock.NewRows([]string{"consent_id", "status", "grant_expires_at"}).
		AddRow(id, "approved", expiredTime)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)).
		WithArgs(id, 1). // LIMIT 1
		WillReturnRows(rows)

	// Expect Update to Expired
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consent_records"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := service.GetConsentInternalView(ctx, &idStr, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, string(models.StatusExpired), resp.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateConsentRecord_DurationsAndTypes(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	durations := []models.GrantDuration{
		models.DurationOneHour,
		models.DurationSixHours,
		models.DurationTwelveHours,
		models.DurationSevenDays,
		models.DurationThirtyDays,
	}

	for _, dur := range durations {
		d := string(dur)
		ct := models.TypeOffline
		req := models.CreateConsentRequest{
			AppID: "app-test",
			ConsentRequirement: models.ConsentRequirement{
				OwnerID:    "user-1",
				OwnerEmail: "user-1",
				Fields:     []models.ConsentField{{FieldName: "f", SchemaID: "s", Owner: "citizen"}},
			},
			ConsentType:   &ct,
			GrantDuration: &d,
		}

		// Mock Get Not Found - specific query for owner_id and app_id
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
			WithArgs(req.ConsentRequirement.OwnerID, req.AppID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		// Mock Insert
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
			WillReturnRows(sqlmock.NewRows([]string{"consent_id"}).AddRow(uuid.New()))

		_, err := service.CreateConsentRecord(ctx, req)
		require.NoError(t, err)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeConsent(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	idStr := id.String()
	revokedBy := "user-revoke"

	// Mock Transaction
	mock.ExpectBegin()

	// Find record
	rows := sqlmock.NewRows([]string{"consent_id", "status"}).
		AddRow(id, "approved")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1 ORDER BY "consent_records"."consent_id" LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	// Save (Update)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consent_records"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.RevokeConsent(ctx, idStr, revokedBy)
	require.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConsentInternalView_ByOwnerEmail(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	ownerEmail := "user@example.com"
	appID := "app-1"
	id := uuid.New()

	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "status", "created_at", "updated_at", "grant_duration"}).
		AddRow(id, "user-1", "approved", time.Now(), time.Now(), "P30D")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_email = $1 AND app_id = $2 ORDER BY created_at DESC`)).
		WithArgs(ownerEmail, appID, 1). // LIMIT 1
		WillReturnRows(rows)

	resp, err := service.GetConsentInternalView(ctx, nil, nil, &ownerEmail, &appID)
	require.NoError(t, err)
	assert.Equal(t, id.String(), resp.ConsentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConsentInternalView_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	idStr := id.String()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrInvalidDB)

	_, err := service.GetConsentInternalView(ctx, &idStr, nil, nil, nil)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateConsentStatusByPortalAction_InvalidAction(t *testing.T) {
	db, _ := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	req := models.ConsentPortalActionRequest{
		Action: "INVALID",
	}

	err := service.UpdateConsentStatusByPortalAction(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestUpdateConsentStatusByPortalAction_InvalidUUID(t *testing.T) {
	db, _ := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	req := models.ConsentPortalActionRequest{
		ConsentID: "invalid-uuid",
		Action:    models.ActionApprove,
		UpdatedBy: "user@example.com",
	}

	err := service.UpdateConsentStatusByPortalAction(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid consent ID")
}

func TestUpdateConsentStatusByPortalAction_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	req := models.ConsentPortalActionRequest{
		ConsentID: id.String(),
		Action:    models.ActionApprove,
		UpdatedBy: "user@example.com",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)+".*"+regexp.QuoteMeta(`LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	err := service.UpdateConsentStatusByPortalAction(ctx, req)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateConsentStatusByPortalAction_Reject(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "status", "grant_duration"}).
		AddRow(id, "user-1", "pending", "P30D")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)+".*"+regexp.QuoteMeta(`LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consent_records"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := models.ConsentPortalActionRequest{
		ConsentID: id.String(),
		Action:    models.ActionReject,
		UpdatedBy: "user@example.com",
	}

	err := service.UpdateConsentStatusByPortalAction(ctx, req)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeConsent_InvalidUUID(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	// UUID parsing happens inside transaction, so transaction begins then errors
	mock.ExpectBegin()
	mock.ExpectRollback()

	err := service.RevokeConsent(ctx, "invalid-uuid", "user@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid consent ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeConsent_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)+".*"+regexp.QuoteMeta(`LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := service.RevokeConsent(ctx, id.String(), "user@example.com")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRevokeConsent_WrongStatus(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"consent_id", "status"}).
		AddRow(id, "rejected")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)+".*"+regexp.QuoteMeta(`LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnRows(rows)
	mock.ExpectRollback()

	err := service.RevokeConsent(ctx, id.String(), "user@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only approved or pending consents can be revoked")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConsentPortalView_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "owner_email", "app_id", "status", "type", "created_at", "updated_at", "grant_duration", "fields"}).
		AddRow(id, "user-1", "user@example.com", "app-1", "approved", "realtime", time.Now(), time.Now(), "P30D", "[]")

	// GORM First() adds LIMIT 1
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)+".*"+regexp.QuoteMeta(`LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnRows(rows)

	resp, err := service.GetConsentPortalView(ctx, id.String())
	require.NoError(t, err)
	assert.Equal(t, "user-1", resp.OwnerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetConsentPortalView_InvalidUUID(t *testing.T) {
	db, _ := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	_, err := service.GetConsentPortalView(ctx, "invalid-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid consent ID")
}

func TestGetConsentPortalView_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	service, _ := NewConsentService(db, "http://portal")
	ctx := context.Background()

	id := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE consent_id = $1`)+".*"+regexp.QuoteMeta(`LIMIT $2`)).
		WithArgs(id, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := service.GetConsentPortalView(ctx, id.String())
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestParseGrantDuration_AllCases(t *testing.T) {
	// Test valid grant durations only (invalid ones are rejected by validation)
	validDurations := []models.GrantDuration{
		models.DurationOneHour,
		models.DurationSixHours,
		models.DurationTwelveHours,
		models.DurationOneDay,
		models.DurationSevenDays,
		models.DurationThirtyDays,
	}

	for _, dur := range validDurations {
		db, mock := setupMockDB(t)
		service, _ := NewConsentService(db, "http://portal")
		ctx := context.Background()

		grantDuration := string(dur)
		req := models.CreateConsentRequest{
			AppID: "app-1",
			ConsentRequirement: models.ConsentRequirement{
				OwnerID:    "user-1",
				OwnerEmail: "user@example.com",
				Fields:     []models.ConsentField{{FieldName: "email", SchemaID: "schema-1", Owner: "citizen"}},
			},
			GrantDuration: &grantDuration,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
			WithArgs("user-1", "app-1", 1).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
			WillReturnRows(sqlmock.NewRows([]string{"consent_id"}).AddRow(uuid.New()))

		_, err := service.CreateConsentRecord(ctx, req)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
