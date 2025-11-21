package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCORSConfig(t *testing.T) {
	t.Run("DefaultOrigins", func(t *testing.T) {
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		config := DefaultCORSConfig()

		assert.Contains(t, config.AllowedOrigins, "http://localhost:5173")
		assert.Contains(t, config.AllowedMethods, "GET")
		assert.Contains(t, config.AllowedMethods, "POST")
		assert.True(t, config.AllowCredentials)
		assert.Equal(t, 86400, config.MaxAge)
	})

	t.Run("CustomOriginsFromEnv", func(t *testing.T) {
		os.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com,https://test.com")
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		config := DefaultCORSConfig()

		assert.Contains(t, config.AllowedOrigins, "http://example.com")
		assert.Contains(t, config.AllowedOrigins, "https://test.com")
	})

	t.Run("WildcardOrigin", func(t *testing.T) {
		os.Setenv("CORS_ALLOWED_ORIGINS", "*")
		defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

		config := DefaultCORSConfig()

		assert.Contains(t, config.AllowedOrigins, "*")
	})
}

func TestCORSMiddleware(t *testing.T) {
	t.Run("AllowedOrigin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:   []string{"http://localhost:5173"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           3600,
		}

		middleware := CORSMiddleware(config)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("WildcardOrigin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET"},
			AllowCredentials: false,
		}

		middleware := CORSMiddleware(config)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://any-origin.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("WildcardWithCredentials", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET"},
			AllowCredentials: true, // Cannot use wildcard with credentials
		}

		middleware := CORSMiddleware(config)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should reflect the actual origin when credentials are required
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("PreflightRequest", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"http://localhost:5173"},
			AllowedMethods: []string{"GET", "POST"},
		}

		middleware := CORSMiddleware(config)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "POST")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Check that Vary header contains the expected values
		// HTTP headers can have multiple values, so we need to check all values
		varyHeaders := w.Header().Values("Vary")
		var varyHeaderStr string
		for _, v := range varyHeaders {
			varyHeaderStr += v + ", "
		}
		assert.Contains(t, varyHeaderStr, "Access-Control-Request-Method")
		assert.Contains(t, varyHeaderStr, "Access-Control-Request-Headers")
	})

	t.Run("DisallowedOrigin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"http://localhost:5173"},
		}

		middleware := CORSMiddleware(config)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("NoOriginHeader", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"http://localhost:5173"},
		}

		middleware := CORSMiddleware(config)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/test", nil)
		// No Origin header
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestNewCORSMiddleware(t *testing.T) {
	middleware := NewCORSMiddleware()
	assert.NotNil(t, middleware)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}
