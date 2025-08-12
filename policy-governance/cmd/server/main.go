// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"policy-governance/internal/database"
	"policy-governance/internal/policies"
	"policy-governance/internal/repository"
)

func main() {
    database.InitDB()

    repo := repository.NewPolicyRepository()
    policyHandler := policies.NewPolicyHandler(repo) 

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/policies/access-policy/{consumerID}/{providerID}", policyHandler.GetAccessPolicy)

	port := ":3001"
	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Policy Governance Service running on port %s. 🛡️", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	<-stop
	log.Println("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server gracefully stopped")
}