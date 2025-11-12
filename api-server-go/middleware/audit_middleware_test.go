package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuditMiddleware(t *testing.T) {
	middleware := NewAuditMiddleware("http://localhost:8080")
	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.auditService)
}

func TestAuditMiddleware_shouldSkipAudit(t *testing.T) {
	middleware := NewAuditMiddleware("http://localhost:8080")

	t.Run("SkipInTestEnvironment", func(t *testing.T) {
		originalEnv := os.Getenv("GO_ENV")
		originalTesting := os.Getenv("TESTING")
		defer func() {
			if originalEnv != "" {
				os.Setenv("GO_ENV", originalEnv)
			} else {
				os.Unsetenv("GO_ENV")
			}
			if originalTesting != "" {
				os.Setenv("TESTING", originalTesting)
			} else {
				os.Unsetenv("TESTING")
			}
		}()

		os.Setenv("GO_ENV", "test")
		assert.True(t, middleware.shouldSkipAudit("/api/v1/schemas"))

		os.Unsetenv("GO_ENV")
		os.Setenv("TESTING", "true")
		assert.True(t, middleware.shouldSkipAudit("/api/v1/schemas"))
	})

	t.Run("SkipHealthCheck", func(t *testing.T) {
		os.Unsetenv("GO_ENV")
		os.Unsetenv("TESTING")
		assert.True(t, middleware.shouldSkipAudit("/health"))
		assert.True(t, middleware.shouldSkipAudit("/health/status"))
	})

	t.Run("SkipDebug", func(t *testing.T) {
		assert.True(t, middleware.shouldSkipAudit("/debug"))
		assert.True(t, middleware.shouldSkipAudit("/debug/pprof"))
	})

	t.Run("SkipOpenAPI", func(t *testing.T) {
		assert.True(t, middleware.shouldSkipAudit("/openapi.yaml"))
	})

	t.Run("SkipFavicon", func(t *testing.T) {
		assert.True(t, middleware.shouldSkipAudit("/favicon.ico"))
	})

	t.Run("DoNotSkipRegularPaths", func(t *testing.T) {
		assert.False(t, middleware.shouldSkipAudit("/api/v1/schemas"))
		assert.False(t, middleware.shouldSkipAudit("/api/v1/members"))
	})
}

func TestAuditMiddleware_getClientIP(t *testing.T) {
	middleware := NewAuditMiddleware("http://localhost:8080")

	t.Run("FromXForwardedFor", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		ip := middleware.getClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("FromXForwardedFor_MultipleIPs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
		ip := middleware.getClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("FromXRealIP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")
		ip := middleware.getClientIP(req)
		assert.Equal(t, "10.0.0.1", ip)
	})

	t.Run("FromRemoteAddr", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		ip := middleware.getClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("UnknownWhenNoHeaders", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "" // Clear RemoteAddr to test unknown case
		ip := middleware.getClientIP(req)
		assert.Equal(t, "unknown", ip)
	})
}

func TestAuditMiddleware_determineDefaultEntityID(t *testing.T) {
	middleware := NewAuditMiddleware("http://localhost:8080")

	t.Run("ConsumerPath", func(t *testing.T) {
		assert.Equal(t, "system_consumer", middleware.determineDefaultEntityID("/consumers/123"))
		assert.Equal(t, "system_consumer", middleware.determineDefaultEntityID("/consumer-applications/456"))
	})

	t.Run("ProviderPath", func(t *testing.T) {
		assert.Equal(t, "system_provider", middleware.determineDefaultEntityID("/providers/123"))
		assert.Equal(t, "system_provider", middleware.determineDefaultEntityID("/provider-submissions/456"))
	})

	t.Run("DefaultAdmin", func(t *testing.T) {
		assert.Equal(t, "system_admin", middleware.determineDefaultEntityID("/api/v1/schemas"))
		assert.Equal(t, "system_admin", middleware.determineDefaultEntityID("/other/path"))
	})
}

func TestAuditMiddleware_ensureValidJSON(t *testing.T) {
	middleware := NewAuditMiddleware("http://localhost:8080")

	t.Run("ValidJSON", func(t *testing.T) {
		validJSON := []byte(`{"key": "value"}`)
		result := middleware.ensureValidJSON(validJSON)
		assert.Equal(t, validJSON, result)
	})

	t.Run("EmptyData", func(t *testing.T) {
		result := middleware.ensureValidJSON([]byte{})
		assert.Equal(t, []byte("{}"), result)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := []byte("not json")
		result := middleware.ensureValidJSON(invalidJSON)
		// Should wrap in JSON object
		var wrapped map[string]interface{}
		err := json.Unmarshal(result, &wrapped)
		assert.NoError(t, err)
		assert.Contains(t, wrapped, "raw_data")
	})

	t.Run("ValidJSONArray", func(t *testing.T) {
		validJSON := []byte(`[{"key": "value"}]`)
		result := middleware.ensureValidJSON(validJSON)
		assert.Equal(t, validJSON, result)
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("Write", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		data := []byte("test data")
		n, err := rw.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, rw.body.Bytes())
		assert.Equal(t, data, recorder.Body.Bytes())
	})

	t.Run("WriteHeader", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		rw := &responseWriter{
			ResponseWriter: recorder,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		rw.WriteHeader(http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, rw.statusCode)
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
}

func TestAuditMiddleware_AuditLoggingMiddleware(t *testing.T) {
	// Set test environment to skip actual audit logging
	originalEnv := os.Getenv("GO_ENV")
	defer func() {
		if originalEnv != "" {
			os.Setenv("GO_ENV", originalEnv)
		} else {
			os.Unsetenv("GO_ENV")
		}
	}()

	os.Setenv("GO_ENV", "test")
	middleware := NewAuditMiddleware("http://localhost:8080")

	t.Run("SkipAuditForHealthCheck", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ProcessRegularRequest", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("POST", "/api/v1/schemas", bytes.NewBufferString(`{"name": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.NewDecoder(w.Body).Decode(&response)
		assert.Equal(t, "ok", response["status"])
	})

	t.Run("CaptureResponseBody", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "123"}`))
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("POST", "/api/v1/members", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "123")
	})

	t.Run("WithGraphQLQuery", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"data": "result"})
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		body := map[string]string{"query": "query { test }"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/graphql", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("WithConsumerIDInPath", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("GET", "/consumers/consumer-123", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("WithProviderIDInPath", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("GET", "/providers/provider-456", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("WithNonJSONResponse", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("plain text response"))
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "plain text response", w.Body.String())
	})

	t.Run("WithEmptyRequestBody", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("WithErrorStatusCode", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		})

		wrapped := middleware.AuditLoggingMiddleware(handler)
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
