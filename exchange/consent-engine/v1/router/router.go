package router

import (
	"net/http"

	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/auth"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/handlers"
	"github.com/gov-dx-sandbox/exchange/consent-engine/v1/middleware"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// V1Router handles all V1 API route registration
type V1Router struct {
	internalHandler *handlers.InternalHandler
	portalHandler   *handlers.PortalHandler
	authMiddleware  *middleware.JWTAuthMiddleware
	corsMiddleware  func(http.Handler) http.Handler
}

// NewV1Router creates a new V1 router with all dependencies
func NewV1Router(
	internalHandler *handlers.InternalHandler,
	portalHandler *handlers.PortalHandler,
	jwtVerifier *auth.JWTVerifier,
) *V1Router {
	return &V1Router{
		internalHandler: internalHandler,
		portalHandler:   portalHandler,
		authMiddleware:  middleware.NewJWTAuthMiddleware(jwtVerifier),
		corsMiddleware:  middleware.NewCORSMiddleware(),
	}
}

// RegisterRoutes registers all V1 API routes to the provided mux
func (r *V1Router) RegisterRoutes(mux *http.ServeMux) {
	r.registerInternalRoutes(mux)
	r.registerPortalRoutes(mux)
}

// registerInternalRoutes registers internal API routes (no authentication required)
func (r *V1Router) registerInternalRoutes(mux *http.ServeMux) {
	// Health check
	mux.Handle("/internal/api/v1/health",
		utils.PanicRecoveryMiddleware(http.HandlerFunc(r.internalHandler.HealthCheck)))

	// Consents endpoint
	mux.Handle("/internal/api/v1/consents",
		utils.PanicRecoveryMiddleware(http.HandlerFunc(r.handleInternalConsents)))
}

// registerPortalRoutes registers portal API routes (authentication required for protected endpoints)
func (r *V1Router) registerPortalRoutes(mux *http.ServeMux) {
	// Health check endpoint (public - no authentication per OpenAPI spec)
	mux.Handle("/api/v1/health",
		utils.PanicRecoveryMiddleware(http.HandlerFunc(r.portalHandler.HealthCheck)))

	// Consent endpoints (authentication required)
	mux.Handle("/api/v1/consents/",
		utils.PanicRecoveryMiddleware(
			r.authMiddleware.Authenticate(http.HandlerFunc(r.handlePortalConsents))))
}

// handleInternalConsents routes internal consent requests to appropriate handlers
func (r *V1Router) handleInternalConsents(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.internalHandler.GetConsent(w, req)
	case http.MethodPost:
		r.internalHandler.CreateConsent(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePortalConsents routes portal consent requests to appropriate handlers
func (r *V1Router) handlePortalConsents(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.portalHandler.GetConsent(w, req)
	case http.MethodPut:
		r.portalHandler.UpdateConsent(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ApplyCORS wraps a handler with CORS middleware
func (r *V1Router) ApplyCORS(handler http.Handler) http.Handler {
	return r.corsMiddleware(handler)
}
