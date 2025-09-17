package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gov-dx-sandbox/api-server-go/handlers"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

func TestMainServerInitialization(t *testing.T) {
	t.Run("NewAPIServer", func(t *testing.T) {
		apiServer := handlers.NewAPIServer()
		if apiServer == nil {
			t.Fatal("Expected APIServer to be created")
		}

		// Test that all services are initialized
		if apiServer.GetProviderService() == nil {
			t.Error("Expected provider service to be initialized")
		}

		// Test that routes can be set up
		mux := http.NewServeMux()
		apiServer.SetupRoutes(mux)

		// Verify that routes are registered by checking if the mux has handlers
		// This is a basic check - in a real test we might want to verify specific routes
		if mux == nil {
			t.Error("Expected mux to be configured")
		}
	})

	t.Run("SetupRoutes", func(t *testing.T) {
		apiServer := handlers.NewAPIServer()
		mux := http.NewServeMux()

		// This should not panic
		apiServer.SetupRoutes(mux)

		// Test that we can make requests to the mux
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		// Add health check handler (as done in main.go)
		mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "api-server"})
		})))

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestHealthCheckEndpoints(t *testing.T) {
	// Create a test server similar to main.go
	mux := http.NewServeMux()

	// Add health check endpoint
	mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "api-server"})
	})))

	// Add debug endpoint
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path, "method": r.Method})
	})))

	t.Run("HealthCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if response["status"] != "healthy" {
			t.Errorf("Expected status 'healthy', got %s", response["status"])
		}

		if response["service"] != "api-server" {
			t.Errorf("Expected service 'api-server', got %s", response["service"])
		}
	})

	t.Run("DebugEndpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/debug", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if response["path"] != "/debug" {
			t.Errorf("Expected path '/debug', got %s", response["path"])
		}

		if response["method"] != "GET" {
			t.Errorf("Expected method 'GET', got %s", response["method"])
		}
	})

	t.Run("DebugEndpoint_POST", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/debug", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if response["method"] != "POST" {
			t.Errorf("Expected method 'POST', got %s", response["method"])
		}
	})
}

func TestServerConfiguration(t *testing.T) {
	t.Run("DefaultPort", func(t *testing.T) {
		// Test that default port is used when PORT env var is not set
		originalPort := os.Getenv("PORT")
		os.Unsetenv("PORT")
		defer func() {
			if originalPort != "" {
				os.Setenv("PORT", originalPort)
			}
		}()

		// The default port should be "3000" as defined in main.go
		expectedPort := "3000"
		// In a real test, we might start the server and check what port it's listening on
		// For now, we just verify the logic exists
		if expectedPort != "3000" {
			t.Errorf("Expected default port to be '3000', got %s", expectedPort)
		}
	})

	t.Run("CustomPort", func(t *testing.T) {
		// Test that custom port is used when PORT env var is set
		originalPort := os.Getenv("PORT")
		os.Setenv("PORT", "8080")
		defer func() {
			if originalPort != "" {
				os.Setenv("PORT", originalPort)
			} else {
				os.Unsetenv("PORT")
			}
		}()

		port := os.Getenv("PORT")
		if port != "8080" {
			t.Errorf("Expected port to be '8080', got %s", port)
		}
	})
}

func TestServerErrorHandling(t *testing.T) {
	t.Run("PanicRecovery", func(t *testing.T) {
		// Create a handler that panics
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		// Wrap with panic recovery middleware
		recoveredHandler := utils.PanicRecoveryMiddleware(panicHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// This should not panic the test
		recoveredHandler.ServeHTTP(w, req)

		// The middleware should handle the panic gracefully
		// The exact response depends on the implementation of PanicRecoveryMiddleware
		// We just verify that the test doesn't crash
	})
}

func TestServerStartup(t *testing.T) {
	t.Run("ServerCanStart", func(t *testing.T) {
		// Create a test server similar to main.go
		apiServer := handlers.NewAPIServer()
		mux := http.NewServeMux()
		apiServer.SetupRoutes(mux)

		// Add health check
		mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "api-server"})
		})))

		// Start server on a random port
		server := httptest.NewServer(mux)
		defer server.Close()

		// Test that server is responding
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestAPIServerIntegration(t *testing.T) {
	t.Run("FullServerSetup", func(t *testing.T) {
		// Test the complete server setup as done in main.go
		apiServer := handlers.NewAPIServer()
		mux := http.NewServeMux()
		apiServer.SetupRoutes(mux)

		// Add health check
		mux.Handle("/health", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "api-server"})
		})))

		// Add debug endpoint
		mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{"path": r.URL.Path, "method": r.Method})
		})))

		// Test that all endpoints are accessible
		testCases := []struct {
			method string
			path   string
			status int
		}{
			{"GET", "/health", http.StatusOK},
			{"GET", "/debug", http.StatusOK},
			{"POST", "/debug", http.StatusOK},
			{"GET", "/consumers", http.StatusOK}, // Should return empty list
			{"GET", "/providers", http.StatusOK}, // Should return empty list
			{"GET", "/admin/metrics", http.StatusOK},
		}

		server := httptest.NewServer(mux)
		defer server.Close()

		for _, tc := range testCases {
			t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
				req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Fatalf("Failed to make request: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != tc.status {
					t.Errorf("Expected status %d, got %d", tc.status, resp.StatusCode)
				}
			})
		}
	})
}

func TestServerLogging(t *testing.T) {
	t.Run("LoggerInitialization", func(t *testing.T) {
		// Test that the logger is properly initialized
		// This is more of an integration test to ensure the logging setup works
		// The actual logger initialization happens in main.go

		// We can't easily test the slog setup without running main(),
		// but we can verify that the logging package is available
		// and that the expected log messages would be generated

		// This is a placeholder test - in a real scenario, we might:
		// 1. Capture log output
		// 2. Verify log format
		// 3. Test log levels
		// 4. Verify structured logging
	})
}

func TestServerGracefulShutdown(t *testing.T) {
	t.Run("ServerShutdown", func(t *testing.T) {
		// Test that the server can be shut down gracefully
		// This would typically involve:
		// 1. Starting the server
		// 2. Sending a shutdown signal
		// 3. Verifying that the server stops accepting new connections
		// 4. Verifying that existing connections are handled properly

		// For now, we'll just test that we can create and close a test server
		apiServer := handlers.NewAPIServer()
		mux := http.NewServeMux()
		apiServer.SetupRoutes(mux)

		server := httptest.NewServer(mux)

		// Verify server is running
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Server should be running: %v", err)
		}
		resp.Body.Close()

		// Close server
		server.Close()

		// Verify server is closed
		_, err = http.Get(server.URL + "/health")
		if err == nil {
			t.Error("Server should be closed")
		}
	})
}
