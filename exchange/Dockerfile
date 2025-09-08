# exchange/Dockerfile

# Builder Stage, same for both services
FROM golang:1.24-alpine AS builder

WORKDIR /app
ARG SERVICE_PATH
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 go build -o /app/service_binary ./$SERVICE_PATH

# Final Stage for Consent Engine (ce_app)
FROM alpine:latest AS ce-final

WORKDIR /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# This service only needs the binary
COPY --from=builder /app/service_binary .
CMD ["./service_binary"]

# Final Stage for Policy Decision Point (pdp_app)
FROM alpine:latest AS pdp-final

WORKDIR /app
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# This service needs the binary and the policies/data directories
COPY --from=builder /app/service_binary .
COPY policy-decision-point/policies ./policies
COPY policy-decision-point/data ./data
CMD ["./service_binary"]