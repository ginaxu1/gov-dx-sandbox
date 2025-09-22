package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// JWTVerifierInterface defines the interface for JWT verification
type JWTVerifierInterface interface {
	VerifyToken(tokenString string) (*jwt.Token, error)
}

// hybridAuthMiddlewareForTesting creates a testing version of the hybrid middleware
func hybridAuthMiddlewareForTesting(jwtVerifier JWTVerifierInterface, engine ConsentEngine, userTokenConfig UserTokenValidationConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract consent ID from the URL path
			consentID := strings.TrimPrefix(r.URL.Path, "/consents/")
			if consentID == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
				return
			}

			// Get the consent record to check permissions
			consentRecord, err := engine.GetConsentStatus(consentID)
			if err != nil {
				utils.RespondWithJSON(w, http.StatusNotFound, utils.ErrorResponse{Error: "Consent record not found"})
				return
			}

			// Check if this is a frontend request (has browser-like headers)
			isFrontendRequest := r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
				strings.Contains(r.Header.Get("User-Agent"), "Mozilla") ||
				strings.Contains(r.Header.Get("User-Agent"), "Chrome") ||
				strings.Contains(r.Header.Get("User-Agent"), "Safari") ||
				strings.Contains(r.Header.Get("User-Agent"), "Firefox")

			// Extract the Authorization header
			authHeader := r.Header.Get("Authorization")

			// If it's a frontend request, JWT authentication is required
			if isFrontendRequest {
				if authHeader == "" {
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Authorization header is required"})
					return
				}

				// Check if it's a Bearer token
				const bearerPrefix = "Bearer "
				if !strings.HasPrefix(authHeader, bearerPrefix) {
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid authorization format. Expected 'Bearer <token>'"})
					return
				}

				// Extract the token
				tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
				if tokenString == "" {
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Token is required"})
					return
				}

				// Verify the token
				token, err := jwtVerifier.VerifyToken(tokenString)
				if err != nil {
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid or expired token"})
					return
				}

				// Extract token information and determine type
				tokenInfo, err := extractTokenInfo(token)
				if err != nil {
					utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid token format"})
					return
				}

				// For user tokens, check if the email matches the consent owner
				if tokenInfo.Type == TokenTypeUser {
					if consentRecord.OwnerEmail != tokenInfo.Email {
						utils.RespondWithJSON(w, http.StatusForbidden, utils.ErrorResponse{Error: "Access denied: email does not match consent owner"})
						return
					}
				}

				// Add token information to the request context
				ctx := context.WithValue(r.Context(), "token_info", tokenInfo)
				ctx = context.WithValue(ctx, "auth_type", string(tokenInfo.Type))
				if tokenInfo.Email != "" {
					ctx = context.WithValue(ctx, userEmailKey, tokenInfo.Email)
				}
				r = r.WithContext(ctx)

			} else {
				// This is an M2M request - JWT authentication is optional
				if authHeader != "" {
					// If JWT is provided, verify it
					const bearerPrefix = "Bearer "
					if strings.HasPrefix(authHeader, bearerPrefix) {
						tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
						if tokenString != "" {
							// Verify the token
							token, err := jwtVerifier.VerifyToken(tokenString)
							if err == nil {
								// Extract token information and determine type
								tokenInfo, err := extractTokenInfo(token)
								if err == nil && tokenInfo.Type == TokenTypeM2M {
									// For M2M tokens, no additional scope validation required
									// M2M tokens are trusted for all consent operations

									// Add token information to the request context
									ctx := context.WithValue(r.Context(), "token_info", tokenInfo)
									ctx = context.WithValue(ctx, "auth_type", string(tokenInfo.Type))
									r = r.WithContext(ctx)
								}
							}
						}
					}
				}

				// For M2M without JWT, set auth type to M2M
				if r.Context().Value("auth_type") == nil {
					ctx := context.WithValue(r.Context(), "auth_type", "m2m")
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TestHybridAuthMiddleware tests the hybrid authentication middleware
func TestHybridAuthMiddleware(t *testing.T) {
	cleanup := SetupTestWithCleanup(t)
	defer cleanup()

	engine := setupPostgresTestEngine(t)

	// Create a test consent first
	consentID := createTestConsent(t, engine)

	// Create a mock JWT verifier
	mockVerifier := &MockJWTVerifier{
		validEmail: "test@example.com",
	}

	// Update the consent to have the matching email
	consentRecord, _ := engine.GetConsentStatus(consentID)
	consentRecord.OwnerEmail = "test@example.com"
	engine.UpdateConsent(consentID, UpdateConsentRequest{
		Status:    ConsentStatus(consentRecord.Status),
		UpdatedBy: "test@example.com",
		Reason:    "Test update",
	})

	// Configure user token validation for testing
	userTokenConfig := UserTokenValidationConfig{
		ExpectedIssuer:   getEnvOrDefault("TEST_ASGARDEO_ISSUER", "https://api.asgardeo.io/t/YOUR_TENANT/oauth2/token"),
		ExpectedAudience: getEnvOrDefault("TEST_ASGARDEO_AUDIENCE", "YOUR_AUDIENCE"),
		ExpectedOrgName:  getEnvOrDefault("TEST_ASGARDEO_ORG_NAME", "YOUR_ORG_NAME"),
		RequiredScopes:   []string{}, // No required scopes for testing
	}

	// Test hybrid authentication middleware
	middleware := hybridAuthMiddlewareForTesting(mockVerifier, engine, userTokenConfig)

	t.Run("UserToken_Success", func(t *testing.T) {
		// Create a mock user token
		userToken := createMockUserToken(t, "test@example.com", "user123")

		// Test token verification directly first
		token, err := mockVerifier.VerifyToken(userToken)
		if err != nil {
			t.Fatalf("Token verification failed: %v", err)
		}

		// Test token info extraction
		tokenInfo, err := extractTokenInfo(token)
		if err != nil {
			t.Fatalf("Token info extraction failed: %v", err)
		}

		if tokenInfo.Type != TokenTypeUser {
			t.Errorf("Expected token type 'user', got %v", tokenInfo.Type)
		}
		if tokenInfo.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %v", tokenInfo.Email)
		}
		if tokenInfo.Subject != "user123" {
			t.Errorf("Expected subject 'user123', got %v", tokenInfo.Subject)
		}

		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		// Add frontend headers to identify this as a frontend request
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that token info is set correctly
			tokenInfo := r.Context().Value("token_info").(*TokenInfo)
			if tokenInfo.Type != TokenTypeUser {
				t.Errorf("Expected token type 'user', got %v", tokenInfo.Type)
			}
			if tokenInfo.Email != "test@example.com" {
				t.Errorf("Expected email 'test@example.com', got %v", tokenInfo.Email)
			}
			if tokenInfo.Subject != "user123" {
				t.Errorf("Expected subject 'user123', got %v", tokenInfo.Subject)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("M2MToken_Success", func(t *testing.T) {
		// Create a mock M2M token
		m2mToken := createMockM2MToken(t, "client123", []string{"consent:read", "consent:write"})

		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		req.Header.Set("Authorization", "Bearer "+m2mToken)
		// No frontend headers - this should be treated as M2M
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that token info is set correctly
			tokenInfo := r.Context().Value("token_info").(*TokenInfo)
			if tokenInfo.Type != TokenTypeM2M {
				t.Errorf("Expected token type 'm2m', got %v", tokenInfo.Type)
			}
			if tokenInfo.ClientID != "client123" {
				t.Errorf("Expected client_id 'client123', got %v", tokenInfo.ClientID)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("UserToken_EmailMismatch_Failure", func(t *testing.T) {
		// Create a mock user token with different email
		userToken := createMockUserToken(t, "other@example.com", "user456")

		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		// Add frontend headers to identify this as a frontend request
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called with email mismatch")
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		expectedError := "Access denied: email does not match consent owner"
		if response["error"] != expectedError {
			t.Errorf("Expected '%s' error, got %v", expectedError, response["error"])
		}
	})

	t.Run("InvalidToken_Failure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		req.Header.Set("Authorization", "Bearer invalid.token")
		// Add frontend headers to identify this as a frontend request
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called with invalid token")
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("MissingAuthorizationHeader_Failure", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		// Add frontend headers to identify this as a frontend request
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called without Authorization header")
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	// New test cases for M2M without JWT and frontend with JWT
	t.Run("M2M_NoJWT_Success", func(t *testing.T) {
		// M2M calls should work without JWT authentication
		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		// No Authorization header - this should work for M2M
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that no token info is set (M2M without JWT)
			tokenInfo := r.Context().Value("token_info")
			if tokenInfo != nil {
				t.Errorf("Expected no token info for M2M without JWT, got %v", tokenInfo)
			}

			// Check that auth type is set to M2M
			authType := r.Context().Value("auth_type")
			if authType != "m2m" {
				t.Errorf("Expected auth_type 'm2m', got %v", authType)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})

	t.Run("Frontend_NoJWT_Failure", func(t *testing.T) {
		// Frontend calls should fail without JWT authentication
		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		// Add a header to identify this as a frontend call
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler should not be called for frontend without JWT")
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		expectedError := "Authorization header is required"
		if response["error"] != expectedError {
			t.Errorf("Expected '%s' error, got %v", expectedError, response["error"])
		}
	})

	t.Run("M2M_WithJWT_Success", func(t *testing.T) {
		// M2M calls with JWT should also work
		m2mToken := createMockM2MToken(t, "client789", []string{"consent:read", "consent:write"})

		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		req.Header.Set("Authorization", "Bearer "+m2mToken)
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that token info is set correctly
			tokenInfo := r.Context().Value("token_info").(*TokenInfo)
			if tokenInfo.Type != TokenTypeM2M {
				t.Errorf("Expected token type 'm2m', got %v", tokenInfo.Type)
			}
			if tokenInfo.ClientID != "client789" {
				t.Errorf("Expected client_id 'client789', got %v", tokenInfo.ClientID)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("Frontend_WithJWT_Success", func(t *testing.T) {
		// Frontend calls with JWT should work
		userToken := createMockUserToken(t, "test@example.com", "user999")

		req := httptest.NewRequest("GET", "/consents/"+consentID, nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		w := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that token info is set correctly
			tokenInfo := r.Context().Value("token_info").(*TokenInfo)
			if tokenInfo.Type != TokenTypeUser {
				t.Errorf("Expected token type 'user', got %v", tokenInfo.Type)
			}
			if tokenInfo.Email != "test@example.com" {
				t.Errorf("Expected email 'test@example.com', got %v", tokenInfo.Email)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}))

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
		}
	})
}

// Helper function to create a test consent
func createTestConsent(t *testing.T, engine ConsentEngine) string {
	req := ConsentRequest{
		AppID: "test-app",
		DataFields: []DataField{
			{

				OwnerID:    "test-owner-123",
				OwnerEmail: "test@example.com",
				Fields:     []string{"person.name", "person.email"},
			},
		},

		SessionID: "test-session-123",
	}

	response, err := engine.ProcessConsentRequest(req)
	if err != nil {
		t.Fatalf("Failed to create test consent: %v", err)
	}

	return response.ConsentID
}

// createMockUserToken creates a mock JWT token for a user
func createMockUserToken(t *testing.T, email, subject string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      subject,
		"email":    email,
		"aut":      "APPLICATION_USER",
		"iss":      getEnvOrDefault("TEST_ASGARDEO_ISSUER", "https://api.asgardeo.io/t/YOUR_TENANT/oauth2/token"),
		"aud":      getEnvOrDefault("TEST_ASGARDEO_AUDIENCE", "YOUR_AUDIENCE"),
		"org_name": getEnvOrDefault("TEST_ASGARDEO_ORG_NAME", "YOUR_ORG_NAME"),
		"iat":      now.Unix(),
		"exp":      now.Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create mock user token: %v", err)
	}

	return tokenString
}

// createMockM2MToken creates a mock JWT token for M2M authentication
func createMockM2MToken(t *testing.T, clientID string, scopes []string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       clientID,
		"client_id": clientID,
		"aut":       "APPLICATION",
		"scope":     joinStrings(scopes, " "),
		"iss":       getEnvOrDefault("TEST_ASGARDEO_ISSUER", "https://api.asgardeo.io/t/YOUR_TENANT/oauth2/token"),
		"aud":       getEnvOrDefault("TEST_ASGARDEO_AUDIENCE", "YOUR_AUDIENCE"),
		"iat":       now.Unix(),
		"exp":       now.Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create mock M2M token: %v", err)
	}

	return tokenString
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// MockJWTVerifier for testing
type MockJWTVerifier struct {
	validEmail string
	shouldFail bool
}

func (m *MockJWTVerifier) VerifyToken(tokenString string) (*jwt.Token, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock JWT verification failed")
	}

	// Check for invalid tokens
	if tokenString == "invalid.token" || tokenString == "invalid.jwt.token" {
		return nil, fmt.Errorf("invalid token")
	}

	// Parse the token with proper verification for testing
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("test-secret"), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return token, nil
}

func (m *MockJWTVerifier) ExtractEmailFromToken(token *jwt.Token) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock email extraction failed")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	if email, ok := claims["email"].(string); ok {
		return email, nil
	}

	return m.validEmail, nil
}

func (m *MockJWTVerifier) VerifyAndExtractEmail(tokenString string) (string, error) {
	token, err := m.VerifyToken(tokenString)
	if err != nil {
		return "", err
	}

	return m.ExtractEmailFromToken(token)
}
