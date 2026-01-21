# Multi-stage Dockerfile for Echo Server (Golang/Fiber)

## Stage 1: Build the Go binary
FROM golang:1.25.6-alpine AS build

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with version from git
RUN VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev") && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=$VERSION" -o echo-server .

## Stage 2: Create the runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 app && adduser -D -u 1001 -G app app

WORKDIR /app

# Copy binary from build stage
COPY --from=build --chown=app:app /app/echo-server .

# Copy templates and public directories
COPY --from=build --chown=app:app /app/templates ./templates
COPY --from=build --chown=app:app /app/public ./public

# Use non-root user
USER app

# Expose ports
EXPOSE 8080
EXPOSE 8443

# Run the application
ENTRYPOINT ["./echo-server"]
