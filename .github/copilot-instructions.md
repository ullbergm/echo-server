# GitHub Copilot Instructions for Echo Server

## Project Overview

This is a **Fiber-based HTTP echo server** written in Go 1.25+ that responds with detailed information about incoming requests, server details, and Kubernetes pod information. The application is designed to be compiled into a native binary with fast startup times and low memory footprint.

## Technology Stack

- **Language**: Go 1.25+
- **Framework**: Fiber v2 (Express-like, built on fasthttp)
- **Build Tool**: Go modules
- **Testing**: Go standard testing package
- **Containerization**: Docker with multi-stage builds (Alpine-based)
- **Kubernetes**: Environment variables via Downward API for pod metadata
- **Template Engine**: Fiber template/html
- **Metrics**: Prometheus client_golang
- **Health**: Fiber healthcheck middleware
- **Security**: JWT decoding (inspection only, no validation)

## Current Feature Set

The echo server provides the following capabilities:

### Core Features
- ✅ All HTTP methods support (GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD)
- ✅ Wildcard path matching - responds to any URL path
- ✅ Content negotiation - JSON for APIs, HTML for browsers
- ✅ Request echoing - complete request info (method, path, query, headers, remote IP)
- ✅ Server information - hostname, IP address, environment variables
- ✅ Proxy header support - X-Forwarded-For, X-Real-IP

### Advanced Features
- ✅ **Kubernetes Integration**: Pod metadata (namespace, name, IP, node), labels, and annotations from environment variables
- ✅ **JWT Token Decoding**: Automatic detection and decoding from multiple headers
- ✅ **Custom Status Codes**: Control response status via `x-set-response-status-code` header
- ✅ **Prometheus Metrics**: Request counters and latency histograms per endpoint
- ✅ **Health Checks**: Liveness and readiness probes with configurable startup delay
- ✅ **Native Compilation**: Fast startup times (<100ms)

### Configuration
- JWT header detection (configurable list of headers to check)
- Readiness delay (configurable startup delay for gradual rollouts)
- Environment variable-based configuration

## Build & Test Commands

All commands should be executed from the repository root using Task (Taskfile.yml).

**Note:** The Taskfile uses `go run <package>@latest` for tools like golangci-lint, gosec, and air.
This ensures consistent versions across all developers and CI without requiring manual tool installation.
The tools are automatically downloaded and cached by Go when first used.

### Development
```bash
# Run in development mode
task run

# Run with hot reload
task dev

# Run with custom port (using go run directly)
go run main.go -port 8087

# Access the application at http://localhost:8080
```

### Testing
```bash
# Run all tests
task test

# Run tests with coverage
task test-coverage

# Run tests with verbose output
task test-verbose

# Run tests with race detection
task test-race

# Generate and view coverage report
task coverage-view

# Or use go commands directly
go test ./...
go test -cover ./...
```

### Building
```bash
# Build binary (with version from git)
task build

# Build with optimizations (smaller binary)
task build-optimized

# Check current version
task version

# Or use go commands directly
go build -o echo-server
GOOS=linux GOARCH=amd64 go build -o echo-server-linux
```

### Linting & Security
```bash
# Run linter (requires golangci-lint installed)
task lint

# Fix linting issues automatically
task lint-fix

# Run security checks (requires gosec installed)
task security

# Tools are run via 'go run <package>@latest' for consistency
# No manual installation required - Go handles it automatically
```

### Docker
```bash
# Build Docker image (with version from git)
task docker-build

# Run container
task docker-run

# Build and run
task docker-all

# Or use docker commands directly
docker build -t echo-server:latest .
docker run -i --rm -p 8080:8080 echo-server:latest
```

## Project Structure

```
echo-server/
├── README.md                    # Main user documentation
├── Taskfile.yml                 # Task runner configuration (preferred over Makefile)
├── Makefile                     # Legacy make commands (deprecated)
├── kubernetes-deployment.yaml   # K8s deployment manifest
├── Dockerfile                   # Multi-stage build
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
├── main.go                      # Application entry point
├── handlers/
│   ├── echo.go                  # Main echo endpoint handler
│   └── metrics.go               # Metrics middleware
├── models/
│   └── response.go              # Response model structs
├── services/
│   ├── jwt.go                   # JWT token decoder service
│   └── metrics.go               # Prometheus metrics service
├── templates/
│   └── echo.html                # HTML template
└── public/
    └── favicon.ico              # Favicon asset
```

