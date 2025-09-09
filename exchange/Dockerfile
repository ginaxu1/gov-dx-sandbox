# Multi-stage Dockerfile for Go services
# Supports building both consent-engine (port 8081) and policy-decision-point (port 8082)
FROM golang:1.24-alpine AS builder

WORKDIR /app
ARG SERVICE_PATH
ARG BUILD_VERSION=latest

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION}" \
    -o /app/service_binary ./$SERVICE_PATH

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/service_binary /app/

# Copy configuration files
COPY --from=builder /app/config/ /app/config/

# Copy service-specific files (only for policy-decision-point)
ARG SERVICE_PATH
RUN if [ "$SERVICE_PATH" = "policy-decision-point" ]; then \
        mkdir -p /app/policies /app/data && \
        cp -r /app/$SERVICE_PATH/policies/* /app/policies/ 2>/dev/null || true && \
        cp -r /app/$SERVICE_PATH/data/* /app/data/ 2>/dev/null || true; \
    fi

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (will be overridden by service-specific ports)
EXPOSE 8080

# Health check (port will be determined by service configuration)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Set environment variables
ENV CONFIG_DIR=/app/config
ENV ENVIRONMENT=production

# Run the application with environment-based flags
CMD ["./service_binary", "-env=${ENVIRONMENT:-production}", "-port=${PORT:-8080}"]