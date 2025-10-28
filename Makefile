# Gov DX Sandbox - Unified Service Management
# Manages Exchange Services and API Server

.PHONY: help start-all stop-all status logs clean build

# Default target
help:
	@echo "Gov DX Sandbox - Service Management"
	@echo "===================================="
	@echo ""
	@echo "Commands:"
	@echo "  make start-all  - Start all services (5 services total)"
	@echo "  make stop-all   - Stop all services"
	@echo "  make status     - Check status of all services"
	@echo "  make logs       - View logs for all services"
	@echo "  make clean      - Clean up all services and containers"
	@echo "  make build      - Build all services"
	@echo ""
	@echo "Service Ports:"
	@echo "  • API Server:        http://localhost:3000"
	@echo "  • Audit Service:     http://localhost:3001"
	@echo "  • Consent Engine:   http://localhost:8081"
	@echo "  • Policy Decision Point: http://localhost:8082"
	@echo "  • Orchestration Engine: http://localhost:4000"
	@echo ""
	@echo ""
	@echo "Prerequisites:"
	@echo "  • PostgreSQL database (optional but recommended)"
	@echo "  • All Go services should compile without errors"
	@echo ""
	@echo "Quick Start:"
	@echo "  make start-all && sleep 15 && make status"
	@echo ""
	@echo "To run integration tests:"
	@echo "  cd integration-tests && ./run-all-tests.sh"

# Start all services
start-all:
	@echo "Starting all services..."
	@echo "========================="
	@echo ""
	
	# Check for PostgreSQL
	@if ! pgrep -x postgres > /dev/null; then \
		echo "⚠️  WARNING: PostgreSQL database is not running"; \
		echo "   Some services require a PostgreSQL database."; \
		echo "   Services may fail to start without it."; \
		echo ""; \
	fi
	
	# Start Exchange Services (PDP, CE, OE via Go)
	@echo "Starting Policy Decision Point (PDP) on port 8082..."
	@cd exchange/policy-decision-point && nohup go run . > /tmp/pdp.log 2>&1 & echo $$! > /tmp/pdp.pid || echo "Failed to start PDP"
	@sleep 2
	
	@echo "Starting Consent Engine (CE) on port 8081..."
	@cd exchange/consent-engine && nohup go run . > /tmp/consent-engine.log 2>&1 & echo $$! > /tmp/consent-engine.pid || echo "Failed to start Consent Engine"
	@sleep 2
	
	@echo "Starting Orchestration Engine (OE) on port 4000..."
	@cd exchange/orchestration-engine-go && nohup go run . > /tmp/orchestration-engine.log 2>&1 & echo $$! > /tmp/orchestration-engine.pid || echo "Failed to start Orchestration Engine"
	@sleep 2
	
	@echo "Starting Audit Service on port 3001..."
	@cd audit-service && nohup go run . > /tmp/audit-service.log 2>&1 & echo $$! > /tmp/audit-service.pid || echo "Failed to start Audit Service"
	@sleep 2
	
	@echo "Starting API Server on port 3000..."
	@cd api-server-go && nohup go run . > /tmp/api-server.log 2>&1 & echo $$! > /tmp/api-server.pid || echo "Failed to start API Server"
	@sleep 2
	
	@echo ""
	@echo "Waiting for services to be ready (10 seconds)..."
	@sleep 10
	
	@echo ""
	@echo "Checking service health..."
	@make status
	
	@echo ""
	@echo "✅ All services started!"
	@echo "====================="
	@echo "Service URLs:"
	@echo "  - API Server: http://localhost:3000"
	@echo "  - Audit Service: http://localhost:3001"
	@echo "  - Policy Decision Point: http://localhost:8082"
	@echo "  - Consent Engine: http://localhost:8081"
	@echo "  - Orchestration Engine: http://localhost:4000"
	@echo ""
	@echo "To view logs: make logs"
	@echo "To stop all: make stop-all"
	@echo ""
	@echo "Process IDs saved to /tmp/*.pid"

