package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"syscall"
	"time"

	"github.com/gov-dx-sandbox/exchange/consent-engine/handlers"
	"github.com/gov-dx-sandbox/exchange/consent-engine/middleware"
	"github.com/gov-dx-sandbox/exchange/consent-engine/service"
	"github.com/gov-dx-sandbox/exchange/consent-engine/store"

	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialize logger
	logLevel := utils.GetEnvOrDefault("LOG_LEVEL", "info")
	utils.SetupLogging("text", logLevel)

	slog.Info("Starting Consent Engine",
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit)

	// Initialize database connection
	dbConfig := store.NewDatabaseConfig()
	db, err := store.ConnectDB(dbConfig)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize database schema
	if err := store.InitDatabase(db); err != nil {
		slog.Error("Failed to initialize database schema", "error", err)
		os.Exit(1)
	}

	// Initialize consent engine service
	consentPortalURL := utils.GetEnvOrDefault("CONSENT_PORTAL_URL", "http://localhost:3000")
	consentEngine := service.NewPostgresConsentEngine(db, consentPortalURL)

	// Start background process for expiry check
	expiryCheckIntervalStr := utils.GetEnvOrDefault("CONSENT_EXPIRY_CHECK_INTERVAL", "1h")
	expiryCheckInterval, err := time.ParseDuration(expiryCheckIntervalStr)
	if err != nil {
		slog.Warn("Invalid expiry check interval, using default", "error", err, "default", "1h")
		expiryCheckInterval = time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consentEngine.StartBackgroundExpiryProcess(ctx, expiryCheckInterval)
	defer consentEngine.StopBackgroundExpiryProcess()

	// Initialize JWT verifier for user tokens
	asgardeoBaseURL := utils.GetEnvOrDefault("ASGARDEO_BASE_URL", "")
	asgardeoIssuer := utils.GetEnvOrDefault("ASGARDEO_ISSUER", "")
	if asgardeoIssuer == "" && asgardeoBaseURL != "" {
		asgardeoIssuer = fmt.Sprintf("%s/oauth2/token", asgardeoBaseURL)
	}
	asgardeoJWKSURL := utils.GetEnvOrDefault("ASGARDEO_JWKS_URL", "")
	if asgardeoJWKSURL == "" && asgardeoBaseURL != "" {
		asgardeoJWKSURL = fmt.Sprintf("%s/oauth2/jwks", asgardeoBaseURL)
	}

	userJWTVerifier := middleware.NewJWTVerifier(asgardeoIssuer, asgardeoJWKSURL, utils.GetEnvOrDefault("ASGARDEO_ORG_NAME", ""))

	// Configure user token validation
	userTokenConfig := middleware.UserTokenValidationConfig{
		ExpectedIssuer:   asgardeoIssuer,
		ExpectedAudience: utils.GetEnvOrDefault("ASGARDEO_AUDIENCE", ""),
		ExpectedOrgName:  utils.GetEnvOrDefault("ASGARDEO_ORG_NAME", ""),
	}

	// Initialize handlers
	consentHandler := handlers.NewConsentHandler(consentEngine)

	// Setup router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"version": Version,
		})
	})

	// Public endpoints (no auth required)
	// POST /consents - Create consent
	mux.HandleFunc("/consents", consentHandler.ConsentHandler)

	// Protected endpoints (require user authentication)
	// GET /consents/{id} - Get consent status
	// PUT /consents/{id} - Update consent
	// PATCH /consents/{id} - Partial update consent
	// DELETE /consents/{id} - Revoke consent
	// Note: Some methods on /consents/{id} might be public or have different auth requirements
	// We'll use a selective auth middleware wrapper
	consentIDHandler := http.HandlerFunc(consentHandler.ConsentHandlerWithID)
	// GET and PUT require auth, PATCH and DELETE do not (based on original implementation)
	// Wait, original implementation had:
	// - GET /consents/{id}: User Auth
	// - PUT /consents/{id}: User Auth
	// - PATCH /consents/{id}: No Auth (Internal/System use?)
	// - DELETE /consents/{id}: No Auth (Internal/System use?)
	// Let's replicate that logic
	protectedMethods := []string{http.MethodGet, http.MethodPut}
	mux.Handle("/consents/", middleware.SelectiveAuthMiddleware(userJWTVerifier, consentEngine, userTokenConfig, protectedMethods)(consentIDHandler))

	// Data Owner endpoints
	// GET /data-owner/{id}
	mux.HandleFunc("/data-owner/", consentHandler.DataOwnerHandler)

	// Consumer endpoints
	// GET /consumer/{id}
	mux.HandleFunc("/consumer/", consentHandler.ConsumerHandler)

	// Data Info endpoint (used by consent portal)
	// GET /data-info/{id}
	mux.HandleFunc("/data-info/", consentHandler.DataInfoHandler)

	// Admin endpoints
	// POST /admin/expiry-check
	mux.HandleFunc("/admin/", consentHandler.AdminHandler)

	// Setup server with middleware
	// Chain: CORS -> Router
	handler := middleware.CorsMiddleware(mux)

	// Start server
	port := utils.GetEnvOrDefault("PORT", "8081")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		slog.Info("Server starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server shutting down...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited properly")
}
