# Multi-stage Dockerfile for Audit Service
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Build arguments with defaults
ARG BUILD_VERSION=latest
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching  
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./
COPY consumer/ ./consumer/
COPY handlers/ ./handlers/
COPY models/ ./models/
COPY redis/ ./redis/
COPY services/ ./services/

# Build the application with build info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/audit-service .

# Final stage
FROM alpine:3.17

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user with specific UID for WSO2 Choreo compliance
RUN adduser -D -s /bin/sh -u 10001 appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/audit-service /app/

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user with specific UID
USER 10001

# Expose port
EXPOSE 3001

# Set environment variables
ENV ENVIRONMENT=production

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

CMD ["./audit-service", "-env=${ENVIRONMENT:-production}", "-port=3001"]