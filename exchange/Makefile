# Exchange Services Makefile

.PHONY: start-local start-prod stop logs test clean build help

# Default to local environment
start: start-local

start-local:
	@./scripts/start-local.sh

start-prod:
	@./scripts/start-production.sh

stop:
	@./scripts/stop.sh

logs:
	@./scripts/logs.sh

test:
	@./scripts/test.sh

clean:
	@docker compose down -v --remove-orphans
	@docker system prune -f

build:
	@docker compose build

# Environment switching
switch-local:
	@echo "Switching to local environment..."
	@docker compose --env-file .env.local up --build -d

switch-prod:
	@echo "Switching to production environment..."
	@docker compose --env-file .env.production up --build -d

# Direct Docker Compose commands
local:
	@docker compose --env-file .env.local up --build

prod:
	@docker compose --env-file .env.production up --build

help:
	@echo "Available commands:"
	@echo "  make start        - Start services (local environment)"
	@echo "  make start-local  - Start services in local environment"
	@echo "  make start-prod   - Start services in production environment"
	@echo "  make stop         - Stop all services"
	@echo "  make logs         - View logs"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean up containers and volumes"
	@echo "  make build        - Build services"
	@echo "  make switch-local - Switch to local environment"
	@echo "  make switch-prod  - Switch to production environment"
	@echo "  make local        - Direct: docker compose --env-file .env.local up --build"
	@echo "  make prod         - Direct: docker compose --env-file .env.production up --build"