FROM golang:1.24-alpine

WORKDIR /app

# Copy both go.mod and go.sum
COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /main .

EXPOSE 8082

CMD [ "/main" ]