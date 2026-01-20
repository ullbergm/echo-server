# Echo Server (Golang/Fiber)

A cloud-native HTTP echo server built with Golang and Fiber that reflects request information, making it perfect for debugging HTTP clients, testing APIs, and learning about Kubernetes.

## Features

- **🔄 Request Echo** - See your complete HTTP request (method, path, headers, query params)
- **📦 Body Echo** - Capture and display request body with Content-Type aware parsing (JSON, XML, form-data, plain text)
- **🌐 Dual Format** - Beautiful HTML for browsers, clean JSON for APIs  
- **☸️ Kubernetes Native** - Shows pod metadata via environment variables when running in K8s
- **🔐 JWT Decoder** - Automatically decodes JWT tokens from your requests
- **📊 Prometheus Metrics** - Built-in metrics endpoint for monitoring
- **📈 Monitor Dashboard** - Real-time server metrics (CPU, RAM, connections)
- **🎯 Custom Status Codes** - Test error handling by controlling response status
- **🗜️ Response Compression** - Automatic gzip/deflate/brotli compression based on Accept-Encoding header
- **🚀 Lightning Fast** - Native Go binary with instant startup and high-performance JSON encoding
- **⚡ Performance Optimized** - Uses goccy/go-json for faster JSON operations and zero-allocation utilities

## Quick Start

### Run with Docker

```bash
docker build -t echo-server:latest .
docker run -p 8080:8080 echo-server:latest
```

Then open http://localhost:8080 in your browser!

### Run Locally

```bash
# Using Make (recommended)
make run

# Or directly with go run
go run main.go

# Or build with version
make build
./echo-server
```

### Build Binary

```bash
# Build with automatic version from git tags
make build

# Or manually specify version
go build -ldflags "-X main.Version=1.0.0" -o echo-server

# Build optimized (smaller binary)
make build-optimized
```

The version is automatically injected at build time from git tags/commits. If not in a git repository, it defaults to "dev".

## Configuration

Environment variables:

- `PORT` - Server port (default: 8080)
- `FIBER_PREFORK` - Enable prefork mode for multi-core scalability (default: false) - Linux only
- `ECHO_PAGE_TITLE` - Custom page title for HTML interface
- `ECHO_ENVIRONMENT_VARIABLES_DISPLAY` - Comma-separated list of env vars to display
- `MAX_BODY_SIZE` - Maximum request body size in bytes (default: 10485760 = 10MB)
- `JWT_HEADER_NAMES` - Comma-separated list of headers to check for JWT (default: Authorization,X-JWT-Token,X-Auth-Token,JWT-Token)
- `HEALTH_READINESS_DELAY_SECONDS` - Delay before readiness probe returns healthy (default: 0)
- `LOG_HEALTHCHECKS` - Enable logging of healthcheck requests (default: false)

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

### Example with Prefork:
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

## License

This project is open source and available under the MIT License.
