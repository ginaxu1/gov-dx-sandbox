# OpenDIF - Comprehensive Makefile
# This Makefile provides standardized commands for all services in the repository

.PHONY: help install-hooks setup validate-build validate-test validate-docker-build check-lint run clean setup-all validate-build-all validate-test-all start-all stop-all status logs

# Default target
help:
	@echo "OpenDIF - Available Commands"
	@echo "=========================================="
	@echo ""
	@echo "Usage: make [COMMAND] [SERVICE]"
	@echo ""
	@echo "Services:"
	@echo "  Go Services:"
	@echo "    - portal-backend"
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
	@echo "  quality-check [SERVICE]        - Run all quality checks (format, lint, security)"
	@echo "  format [SERVICE]               - Format Go code with gofumpt and goimports"
	@echo "  lint [SERVICE]                 - Run lint checks (go vet, gofmt)"
	@echo "  security [SERVICE]             - Run security checks with gosec"
	@echo "  staticcheck [SERVICE]          - Run staticcheck analysis"
	@echo "  install-tools                  - Install all required Go quality tools"
	@echo "  run [SERVICE]                  - Run the service locally"
	@echo "  clean                          - Clean all build artifacts"
	@echo ""
	@echo "Batch Commands:"
	@echo "  setup-all                      - Setup all services"
	@echo "  validate-build-all             - Build all services"
	@echo "  validate-test-all              - Test all services"
	@echo "  quality-check-all              - Run quality checks on all Go services"
	@echo "  format-all                     - Format all Go services"
	@echo "  lint-all                       - Lint all Go services"
	@echo ""
	@echo "Runtime Commands (Legacy/Convenience):"
	@echo "  start-all                      - Start all services (5 services total)"
	@echo "  stop-all                       - Stop all services"
	@echo "  status                         - Check status of all services"
	@echo "  logs                           - View logs for all services"
	@echo ""
	@echo "Examples:"
	@echo "  make setup portal-backend"
	@echo "  make validate-build orchestration-engine"
	@echo "  make validate-test consent-engine"
	@echo "  make run member-portal"
	@echo "  make start-all"

# Variables
ROOT_DIR := $(shell pwd)
BIN_DIR := $(ROOT_DIR)/bin
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Service paths
portal_backend_PATH := portal-backend
AUDIT_SERVICE_PATH := audit-service
ORCHESTRATION_ENGINE_PATH := exchange/orchestration-engine
CONSENT_ENGINE_PATH := exchange/consent-engine
POLICY_DECISION_POINT_PATH := exchange/policy-decision-point

MEMBER_PORTAL_PATH := portals/member-portal
ADMIN_PORTAL_PATH := portals/admin-portal
CONSENT_PORTAL_PATH := portals/consent-portal

# Go services list
GO_SERVICES := portal-backend audit-service orchestration-engine consent-engine policy-decision-point
FRONTEND_SERVICES := member-portal admin-portal consent-portal

# =============================================================================
# ROUTER HELPER FUNCTIONS
# =============================================================================

# Common shell function to resolve service paths and route commands
# This consolidates the repeated pattern across all router targets

# Create bin directory
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# =============================================================================
# SETUP COMMANDS
# =============================================================================

