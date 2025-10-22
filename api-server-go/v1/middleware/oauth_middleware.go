package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gov-dx-sandbox/api-server-go/models"
	"github.com/gov-dx-sandbox/api-server-go/services"
	"github.com/gov-dx-sandbox/api-server-go/shared"
	"github.com/gov-dx-sandbox/api-server-go/shared/utils"
)

// OAuth2Middleware provides OAuth 2.0 token validation middleware
type OAuth2Middleware struct {
	oauthService *services.OAuth2Service
}

// NewOAuth2Middleware creates a new OAuth 2.0 middleware
func NewOAuth2Middleware(oauthService *services.OAuth2Service) *OAuth2Middleware {
	return &OAuth2Middleware{oauthService: oauthService}
}

// Context key types to avoid collisions
type contextKey string

const (
	userInfoKey contextKey = "user_info"
	clientIDKey contextKey = "client_id"
	scopesKey   contextKey = "scopes"
)

// UserInfoFromContext extracts user information from request context
func UserInfoFromContext(ctx context.Context) (*models.UserInfo, bool) {
	userInfo, ok := ctx.Value(userInfoKey).(*models.UserInfo)
	return userInfo, ok
}

// ClientIDFromContext extracts client ID from request context
func ClientIDFromContext(ctx context.Context) (string, bool) {
	clientID, ok := ctx.Value(clientIDKey).(string)
	return clientID, ok
}

// ScopesFromContext extracts scopes from request context
func ScopesFromContext(ctx context.Context) ([]string, bool) {
	scopes, ok := ctx.Value(scopesKey).([]string)
	return scopes, ok
}

// RequireOAuth2Token validates OAuth 2.0 access tokens
func (m *OAuth2Middleware) RequireOAuth2Token(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract access token from Authorization header
		accessToken, err := shared.ExtractAccessToken(r)
		if err != nil {
			slog.Warn("Missing or invalid access token", "error", err, "path", r.URL.Path)
			utils.RespondWithError(w, http.StatusUnauthorized, "Missing or invalid access token")
			return
		}

		// Validate access token and get user information
		userInfo, err := m.oauthService.ValidateAccessToken(accessToken)
		if err != nil {
			slog.Warn("Access token validation failed", "error", err, "path", r.URL.Path)
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired access token")
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
		ctx = context.WithValue(ctx, clientIDKey, userInfo.ClientID)
		ctx = context.WithValue(ctx, scopesKey, userInfo.Scopes)
		r = r.WithContext(ctx)

		slog.Info("OAuth2 token validated",
			"user_id", userInfo.UserID,
			"client_id", userInfo.ClientID,
			"scopes", userInfo.Scopes,
			"path", r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// RequireScope validates that the user has the required scope
func (m *OAuth2Middleware) RequireScope(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get scopes from context
			scopes, ok := ScopesFromContext(r.Context())
			if !ok {
				slog.Warn("No scopes found in context", "path", r.URL.Path)
				utils.RespondWithError(w, http.StatusInternalServerError, "No scopes found in context")
				return
			}

			// Check if user has required scope
			if !shared.HasRequiredScope(scopes, requiredScope) {
				slog.Warn("Insufficient scope",
					"required_scope", requiredScope,
					"user_scopes", scopes,
					"path", r.URL.Path)
				utils.RespondWithError(w, http.StatusForbidden, "Insufficient scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyScope validates that the user has at least one of the required scopes
func (m *OAuth2Middleware) RequireAnyScope(requiredScopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get scopes from context
			scopes, ok := ScopesFromContext(r.Context())
			if !ok {
				slog.Warn("No scopes found in context", "path", r.URL.Path)
				utils.RespondWithError(w, http.StatusInternalServerError, "No scopes found in context")
				return
			}

			// Check if user has any of the required scopes
			if !shared.HasAnyScope(scopes, requiredScopes...) {
				slog.Warn("Insufficient scope",
					"required_scopes", requiredScopes,
					"user_scopes", scopes,
					"path", r.URL.Path)
				utils.RespondWithError(w, http.StatusForbidden, "Insufficient scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllScopes validates that the user has all of the required scopes
func (m *OAuth2Middleware) RequireAllScopes(requiredScopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get scopes from context
			scopes, ok := ScopesFromContext(r.Context())
			if !ok {
				slog.Warn("No scopes found in context", "path", r.URL.Path)
				utils.RespondWithError(w, http.StatusInternalServerError, "No scopes found in context")
				return
			}

			// Check if user has all required scopes
			if !shared.HasAllScopes(scopes, requiredScopes...) {
				slog.Warn("Insufficient scope",
					"required_scopes", requiredScopes,
					"user_scopes", scopes,
					"path", r.URL.Path)
				utils.RespondWithError(w, http.StatusForbidden, "Insufficient scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalOAuth2Token validates OAuth 2.0 access tokens if present
func (m *OAuth2Middleware) OptionalOAuth2Token(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract access token from Authorization header
		accessToken, err := shared.ExtractAccessToken(r)
		if err != nil {
			// No token present, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate access token and get user information
		userInfo, err := m.oauthService.ValidateAccessToken(accessToken)
		if err != nil {
			// Invalid token, continue without authentication
			slog.Warn("Optional OAuth2 token validation failed", "error", err, "path", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
		ctx = context.WithValue(ctx, clientIDKey, userInfo.ClientID)
		ctx = context.WithValue(ctx, scopesKey, userInfo.Scopes)
		r = r.WithContext(ctx)

		slog.Info("Optional OAuth2 token validated",
			"user_id", userInfo.UserID,
			"client_id", userInfo.ClientID,
			"scopes", userInfo.Scopes,
			"path", r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

// Helper methods

// Middleware composition helpers

// OAuth2WithScope creates a middleware that requires OAuth 2.0 token and specific scope
func OAuth2WithScope(oauthService *services.OAuth2Service, requiredScope string) func(http.Handler) http.Handler {
	middleware := NewOAuth2Middleware(oauthService)
	return func(next http.Handler) http.Handler {
		return middleware.RequireOAuth2Token(middleware.RequireScope(requiredScope)(next))
	}
}

// OAuth2WithAnyScope creates a middleware that requires OAuth 2.0 token and any of the specified scopes
func OAuth2WithAnyScope(oauthService *services.OAuth2Service, requiredScopes ...string) func(http.Handler) http.Handler {
	middleware := NewOAuth2Middleware(oauthService)
	return func(next http.Handler) http.Handler {
		return middleware.RequireOAuth2Token(middleware.RequireAnyScope(requiredScopes...)(next))
	}
}

// OAuth2WithAllScopes creates a middleware that requires OAuth 2.0 token and all of the specified scopes
func OAuth2WithAllScopes(oauthService *services.OAuth2Service, requiredScopes ...string) func(http.Handler) http.Handler {
	middleware := NewOAuth2Middleware(oauthService)
	return func(next http.Handler) http.Handler {
		return middleware.RequireOAuth2Token(middleware.RequireAllScopes(requiredScopes...)(next))
	}
}
