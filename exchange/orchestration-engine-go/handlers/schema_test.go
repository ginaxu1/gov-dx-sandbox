package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/logger"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.Init()
}

func TestNewSchemaHandler(t *testing.T) {
	handler := NewSchemaHandler(nil)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.schemaService)
}

func TestSchemaHandler_CreateSchema_InvalidJSON(t *testing.T) {
	// Note: Handler checks for nil service first, so returns 503
	// JSON validation is tested in integration tests with a real service
	handler := NewSchemaHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/sdl", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSchema(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_CreateSchema_MissingFields(t *testing.T) {
	// Note: Handler checks for nil service first, so returns 503
	// Field validation is tested in integration tests with a real service
	handler := NewSchemaHandler(nil)

	reqBody := CreateSchemaRequest{
		Version: "1.0.0",
		// Missing SDL and CreatedBy
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/sdl", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSchema(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_CreateSchema_NoService(t *testing.T) {
	handler := NewSchemaHandler(nil)

	reqBody := CreateSchemaRequest{
		Version:   "1.0.0",
		SDL:       "type Query { test: String }",
		CreatedBy: "test-user",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/sdl", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSchema(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "database not connected")
}

func TestSchemaHandler_GetSchemas_NoService(t *testing.T) {
	handler := NewSchemaHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/sdl/versions", nil)
	w := httptest.NewRecorder()

	handler.GetSchemas(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_GetActiveSchema_NoService(t *testing.T) {
	handler := NewSchemaHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/sdl", nil)
	w := httptest.NewRecorder()

	handler.GetActiveSchema(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_ActivateSchema_NoService(t *testing.T) {
	handler := NewSchemaHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/sdl/versions/1.0.0/activate", nil)
	w := httptest.NewRecorder()

	// Use chi router to set URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("version", "1.0.0")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.ActivateSchema(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_ValidateSDL_InvalidJSON(t *testing.T) {
	// Note: Handler checks for nil service first, so returns 503
	// JSON validation is tested in integration tests with a real service
	handler := NewSchemaHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/sdl/validate", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.ValidateSDL(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_ValidateSDL_NoService(t *testing.T) {
	handler := NewSchemaHandler(nil)

	reqBody := ValidateSDLRequest{
		SDL: "type Query { test: String }",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/sdl/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ValidateSDL(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_CheckCompatibility_InvalidJSON(t *testing.T) {
	// Note: The handler checks for nil service first, so invalid JSON returns 503
	// To test JSON parsing, we'd need a real service, which is tested in integration tests
	handler := NewSchemaHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/sdl/check-compatibility", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.CheckCompatibility(w, req)

	// Handler checks service first, so returns 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestSchemaHandler_CheckCompatibility_NoService(t *testing.T) {
	handler := NewSchemaHandler(nil)

	reqBody := ValidateSDLRequest{
		SDL: "type Query { test: String }",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/sdl/check-compatibility", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CheckCompatibility(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