# Install Git hooks
install-hooks:
	@echo "Installing git hooks..."
	@if [ ! -d ".githooks" ]; then \
		echo "❌ Error: .githooks directory not found"; \
		exit 1; \
	fi
	@mkdir -p .git/hooks
	@cp .githooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed successfully"
	@echo "📍 Pre-commit hook will now run automatically on every commit"
	@echo "💡 To bypass temporarily, use: git commit --no-verify"
	
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
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)"; TARGET="setup-go-service" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)"; TARGET="setup-go-service" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)"; TARGET="setup-go-service" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)"; TARGET="setup-go-service" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)"; TARGET="setup-go-service" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)"; TARGET="setup-frontend-service" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)"; TARGET="setup-frontend-service" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)"; TARGET="setup-frontend-service" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) $$TARGET SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

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
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)"; TARGET="validate-build-go-service" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)"; TARGET="validate-build-go-service" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)"; TARGET="validate-build-go-service" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)"; TARGET="validate-build-go-service" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)"; TARGET="validate-build-go-service" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)"; TARGET="validate-build-frontend-service" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)"; TARGET="validate-build-frontend-service" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)"; TARGET="validate-build-frontend-service" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) $$TARGET SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

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
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)"; TARGET="validate-test-go-service" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)"; TARGET="validate-test-go-service" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)"; TARGET="validate-test-go-service" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)"; TARGET="validate-test-go-service" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)"; TARGET="validate-test-go-service" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)"; TARGET="validate-test-frontend-service" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)"; TARGET="validate-test-frontend-service" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)"; TARGET="validate-test-frontend-service" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) $$TARGET SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

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
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) validate-docker-build-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# =============================================================================
# CODE QUALITY COMMANDS
# =============================================================================

# Install essential Go quality tools (minimal set)
install-tools:
	@echo "Installing essential Go quality tools..."
	@go install mvdan.cc/gofumpt@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "✅ Essential Go quality tools installed"
	@echo "💡 Configure your IDE (VS Code/GoLand) for real-time linting!"
	@echo "ℹ️  For security scanning, install gosec or use 'make security <service>'"

# Format Go code
format-go-service:
	@echo "Formatting Go service: $(SERVICE)"
	@if [ -f "$(SERVICE_PATH)/go.mod" ]; then \
		cd $(SERVICE_PATH) && go mod tidy; \
	else \
		echo "⚠️  No go.mod found in $(SERVICE_PATH), skipping go mod tidy"; \
	fi
	@echo "Running gofumpt..."
	@if command -v $$(go env GOPATH)/bin/gofumpt > /dev/null 2>&1; then \
		cd $(SERVICE_PATH) && $$(go env GOPATH)/bin/gofumpt -w .; \
	else \
		echo "⚠️  gofumpt not available, using gofmt"; \
		cd $(SERVICE_PATH) && gofmt -w .; \
	fi
	@echo "Running goimports..."
	@if command -v $$(go env GOPATH)/bin/goimports > /dev/null 2>&1; then \
		cd $(SERVICE_PATH) && $$(go env GOPATH)/bin/goimports -w .; \
	else \
		echo "⚠️  goimports not available, skipping import organization"; \
	fi
	@echo "✅ Code formatted for Go service $(SERVICE)"

# Run basic Go linting (using built-in tools)
lint-go-service:
	@echo "Running basic lint checks for Go service: $(SERVICE)"
	@if [ -f "$(SERVICE_PATH)/go.mod" ]; then \
		cd $(SERVICE_PATH) && go mod tidy; \
	else \
		echo "⚠️  No go.mod found in $(SERVICE_PATH), skipping go mod tidy"; \
	fi
	@echo "Running go vet..."
	@cd $(SERVICE_PATH) && go vet ./... || (echo "❌ go vet found issues in $(SERVICE)" && exit 1)
	@echo "Running gofmt check..."
	@cd $(SERVICE_PATH) && test -z "$$(gofmt -l .)" || (echo "❌ Code needs formatting. Run: make format $(SERVICE)" && gofmt -l . && exit 1)
	@echo "✅ Basic lint checks completed for Go service $(SERVICE)"

# Run staticcheck
staticcheck-go-service:
	@echo "Running staticcheck for Go service: $(SERVICE)"
	@if [ -f "$(SERVICE_PATH)/go.mod" ]; then \
		cd $(SERVICE_PATH) && go mod tidy; \
	else \
		echo "⚠️  No go.mod found in $(SERVICE_PATH), skipping go mod tidy"; \
	fi
	@if command -v $$(go env GOPATH)/bin/staticcheck > /dev/null 2>&1; then \
		cd $(SERVICE_PATH) && $$(go env GOPATH)/bin/staticcheck ./... || echo "⚠️  Staticcheck found issues in $(SERVICE) (non-blocking)"; \
	else \
		echo "ℹ️  staticcheck not installed, skipping analysis for $(SERVICE)"; \
	fi
	@echo "✅ Staticcheck completed for Go service $(SERVICE)"

