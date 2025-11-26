# OpenDIF - Comprehensive Makefile
# This Makefile provides standardized commands for all services in the repository

.PHONY: help setup validate-build validate-test validate-docker-build check-lint run clean setup-all validate-build-all validate-test-all

# Default target
help:
	@echo "OpenDIF - Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "Usage: make [COMMAND] [SERVICE]"
	@echo ""
	@echo "Services:"
	@echo "  Go Services:"
	@echo "    - api-server-go"
	@echo "    - audit-service" 
	@echo "    - orchestration-engine"
	@echo "    - consent-engine"
	@echo "    - policy-decision-point"
	@echo ""
	@echo "  Frontend Services:"
	@echo "    - member-portal"
	@echo "    - admin-portal"
	@echo "    - consent-portal"
	@echo ""
	@echo "Commands:"
	@echo "  setup [SERVICE]                - Install required modules/dependencies"
	@echo "  validate-build [SERVICE]       - Build the service and validate"
	@echo "  validate-test [SERVICE]        - Run unit tests with coverage"
	@echo "  validate-docker-build [SERVICE] - Validate Docker image builds"
	@echo "  check-lint [SERVICE]           - Run lint checks"
	@echo "  run [SERVICE]                  - Run the service locally"
	@echo "  clean                          - Clean all build artifacts"
	@echo ""
	@echo "Batch Commands:"
	@echo "  setup-all                      - Setup all services"
	@echo "  validate-build-all             - Build all services"
	@echo "  validate-test-all              - Test all services"
	@echo ""
	@echo "Examples:"
	@echo "  make setup api-server-go"
	@echo "  make validate-build orchestration-engine"
	@echo "  make validate-test consent-engine"
	@echo "  make run member-portal"

# Variables
ROOT_DIR := $(shell pwd)
BIN_DIR := $(ROOT_DIR)/bin
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Service paths
API_SERVER_PATH := api-server-go
AUDIT_SERVICE_PATH := audit-service
ORCHESTRATION_ENGINE_PATH := exchange/orchestration-engine-go
CONSENT_ENGINE_PATH := exchange/consent-engine
POLICY_DECISION_POINT_PATH := exchange/policy-decision-point

MEMBER_PORTAL_PATH := portals/member-portal
ADMIN_PORTAL_PATH := portals/admin-portal
CONSENT_PORTAL_PATH := portals/consent-portal

# Go services list
GO_SERVICES := api-server-go audit-service orchestration-engine consent-engine policy-decision-point
FRONTEND_SERVICES := member-portal admin-portal consent-portal

# Create bin directory
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# =============================================================================
# SETUP COMMANDS
# =============================================================================

# Setup for Go services
setup-go-service:
	@echo "Setting up Go service: $(SERVICE)"
	@cd $(SERVICE_PATH) && go mod tidy
	@cd $(SERVICE_PATH) && go mod download
	@echo "✅ Go service $(SERVICE) dependencies installed"

# Setup for Frontend services
setup-frontend-service:
	@echo "Setting up Frontend service: $(SERVICE)"
	@cd $(SERVICE_PATH) && npm ci
	@echo "✅ Frontend service $(SERVICE) dependencies installed"

# Setup command router
setup:
	@if [ "$(filter $(word 2,$(MAKECMDGOALS)),$(GO_SERVICES))" ]; then \
		$(MAKE) setup-go-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	elif [ "$(filter $(word 2,$(MAKECMDGOALS)),$(FRONTEND_SERVICES))" ]; then \
		$(MAKE) setup-frontend-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	else \
		echo "❌ Unknown service: $(word 2,$(MAKECMDGOALS))"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi

# =============================================================================
# BUILD VALIDATION COMMANDS
# =============================================================================

# Validate build for Go services
validate-build-go-service: $(BIN_DIR)
	@echo "Building Go service: $(SERVICE)"
	@cd $(SERVICE_PATH) && go mod tidy
	@cd $(SERVICE_PATH) && CGO_ENABLED=0 go build \
		-ldflags="-w -s -X main.Version=dev -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
		-o $(BIN_DIR)/$(SERVICE) .
	@echo "✅ Go service $(SERVICE) built successfully -> $(BIN_DIR)/$(SERVICE)"

# Validate build for Frontend services
validate-build-frontend-service:
	@echo "Building Frontend service: $(SERVICE)"
	@cd $(SERVICE_PATH) && npm ci
	@cd $(SERVICE_PATH) && npm run build
	@echo "✅ Frontend service $(SERVICE) built successfully -> $(SERVICE_PATH)/dist/"