### Key Files

#### Application Code
- **main.go**: Application entry point with Fiber setup, middleware configuration, and template functions
- **handlers/echo.go**: Main echo handler with endpoints for all HTTP methods and both JSON/HTML responses; includes `getKubernetesInfo()` for reading pod metadata from environment variables
- **models/response.go**: Data structures for request, server, Kubernetes, and JWT information
- **services/jwt.go**: Service for detecting and decoding JWT tokens from request headers
- **services/metrics.go**: Prometheus metrics collection and registration
- **templates/echo.html**: Go HTML template for browser interface
- **go.mod**: Go module dependencies

#### Documentation Files
- **README.md**: Main user-facing documentation covering all features, API endpoints, quick start, configuration, use cases, testing, troubleshooting, and performance

## API Endpoints

The echo server supports **all HTTP methods** (GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD) on **all paths**.

### Main Echo Endpoint
- **Path**: `/{any-path}` - Wildcard path matching, responds to any URL path
- **Methods**: GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD
- **Response Format**: JSON (default) or HTML (based on `Accept` header)
- **Features**:
  - Reflects complete request information (method, path, query, headers, remote address)
  - Returns server information (hostname, IP, environment variables)
  - Returns Kubernetes pod metadata when running in K8s (namespace, pod name, labels, annotations)
  - Automatically detects and decodes JWT tokens from configured headers
  - Supports custom status codes via `x-set-response-status-code` header

### Health Check Endpoints
- **Liveness**: `GET /healthz/live` - Returns 200 OK when application is running
- **Readiness**: `GET /healthz/ready` - Returns 200 OK when ready to receive traffic (supports configurable startup delay)

### Metrics Endpoint
- **Prometheus**: `GET /metrics` - Exposes Prometheus-compatible metrics including:
  - `echo_requests_total{method, uri}` - Request counters per endpoint
  - `http_server_requests_seconds{method, uri}` - Request latency histograms

### Key Features
- **Content Negotiation**: Returns JSON for `Accept: application/json` or HTML for browsers
- **JWT Decoding**: Automatically decodes JWT tokens from Authorization, X-JWT-Token, X-Auth-Token, and JWT-Token headers
- **Custom Status Codes**: Set response status code with `x-set-response-status-code` header (200-599)
- **Proxy Header Support**: Correctly extracts remote IP from X-Forwarded-For and X-Real-IP headers
- **Metrics Collection**: Automatic request counting and latency tracking per endpoint


## Code Style & Conventions

### Go Conventions
- Use **Go 1.25+** features where appropriate
- Follow standard Go naming conventions (camelCase for unexported, PascalCase for exported)
- Use **gofmt** for code formatting (run automatically by most editors)
- Use **golangci-lint** for code quality checks (if configured)
- Follow Go project layout standards (handlers, models, services packages)

### Code Formatting
- Run `gofmt -w .` to format all Go files
- Run `go mod tidy` after adding/removing dependencies
- Most editors (VS Code, GoLand) format automatically on save
- Use meaningful variable names (avoid single-letter except for short scopes)

### Code Organization
- Keep HTTP handlers in `handlers` package
- Place data models in `models` package
- Place business logic in `services` package
- Use private (unexported) helper functions to keep handlers clean
- Group related functionality in the same file

### Error Handling
- Always check and handle errors explicitly
- Log errors appropriately (consider structured logging)
- Return sensible defaults (e.g., "unknown") when information is unavailable
- Use `fmt.Errorf` or `errors.New` for error creation

### Template Usage
- Use Go's `html/template` via Fiber template engine
- Template automatically escapes HTML to prevent XSS
- Register custom functions with `engine.AddFunc()`
- Keep templates in `templates` directory

