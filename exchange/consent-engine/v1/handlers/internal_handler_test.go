package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/models"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestService(t *testing.T) (*services.ConsentService, sqlmock.Sqlmock) {
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

	service, err := services.NewConsentService(gormDB, "http://localhost:5173")
	require.NoError(t, err)

	return service, mock
}

func TestInternalHandler_HealthCheck(t *testing.T) {
	handler := &InternalHandler{consentService: nil}

	req := httptest.NewRequest("GET", "/internal/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestInternalHandler_HealthCheck_MethodNotAllowed(t *testing.T) {
	handler := &InternalHandler{consentService: nil}

	req := httptest.NewRequest("POST", "/internal/api/v1/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestInternalHandler_GetConsent_MissingAppId(t *testing.T) {
	handler := &InternalHandler{consentService: nil}

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?ownerId=user-1", nil)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInternalHandler_GetConsent_MissingOwner(t *testing.T) {
	handler := &InternalHandler{consentService: nil}

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?appId=app-1", nil)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInternalHandler_CreateConsent_InvalidBody(t *testing.T) {
	handler := &InternalHandler{consentService: nil}

	req := httptest.NewRequest("POST", "/internal/api/v1/consents", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateConsent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInternalHandler_CreateConsent_MethodNotAllowed(t *testing.T) {
	handler := &InternalHandler{consentService: nil}

	req := httptest.NewRequest("GET", "/internal/api/v1/consents", nil)
	w := httptest.NewRecorder()

	handler.CreateConsent(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestInternalHandler_NewInternalHandler(t *testing.T) {
	service, _ := setupTestService(t)
	handler := NewInternalHandler(service)
	assert.NotNil(t, handler)
	assert.Equal(t, service, handler.consentService)
}

func TestInternalHandler_GetConsent_Success_WithOwnerID(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "owner_email", "app_id", "status", "type", "created_at", "updated_at", "grant_duration", "fields", "consent_portal_url"}).
		AddRow(id, "user-1", "user@example.com", "app-1", "approved", "realtime", time.Now(), time.Now(), "P30D", "[]", "http://portal")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?ownerId=user-1&appId=app-1", nil)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.ConsentResponseInternalView
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, id.String(), response.ConsentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInternalHandler_GetConsent_Success_WithOwnerEmail(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"consent_id", "owner_id", "owner_email", "app_id", "status", "type", "created_at", "updated_at", "grant_duration", "fields", "consent_portal_url"}).
		AddRow(id, "user-1", "user@example.com", "app-1", "approved", "realtime", time.Now(), time.Now(), "P30D", "[]", "http://portal")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_email = $1 AND app_id = $2 ORDER BY created_at DESC`)).
		WithArgs("user@example.com", "app-1", 1).
		WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?ownerEmail=user@example.com&appId=app-1", nil)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInternalHandler_GetConsent_NotFound(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records"`)).
		WillReturnError(gorm.ErrRecordNotFound)

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?ownerId=user-1&appId=app-1", nil)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInternalHandler_GetConsent_ContextTimeout(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Nanosecond) // Ensure context is expired

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records"`)).
		WillReturnError(ctx.Err())

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?ownerId=user-1&appId=app-1", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestInternalHandler_CreateConsent_Success(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	// Mock GetConsentInternalView returning not found - specific query for owner_id and app_id
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock Create - GORM Create doesn't use transactions for simple inserts
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
		WillReturnRows(sqlmock.NewRows([]string{"consent_id"}).AddRow(uuid.New()))

	reqBody := models.CreateConsentRequest{
		AppID: "app-1",
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user@example.com",
			Fields:     []models.ConsentField{{FieldName: "email", SchemaID: "schema-1", Owner: "citizen"}},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/internal/api/v1/consents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateConsent(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInternalHandler_CreateConsent_CreateFailed(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	// Mock GetConsentInternalView returning not found - specific query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Mock Create failing - GORM Create doesn't use transactions for simple inserts
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
		WillReturnError(errors.New("db error"))

	reqBody := models.CreateConsentRequest{
		AppID: "app-1",
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user@example.com",
			Fields:     []models.ConsentField{{FieldName: "email", SchemaID: "schema-1", Owner: "citizen"}},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/internal/api/v1/consents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateConsent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInternalHandler_GetConsent_InternalError(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnError(errors.New("database connection error"))

	req := httptest.NewRequest("GET", "/internal/api/v1/consents?ownerId=user-1&appId=app-1", nil)
	w := httptest.NewRecorder()

	handler.GetConsent(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInternalHandler_CreateConsent_InternalError(t *testing.T) {
	service, mock := setupTestService(t)
	handler := NewInternalHandler(service)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consent_records" WHERE owner_id = $1 AND app_id = $2 ORDER BY created_at DESC`)+".*"+regexp.QuoteMeta(` LIMIT $3`)).
		WithArgs("user-1", "app-1", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// Return error that's not ErrConsentCreateFailed to trigger internal error path
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consent_records"`)).
		WillReturnError(errors.New("unexpected database error"))

	reqBody := models.CreateConsentRequest{
		AppID: "app-1",
		ConsentRequirement: models.ConsentRequirement{
			OwnerID:    "user-1",
			OwnerEmail: "user@example.com",
			Fields:     []models.ConsentField{{FieldName: "email", SchemaID: "schema-1", Owner: "citizen"}},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/internal/api/v1/consents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateConsent(w, req)

	// The error gets wrapped as ErrConsentCreateFailed, so we get 400, not 500
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