# Build validation router
validate-build:
	@if [ "$(filter $(word 2,$(MAKECMDGOALS)),$(GO_SERVICES))" ]; then \
		$(MAKE) validate-build-go-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	elif [ "$(filter $(word 2,$(MAKECMDGOALS)),$(FRONTEND_SERVICES))" ]; then \
		$(MAKE) validate-build-frontend-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	else \
		echo "❌ Unknown service: $(word 2,$(MAKECMDGOALS))"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi

# =============================================================================
# TEST VALIDATION COMMANDS
# =============================================================================

# Validate tests for Go services
validate-test-go-service:
	@echo "Running tests for Go service: $(SERVICE)"
	@cd $(SERVICE_PATH) && go mod tidy
	@echo "Running unit tests with coverage..."
	@cd $(SERVICE_PATH) && go test -v -race -coverprofile=coverage.out -covermode=atomic ./... || (echo "❌ Tests failed for $(SERVICE)" && exit 1)
	@cd $(SERVICE_PATH) && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: $(SERVICE_PATH)/coverage.html"
	@cd $(SERVICE_PATH) && go tool cover -func=coverage.out | tail -1
	@echo "✅ Tests passed for Go service $(SERVICE)"

# Validate tests for Frontend services (lint + type check as test equivalent)
validate-test-frontend-service:
	@echo "Running tests for Frontend service: $(SERVICE)"
	@cd $(SERVICE_PATH) && npm ci
	@echo "Running TypeScript compilation check..."
	@cd $(SERVICE_PATH) && npx tsc --noEmit || (echo "❌ TypeScript compilation failed for $(SERVICE)" && exit 1)
	@echo "Running lint checks..."
	@cd $(SERVICE_PATH) && npm run lint || (echo "❌ Lint checks failed for $(SERVICE)" && exit 1)
	@echo "✅ Tests passed for Frontend service $(SERVICE)"

# Test validation router
validate-test:
	@if [ "$(filter $(word 2,$(MAKECMDGOALS)),$(GO_SERVICES))" ]; then \
		$(MAKE) validate-test-go-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	elif [ "$(filter $(word 2,$(MAKECMDGOALS)),$(FRONTEND_SERVICES))" ]; then \
		$(MAKE) validate-test-frontend-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	else \
		echo "❌ Unknown service: $(word 2,$(MAKECMDGOALS))"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi

# =============================================================================
# DOCKER BUILD VALIDATION COMMANDS
# =============================================================================

# Validate Docker build
validate-docker-build-service:
	@echo "Validating Docker build for service: $(SERVICE)"
	@cd $(SERVICE_PATH) && docker build -t $(SERVICE):test \
		--build-arg BUILD_VERSION=test \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		. || (echo "❌ Docker build failed for $(SERVICE)" && exit 1)
	@echo "✅ Docker build successful for $(SERVICE)"
	@docker rmi $(SERVICE):test 2>/dev/null || true

# Docker validation router
validate-docker-build:
	@if [ "$(filter $(word 2,$(MAKECMDGOALS)),$(GO_SERVICES) $(FRONTEND_SERVICES))" ]; then \
		$(MAKE) validate-docker-build-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	else \
		echo "❌ Unknown service: $(word 2,$(MAKECMDGOALS))"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi

# =============================================================================
# LINT CHECK COMMANDS
# =============================================================================

# Lint check for Go services
check-lint-go-service:
	@echo "Running lint checks for Go service: $(SERVICE)"
	@cd $(SERVICE_PATH) && go mod tidy
	@echo "Running go fmt..."
	@cd $(SERVICE_PATH) && gofmt -l . | tee /tmp/gofmt-$(SERVICE).out
	@if [ -s /tmp/gofmt-$(SERVICE).out ]; then \
		echo "❌ Files need formatting. Run: cd $(SERVICE_PATH) && gofmt -w ."; \
		exit 1; \
	fi
	@echo "Running go vet..."
	@cd $(SERVICE_PATH) && go vet ./... || (echo "❌ go vet failed for $(SERVICE)" && exit 1)
	@echo "Checking for golint..."
	@which golint > /dev/null 2>&1 || (echo "Installing golint..." && go install golang.org/x/lint/golint@latest)
	@echo "Running golint..."
	@cd $(SERVICE_PATH) && golint ./... | tee /tmp/golint-$(SERVICE).out
	@if [ -s /tmp/golint-$(SERVICE).out ]; then \
		echo "⚠️  Lint suggestions found (not blocking)"; \
	fi
	@echo "✅ Lint checks completed for Go service $(SERVICE)"

