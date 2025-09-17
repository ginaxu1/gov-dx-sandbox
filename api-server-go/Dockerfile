# Build stage
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy shared modules to the correct structure
COPY exchange/shared/ ./exchange/shared/

# Copy API server source code
COPY api-server-go/ ./api-server-go/

# Set working directory to API server
WORKDIR /app/api-server-go

# Build the application
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service_binary .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates wget

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/api-server-go/service_binary .

# Copy OpenAPI specification
COPY --from=builder /app/api-server-go/openapi.yaml .

# Make binary executable
RUN chmod +x service_binary

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run the binary
CMD ["./service_binary"]