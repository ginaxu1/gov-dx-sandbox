package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/config"
	"github.com/gov-dx-sandbox/exchange/shared/utils"

	// V1 API imports
	v1auth "github.com/gov-dx-sandbox/exchange/consent-engine/v1/auth"
	v1db "github.com/gov-dx-sandbox/exchange/consent-engine/v1/database"
	v1handlers "github.com/gov-dx-sandbox/exchange/consent-engine/v1/handlers"
	v1router "github.com/gov-dx-sandbox/exchange/consent-engine/v1/router"
	v1services "github.com/gov-dx-sandbox/exchange/consent-engine/v1/services"
)

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load configuration using flags
	cfg := config.LoadConfig("consent-engine")

	// Setup logging
	utils.SetupLogging(cfg.Logging.Format, cfg.Logging.Level)

	slog.Info("Starting consent engine",
		"environment", cfg.Environment,
		"port", cfg.Service.Port,
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit)

	// Initialize V1 database connection
	slog.Info("Initializing V1 database connection...")
	v1DBConfig := v1db.NewDatabaseConfig()
	v1DB, err := v1db.ConnectGormDB(v1DBConfig)
	if err != nil {
		slog.Error("Failed to connect to V1 database", "error", err)
		os.Exit(1)
	}
	slog.Info("V1 database connected successfully")

	// Get underlying SQL DB for proper cleanup
	v1SqlDB, err := v1DB.DB()
	if err != nil {
		slog.Error("Failed to get V1 database connection", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := v1SqlDB.Close(); err != nil {
			slog.Error("Failed to close V1 database connection", "error", err)
		} else {
			slog.Info("V1 database connection closed successfully")
		}
	}()

	// Get consent portal URL from environment
	consentPortalUrl := getEnvOrDefault("CONSENT_PORTAL_URL", "http://localhost:5173")
	slog.Info("Using consent portal URL", "url", consentPortalUrl)

	// Initialize V1 consent service
	v1ConsentService, err := v1services.NewConsentService(v1DB, consentPortalUrl)
	if err != nil {
		slog.Error("Failed to initialize V1 consent service", "error", err)
		os.Exit(1)
	}

	// Initialize V1 handlers
	v1InternalHandler := v1handlers.NewInternalHandler(v1ConsentService)
	v1PortalHandler := v1handlers.NewPortalHandler(v1ConsentService)

	// Initialize V1 JWT verifier
	orgName := getEnvOrDefault("ASGARDEO_ORG_NAME", "YOUR_ORG_NAME")
	userIssuer := getEnvOrDefault("ASGARDEO_ISSUER", "https://api.asgardeo.io/t/"+orgName+"/oauth2/token")
	userAudience := getEnvOrDefault("ASGARDEO_AUDIENCE", "YOUR_AUDIENCE")
	userJwksURL := getEnvOrDefault("ASGARDEO_JWKS_URL", "https://api.asgardeo.io/t/"+orgName+"/oauth2/jwks")

	slog.Info("JWT verifier configuration",
		"org_name", orgName,
		"issuer", userIssuer,
		"audience", userAudience,
		"jwks_url", userJwksURL)

	v1JWTVerifier, err := v1auth.NewJWTVerifier(v1auth.JWTVerifierConfig{
		JWKSUrl:      userJwksURL,
		Issuer:       userIssuer,
		Audience:     userAudience,
		Organization: orgName,
	})
	if err != nil {
		slog.Error("Failed to initialize V1 JWT verifier", "error", err)
		os.Exit(1)
	}

	// Initialize V1 router and register all V1 routes
	v1Router := v1router.NewV1Router(v1InternalHandler, v1PortalHandler, v1JWTVerifier)
	mux := http.NewServeMux()

	slog.Info("Registering V1 API routes")
	v1Router.RegisterRoutes(mux)
	slog.Info("V1 API routes registered successfully")

	// Register legacy /health endpoint for compatibility with health checks
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("consent-engine")))

	// Create server configuration
	serverConfig := &utils.ServerConfig{
		Port:         cfg.Service.Port,
		ReadTimeout:  cfg.Service.Timeout,
		WriteTimeout: cfg.Service.Timeout,
		IdleTimeout:  60 * time.Second,
	}

	// Apply CORS middleware and create server
	handler := v1Router.ApplyCORS(mux)
	httpServer := utils.CreateServer(serverConfig, handler)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(httpServer, "consent-engine"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
