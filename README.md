# Echo Server (Golang/Fiber)

A cloud-native HTTP echo server built with Golang and Fiber that reflects request information, making it perfect for debugging HTTP clients, testing APIs, and learning about Kubernetes.

## Features

- **🔄 Request Echo** - See your complete HTTP request (method, path, headers, query params)
- **📦 Body Echo** - Capture and display request body with Content-Type aware parsing (JSON, XML, form-data, plain text)
- **🌐 Dual Format** - Beautiful HTML for browsers, clean JSON for APIs  
- **☸️ Kubernetes Native** - Shows pod metadata via environment variables when running in K8s
- **🔐 JWT Decoder** - Automatically decodes JWT tokens from your requests
- **🔒 TLS/HTTPS Support** - Optional HTTPS with automatic self-signed certificate generation or custom certificates
- **📊 Prometheus Metrics** - Built-in metrics endpoint for monitoring
- **📈 Monitor Dashboard** - Real-time server metrics (CPU, RAM, connections)
- **🎯 Custom Status Codes** - Test error handling by controlling response status
- **🗜️ Response Compression** - Automatic gzip/deflate/brotli compression based on Accept-Encoding header
- **🚀 Lightning Fast** - Native Go binary with instant startup and high-performance JSON encoding
- **⚡ Performance Optimized** - Uses goccy/go-json for faster JSON operations and zero-allocation utilities

## Quick Start

### Run with Docker

```bash
# Build the image
docker build -t echo-server:latest .

# Run with HTTP only
docker run -p 8080:8080 echo-server:latest

# Run with HTTPS enabled (self-signed certificate)
docker run -p 8080:8080 -p 8443:8443 \
  -e TLS_ENABLED=true \
  echo-server:latest

# Run with custom certificates
docker run -p 8080:8080 -p 8443:8443 \
  -e TLS_ENABLED=true \
  -v /path/to/certs:/certs:ro \
  echo-server:latest
```

Then open <http://localhost:8080> (or <https://localhost:8443>) in your browser!

### Run Locally

```bash
# Using Task (recommended)
task run

# Or directly with go run
go run main.go

# Or build with version
task build
./echo-server
```

### Build Binary

```bash
# Build with automatic version from git tags
task build

# Or manually specify version
go build -ldflags "-X main.Version=1.0.0" -o echo-server

# Build optimized (smaller binary)
task build-optimized
```

The version is automatically injected at build time from git tags/commits. If not in a git repository, it defaults to "dev".

### Pre-commit Hooks

