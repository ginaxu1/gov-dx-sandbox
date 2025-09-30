package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gov-dx-sandbox/exchange/shared/config"
	"github.com/gov-dx-sandbox/exchange/shared/utils"
)

// Build information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig("policy-decision-point")

	// Setup logging
	utils.SetupLogging(cfg.Logging.Format, cfg.Logging.Level)

	slog.Info("Starting policy decision point",
		"environment", cfg.Environment,
		"port", cfg.Service.Port,
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize policy evaluator (which includes database service)
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		slog.Error("Failed to initialize policy evaluator", "error", err)
		os.Exit(1)
	}

	// Initialize metadata handler with database service and evaluator
	metadataHandler := NewMetadataHandler(evaluator.dbService, evaluator)

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/decide", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.policyDecisionHandler)))
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.debugHandler)))
	mux.Handle("/metadata/update", utils.PanicRecoveryMiddleware(http.HandlerFunc(metadataHandler.UpdateProviderMetadata)))
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("policy-decision-point")))

	// Create server using utils
	serverConfig := &utils.ServerConfig{
		Port:         cfg.Service.Port,
		ReadTimeout:  cfg.Service.Timeout,
		WriteTimeout: cfg.Service.Timeout,
		IdleTimeout:  60 * time.Second,
	}
	server := utils.CreateServer(serverConfig, mux)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(server, "policy-decision-point"); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}

	// Cleanup database connection on shutdown
	defer func() {
		if err := evaluator.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
	}()
}
