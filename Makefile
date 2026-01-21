.PHONY: build run clean test lint lint-fix test-race test-verbose coverage coverage-view security pre-commit dev docker-build docker-run deps-check deps-update deps-verify bench fmt version

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

# Run linter
lint:
	@echo "Running golangci-lint..."
	go tool golangci-lint run

# Fix linting issues automatically
lint-fix:
	@echo "Running golangci-lint with fixes..."
	go tool golangci-lint run --fix

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests and open coverage in browser
coverage-view: coverage
	@echo "Opening coverage report in browser..."
	@if command -v open > /dev/null; then \
		open coverage.html; \
	elif command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
	elif command -v start > /dev/null; then \
		start coverage.html; \
	else \
		echo "Please open coverage.html manually"; \
	fi

# Check for security issues (requires gosec)
security:
	@echo "Running security checks..."
	go tool gosec ./...

# Run all pre-commit checks
pre-commit: fmt lint test
	@echo "All pre-commit checks passed!"

# Run with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	go tool air

# Docker build with version tag
docker-build:
	@echo "Building Docker image version $(VERSION)..."
	docker build -t echo-server:$(VERSION) -t echo-server:latest .

# Docker run for local testing
docker-run:
	@echo "Running Docker container..."
	docker run -it --rm -p 8080:8080 -p 8443:8443 echo-server:latest

# Check dependencies for updates
deps-check:
	@echo "Checking for dependency updates..."
	go list -u -m all

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	go mod verify

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...
