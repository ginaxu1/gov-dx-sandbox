# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orchestration-engine-go .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /workspace

# Copy the binary from builder stage
COPY --from=builder /workspace/orchestration-engine-go .

# Copy production configuration files
COPY config.production.json config.json
COPY schema.production.graphql schema.graphql

# Expose port
EXPOSE 8081

# Run the application
CMD ["./orchestration-engine-go"]
