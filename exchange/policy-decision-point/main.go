package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/gov-dx-sandbox/exchange/utils"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	// Setup server with default configuration
	config := utils.DefaultServerConfig()
	config.Port = utils.GetEnvOrDefault("PORT", "8080")
	server := utils.CreateServer(config, mux)

	// Start server with graceful shutdown
	if err := utils.StartServerWithGracefulShutdown(server, "policy-decision-point"); err != nil {
		os.Exit(1)
	}
}