This project uses [pre-commit](https://pre-commit.com/) to ensure code quality before commits:

**Installation:**

```bash
# Install pre-commit
pip install pre-commit
# Or on macOS:
brew install pre-commit

# Install the git hooks
task pre-commit-install
```

**Usage:**

```bash
# Hooks run automatically on git commit
git commit -m "feat(api): add new endpoint"

# Run hooks manually on all files
task pre-commit-run

# Update hook versions
task pre-commit-update

# Skip hooks (not recommended)
git commit --no-verify -m "message"
```

**What hooks run:**

- Code formatting (gofmt)
- Linting (golangci-lint)
- Tests (go test with race detection)
- YAML/Markdown linting
- Dockerfile linting (hadolint)
- Commit message validation (conventional commits)
- Secret detection
- Trailing whitespace and file endings

### Commit Message Convention

This project follows [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

**Format:**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test changes
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Other changes

**Examples:**

```bash
feat(handlers): add custom status code support
fix(jwt): handle missing authorization header
docs: update README with TLS examples
test(services): add body parser edge cases
```

**Validation:**

```bash
# Validate a commit message
echo "feat(api): add new endpoint" | npx @commitlint/cli --config .commitlintrc.yml

# Validate the last commit
git log -1 --pretty=%B | npx @commitlint/cli --config .commitlintrc.yml

# Show usage examples
task commit-lint

# Commit messages are auto-validated by pre-commit hook
git commit -m "feat(api): add new endpoint"
```

See [COMMIT_CONVENTION.md](COMMIT_CONVENTION.md) for detailed guidelines.

### Dev Container (Codespaces / VS Code)

This project includes a complete dev container configuration with all tools pre-installed:

**Open in GitHub Codespaces:**

- Click "Code" → "Create codespace on main"
- Everything is pre-configured and ready to use

**Open in VS Code Dev Container:**

```bash
# Prerequisites: Docker and VS Code with "Dev Containers" extension
# Open in VS Code, then:
# F1 → "Dev Containers: Reopen in Container"
```

**Pre-installed tools:**

- Go 1.25, Node.js LTS, Python 3.11
- Task runner, pre-commit, commitlint
- golangci-lint, hadolint, yamllint, markdownlint
- All VS Code extensions (Go, Docker, Kubernetes, etc.)
- Pre-commit hooks automatically configured

See [.devcontainer/README.md](.devcontainer/README.md) for details.

### Using Task (Task Runner)

This project uses [Task](https://taskfile.dev) as its command runner for better developer experience.

**Installation:**

```bash
# macOS
brew install go-task

# Linux (installs to ~/.local/bin)
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Or using go
go install github.com/go-task/task/v3/cmd/task@latest

# Windows (using Chocolatey)
choco install go-task
```

**Quick commands:**

```bash
task                # Show available tasks (runs --list)
task --list         # List all tasks with descriptions
task help           # Show detailed help

# Development workflow
task run            # Run the server
task dev            # Run with hot reload
task test           # Run tests
task lint           # Run linter
task pre-commit-run # Run all pre-commit checks

# Quick Start
task run               # Run the server
task dev               # Run with hot reload
task test              # Run tests
task build             # Build binary

# Development
task test-coverage     # Run tests with coverage
task test-race         # Run with race detection
task coverage-view     # Generate and view coverage HTML
task lint              # Run linter
task lint-fix          # Auto-fix linting issues
task fmt               # Format code
task pre-commit-run    # Run all pre-commit checks
task pre-commit-install # Install pre-commit hooks

# Docker
task docker-build      # Build Docker image
task docker-run        # Run Docker container
task docker-all        # Build and run
task docker-run     # Run container
task docker-all     # Build and run

# More
task ci             # Run all CI checks
task pre-commit     # Run pre-commit checks
task version        # Show current version
```

> **Note:** The Makefile is still present for backward compatibility, but Task is recommended for new development.

## Configuration

Environment variables:

### General Configuration

- `PORT` - HTTP server port (default: 8080)
- `FIBER_PREFORK` - Enable prefork mode for multi-core scalability (default: false) - Linux only
- `ECHO_PAGE_TITLE` - Custom page title for HTML interface
- `ECHO_ENVIRONMENT_VARIABLES_DISPLAY` - Comma-separated list of env vars to display
- `MAX_BODY_SIZE` - Maximum request body size in bytes (default: 10485760 = 10MB)
- `JWT_HEADER_NAMES` - Comma-separated list of headers to check for JWT (default: Authorization,X-JWT-Token,X-Auth-Token,JWT-Token)
- `HEALTH_READINESS_DELAY_SECONDS` - Delay before readiness probe returns healthy (default: 0)
- `LOG_HEALTHCHECKS` - Enable logging of healthcheck requests (default: false)

### TLS/HTTPS Configuration

- `TLS_ENABLED` - Enable TLS/HTTPS support (default: false)
- `TLS_PORT` - HTTPS server port (default: 8443)
- `TLS_CERT_FILE` - Path to TLS certificate file (default: /certs/tls.crt)
- `TLS_KEY_FILE` - Path to TLS private key file (default: /certs/tls.key)

When `TLS_ENABLED=true`:

- If certificate files exist at the specified paths, they will be loaded
- If certificate files don't exist, a self-signed certificate is automatically generated in memory
- Both HTTP (PORT) and HTTPS (TLS_PORT) servers run simultaneously
- Self-signed certificates are valid for 365 days with 2048-bit RSA keys

## API Endpoints

- `/*` - Echo endpoint (all HTTP methods: GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD)
- `/monitor` - Monitor dashboard with real-time server metrics (CPU, RAM, connections)
- `/healthz/live` - Liveness probe
- `/healthz/ready` - Readiness probe
- `/metrics` - Prometheus metrics

### Response Compression

The server automatically compresses responses when the client sends an `Accept-Encoding` header with supported compression methods (gzip, deflate, or brotli).

**Example:**

```bash
# With compression
curl -H "Accept-Encoding: gzip" http://localhost:8080/

# Without compression
curl http://localhost:8080/
```

Compression is automatically skipped for:

- Healthcheck endpoints (`/healthz/live`, `/healthz/ready`)
- Responses smaller than 200 bytes
- Already encoded responses

### Request Body Echo

The echo server automatically captures and parses request bodies for POST, PUT, PATCH, and DELETE methods:

**Supported Content Types:**

- `application/json` - Parsed and displayed as JSON
- `application/xml` / `text/xml` - Parsed and displayed as XML (or raw text if parsing fails)
- `application/x-www-form-urlencoded` - Parsed as form data
- `multipart/form-data` - Parsed with file upload support
- `text/*` - Displayed as plain text
- Binary data - Automatically detected and base64 encoded

**Examples:**

```bash
# POST with JSON body
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":30}'

# POST with form data
curl -X POST http://localhost:8080/api/form \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=johndoe&email=john@example.com"

# PUT with JSON body
curl -X PUT http://localhost:8080/api/resource/123 \
  -H "Content-Type: application/json" \
  -d '{"status":"updated"}'

# POST with binary data (will be base64 encoded)
curl -X POST http://localhost:8080/api/upload \
  -H "Content-Type: application/octet-stream" \
  --data-binary @file.bin
```

**Body Safety Features:**

- Maximum body size limit (default 10MB, configurable via `MAX_BODY_SIZE`)
- Binary data detection and automatic base64 encoding
- Truncation indicator when body exceeds size limit
- Content-Type aware parsing with fallback to text/base64

### Monitor Dashboard

The `/monitor` endpoint provides a real-time dashboard showing server metrics:

- **CPU Usage** - Process and system CPU utilization
- **Memory Usage** - RAM consumption by the process
- **Response Time** - Average response latency
- **Open Connections** - Number of active connections

Access the dashboard at: `http://localhost:8080/monitor`

You can also retrieve metrics as JSON:

```bash
curl -H "Accept: application/json" http://localhost:8080/monitor
```

## TLS/HTTPS Support

The echo server supports TLS/HTTPS with automatic self-signed certificate generation or custom certificates.

### Quick Start with TLS

```bash
# Enable HTTPS with auto-generated self-signed certificate
TLS_ENABLED=true ./echo-server

# Both servers will start:
# - HTTP on port 8080
# - HTTPS on port 8443

# Test HTTP endpoint
curl http://localhost:8080/

# Test HTTPS endpoint (with self-signed cert)
curl -k https://localhost:8443/
```

### Using Custom Certificates

```bash
# Option 1: Using file paths
TLS_ENABLED=true \
TLS_CERT_FILE=/path/to/cert.pem \
TLS_KEY_FILE=/path/to/key.pem \
./echo-server

# Option 2: Using Docker with mounted certificates
docker run -p 8080:8080 -p 8443:8443 \
  -e TLS_ENABLED=true \
  -v /path/to/certs:/certs:ro \
  echo-server:latest
```

### TLS Features

- **Dual-Stack Operation**: HTTP and HTTPS servers run simultaneously when TLS is enabled
- **Auto-Generated Certificates**: Self-signed certificates are automatically created when no certificate files are found
- **Certificate Information**: View certificate details (subject, issuer, expiry, serial number) in the echo response
- **TLS Connection Info**: Each request shows whether it was received via HTTP or HTTPS
- **Protocol Metrics**: Prometheus metrics track HTTP vs HTTPS requests separately

### Kubernetes/OpenShift TLS Integration

#### Using Kubernetes Secrets

```bash
# Create TLS secret from certificate files
kubectl create secret tls echo-server-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key

# Apply deployment with TLS enabled
kubectl apply -f deploy/kubernetes/deployment.yaml
```

#### Using cert-manager (Automated Certificates)

cert-manager automatically provisions and renews TLS certificates from Let's Encrypt or other ACME providers.

**Prerequisites:**

- cert-manager installed in your cluster ([installation guide](https://cert-manager.io/docs/installation/))
- ClusterIssuer or Issuer configured

**Deploy with cert-manager:**

```bash
# 1. Apply the Certificate resource
kubectl apply -f deploy/kubernetes/certificate.yaml

# 2. cert-manager will automatically create the echo-server-tls secret

# 3. Deploy the application with TLS enabled
kubectl apply -f deploy/kubernetes/deployment.yaml

# 4. Apply the Ingress with TLS
kubectl apply -f deploy/kubernetes/ingress.yaml
```

The Certificate resource (`deploy/kubernetes/certificate.yaml`) specifies:

- Certificate duration and renewal periods
- DNS names the certificate is valid for
- Reference to the Issuer/ClusterIssuer
- Secret name where the certificate will be stored

cert-manager will automatically:

- Request a certificate from the configured issuer
- Store the certificate in the specified Kubernetes secret
- Renew the certificate before it expires

#### OpenShift Routes with TLS

OpenShift provides three TLS termination options:

**1. Edge Termination (Recommended)**

- TLS is terminated at the router
- Traffic to the service is HTTP
- No application-side TLS configuration needed

```bash
# Apply edge termination route (default)
kubectl apply -f deploy/openshift/route.yaml
```

**2. Passthrough Termination**

- TLS traffic passes through the router to the service
- Application handles TLS termination (requires `TLS_ENABLED=true`)
- End-to-end encryption

```yaml
# Uncomment the passthrough route in deploy/openshift/route.yaml
tls:
  termination: passthrough
  insecureEdgeTerminationPolicy: Redirect
```

**3. Re-encrypt Termination**

- TLS is terminated at the router and re-encrypted to the service
- Requires `TLS_ENABLED=true` in the application
- Provides end-to-end encryption with router-based certificate management

```yaml
# Uncomment the re-encrypt route in deploy/openshift/route.yaml
tls:
  termination: reencrypt
  insecureEdgeTerminationPolicy: Redirect
```

### Self-Signed Certificate Details

When TLS is enabled without certificate files, the server automatically generates a self-signed certificate with:

- **Algorithm**: RSA 2048-bit
- **Validity**: 365 days from generation
- **Subject**: CN=<hostname>, O=Echo Server
- **DNS Names**: <hostname>, localhost
- **Key Usage**: Key Encipherment, Digital Signature
- **Extended Key Usage**: Server Authentication

**Note:** Self-signed certificates are suitable for testing and development. For production use, use certificates from a trusted Certificate Authority or cert-manager.

## Middleware Stack

1. **Recovery** - Panic handling
2. **Compress** - Response compression (gzip/deflate/brotli) based on Accept-Encoding header
3. **Favicon** - Serves favicon.ico
4. **Logger** - Request logging with IP and User-Agent
5. **Healthcheck** - Liveness and readiness probes
6. **Metrics** - Prometheus metrics collection
7. **Monitor** - Real-time server monitoring dashboard

## Technology Stack

- **Language**: Go 1.25+
- **Framework**: Fiber v2 (Express-like, built on fasthttp)
- **Template Engine**: Fiber template/html
- **Metrics**: Prometheus client_golang
- **JSON Encoding**: goccy/go-json (faster than standard encoding/json)

## Performance Optimizations

This echo server implements several performance optimizations based on the [Fiber "Faster Fiber" guide](https://docs.gofiber.io/guide/faster-fiber):

### 1. Fast JSON Encoding/Decoding

- Uses **goccy/go-json** instead of Go's standard `encoding/json`
- Significantly faster JSON marshaling/unmarshaling for API responses
- Configured in the Fiber app config with `JSONEncoder` and `JSONDecoder`

### 2. Zero-Allocation String Conversions

- Uses `gofiber/utils.UnsafeString()` for byte-to-string conversions
- Reduces memory allocations when processing headers and query strings
- Particularly effective in high-throughput scenarios

### 3. Optional Prefork Mode

- Enable with `FIBER_PREFORK=true` environment variable
- Spawns multiple processes to utilize all CPU cores
- Best for production deployments on Linux systems
- Not recommended for Windows environments

### Example with Prefork

```bash
# Enable prefork for multi-core performance
FIBER_PREFORK=true ./echo-server

# Or with Docker
docker run -e FIBER_PREFORK=true -p 8080:8080 echo-server:latest
```

### Performance Benefits

- **Faster JSON operations**: 2-3x improvement in JSON encoding/decoding
- **Reduced allocations**: Lower memory footprint and GC pressure
- **Better scalability**: Prefork mode enables true multi-core utilization

## TLS/HTTPS Support

The echo server supports optional TLS/HTTPS with flexible certificate management.

### Enable TLS

Enable HTTPS by setting the `TLS_ENABLED` environment variable:

```bash
# Run with TLS enabled (using self-signed certificate)
TLS_ENABLED=true ./echo-server

# With Docker
docker run -e TLS_ENABLED=true -p 8080:8080 -p 8443:8443 echo-server:latest
```

When TLS is enabled:

- **HTTP server** continues running on port 8080 (or custom `PORT`)
- **HTTPS server** runs on port 8443 (or custom `TLS_PORT`)
- Both servers share the same routes, middleware, and handlers

### Certificate Management

#### Option 1: Self-Signed Certificates (Automatic)

If no certificate files are provided, the server automatically generates a self-signed certificate:

```bash
# Automatically generates self-signed certificate
TLS_ENABLED=true ./echo-server
```

**Self-signed certificate properties:**

- 2048-bit RSA key
- Valid for 365 days
- Includes hostname and localhost in DNS names
- Generated in memory (not written to disk)

#### Option 2: Custom Certificates

Provide your own certificates via environment variables:

```bash
# Using custom certificate files
TLS_ENABLED=true \
TLS_CERT_FILE=/path/to/cert.pem \
TLS_KEY_FILE=/path/to/key.pem \
./echo-server

# With Docker (mount certificate directory)
docker run \
  -e TLS_ENABLED=true \
  -e TLS_CERT_FILE=/certs/tls.crt \
  -e TLS_KEY_FILE=/certs/tls.key \
  -v /local/certs:/certs:ro \
  -p 8080:8080 -p 8443:8443 \
  echo-server:latest
```

### Kubernetes Integration

Example Kubernetes deployment with TLS:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: echo-server-tls
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-server
spec:
  template:
    spec:
      containers:
      - name: echo-server
        image: echo-server:latest
        env:
        - name: TLS_ENABLED
          value: "true"
        - name: TLS_CERT_FILE
          value: "/certs/tls.crt"
        - name: TLS_KEY_FILE
          value: "/certs/tls.key"
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8443
          name: https
        volumeMounts:
        - name: tls-certs
          mountPath: /certs
          readOnly: true
      volumes:
      - name: tls-certs
        secret:
          secretName: echo-server-tls
```

### Testing HTTPS

```bash
# Test HTTPS endpoint (self-signed certificate)
curl -k https://localhost:8443/

# Test with certificate verification disabled
curl --insecure https://localhost:8443/api/test

# View certificate information
openssl s_client -connect localhost:8443 -showcerts
```

### TLS Response Information

When TLS is enabled, the echo response includes TLS information:

**Server TLS Info** (`server.tls`):

- Certificate subject and issuer
- Certificate validity period (notBefore, notAfter)
- Certificate serial number
- DNS names in certificate

**Request TLS Info** (`request.tls`):

- Whether the specific request used TLS
- TLS version (when available)
- Cipher suite information (when available)

**Example JSON response:**

```json
{
  "request": {
    "tls": {
      "enabled": true,
      "version": "TLS 1.3"
    }
  },
  "server": {
    "tls": {
      "enabled": true,
      "subject": "CN=echo-server,O=Echo Server",
      "issuer": "CN=echo-server,O=Echo Server",
      "notBefore": "2024-01-20T12:00:00Z",
      "notAfter": "2025-01-20T12:00:00Z",
      "serialNumber": "123456789",
      "dnsNames": ["echo-server", "localhost"]
    }
  }
}
```

### Metrics

When TLS is enabled, Prometheus metrics include a `protocol` label to distinguish HTTP from HTTPS requests:

```
# HTTP requests
echo_requests_total{method="GET",uri="/",protocol="http"} 42

# HTTPS requests  
echo_requests_total{method="GET",uri="/",protocol="https"} 58
```

## License

This project is open source and available under the MIT License.
