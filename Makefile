# Gov DX Sandbox - Unified Service Management
# Manages Exchange Services and API Server

.PHONY: help start-all stop-all status logs clean build

# Default target
help:
	@echo "Gov DX Sandbox - Service Management"
	@echo "===================================="
	@echo ""
	@echo "Commands:"
	@echo "  start-all      - Start all services (Exchange + API Server)"
	@echo "  stop-all       - Stop all services"
	@echo "  status         - Check status of all services"
	@echo "  logs           - View logs for all services"
	@echo "  clean          - Clean up all services and containers"
	@echo "  build          - Build all services"
	@echo ""
	@echo "Service Ports:"
	@echo "  Exchange Services:"
	@echo "    - Policy Decision Point (PDP): http://localhost:8082"
	@echo "    - Consent Engine (CE): http://localhost:8081"
	@echo "    - Orchestration Engine (OE): http://localhost:4000"
	@echo "  API Server: http://localhost:3000"

# Start all services
start-all:
	@echo "Starting all services..."
	@echo "========================="
	@echo ""
	
	# Start Exchange Services
	@echo "Starting Exchange Services..."
	@cd exchange && docker compose --env-file .env.local up --build -d
	@echo "✅ Exchange services started"
	@echo ""
	
	# Start API Server
	@echo "Starting API Server..."
	@cd api-server-go && go run main.go &
	@echo "✅ API Server started"
	@echo ""
	
	@echo "All services started!"
	@echo "====================="
	@echo "Exchange Services:"
	@echo "  - Policy Decision Point: http://localhost:8082"
	@echo "  - Consent Engine: http://localhost:8081"
	@echo "  - Orchestration Engine: http://localhost:4000"
	@echo "API Server: http://localhost:3000"
	@echo ""
	@echo "To view logs: make logs"
	@echo "To stop all: make stop-all"

# Stop all services
stop-all:
	@echo "Stopping all services..."
	@echo "========================"
	@echo ""
	
	# Stop Exchange Services
	@echo "Stopping Exchange Services..."
	@cd exchange && docker compose down
	@echo "✅ Exchange services stopped"
	@echo ""
	
	# Stop API Server
	@echo "Stopping API Server..."
	@pkill -f "go run main.go" || true
	@pkill -f "api-server-go" || true
	@echo "✅ API Server stopped"
	@echo ""
	
	@echo "All services stopped!"

# Check status of all services
status:
	@echo "Service Status"
	@echo "=============="
	@echo ""
	
	# Check Exchange Services
	@echo "Exchange Services:"
	@echo "  Policy Decision Point: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8082/health 2>/dev/null || echo 'DOWN')"
	@echo "  Consent Engine: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8081/health 2>/dev/null || echo 'DOWN')"
	@echo "  Orchestration Engine: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:4000/health 2>/dev/null || echo 'DOWN')"
	@echo ""
	
	# Check API Server
	@echo "API Server:"
	@echo "  API Server: $$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/health 2>/dev/null || echo 'DOWN')"
	@echo ""
	
	# Show running processes
	@echo "Running processes:"
	@ps aux | grep -E "(consent-engine|policy-decision-point|orchestration|api-server)" | grep -v grep || echo "No services running"

# View logs for all services
logs:
	@echo "Service Logs"
	@echo "============"
	@echo ""
	@echo "Exchange Services logs:"
	@cd exchange && docker compose logs --tail=50
	@echo ""
	@echo "API Server logs (if running):"
	@ps aux | grep -E "go run main.go|api-server" | grep -v grep || echo "API Server not running"

# Clean up all services and containers
clean:
	@echo "Cleaning up all services..."
	@echo "==========================="
	@echo ""
	
	# Clean Exchange Services
	@echo "Cleaning Exchange Services..."
	@cd exchange && docker compose down -v --remove-orphans
	@echo "✅ Exchange services cleaned"
	@echo ""
	
	# Clean API Server
	@echo "Cleaning API Server..."
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
	
	# Build Exchange Services
	@echo "Building Exchange Services..."
	@cd exchange && docker compose build
	@echo "✅ Exchange services built"
	@echo ""
	
	# Build API Server (if needed)
	@echo "Building API Server..."
	@cd api-server-go && go mod tidy
	@echo "✅ API Server built"
	@echo ""
	
	@echo "All services built!"