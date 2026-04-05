# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-api .

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates curl bash docker-cli

WORKDIR /app
COPY --from=builder /app/agent-api .
COPY --from=builder /app/ui ./ui

# Create pages directory
RUN mkdir -p /app/pages

EXPOSE 8080

CMD ["./agent-api"]
