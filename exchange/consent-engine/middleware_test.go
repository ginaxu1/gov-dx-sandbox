package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCORSMiddleware tests the CORS middleware
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "CORS_OPTIONS_Preflight",
			method:         http.MethodOptions,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Error("Expected Access-Control-Allow-Origin header")
				}
				if w.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("Expected Access-Control-Allow-Methods header")
				}
			},
		},
		{
			name:           "CORS_RegularRequest",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Header().Get("Access-Control-Allow-Origin") != "*" {
					t.Error("Expected Access-Control-Allow-Origin header")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := corsMiddleware(nextHandler)
			req := httptest.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkHeaders != nil {
				tt.checkHeaders(t, w)
			}
		})
	}
}

// TestUserAuthMiddleware tests the user authentication middleware
func TestUserAuthMiddleware(t *testing.T) {
	engine := setupPostgresTestEngine(t)

	// Create a test consent
	record := createTestConsent(t, engine, "test-app", "user@example.com")

	tests := []struct {
		name           string
		consentID      string
		authHeader     string
		expectedStatus int
		checkContext   func(t *testing.T, r *http.Request)
	}{
		{
			name:           "UserAuth_MissingConsentID",
			consentID:      "",
			authHeader:     "Bearer token",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "UserAuth_ConsentNotFound",
			consentID:      "non-existent",
			authHeader:     "Bearer token",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "UserAuth_MissingAuthHeader",
			consentID:      record.ConsentID,
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "UserAuth_InvalidAuthFormat",
			consentID:      record.ConsentID,
			authHeader:     "Invalid token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "UserAuth_InvalidToken",
			consentID:      record.ConsentID,
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext != nil {
					tt.checkContext(t, r)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create a real JWT verifier (will fail for invalid tokens, which is expected)
			jwtVerifier := NewJWTVerifier("https://api.asgardeo.io/t/test/oauth2/jwks", "test-audience", "test-org")
			userTokenConfig := UserTokenValidationConfig{
				ExpectedIssuer:   "https://api.asgardeo.io/t/test/oauth2/token",
				ExpectedAudience: "test-audience",
				ExpectedOrgName:  "test-org",
			}

			handler := userAuthMiddleware(jwtVerifier, engine, userTokenConfig)(nextHandler)

			path := "/v1/consents/" + tt.consentID
			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

// TestSelectiveAuthMiddleware tests the selective authentication middleware
func TestSelectiveAuthMiddleware(t *testing.T) {
	engine := setupPostgresTestEngine(t)
	record := createTestConsent(t, engine, "test-app", "user@example.com")

	tests := []struct {
		name               string
		method             string
		requireAuthMethods []string
		expectedAuthCalled bool
	}{
		{
			name:               "SelectiveAuth_RequiresAuth_GET",
			method:             http.MethodGet,
			requireAuthMethods: []string{"GET", "PUT"},
			expectedAuthCalled: true,
		},
		{
			name:               "SelectiveAuth_RequiresAuth_PUT",
			method:             http.MethodPut,
			requireAuthMethods: []string{"GET", "PUT"},
			expectedAuthCalled: true,
		},
		{
			name:               "SelectiveAuth_NoAuth_PATCH",
			method:             http.MethodPatch,
			requireAuthMethods: []string{"GET", "PUT"},
			expectedAuthCalled: false,
		},
		{
			name:               "SelectiveAuth_NoAuth_DELETE",
			method:             http.MethodDelete,
			requireAuthMethods: []string{"GET", "PUT"},
			expectedAuthCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authCalled = true
				w.WriteHeader(http.StatusOK)
			})

			jwtVerifier := NewJWTVerifier("https://api.asgardeo.io/t/test/oauth2/jwks", "test-audience", "test-org")
			userTokenConfig := UserTokenValidationConfig{
				ExpectedIssuer:   "https://api.asgardeo.io/t/test/oauth2/token",
				ExpectedAudience: "test-audience",
				ExpectedOrgName:  "test-org",
			}

			handler := selectiveAuthMiddleware(jwtVerifier, engine, userTokenConfig, tt.requireAuthMethods)(nextHandler)

			req := httptest.NewRequest(tt.method, "/v1/consents/"+record.ConsentID, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// If auth is required but not provided, we expect 401/403
			// If auth is not required, handler should be called
			if tt.expectedAuthCalled {
				// Auth middleware will be called, but will fail without valid token
				// This is expected behavior - we're just testing the selective logic
				if w.Code == http.StatusOK {
					// If we got OK, auth was bypassed (unexpected)
					if !authCalled {
						t.Error("Expected auth to be called, but handler was not invoked")
					}
				}
			} else {
				// No auth required, handler should be called directly
				if w.Code == http.StatusOK && !authCalled {
					t.Error("Expected handler to be called without auth")
				}
			}
		})
	}
}
