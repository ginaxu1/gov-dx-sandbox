# Monolithic Dockerfile for Exchange Services
# Builds all three services: Policy Decision Point, Consent Engine, and Orchestration Engine

FROM golang:1.25-alpine AS builder

WORKDIR /app

# Build arguments
ARG BUILD_VERSION=latest
ARG BUILD_TIME
ARG GIT_COMMIT

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy shared packages first
COPY shared/ /app/shared/

# Copy all service source code
COPY policy-decision-point/ /app/policy-decision-point/
COPY consent-engine/ /app/consent-engine/
COPY orchestration-engine-go/ /app/orchestration-engine-go/

# Copy shared modules to each service directory for local module resolution
RUN cp -r /app/shared /app/policy-decision-point/ && \
    cp -r /app/shared /app/consent-engine/ && \
    cp -r /app/shared /app/orchestration-engine-go/

# Set working directory for each service and build
WORKDIR /app/policy-decision-point
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/bin/pdp .

WORKDIR /app/consent-engine
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/bin/ce .

WORKDIR /app/orchestration-engine-go
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/bin/oe .

# Final stage
FROM alpine:3.17

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user with specific UID for WSO2 Choreo compliance
RUN adduser -D -s /bin/sh -u 10001 appuser

# Set working directory
WORKDIR /app

# Copy all binaries
COPY --from=builder /app/bin/ /app/bin/

# Copy service-specific files
COPY --from=builder /app/policy-decision-point/policies/ /app/policies/
COPY --from=builder /app/policy-decision-point/data/ /app/data/
COPY --from=builder /app/orchestration-engine-go/schema.graphql /app/
COPY --from=builder /app/orchestration-engine-go/config.json /app/

# Create startup script
RUN cat > /app/start-services.sh << 'EOF'
#!/bin/sh
set -e

# Start Policy Decision Point
echo "Starting Policy Decision Point on port 8082..."
/app/bin/pdp -env=${ENVIRONMENT:-production} -port=8082 &
PDP_PID=$!

# Start Consent Engine
echo "Starting Consent Engine on port 8081..."
/app/bin/ce -env=${ENVIRONMENT:-production} -port=8081 &
CE_PID=$!

# Start Orchestration Engine
echo "Starting Orchestration Engine on port 4000..."
/app/bin/oe &
OE_PID=$!

# Function to handle shutdown
shutdown() {
    kill $PDP_PID $CE_PID $OE_PID 2>/dev/null || true
    wait
    exit 0
}

# Set up signal handlers
trap shutdown SIGTERM SIGINT

# Wait for all processes
wait
EOF

RUN chmod +x /app/start-services.sh

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user with specific UID
USER 10001

# Expose all ports
EXPOSE 4000 8081 8082

# Set environment variables
ENV CONFIG_DIR=/app/config
ENV ENVIRONMENT=production

# Health check for all services
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4000/health && \
        wget --no-verbose --tries=1 --spider http://localhost:8081/health && \
        wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Start all services
CMD ["/app/start-services.sh"]
 