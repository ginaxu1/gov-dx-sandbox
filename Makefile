# Gov DX Sandbox - Unified Service Management
# Manages both Exchange Services and Sri Lanka Passport Application

.PHONY: help start-all stop-all status logs clean build dev start-passport stop-passport start-exchange stop-exchange restart-all

# Default target
help:
	@echo "Gov DX Sandbox - Service Management"
	@echo "===================================="
	@echo ""
	@echo "Main Commands:"
	@echo "  start-all      - Start all services (Exchange + Passport App)"
	@echo "  stop-all       - Stop all services"
	@echo "  restart-all    - Restart all services"
	@echo "  status         - Check status of all services"
	@echo "  logs           - View logs for all services"
	@echo "  clean          - Clean up all services and containers"
	@echo ""
	@echo "Exchange Services:"
	@echo "  start-exchange - Start only Exchange services (PDP, CE, OE)"
	@echo "  stop-exchange  - Stop only Exchange services"
	@echo "  logs-exchange  - View Exchange services logs"
	@echo ""
	@echo "Sri Lanka Passport App:"
	@echo "  start-passport - Start only Passport application"
	@echo "  stop-passport  - Stop only Passport application"
	@echo "  logs-passport  - View Passport application logs"
	@echo "  dev-passport   - Start Passport app in development mode"
	@echo ""
	@echo "Development:"
	@echo "  build          - Build all services"
	@echo "  test           - Run all tests"
	@echo ""
	@echo "Service Ports:"
	@echo "  Exchange Services:"
	@echo "    - Policy Decision Point (PDP): http://localhost:8082"
	@echo "    - Consent Engine (CE): http://localhost:8081"
	@echo "    - Orchestration Engine (OE): http://localhost:8080"
	@echo "  Sri Lanka Passport: http://localhost:3000"

# Main unified commands
start-all: start-exchange start-passport
	@echo ""
	@echo "All services started!"
	@echo "========================="
	@echo "Exchange Services:"
	@echo "  - Policy Decision Point: http://localhost:8082"
	@echo "  - Consent Engine: http://localhost:8081"
	@echo "  - Orchestration Engine: http://localhost:8080"
	@echo ""
	@echo "Sri Lanka Passport App:"
	@echo "  - Application: http://localhost:3000"
	@echo ""
	@echo "Use 'make status' to check service health"
	@echo "Use 'make logs' to view service logs"

stop-all: stop-exchange stop-passport
	@echo "🛑 All services stopped"

restart-all: stop-all start-all

status:
	@echo "Service Status"
	@echo "=============="
	@echo ""
	@echo "Exchange Services (Monolithic):"
	@cd exchange && docker compose ps 2>/dev/null || echo "Exchange services not running"
	@echo ""
	@echo "Sri Lanka Passport App:"
	@if pgrep -f "next dev" > /dev/null; then \
		echo "✅ Passport App (Port 3000) - Running"; \
	else \
		echo "❌ Passport App (Port 3000) - Not running"; \
	fi
	@echo ""
	@echo "Health Checks:"
	@curl -s http://localhost:8082/health > /dev/null && echo "✅ PDP (8082)" || echo "❌ PDP (8082)"
	@curl -s http://localhost:8081/health > /dev/null && echo "✅ CE (8081)" || echo "❌ CE (8081)"
	@curl -s http://localhost:8080/health > /dev/null && echo "✅ OE (8080)" || echo "❌ OE (8080)"
	@curl -s http://localhost:3000 > /dev/null && echo "✅ Passport App (3000)" || echo "❌ Passport App (3000)"

logs:
	@echo "Service Logs"
	@echo "============"
	@echo ""
	@echo "Exchange Services Logs:"
	@cd exchange && docker compose logs --tail=20
	@echo ""
	@echo "Passport App Logs:"
	@if pgrep -f "next dev" > /dev/null; then \
		echo "Passport app is running. Use 'make logs-passport' for detailed logs."; \
	else \
		echo "Passport app is not running."; \
	fi

clean:
	@echo "🧹 Cleaning up all services..."
	@cd exchange && docker compose down -v --remove-orphans
	@docker system prune -f
	@if pgrep -f "next dev" > /dev/null; then \
		pkill -f "next dev"; \
		echo "Stopped Passport app"; \
	fi
	@echo "✅ Cleanup complete"

# Exchange Services Management
start-exchange:
	@echo "Starting Exchange Services..."
	@cd exchange && make start-local
	@echo "✅ Exchange services started"

stop-exchange:
	@echo "Stopping Exchange Services..."
	@cd exchange && make stop
	@echo "✅ Exchange services stopped"

logs-exchange:
	@echo "Exchange Services Logs:"
	@cd exchange && make logs

# Sri Lanka Passport App Management
start-passport:
	@echo "Starting Sri Lanka Passport Application..."
	@cd sri-lanka-passport && npm run dev > /dev/null 2>&1 &
	@sleep 3
	@if curl -s http://localhost:3000 > /dev/null; then \
		echo "✅ Passport app started at http://localhost:3000"; \
	else \
		echo "⏳ Passport app starting... (may take a moment)"; \
		echo "Check status with: make status"; \
	fi

stop-passport:
	@echo "Stopping Sri Lanka Passport Application..."
	@if pgrep -f "next dev" > /dev/null; then \
		pkill -f "next dev"; \
		echo "✅ Passport app stopped"; \
	else \
		echo "Passport app was not running"; \
	fi

dev-passport:
	@echo "Starting Passport App in Development Mode..."
	@cd sri-lanka-passport && npm run dev

logs-passport:
	@echo "Passport App Logs:"
	@if pgrep -f "next dev" > /dev/null; then \
		echo "Passport app is running. Check terminal where it was started for logs."; \
		echo "Or run 'make dev-passport' to see logs in real-time."; \
	else \
		echo "Passport app is not running. Start it with 'make start-passport'"; \
	fi

# Build and Test Commands
build:
	@echo "Building all services..."
	@cd exchange && make build
	@cd sri-lanka-passport && npm install
	@echo "✅ All services built"

test:
	@echo "Running all tests..."
	@cd exchange && make test
	@echo "✅ Tests completed"

# Quick development setup
dev-setup:
	@echo "Setting up development environment..."
	@cd sri-lanka-passport && npm install
	@cd exchange && make build-local
	@echo "✅ Development environment ready"
	@echo "Run 'make start-all' to start all services"

# Production commands
start-prod:
	@echo "Starting all services in production mode..."
	@cd exchange && make start-prod
	@cd sri-lanka-passport && npm run build
	@cd sri-lanka-passport && npm start > /dev/null 2>&1 &
	@echo "✅ All services started in production mode"

# Individual service health checks
health-check:
	@echo "Performing health checks..."
	@echo "=========================="
	@echo ""
	@echo "Exchange Services:"
	@curl -s http://localhost:8082/health && echo " - PDP healthy" || echo " - PDP unhealthy"
	@curl -s http://localhost:8081/health && echo " - CE healthy" || echo " - CE unhealthy"
	@curl -s http://localhost:8080/health && echo " - OE healthy" || echo " - OE unhealthy"
	@echo ""
	@echo "Passport App:"
	@curl -s http://localhost:3000 > /dev/null && echo " - Passport App healthy" || echo " - Passport App unhealthy"
