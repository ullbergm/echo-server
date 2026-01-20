.PHONY: build run clean test

# Get version from git tags, or use 'dev' if not in a git repo
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# Build the application
build:
	@echo "Building echo-server version $(VERSION)..."
	go build $(LDFLAGS) -o echo-server

# Build with optimizations (smaller binary)
build-optimized:
	@echo "Building optimized echo-server version $(VERSION)..."
	go build $(LDFLAGS) -ldflags="-s -w" -o echo-server

# Run the application in development mode
run:
	@echo "Running echo-server version $(VERSION)..."
	go run $(LDFLAGS) main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f echo-server echo-server.exe

# Format code
fmt:
	go fmt ./...

# Show version that would be used
version:
	@echo $(VERSION)
