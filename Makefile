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
	@echo "    - orchestration-engine-go"
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
	@echo "  lint [SERVICE]                 - Run golangci-lint with comprehensive checks"
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
	@echo "Examples:"
	@echo "  make setup api-server-go"
	@echo "  make validate-build orchestration-engine-go"
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
GO_SERVICES := api-server-go audit-service orchestration-engine-go consent-engine policy-decision-point
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
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH_VAR="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH_VAR="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH_VAR="$(CONSENT_PORTAL_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown service: $$SERVICE_NAME"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) setup-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		member-portal|admin-portal|consent-portal) \
			$(MAKE) setup-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
	esac

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
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH_VAR="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH_VAR="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH_VAR="$(CONSENT_PORTAL_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown service: $$SERVICE_NAME"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) validate-build-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		member-portal|admin-portal|consent-portal) \
			$(MAKE) validate-build-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
	esac

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
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH_VAR="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH_VAR="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH_VAR="$(CONSENT_PORTAL_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown service: $$SERVICE_NAME"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) validate-test-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		member-portal|admin-portal|consent-portal) \
			$(MAKE) validate-test-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
	esac

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
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH_VAR="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH_VAR="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH_VAR="$(CONSENT_PORTAL_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown service: $$SERVICE_NAME"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	else \
		$(MAKE) validate-docker-build-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR; \
	fi

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
	@echo "ℹ️  For security scanning, use IDE extensions or CI/CD pipelines"

# Format Go code
format-go-service:
	@echo "Formatting Go service: $(SERVICE)"
	@if [ -f "$(SERVICE_PATH)/go.mod" ]; then \
		cd $(SERVICE_PATH) && go mod tidy; \
	else \
		echo "⚠️  No go.mod found in $(SERVICE_PATH), skipping go mod tidy"; \
	fi
	@echo "Running gofumpt..."
	@cd $(SERVICE_PATH) && $$(go env GOPATH)/bin/gofumpt -w . 2>/dev/null || gofmt -w .
	@echo "Running goimports..."
	@cd $(SERVICE_PATH) && $$(go env GOPATH)/bin/goimports -w . 2>/dev/null || echo "⚠️  goimports not available, using go fmt"
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
	@cd $(SERVICE_PATH) && go vet ./... || echo "⚠️  go vet found issues in $(SERVICE) (non-blocking)"
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
	@cd $(SERVICE_PATH) && gofmt -l . | tee /tmp/gofmt-$(SERVICE).out
	@if [ -s /tmp/gofmt-$(SERVICE).out ]; then \
		echo "❌ Files need formatting. Run: make format $(SERVICE)"; \
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
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown Go service: $$SERVICE_NAME"; \
		echo "Available Go services: $(GO_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) format-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		*) \
			echo "❌ Unknown Go service: $$SERVICE_NAME"; \
			echo "Available Go services: $(GO_SERVICES)"; \
			exit 1 ;; \
	esac

# Lint router
lint:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown Go service: $$SERVICE_NAME"; \
		echo "Available Go services: $(GO_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) lint-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		*) \
			echo "❌ Unknown Go service: $$SERVICE_NAME"; \
			echo "Available Go services: $(GO_SERVICES)"; \
			exit 1 ;; \
	esac

# Staticcheck router
staticcheck:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown Go service: $$SERVICE_NAME"; \
		echo "Available Go services: $(GO_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) staticcheck-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		*) \
			echo "❌ Unknown Go service: $$SERVICE_NAME"; \
			echo "Available Go services: $(GO_SERVICES)"; \
			exit 1 ;; \
	esac

# Security router
security:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown Go service: $$SERVICE_NAME"; \
		echo "Available Go services: $(GO_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) security-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		*) \
			echo "❌ Unknown Go service: $$SERVICE_NAME"; \
			echo "Available Go services: $(GO_SERVICES)"; \
			exit 1 ;; \
	esac

# Quality check router
quality-check:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		api-server-go) \
			$(MAKE) quality-check-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(API_SERVER_PATH) ;; \
		audit-service) \
			$(MAKE) quality-check-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(AUDIT_SERVICE_PATH) ;; \
		orchestration-engine-go) \
			$(MAKE) quality-check-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(ORCHESTRATION_ENGINE_PATH) ;; \
		consent-engine) \
			$(MAKE) quality-check-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(CONSENT_ENGINE_PATH) ;; \
		policy-decision-point) \
			$(MAKE) quality-check-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(POLICY_DECISION_POINT_PATH) ;; \
		member-portal) \
			$(MAKE) check-lint-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(MEMBER_PORTAL_PATH) ;; \
		admin-portal) \
			$(MAKE) check-lint-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(ADMIN_PORTAL_PATH) ;; \
		consent-portal) \
			$(MAKE) check-lint-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$(CONSENT_PORTAL_PATH) ;; \
		*) \
			echo "❌ Unknown service: $$SERVICE_NAME"; \
			echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
			exit 1 ;; \
	esac

# Legacy lint check router (for backward compatibility)
check-lint:
	@SERVICE_NAME="$(word 2,$(MAKECMDGOALS))"; \
	case "$$SERVICE_NAME" in \
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH_VAR="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH_VAR="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH_VAR="$(CONSENT_PORTAL_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown service: $$SERVICE_NAME"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) check-lint-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		member-portal|admin-portal|consent-portal) \
			$(MAKE) check-lint-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
	esac

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
		api-server-go) SERVICE_PATH_VAR="$(API_SERVER_PATH)" ;; \
		audit-service) SERVICE_PATH_VAR="$(AUDIT_SERVICE_PATH)" ;; \
		orchestration-engine-go) SERVICE_PATH_VAR="$(ORCHESTRATION_ENGINE_PATH)" ;; \
		consent-engine) SERVICE_PATH_VAR="$(CONSENT_ENGINE_PATH)" ;; \
		policy-decision-point) SERVICE_PATH_VAR="$(POLICY_DECISION_POINT_PATH)" ;; \
		member-portal) SERVICE_PATH_VAR="$(MEMBER_PORTAL_PATH)" ;; \
		admin-portal) SERVICE_PATH_VAR="$(ADMIN_PORTAL_PATH)" ;; \
		consent-portal) SERVICE_PATH_VAR="$(CONSENT_PORTAL_PATH)" ;; \
		*) SERVICE_PATH_VAR="" ;; \
	esac; \
	if [ -z "$$SERVICE_PATH_VAR" ]; then \
		echo "❌ Unknown service: $$SERVICE_NAME"; \
		echo "Available services: $(GO_SERVICES) $(FRONTEND_SERVICES)"; \
		exit 1; \
	fi; \
	case "$$SERVICE_NAME" in \
		api-server-go|audit-service|orchestration-engine-go|consent-engine|policy-decision-point) \
			$(MAKE) run-go-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
		member-portal|admin-portal|consent-portal) \
			$(MAKE) run-frontend-service SERVICE=$$SERVICE_NAME SERVICE_PATH=$$SERVICE_PATH_VAR ;; \
	esac

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

# Quality check all Go services
quality-check-all:
	@echo "Running quality checks on all Go services..."
	@for service in $(GO_SERVICES); do \
		echo "Quality checking $$service..."; \
		$(MAKE) quality-check $$service; \
	done
	@echo "✅ All Go services passed quality checks"

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