### Testing
- Use Go's standard `testing` package
- Test files should be named `*_test.go`
- Use `go test ./...` to run all tests
- Use table-driven tests for multiple test cases

## Kubernetes Integration

The application retrieves Kubernetes pod metadata exclusively from environment variables injected via the Kubernetes Downward API. This approach is simpler, more secure, and doesn't require RBAC permissions:

- **No Kubernetes API access required** - No ServiceAccount, Role, or RoleBinding needed
- **Environment variable based** - Uses Downward API to inject pod metadata
- **Simpler deployment** - Works in restricted environments
- **More secure** - No API server communication or RBAC permissions
- Gracefully degrades when not running in Kubernetes (returns null)

### Environment Variables
The application reads these environment variables (injected by Kubernetes Downward API):
- `HOSTNAME` - Pod hostname
- `K8S_NAMESPACE` - Kubernetes namespace (from `metadata.namespace`)
- `K8S_POD_NAME` - Pod name (from `metadata.name`)
- `K8S_POD_IP` - Pod IP address (from `status.podIP`)
- `K8S_NODE_NAME` - Node where pod is running (from `spec.nodeName`)
- `K8S_LABEL_*` - Pod labels (e.g., `K8S_LABEL_app` from `metadata.labels['app']`, `K8S_LABEL_deployment` from `metadata.labels['deployment']`)
- `K8S_ANNOTATION_*` - Pod annotations (e.g., `K8S_ANNOTATION_description` from `metadata.annotations['description']`)
- `KUBERNETES_SERVICE_HOST` / `KUBERNETES_SERVICE_PORT` - K8s API server (shown in Kubernetes section as `serviceHost` and `servicePort`)

## Security Considerations

- **Always escape HTML output** using the `escapeHtml()` method to prevent XSS attacks
- **Never log sensitive information** (credentials, tokens, etc.)
- **Never commit secrets** to the repository
- The application is read-only and doesn't modify any resources
- No Kubernetes API access required - uses only environment variables for pod metadata

## Boundaries & Restrictions

### DO NOT modify:
- Dockerfile multi-stage build structure (unless necessary for the task)
- Existing security measures (HTML template escaping)
- Core endpoint paths (would break existing integrations)

### DO modify carefully:
- `go.mod` - Only add/update dependencies when necessary (use `go get`)
- Environment variable names (documented in README)
- API endpoint paths (would break existing integrations)

### When adding dependencies:
- Use `go get <package>` to add dependencies
- Run `go mod tidy` to clean up
- Verify dependencies are maintained and compatible
- Test the build after adding dependencies

## Pull Request Requirements

### Code Requirements
- All code changes must maintain or improve existing functionality
- New features should include appropriate tests
- Maintain existing code style and conventions
- Run `gofmt -w .` to format code
- Ensure `go build` succeeds
- Verify the Docker build completes successfully
- Test endpoints manually when making changes to handlers

### Documentation Requirements (MANDATORY)
- **All code changes MUST include corresponding documentation updates**
- New features require updates to README.md, USAGE_GUIDE.md, ARCHITECTURE.md, and SCREENSHOTS.md
- Bug fixes that change behavior require README.md updates
- Configuration changes require updates to README.md configuration table
- API changes require updates to README.md and USAGE_GUIDE.md
- See "Documentation Maintenance Guidelines" section below for details
- PRs without documentation updates will be rejected

### Validation Checklist
Before submitting a PR, ensure:
- [ ] Code formatted with `gofmt -w .`
- [ ] Dependencies tidied with `go mod tidy`
- [ ] Code changes tested (unit tests + manual testing)
- [ ] Documentation updated (see Documentation Review Checklist)
- [ ] All code examples in documentation work correctly
- [ ] Build succeeds: `go build`
- [ ] Docker build succeeds: `docker build -t echo-server:latest .`
- [ ] No security vulnerabilities introduced
- [ ] HTML template escaping maintained for user-provided content

## Common Tasks

### Adding a new feature:
1. **Code Changes**:
   - Add implementation in appropriate service or handler
   - Update data models in `models` package if needed
   - Add tests in corresponding test files
   - Update HTML template if feature affects UI
   - **Run `gofmt -w .` to format the code**