# Lint check for Frontend services
check-lint-frontend-service:
	@echo "Running lint checks for Frontend service: $(SERVICE)"
	@cd $(SERVICE_PATH) && npm ci
	@cd $(SERVICE_PATH) && npm run lint || (echo "❌ Lint checks failed for $(SERVICE)" && exit 1)
	@echo "✅ Lint checks passed for Frontend service $(SERVICE)"

# Lint check router
check-lint:
	@if [ "$(filter $(word 2,$(MAKECMDGOALS)),$(GO_SERVICES))" ]; then \
		$(MAKE) check-lint-go-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	elif [ "$(filter $(word 2,$(MAKECMDGOALS)),$(FRONTEND_SERVICES))" ]; then \
		$(MAKE) check-lint-frontend-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	else \
		echo "❌ Unknown service: $(word 2,$(MAKECMDGOALS))"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi

# =============================================================================
# RUN COMMANDS
# =============================================================================

# Run Go services
run-go-service:
	@echo "Running Go service: $(SERVICE)"
	@echo "Service will run in foreground. Press Ctrl+C to stop."
	@cd $(SERVICE_PATH) && go run .

# Run Frontend services
run-frontend-service:
	@echo "Running Frontend service: $(SERVICE)"
	@echo "Service will run in foreground. Press Ctrl+C to stop."
	@cd $(SERVICE_PATH) && npm run dev

# Run router
run:
	@if [ "$(filter $(word 2,$(MAKECMDGOALS)),$(GO_SERVICES))" ]; then \
		$(MAKE) run-go-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	elif [ "$(filter $(word 2,$(MAKECMDGOALS)),$(FRONTEND_SERVICES))" ]; then \
		$(MAKE) run-frontend-service SERVICE=$(word 2,$(MAKECMDGOALS)) SERVICE_PATH=$(call get-service-path,$(word 2,$(MAKECMDGOALS))); \
	else \
		echo "❌ Unknown service: $(word 2,$(MAKECMDGOALS))"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi

# =============================================================================
# UTILITY COMMANDS
# =============================================================================

# Clean all build artifacts
clean:
	@echo "Cleaning all build artifacts..."
	@rm -rf $(BIN_DIR)
	@find . -name "coverage.out" -delete 2>/dev/null || true
	@find . -name "coverage.html" -delete 2>/dev/null || true
	@find . -name "node_modules" -type d -exec rm -rf {} + 2>/dev/null || true
	@find . -name "dist" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "✅ All build artifacts cleaned"

# Helper function to get service path
define get-service-path
$(if $(filter $1,api-server-go),$(API_SERVER_PATH),\
$(if $(filter $1,audit-service),$(AUDIT_SERVICE_PATH),\
$(if $(filter $1,orchestration-engine),$(ORCHESTRATION_ENGINE_PATH),\
$(if $(filter $1,consent-engine),$(CONSENT_ENGINE_PATH),\
$(if $(filter $1,policy-decision-point),$(POLICY_DECISION_POINT_PATH),\
$(if $(filter $1,member-portal),$(MEMBER_PORTAL_PATH),\
$(if $(filter $1,admin-portal),$(ADMIN_PORTAL_PATH),\
$(if $(filter $1,consent-portal),$(CONSENT_PORTAL_PATH),))))))))
endef

# Allow service names to be used as targets (ignore them)
$(GO_SERVICES) $(FRONTEND_SERVICES):
	@:

# =============================================================================
# BATCH OPERATIONS
# =============================================================================

# Setup all services
setup-all:
	@echo "Setting up all services..."
	@for service in $(GO_SERVICES); do \
		echo "Setting up $$service..."; \
		$(MAKE) setup $$service; \
	done
	@for service in $(FRONTEND_SERVICES); do \
		echo "Setting up $$service..."; \
		$(MAKE) setup $$service; \
	done
	@echo "✅ All services setup complete"

# Build all services
validate-build-all:
	@echo "Building all services..."
	@for service in $(GO_SERVICES); do \
		echo "Building $$service..."; \
		$(MAKE) validate-build $$service; \
	done
	@for service in $(FRONTEND_SERVICES); do \
		echo "Building $$service..."; \
		$(MAKE) validate-build $$service; \
	done
	@echo "✅ All services built successfully"

# Test all services
validate-test-all:
	@echo "Testing all services..."
	@for service in $(GO_SERVICES); do \
		echo "Testing $$service..."; \
		$(MAKE) validate-test $$service; \
	done
	@for service in $(FRONTEND_SERVICES); do \
		echo "Testing $$service..."; \
		$(MAKE) validate-test $$service; \
	done
	@echo "✅ All services tested successfully"

