# Makefile for running microservices

.PHONY: run run-mock-data-drp run-drp run-dmt run-provider-wrappers run-graphql-resolver stop clean logs

# Run all services
run:
	@echo "Starting all microservices..."
	@make run-provider-wrappers &
	@echo "Start Sleep"
	@sleep 20  # Wait for provider wrappers to be fully up
	@echo "Sleep complete"
	@echo "All services started. 'make stop' to stop all services."
	

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

run-provider-wrappers:
	@echo "Starting all provider wrappers..."
	@make run-drp & make run-dmt  # Ensure both DRP and DMT are running


run-graphql-resolver:
	@echo "Starting GraphQL Resolver..."
	@cd graphql-resolver/ && node index.js

# Stop all services (kills processes by name - adjust as needed)
stop:
	@echo "Stopping all services..."
	@pkill -f "bal run" || true
	@pkill -f "node index.js" || true
	@echo "All services stopped."

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

# Health check (optional)
health:
	@echo "Checking service health..."
	@curl -s http://localhost:PORT1/health || echo "DRP service not responding"
	@curl -s http://localhost:PORT2/health || echo "DMT service not responding"
	@curl -s http://localhost:PORT3/health || echo "GraphQL service not responding"