# Run security checks with gosec (if available)
security-go-service:
	@echo "Running security checks for Go service: $(SERVICE)"
	@if [ -f "$(SERVICE_PATH)/go.mod" ]; then \
		cd $(SERVICE_PATH) && go mod tidy; \
	else \
		echo "⚠️  No go.mod found in $(SERVICE_PATH), skipping go mod tidy"; \
	fi
	@if command -v gosec > /dev/null 2>&1; then \
		cd $(SERVICE_PATH) && gosec -quiet ./... || echo "⚠️  Security issues found in $(SERVICE) (non-blocking)"; \
	else \
		echo "ℹ️  gosec not installed, skipping security scan for $(SERVICE)"; \
	fi
	@echo "✅ Security check completed for Go service $(SERVICE)"

# Comprehensive quality check for Go services
quality-check-go-service:
	@echo "Running comprehensive quality checks for Go service: $(SERVICE)"
	@$(MAKE) format-go-service SERVICE=$(SERVICE) SERVICE_PATH=$(SERVICE_PATH)
	@$(MAKE) lint-go-service SERVICE=$(SERVICE) SERVICE_PATH=$(SERVICE_PATH)
	@$(MAKE) staticcheck-go-service SERVICE=$(SERVICE) SERVICE_PATH=$(SERVICE_PATH)
	@$(MAKE) security-go-service SERVICE=$(SERVICE) SERVICE_PATH=$(SERVICE_PATH)
	@$(MAKE) validate-test-go-service SERVICE=$(SERVICE) SERVICE_PATH=$(SERVICE_PATH)
	@echo "✅ All quality checks passed for Go service $(SERVICE)"

# Legacy lint check for Go services (for backward compatibility)
check-lint-go-service:
	@echo "Running basic lint checks for Go service: $(SERVICE)"
	@cd $(SERVICE_PATH) && go mod tidy
	@echo "Running go fmt..."
	@OUTPUT=$$(cd $(SERVICE_PATH) && gofmt -l .); \
	if [ -n "$$OUTPUT" ]; then \
		echo "❌ Files need formatting. Run: make format $(SERVICE)"; \
		echo "$$OUTPUT"; \
		exit 1; \
	fi
	@echo "Running go vet..."
	@cd $(SERVICE_PATH) && go vet ./... || (echo "❌ go vet failed for $(SERVICE)" && exit 1)
	@echo "✅ Basic lint checks completed for Go service $(SERVICE)"

# Frontend service quality checks
check-lint-frontend-service:
	@echo "Running lint checks for Frontend service: $(SERVICE)"
	@cd $(SERVICE_PATH) && npm ci
	@cd $(SERVICE_PATH) && npm run lint || (echo "❌ Lint checks failed for $(SERVICE)" && exit 1)
	@echo "✅ Lint checks passed for Frontend service $(SERVICE)"

# =============================================================================
# QUALITY CHECK ROUTERS
# =============================================================================