# Stop all services
stop-all:
	@echo "Stopping all services..."
	@echo "========================"
	@echo ""
	
	# Stop services using PID files
	@if [ -f /tmp/pdp.pid ]; then \
		echo "Stopping Policy Decision Point (PDP)..."; \
		kill $$(cat /tmp/pdp.pid) 2>/dev/null || true; \
		rm -f /tmp/pdp.pid; \
	fi
	
	@if [ -f /tmp/consent-engine.pid ]; then \
		echo "Stopping Consent Engine (CE)..."; \
		kill $$(cat /tmp/consent-engine.pid) 2>/dev/null || true; \
		rm -f /tmp/consent-engine.pid; \
	fi
	
	@if [ -f /tmp/orchestration-engine.pid ]; then \
		echo "Stopping Orchestration Engine (OE)..."; \
		kill $$(cat /tmp/orchestration-engine.pid) 2>/dev/null || true; \
		rm -f /tmp/orchestration-engine.pid; \
	fi
	
	@if [ -f /tmp/audit-service.pid ]; then \
		echo "Stopping Audit Service..."; \
		kill $$(cat /tmp/audit-service.pid) 2>/dev/null || true; \
		rm -f /tmp/audit-service.pid; \
	fi
	
	@if [ -f /tmp/api-server.pid ]; then \
		echo "Stopping API Server..."; \
		kill $$(cat /tmp/api-server.pid) 2>/dev/null || true; \
		rm -f /tmp/api-server.pid; \
	fi
	
	@echo ""
	@echo "Cleaning up stray processes..."
	@pkill -f "policy-decision-point" || true
	@pkill -f "consent-engine" || true
	@pkill -f "orchestration-engine-go" || true
	@pkill -f "audit-service" || true
	@pkill -f "api-server-go" || true
	
	@echo ""
	@echo "✅ All services stopped!"

# Check status of all services
status:
	@echo "Service Status"
	@echo "=============="
	@echo ""
	
	# Check all services
	@echo "📡 API Server (3000):        $$([ "$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/health 2>/dev/null)" = "200" ] && echo '✅ UP' || echo '❌ DOWN')"
	@echo "📡 Audit Service (3001):     $$([ "$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3001/health 2>/dev/null)" = "200" ] && echo '✅ UP' || echo '❌ DOWN')"
	@echo "📡 Consent Engine (8081):    $$([ "$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8081/health 2>/dev/null)" = "200" ] && echo '✅ UP' || echo '❌ DOWN')"
	@echo "📡 Policy Decision Point (8082): $$([ "$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8082/health 2>/dev/null)" = "200" ] && echo '✅ UP' || echo '❌ DOWN')"
	@echo "📡 Orchestration Engine (4000): $$([ "$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:4000/health 2>/dev/null)" = "200" ] && echo '✅ UP' || echo '❌ DOWN')"
	@echo ""

# View logs for all services
logs:
	@echo "Service Logs"
	@echo "============"
	@echo ""
	@if [ -f /tmp/api-server.log ]; then echo "=== API Server ===" && tail -20 /tmp/api-server.log; fi
	@if [ -f /tmp/audit-service.log ]; then echo "=== Audit Service ===" && tail -20 /tmp/audit-service.log; fi
	@if [ -f /tmp/consent-engine.log ]; then echo "=== Consent Engine ===" && tail -20 /tmp/consent-engine.log; fi
	@if [ -f /tmp/pdp.log ]; then echo "=== Policy Decision Point ===" && tail -20 /tmp/pdp.log; fi
	@if [ -f /tmp/orchestration-engine.log ]; then echo "=== Orchestration Engine ===" && tail -20 /tmp/orchestration-engine.log; fi

# Clean up all services and containers
clean:
	@echo "Cleaning up all services..."
	@echo "==========================="
	@echo ""
	
	# Clean Exchange Services (Docker)
	@echo "Cleaning Exchange Services (Docker)..."
	@cd exchange && docker compose down -v --remove-orphans
	@echo "✅ Exchange services cleaned"
	@echo ""
	
	# Clean API Server (Go)
	@echo "Cleaning API Server (Go)..."
	@pkill -f "go run main.go" || true
	@pkill -f "api-server-go" || true
	@echo "✅ API Server cleaned"
	@echo ""
	
	# Clean Docker system
	@echo "Cleaning Docker system..."
	@docker system prune -f
	@echo "✅ Docker system cleaned"
	@echo ""
	
	@echo "All services cleaned!"

# Build all services
build:
	@echo "Building all services..."
	@echo "========================"
	@echo ""
	
	# Build Exchange Services (Docker)
	@echo "Building Exchange Services (Docker)..."
	@cd exchange && docker compose build
	@echo "✅ Exchange services built"
	@echo ""
	
	# Build API Server (Go)
	@echo "Building API Server (Go)..."
	@cd api-server-go && go mod tidy
	@echo "✅ API Server built"
	@echo ""
	
	@echo "All services built!"