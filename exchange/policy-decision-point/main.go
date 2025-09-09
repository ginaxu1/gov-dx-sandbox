package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gov-dx-sandbox/exchange/config"
	"github.com/gov-dx-sandbox/exchange/utils"
)

func main() {
	// Load configuration using flags
	cfg := config.LoadConfig("policy-decision-point")

	// Setup logging
	setupLogging(cfg)

	slog.Info("Starting policy decision point",
		"environment", cfg.Environment,
		"port", cfg.Service.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize policy evaluator
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		slog.Error("Failed to initialize policy evaluator", "error", err)
		os.Exit(1)
	}

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/decide", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.policyDecisionHandler)))
	mux.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.debugHandler)))
	mux.Handle("/health", utils.PanicRecoveryMiddleware(utils.HealthHandler("policy-decision-point")))

	// Get port from configuration
	port := cfg.Service.Port
	listenAddr := fmt.Sprintf(":%s", port)

	slog.Info("Policy decision point server starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		slog.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}

// setupLogging configures logging based on the configuration
func setupLogging(cfg *config.Config) {
	var handler slog.Handler

	switch cfg.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(cfg.Logging.Level),
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(cfg.Logging.Level),
		})
	}

	slog.SetDefault(slog.New(handler))
}

// getLogLevel converts string level to slog.Level
func getLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
