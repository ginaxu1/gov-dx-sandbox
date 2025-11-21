package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gov-dx-sandbox/exchange/consent-engine/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockConsentEngine is a mock implementation of ConsentEngine for testing
type mockConsentEngine struct {
	mock.Mock
}

func (m *mockConsentEngine) CreateConsent(req models.ConsentRequest) (*models.ConsentRecord, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) FindExistingConsent(consumerAppID, ownerID string) *models.ConsentRecord {
	args := m.Called(consumerAppID, ownerID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.ConsentRecord)
}

func (m *mockConsentEngine) GetConsentStatus(id string) (*models.ConsentRecord, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) UpdateConsent(id string, req models.UpdateConsentRequest) (*models.ConsentRecord, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) ProcessConsentPortalRequest(req models.ConsentPortalRequest) (*models.ConsentRecord, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) GetConsentsByDataOwner(dataOwner string) ([]*models.ConsentRecord, error) {
	args := m.Called(dataOwner)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) GetConsentsByConsumer(consumer string) ([]*models.ConsentRecord, error) {
	args := m.Called(consumer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) CheckConsentExpiry() ([]*models.ConsentRecord, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) StartBackgroundExpiryProcess(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func (m *mockConsentEngine) ProcessConsentRequest(req models.ConsentRequest) (*models.ConsentRecord, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) RevokeConsent(id string, reason string) (*models.ConsentRecord, error) {
	args := m.Called(id, reason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConsentRecord), args.Error(1)
}

func (m *mockConsentEngine) StopBackgroundExpiryProcess() {
	m.Called()
}

func TestNewConsentHandler(t *testing.T) {
	mockEngine := new(mockConsentEngine)
	handler := NewConsentHandler(mockEngine)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockEngine, handler.engine)
}

func TestHandlePortalAction(t *testing.T) {
	mockEngine := new(mockConsentEngine)
	handler := NewConsentHandler(mockEngine)

	tests := []struct {
		name           string
		requestBody    models.ConsentPortalRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "ApproveAction_Success",
			requestBody: models.ConsentPortalRequest{
				ConsentID: "consent_123",
				Action:    "approve",
				DataOwner: "user@example.com",
				Reason:    "User approved",
			},
			mockSetup: func() {
				mockEngine.On("ProcessConsentPortalRequest", mock.AnythingOfType("models.ConsentPortalRequest")).
					Return(&models.ConsentRecord{
						ConsentID: "consent_123",
						Status:    models.StatusApproved,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "DenyAction_Success",
			requestBody: models.ConsentPortalRequest{
				ConsentID: "consent_456",
				Action:    "deny",
				DataOwner: "user@example.com",
				Reason:    "User denied",
			},
			mockSetup: func() {
				mockEngine.On("ProcessConsentPortalRequest", mock.AnythingOfType("models.ConsentPortalRequest")).
					Return(&models.ConsentRecord{
						ConsentID: "consent_456",
						Status:    models.StatusRejected,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "InvalidJSON",
			requestBody: models.ConsentPortalRequest{},
			mockSetup: func() {
				// No mock setup needed for invalid JSON
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "EngineError",
			requestBody: models.ConsentPortalRequest{
				ConsentID: "consent_789",
				Action:    "approve",
				DataOwner: "user@example.com",
			},
			mockSetup: func() {
				mockEngine.On("ProcessConsentPortalRequest", mock.AnythingOfType("models.ConsentPortalRequest")).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine.ExpectedCalls = nil
			mockEngine.Calls = nil
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			var body []byte
			var err error
			if tt.name == "InvalidJSON" {
				body = []byte("invalid json")
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/consents/portal/actions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandlePortalAction(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEngine.AssertExpectations(t)
		})
	}
}

func TestConsentHandler_ErrorCases(t *testing.T) {
	mockEngine := new(mockConsentEngine)
	handler := NewConsentHandler(mockEngine)

	t.Run("CreateConsent_ReadBodyError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/consents", bytes.NewBuffer([]byte{}))
		req.Body = &errorReader{}
		w := httptest.NewRecorder()

		handler.createConsent(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateConsent_EngineError", func(t *testing.T) {
		mockEngine.On("ProcessConsentRequest", mock.AnythingOfType("models.ConsentRequest")).
			Return(nil, assert.AnError)

		body, _ := json.Marshal(models.ConsentRequest{
			AppID: "test-app",
			ConsentRequirements: []models.ConsentRequirement{
				{
					Owner:   "CITIZEN",
					OwnerID: "user@example.com",
					Fields: []models.ConsentField{
						{FieldName: "personInfo.name", SchemaID: "schema-123"},
					},
				},
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/consents", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.createConsent(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("GetConsentByID_NotFound", func(t *testing.T) {
		mockEngine.On("GetConsentStatus", "non-existent").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/consents/non-existent", nil)
		w := httptest.NewRecorder()

		handler.getConsentByID(w, req, "non-existent")
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("GetDataInfo_NotFound", func(t *testing.T) {
		mockEngine.On("GetConsentStatus", "non-existent").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/data-info/non-existent", nil)
		w := httptest.NewRecorder()

		handler.getDataInfo(w, req, "non-existent")
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("UpdateConsentByID_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/consents/consent_123", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.updateConsentByID(w, req, "consent_123")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PatchConsentByID_NotFound", func(t *testing.T) {
		mockEngine.ExpectedCalls = nil
		mockEngine.Calls = nil
		mockEngine.On("GetConsentStatus", "non-existent-patch").
			Return(nil, &notFoundError{message: "consent record not found"})

		body, _ := json.Marshal(map[string]string{"status": "approved"})
		req := httptest.NewRequest(http.MethodPatch, "/consents/non-existent-patch", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.patchConsentByID(w, req, "non-existent-patch")
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("RevokeConsentByID_InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/consents/consent_123", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.revokeConsentByID(w, req, "consent_123")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetConsentsByDataOwner_Error", func(t *testing.T) {
		mockEngine.On("GetConsentsByDataOwner", "owner@example.com").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/data-owner/owner@example.com", nil)
		w := httptest.NewRecorder()

		handler.getConsentsByDataOwner(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("GetConsentsByConsumer_Error", func(t *testing.T) {
		mockEngine.On("GetConsentsByConsumer", "consumer-app").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/consumer/consumer-app", nil)
		w := httptest.NewRecorder()

		handler.getConsentsByConsumer(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("CheckConsentExpiry_Error", func(t *testing.T) {
		mockEngine.On("CheckConsentExpiry").
			Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodPost, "/admin/expiry-check", nil)
		w := httptest.NewRecorder()

		handler.checkConsentExpiry(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("DataInfoHandler_EmptyPath", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/data-info/", nil)
		w := httptest.NewRecorder()

		handler.DataInfoHandler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DataInfoHandler_InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/data-info/consent_123", nil)
		w := httptest.NewRecorder()

		handler.DataInfoHandler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("AdminHandler_InvalidPath", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/admin/invalid", nil)
		w := httptest.NewRecorder()

		handler.AdminHandler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("AdminHandler_InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/expiry-check", nil)
		w := httptest.NewRecorder()

		handler.AdminHandler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("DataOwnerHandler_InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/data-owner/owner@example.com", nil)
		w := httptest.NewRecorder()

		handler.DataOwnerHandler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("ConsumerHandler_InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/consumer/app", nil)
		w := httptest.NewRecorder()

		handler.ConsumerHandler(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("HandlePortalAction_NotFound", func(t *testing.T) {
		mockEngine.On("ProcessConsentPortalRequest", mock.AnythingOfType("models.ConsentPortalRequest")).
			Return(nil, assert.AnError)

		body, _ := json.Marshal(models.ConsentPortalRequest{
			ConsentID: "non-existent",
			Action:    "approve",
			DataOwner: "user@example.com",
		})

		req := httptest.NewRequest(http.MethodPost, "/consents/portal/actions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePortalAction(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockEngine.AssertExpectations(t)
	})

	t.Run("HandlePortalAction_InvalidAction", func(t *testing.T) {
		mockEngine.ExpectedCalls = nil
		mockEngine.Calls = nil
		mockEngine.On("ProcessConsentPortalRequest", mock.AnythingOfType("models.ConsentPortalRequest")).
			Return(nil, &invalidActionError{message: "invalid action: invalid"})

		body, _ := json.Marshal(models.ConsentPortalRequest{
			ConsentID: "consent_invalid_action",
			Action:    "invalid",
			DataOwner: "user@example.com",
		})

		req := httptest.NewRequest(http.MethodPost, "/consents/portal/actions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.HandlePortalAction(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockEngine.AssertExpectations(t)
	})
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}

func (e *errorReader) Close() error {
	return nil
}

// notFoundError is an error that contains "not found"
type notFoundError struct {
	message string
}

func (e *notFoundError) Error() string {
	return e.message
}

// invalidActionError is an error that contains "invalid action"
type invalidActionError struct {
	message string
}

func (e *invalidActionError) Error() string {
	return e.message
}

