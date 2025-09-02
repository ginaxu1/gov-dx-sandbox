FROM golang:1.24-alpine AS builder

WORKDIR /app

# Accept the service path as a build argument
ARG SERVICE_PATH

# Copy and download modules first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

RUN go mod tidy

# Copy the entire source code from the root context
COPY . .

# Build the specific service binary using the ARG
RUN CGO_ENABLED=0 go build -o /app/service_binary ./$SERVICE_PATH

FROM alpine:latest

WORKDIR /app

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the binary from the builder stage and set ownership
COPY --from=builder --chown=appuser:appgroup /app/service_binary .

# Switch to the non-root user
USER appuser

# Run the binary
CMD ["./service_binary"]