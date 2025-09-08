package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gov-dx-sandbox/exchange/utils"
)

// Constants for configuration
const (
	defaultPort = "8080"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	// Create an instance of our evaluator
	evaluator, err := NewPolicyEvaluator(ctx)
	if err != nil {
		slog.Error("Could not initialize policy evaluator", "error", err)
		os.Exit(1)
	}

	// Register the handler method from our evaluator instance
	http.Handle("/decide", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.policyDecisionHandler)))
	http.Handle("/debug", utils.PanicRecoveryMiddleware(http.HandlerFunc(evaluator.debugHandler)))

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	listenAddr := fmt.Sprintf(":%s", port)

	slog.Info("PCE server starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		slog.Error("Could not start server", "error", err)
		os.Exit(1)
	}
}
