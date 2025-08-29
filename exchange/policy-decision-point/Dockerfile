# Use the official Go image to build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies needed for robust DNS resolution in Alpine
# (This ensures Go's net package has necessary underlying libs)
RUN apk update && apk add --no-cache libc-dev

# Copy the Go module and source files
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /pdp_app

# Use a minimal image to run the application
FROM alpine:latest

WORKDIR /

# Copy the Go binary from the builder stage
COPY --from=builder /pdp_app .

# Expose the port the app will run on
EXPOSE 8080

# Set the entry point to run the application
CMD ["./pdp_app"]