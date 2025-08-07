# Makefile for running microservices

.PHONY: run-mock-data-drp run-drp run-dmt run-provider-wrappers run-graphql-resolver clean logs

# Run individual services
run-mock-data-drp:
	@echo "Starting mock data DRP service..."
	@cd mocks/mock-drp/ && bal run

run-drp:
	@make run-mock-data-drp & # Ensure mock data is ready before starting DRP
	@echo "Starting DRP service..."
	@cd provider-wrappers/drp/ && bal run

run-dmt:
	@echo "Starting DMT service..."
	@cd provider-wrappers/dmt/ && bal run

# Run all provider wrappers
run-providers:
	@echo "Starting all provider wrappers..."
	@make run-drp & make run-dmt  # Ensure both DRP and DMT are running

# Run GraphQL Resolver
run-resolver:
	@echo "Starting GraphQL Resolver..."
	@cd graphql-resolver/ && node index.js


# View logs (if you want to redirect logs to files)
logs:
	@echo "Service logs would appear here if redirected to files"

# Clean up any temporary files
clean:
	@echo "Cleaning up..."
	@rm -f *.log

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Install dependencies here if needed"
