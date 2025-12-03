package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger.Init()
}

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := Response{Message: "OpenDIF Server is Healthy!"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "OpenDIF Server is Healthy!", response.Message)
}

func TestHealthEndpoint_WrongMethod(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := Response{Message: "OpenDIF Server is Healthy!"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCorsMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCorsMiddleware_OptionsRequest(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// OPTIONS request should return 200 immediately
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestGetEnv(t *testing.T) {
	t.Run("Returns default when env var not set", func(t *testing.T) {
		value := getEnv("NON_EXISTENT_VAR_12345", "default-value")
		assert.Equal(t, "default-value", value)
	})

	t.Run("Returns env var value when set", func(t *testing.T) {
		key := "TEST_ENV_VAR_12345"
		os.Setenv(key, "test-value")
		defer os.Unsetenv(key)

		value := getEnv(key, "default-value")
		assert.Equal(t, "test-value", value)
	})

	t.Run("Returns default when env var is empty", func(t *testing.T) {
		key := "TEST_EMPTY_VAR_12345"
		os.Setenv(key, "")
		defer os.Unsetenv(key)

		value := getEnv(key, "default-value")
		assert.Equal(t, "default-value", value) // Empty string is treated as not set
	})
}

func TestGetDatabaseConnectionString(t *testing.T) {
	t.Run("Uses Choreo environment variables when set", func(t *testing.T) {
		// Set Choreo variables
		os.Setenv("CHOREO_DB_OE_HOSTNAME", "choreo-host")
		os.Setenv("CHOREO_DB_OE_USERNAME", "choreo-user")
		os.Setenv("CHOREO_DB_OE_PASSWORD", "choreo-pass")
		os.Setenv("CHOREO_DB_OE_DATABASENAME", "choreo-db")
		os.Setenv("CHOREO_DB_OE_PORT", "5433")
		defer func() {
			os.Unsetenv("CHOREO_DB_OE_HOSTNAME")
			os.Unsetenv("CHOREO_DB_OE_USERNAME")
			os.Unsetenv("CHOREO_DB_OE_PASSWORD")
			os.Unsetenv("CHOREO_DB_OE_DATABASENAME")
			os.Unsetenv("CHOREO_DB_OE_PORT")
		}()

		connStr := getDatabaseConnectionString()
		assert.Contains(t, connStr, "host=choreo-host")
		assert.Contains(t, connStr, "user=choreo-user")
		assert.Contains(t, connStr, "password=choreo-pass")
		assert.Contains(t, connStr, "dbname=choreo-db")
		assert.Contains(t, connStr, "port=5433")
		assert.Contains(t, connStr, "sslmode=require")
	})

	t.Run("Falls back to standard env vars when Choreo vars not set", func(t *testing.T) {
		// Unset Choreo vars
		os.Unsetenv("CHOREO_DB_OE_HOSTNAME")
		os.Unsetenv("CHOREO_DB_OE_USERNAME")
		os.Unsetenv("CHOREO_DB_OE_PASSWORD")
		os.Unsetenv("CHOREO_DB_OE_DATABASENAME")

		// Set standard vars
		os.Setenv("DB_HOST", "standard-host")
		os.Setenv("DB_PORT", "5434")
		os.Setenv("DB_USER", "standard-user")
		os.Setenv("DB_PASSWORD", "standard-pass")
		os.Setenv("DB_NAME", "standard-db")
		os.Setenv("DB_SSLMODE", "disable")
		defer func() {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("DB_SSLMODE")
		}()

		connStr := getDatabaseConnectionString()
		assert.Contains(t, connStr, "host=standard-host")
		assert.Contains(t, connStr, "user=standard-user")
		assert.Contains(t, connStr, "password=standard-pass")
		assert.Contains(t, connStr, "dbname=standard-db")
		assert.Contains(t, connStr, "port=5434")
		assert.Contains(t, connStr, "sslmode=disable")
	})

	t.Run("Uses defaults when no env vars set", func(t *testing.T) {
		// Unset all vars
		os.Unsetenv("CHOREO_DB_OE_HOSTNAME")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_SSLMODE")

		connStr := getDatabaseConnectionString()
		assert.NotEmpty(t, connStr)
		assert.Contains(t, connStr, "host=")
		assert.Contains(t, connStr, "port=")
	})
}