2. **Documentation Updates** (REQUIRED):
   - Update **README.md**: Add feature to features list, API documentation, and examples section
   - Update documentation with detailed usage examples

3. **Verification**:
   - Test the feature manually with curl and browser
   - Run unit tests with `go test ./...`
   - Verify documentation examples work correctly

### Adding a new endpoint:
1. Add handler function in `handlers/echo.go` or new file in `handlers` package
2. Register route in `main.go` with `app.Get()`, `app.Post()`, etc.
3. Add corresponding test in `handlers/*_test.go`
4. **Run `gofmt -w .` to format the code**
5. Update **README.md** API Endpoints section with examples
6. Update documentation with detailed usage examples

### Modifying response format:
1. Update model structs in `models` package
2. Update `buildEchoResponse()` or template in `templates/echo.html`
3. **Run `gofmt -w .` to format the code**
4. Update tests to match new format
5. Update **README.md** if user-visible

### Adding configuration option:
1. Add environment variable check in code with `os.Getenv()`
2. Set default value if environment variable is not set
3. **Run `gofmt -w .` to format the code**
4. Update **README.md** Configuration section table
5. Update documentation with configuration examples

### Adding Kubernetes metadata:
1. Modify `getKubernetesInfo()` function in `handlers/echo.go` to read additional environment variables
2. Update deployment manifest (kubernetes-deployment.yaml) to inject new environment variables using Downward API
3. Update `models.KubernetesInfo` struct if adding new fields
4. Update HTML template in `templates/echo.html` to display new metadata
5. **Run `gofmt -w .` to format the code**
6. Update **README.md** Kubernetes Integration section
7. Update documentation with examples of new metadata and how to inject it

### Fixing bugs:
1. Write test that reproduces the bug
2. Fix the bug in code
3. **Run `gofmt -w .` to format the code**
4. Verify tests pass with `go test ./...`
5. Update documentation if behavior changes or if documentation was incorrect
6. Add troubleshooting entry to **README.md** if it's a common issue

## Documentation Maintenance Guidelines

**CRITICAL**: Documentation must be kept in sync with code changes. When making ANY code change:

### Required Documentation Updates

1. **For New Features** - Update:
   - README.md (features list + API section + examples)

2. **For Bug Fixes** - Update as needed:
   - README.md (if behavior changes or troubleshooting needed)

3. **For Configuration Changes** - Update:
   - README.md (configuration table)

4. **For API Changes** - Update:
   - README.md (API endpoints section)

### Documentation Quality Standards

- **Accuracy**: All code examples must work as shown
- **Completeness**: Every feature must have examples
- **Currency**: Update docs in the same PR as code changes
- **Clarity**: Use clear language with step-by-step instructions
- **Examples**: Include curl commands and expected outputs
- **Troubleshooting**: Add common issues to README.md

### Documentation Review Checklist

Before submitting PR with code changes, verify:
- [ ] All new features documented in README.md
- [ ] Configuration table updated in README.md (if config changes)
- [ ] All code examples tested and working

## Example Code Patterns

### Adding a new response field:
```go
// In models/response.go
type EchoResponse struct {
    Request    RequestInfo     `json:"request"`
    Server     ServerInfo      `json:"server"`
    NewField   string          `json:"newField"`
}

// In handlers/echo.go (populate)
response.NewField = computeNewField()
```

### Testing an endpoint:
```go
func TestEchoHandler(t *testing.T) {
    app := fiber.New()
    app.Get("/test", EchoHandler(jwtService))
    
    req := httptest.NewRequest("GET", "/test", nil)
    resp, err := app.Test(req)
    
    if err != nil {
        t.Fatalf("Failed to send request: %v", err)
    }
    if resp.StatusCode != 200 {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }
}
```

## Additional Notes

- The application is designed to be stateless and cloud-native
- Startup time is typically < 100ms
- Memory footprint is minimal (< 50MB)
- The server binds to `0.0.0.0:8080` by default
- All HTTP methods are supported on the root path
