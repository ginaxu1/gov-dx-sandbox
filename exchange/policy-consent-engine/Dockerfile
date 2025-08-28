# Use the official Go image as a base
FROM golang:1.22.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o main .

# Use a minimal base image for the final stage
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built executable from the builder stage
COPY --from=builder /app/main .

# Expose the port your application listens on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]