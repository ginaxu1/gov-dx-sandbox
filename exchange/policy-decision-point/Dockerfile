# Multi-stage Dockerfile for Policy Decision Point
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Build arguments
ARG SERVICE_PATH=policy-decision-point
ARG BUILD_VERSION=latest
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files and source code
COPY policy-decision-point/go.mod policy-decision-point/go.sum ./
COPY policy-decision-point/ ./

# Copy shared packages - for Choreo builds, shared packages should be copied into the service directory
COPY shared/ /app/shared/

# Download dependencies
RUN go mod download

# Build the application with build info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/service_binary .

# Final stage
FROM alpine:3.17

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user with specific UID for WSO2 Choreo compliance
RUN adduser -D -s /bin/sh -u 10001 appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/service_binary /app/

# Copy service-specific files (policies and data)
COPY --from=builder /app/policies/ /app/policies/
COPY --from=builder /app/data/ /app/data/

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user with specific UID
USER 10001

# Expose port
EXPOSE 8082

# Set environment variables
ENV CONFIG_DIR=/app/config
ENV ENVIRONMENT=production

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Run the application with environment-based flags
CMD ["./service_binary", "-env=${ENVIRONMENT:-production}", "-port=8082"]