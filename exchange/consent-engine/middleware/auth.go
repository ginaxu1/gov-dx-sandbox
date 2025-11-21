package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"


	"github.com/gov-dx-sandbox/exchange/consent-engine/service"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// Context key types to avoid collisions
type contextKey string

const (
	userEmailKey contextKey = "user_email"
	authTypeKey  contextKey = "auth_type"
	tokenInfoKey contextKey = "token_info"
)

// Token type detection
type TokenType string

const (
	TokenTypeUser    TokenType = "user"
	TokenTypeUnknown TokenType = "unknown"
)

// TokenInfo contains information about a verified token
type TokenInfo struct {
	Type     TokenType
	Subject  string
	Email    string
	ClientID string
	Issuer   string
	Audience []string
	Scopes   []string
	AuthType string
}

// UserTokenValidationConfig holds configuration for user token validation
type UserTokenValidationConfig struct {
	ExpectedIssuer   string
	ExpectedAudience string
	ExpectedOrgName  string
	RequiredScopes   []string
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UserAuthMiddleware handles user JWT authentication only
func UserAuthMiddleware(userJWTVerifier *JWTVerifier, engine service.ConsentEngine, userTokenConfig UserTokenValidationConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract consent ID from the URL path
			// Handle both /consents/{id} and /consents/{id}/ patterns
			path := strings.TrimPrefix(r.URL.Path, "/consents")
			path = strings.TrimPrefix(path, "/")
			if path == "" {
				utils.RespondWithJSON(w, http.StatusBadRequest, utils.ErrorResponse{Error: "Consent ID is required"})
				return
			}

			// Remove any trailing slashes and additional path segments
			consentID := strings.Split(path, "/")[0]
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

			// Extract the Authorization header
			authHeader := r.Header.Get("Authorization")
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

			// Verify the token using user JWT verifier
			slog.Info("Attempting JWT verification",
				"consent_id", consentID,
				"token_length", len(tokenString),
				"token_preview", tokenString[:min(50, len(tokenString))]+"...")
			token, err := userJWTVerifier.VerifyToken(tokenString)
			if err != nil {
				slog.Error("User token verification failed",
					"error", err,
					"consent_id", consentID,
					"error_type", fmt.Sprintf("%T", err),
					"token_preview", tokenString[:min(50, len(tokenString))]+"...")
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Invalid user token"})
				return
			}
			slog.Info("JWT verification successful", "consent_id", consentID)

			// Extract email from token
			email, err := userJWTVerifier.ExtractEmailFromToken(token)
			if err != nil {
				slog.Warn("Failed to extract email from user token", "error", err, "consent_id", consentID)
				utils.RespondWithJSON(w, http.StatusUnauthorized, utils.ErrorResponse{Error: "Token missing email claim"})
				return
			}

			// Check if the email matches the consent owner email
			if consentRecord.OwnerEmail != email {
				slog.Warn("User email does not match consent owner email",
					"user_email", email,
					"consent_owner_email", consentRecord.OwnerEmail,
					"consent_id", consentID)
				utils.RespondWithJSON(w, http.StatusForbidden, utils.ErrorResponse{Error: "Access denied: email does not match consent owner"})
				return
			}

			// Add user auth type and email to the request context
			ctx := context.WithValue(r.Context(), authTypeKey, "user")
			ctx = context.WithValue(ctx, userEmailKey, email)
			r = r.WithContext(ctx)

			slog.Info("User authentication successful",
				"email", email,
				"consent_id", consentID)

			next.ServeHTTP(w, r)
		})
	}
}

// SelectiveAuthMiddleware applies authentication only to specific HTTP methods
func SelectiveAuthMiddleware(userJWTVerifier *JWTVerifier, engine service.ConsentEngine, userTokenConfig UserTokenValidationConfig, requireAuthMethods []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this method requires authentication
			requiresAuth := false
			for _, method := range requireAuthMethods {
				if r.Method == method {
					requiresAuth = true
					break
				}
			}

			if !requiresAuth {
				// No authentication required for this method
				next.ServeHTTP(w, r)
				return
			}

			// Apply authentication for methods that require it
			authHandler := UserAuthMiddleware(userJWTVerifier, engine, userTokenConfig)(next)
			authHandler.ServeHTTP(w, r)
		})
	}
}