# Format router
format:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) echo "❌ Unknown Go service: $$SERVICE_NAME"; echo "Available Go services: $(GO_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) format-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# Lint router
lint:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) echo "❌ Unknown Go service: $$SERVICE_NAME"; echo "Available Go services: $(GO_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) lint-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# Staticcheck router
staticcheck:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) echo "❌ Unknown Go service: $$SERVICE_NAME"; echo "Available Go services: $(GO_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) staticcheck-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# Security router
security:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) echo "❌ Unknown Go service: $$SERVICE_NAME"; echo "Available Go services: $(GO_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) security-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# Quality check router
quality-check:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)"; TARGET="quality-check-go-service" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)"; TARGET="quality-check-go-service" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)"; TARGET="quality-check-go-service" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)"; TARGET="quality-check-go-service" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)"; TARGET="quality-check-go-service" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)"; TARGET="check-lint-frontend-service" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)"; TARGET="check-lint-frontend-service" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)"; TARGET="check-lint-frontend-service" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) $$TARGET SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# Legacy lint check router (for backward compatibility)
check-lint:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)"; TARGET="check-lint-go-service" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)"; TARGET="check-lint-go-service" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)"; TARGET="check-lint-go-service" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)"; TARGET="check-lint-go-service" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)"; TARGET="check-lint-go-service" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)"; TARGET="check-lint-frontend-service" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)"; TARGET="check-lint-frontend-service" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)"; TARGET="check-lint-frontend-service" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) $$TARGET SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

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
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		portal-backend) SERVICE_PATH="$(portal_backend_PATH)"; TARGET="run-go-service" ;; \
		audit-service) SERVICE_PATH="$(AUDIT_SERVICE_PATH)"; TARGET="run-go-service" ;; \
		orchestration-engine) SERVICE_PATH="$(ORCHESTRATION_ENGINE_PATH)"; TARGET="run-go-service" ;; \
		consent-engine) SERVICE_PATH="$(CONSENT_ENGINE_PATH)"; TARGET="run-go-service" ;; \
		policy-decision-point) SERVICE_PATH="$(POLICY_DECISION_POINT_PATH)"; TARGET="run-go-service" ;; \
		member-portal) SERVICE_PATH="$(MEMBER_PORTAL_PATH)"; TARGET="run-frontend-service" ;; \
		admin-portal) SERVICE_PATH="$(ADMIN_PORTAL_PATH)"; TARGET="run-frontend-service" ;; \
		consent-portal) SERVICE_PATH="$(CONSENT_PORTAL_PATH)"; TARGET="run-frontend-service" ;; \
		*) echo "❌ Unknown service: $$SERVICE_NAME"; echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; exit 1 ;; \
	esac; \
	$(MAKE) $$TARGET SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH

# =============================================================================
# UTILITY COMMANDS
# =============================================================================

# Clean all build artifacts
clean:
	@echo "Cleaning all build artifacts..."
	@rm -rf $(BIN_DIR)
	@find . -name "coverage.out" -delete 2>/dev/null || true
	@find . -name "coverage.html" -delete 2>/dev/null || true
	@find . -type d -name "node_modules" | while read dir; do rm -rf "$$dir"; done 2>/dev/null || true
	@rm -rf portals/member-portal/dist portals/admin-portal/dist portals/consent-portal/dist 2>/dev/null || true
	@echo "✅ All build artifacts cleaned"

# Allow service names to be used as targets (ignore them)
$(GO_SERVICES) $(FRONTEND_SERVICES):
	@:

# =============================================================================
# BATCH OPERATIONS
# =============================================================================

# Setup all services
setup-all:
	@echo "Setting up all services..."
	@$(MAKE) install-hooks
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

# Quality check all Go services
quality-check-all:
	@echo "Running quality checks on all Go services..."
	@set -e; \
	for service in $(GO_SERVICES); do \
		echo "Quality checking $$service..."; \
		$(MAKE) quality-check $$service & \
	done; \
	wait

# Format all Go services
format-all:
	@echo "Formatting all Go services..."
	@for service in $(GO_SERVICES); do \
		echo "Formatting $$service..."; \
		$(MAKE) format $$service; \
	done
	@echo "✅ All Go services formatted"

# Lint all Go services
lint-all:
	@echo "Linting all Go services..."
	@for service in $(GO_SERVICES); do \
		echo "Linting $$service..."; \
		$(MAKE) lint $$service; \
	done
	@echo "✅ All Go services passed lint checks"

# =============================================================================
# RUNTIME COMMANDS (Legacy/Convenience)
# =============================================================================

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
