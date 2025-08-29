# In consent-engine/Dockerfile

FROM golang:1.24-alpine

WORKDIR /app

# Copy both go.mod and go.sum
COPY go.mod go.sum ./

# This command now uses go.sum to download the exact dependencies
RUN go mod download

COPY *.go ./

RUN go build -o /main .

EXPOSE 8082

CMD [ "/main" ]