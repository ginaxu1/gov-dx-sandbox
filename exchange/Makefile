# Exchange Services Makefile

.PHONY: start-local start-prod stop logs test status clean build help

# Default to local environment
start: start-local

# Service management (using consolidated manage.sh)
start-local:
	@./scripts/manage.sh start-local

start-prod:
	@./scripts/manage.sh start-prod

stop:
	@./scripts/manage.sh stop

# Monitoring and testing
logs:
	@./scripts/manage.sh logs

test:
	@./scripts/test.sh

status:
	@./scripts/manage.sh status

# Maintenance
clean:
	@docker compose down -v --remove-orphans
	@docker system prune -f

build:
	@echo "Preparing for Docker build..."
	@./scripts/prepare-docker-build.sh
	@docker compose build
	@echo "Restoring for local development..."
	@./scripts/restore-local-build.sh

build-local:
	@echo "Building for local development..."
	@./scripts/restore-local-build.sh
	@docker compose build

build-prod:
	@echo "Building for production deployment..."
	@./scripts/prepare-docker-build.sh
	@docker compose --env-file .env.production build
	@echo "Restoring for local development..."
	@./scripts/restore-local-build.sh

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

# Help
help:
	@./scripts/manage.sh help