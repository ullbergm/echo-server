# Golang Rewrite Research

## Executive Summary

This document analyzes the feasibility, benefits, and challenges of rewriting the Echo Server from Java/Quarkus to Golang. The research covers code complexity, performance characteristics, ecosystem considerations, and provides a detailed comparison to guide decision-making.

**Quick Verdict**: Golang would be a viable and potentially beneficial alternative, particularly if simplicity, deployment ease, and ecosystem consistency are prioritized. However, the current Java/Quarkus implementation is mature, performant, and well-maintained, making a rewrite a strategic rather than technical necessity.

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Golang Implementation Analysis](#golang-implementation-analysis)
3. [Code Complexity Comparison](#code-complexity-comparison)
4. [Pros and Cons](#pros-and-cons)
5. [Feature Parity Analysis](#feature-parity-analysis)
6. [Performance Comparison](#performance-comparison)
7. [Ecosystem and Tooling](#ecosystem-and-tooling)
8. [Migration Effort Estimate](#migration-effort-estimate)
9. [Recommendations](#recommendations)

---

## Current State Analysis

### Technology Stack

- **Language**: Java 21
- **Framework**: Quarkus 3.30.6 (modern cloud-native framework)
- **Build Tool**: Maven (with wrapper)
- **Testing**: JUnit 5, REST Assured
- **Containerization**: Multi-stage Docker build with native compilation
- **Native Compilation**: GraalVM/Mandrel 23.1
- **Dependencies**: 
  - quarkus-rest-jackson (REST + JSON serialization)
  - quarkus-micrometer-registry-prometheus (metrics)
  - quarkus-smallrye-health (health checks)

### Code Metrics

| Metric | Value |
|--------|-------|
| Total Java Files | 8 files |
| Total Lines of Code | ~2,103 lines |
| Main Handler (EchoResource.java) | 770 lines |
| Data Models (EchoResponse.java) | 293 lines |
| JWT Service (JwtDecoderService.java) | 197 lines |
| Health Check | ~70 lines |
| Test Files | 2 files |

### Key Features Implemented

1. **HTTP Methods Support**: GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD
2. **Content Negotiation**: JSON and HTML responses based on Accept header
3. **Request Echo**: Complete request information (method, path, query, headers, remote IP)
4. **Server Information**: Hostname, IP address, environment variables
5. **Kubernetes Integration**: Pod metadata from environment variables (namespace, pod name, labels, annotations)
6. **JWT Token Decoding**: Automatic detection and decoding from multiple headers
7. **Custom Status Codes**: Via `x-set-response-status-code` header
8. **Prometheus Metrics**: Request counters and latency histograms per endpoint
9. **Health Checks**: Liveness and readiness probes with configurable startup delay
10. **Native Compilation**: Fast startup (<100ms) and low memory footprint (<50MB)

### Build and Quality Tools

- **Code Formatting**: Eclipse formatter with strict validation
- **Static Analysis**: Checkstyle, SpotBugs
- **Maven Enforcer**: Dependency convergence validation
- **Integration Tests**: Native image testing support

---

## Golang Implementation Analysis

### Technology Stack Equivalent

A Golang implementation would likely use:

- **Language**: Go 1.25+ (similar modern language version)
- **HTTP Framework**: 
  - Option 1: Standard library `net/http` (minimal dependencies)
  - Option 2: Gin or Echo framework (popular, feature-rich)
  - Option 3: Fiber (Express-like, high performance)
- **Build Tool**: Go toolchain (built-in, no external tool needed)
- **Testing**: Standard library `testing` package
- **Containerization**: Single-stage or multi-stage Docker build
- **Native Compilation**: Built-in (no special tooling required)
- **Dependencies**:
  - Standard library for HTTP server, JSON, JWT decoding
  - `github.com/prometheus/client_golang` for metrics
  - Optional: `github.com/gin-gonic/gin` or similar framework

### Estimated Code Metrics

Based on typical Go implementations of similar services:

| Metric | Estimated Value |
|--------|-----------------|
| Total Go Files | 6-8 files |
| Total Lines of Code | 800-1,200 lines |
| Main Handler | 200-300 lines |
| Data Models | 100-150 lines |
| JWT Service | 80-120 lines |
| Health Check | 30-50 lines |
| Test Files | 2-4 files |

**Expected Reduction**: 40-60% fewer lines of code compared to Java/Quarkus

### Code Structure

Typical Go implementation structure:
```
echo-server/
├── main.go                 # Entry point and server setup
├── handlers/
│   ├── echo.go            # Main echo handler
│   ├── health.go          # Health check handlers
│   └── metrics.go         # Metrics middleware
├── models/
│   └── response.go        # Response data structures
├── services/
│   └── jwt.go             # JWT decoding service
├── config/
│   └── config.go          # Configuration management
├── go.mod                  # Dependency management
├── go.sum                  # Dependency checksums
├── Dockerfile              # Container build
└── *_test.go              # Test files (co-located with code)
```

---

## Code Complexity Comparison

### Handler Code Comparison

#### Current Java/Quarkus (770 lines total)

**Characteristics**:
- Separate methods for each HTTP method × content type combination (GET JSON, GET HTML, POST JSON, etc.)
- Heavy use of annotations (@GET, @Path, @Produces, @Context)
- Dependency injection with @Inject
- Verbose HTML generation using StringBuilder
- Concurrent map management for metrics (ConcurrentHashMap, Timer, Counter)
- Complex nested class structure for response models

**Example** (simplified):
```java
@GET
@Path("{path:.*}")
@Produces(MediaType.APPLICATION_JSON)
public Response getJson(@Context HttpHeaders headers, @Context UriInfo uriInfo, @Context HttpServerRequest request) {
    String uri = uriInfo.getPath();
    return getTimer("GET", uri).record(() -> {
        getCounter("GET", uri).increment();
        int statusCode = getCustomStatusCode(headers);
        EchoResponse echoResponse = buildEchoResponse("GET", headers, uriInfo, request);
        return Response.status(statusCode).entity(echoResponse).build();
    });
}

// Repeated for POST, PUT, PATCH, DELETE, OPTIONS, HEAD
// Each method has both JSON and HTML variants
```

#### Golang Equivalent (estimated 200-300 lines)

**Characteristics**:
- Single handler function with content negotiation
- Router pattern matching for all paths
- Direct access to request/response objects
- Template-based or simple HTML generation
- Built-in map types (sync.Map or regular maps with mutex for metrics)
- Struct tags for JSON serialization

**Example** (simplified):
```go
func EchoHandler(w http.ResponseWriter, r *http.Request) {
    // Metrics
    startTime := time.Now()
    defer recordMetrics(r.Method, r.URL.Path, time.Since(startTime))
    
    // Get custom status code if provided
    statusCode := getCustomStatusCode(r)
    
    // Build response
    response := buildEchoResponse(r)
    
    // Content negotiation
    if strings.Contains(r.Header.Get("Accept"), "text/html") {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(statusCode)
        renderHTML(w, response)
    } else {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(response)
    }
}

// Single handler handles all HTTP methods and content types
```

### Model Code Comparison

#### Current Java (293 lines)

**Characteristics**:
- Nested static classes for RequestInfo, ServerInfo, KubernetesInfo, JwtInfo
- Getter/setter methods for every field (boilerplate)
- Jackson annotations for JSON serialization
- Collections.unmodifiableMap for immutability
- GraalVM reflection registration annotations

**Example** (simplified):
```java
@JsonInclude(JsonInclude.Include.NON_NULL)
@RegisterForReflection
public static class RequestInfo {
    private String method;
    private String path;
    private Map<String, String> headers;
    
    public String getMethod() {
        return method;
    }
    
    public void setMethod(String method) {
        this.method = method;
    }
    
    // Repeated for all fields...
}
```

#### Golang Equivalent (estimated 100-150 lines)

**Characteristics**:
- Simple struct definitions with JSON tags
- No getter/setter boilerplate
- Built-in JSON marshaling with struct tags
- Pointer types for optional fields (nil = omit)

**Example** (simplified):
```go
type RequestInfo struct {
    Method        string            `json:"method"`
    Path          string            `json:"path"`
    Query         string            `json:"query,omitempty"`
    Headers       map[string]string `json:"headers"`
    RemoteAddress string            `json:"remoteAddress"`
}

type ServerInfo struct {
    Hostname    string            `json:"hostname"`
    HostAddress string            `json:"hostAddress,omitempty"`
    Environment map[string]string `json:"environment"`
}

type KubernetesInfo struct {
    Namespace   string            `json:"namespace"`
    PodName     string            `json:"podName"`
    PodIP       string            `json:"podIp,omitempty"`
    NodeName    string            `json:"nodeName,omitempty"`
    ServiceHost string            `json:"serviceHost,omitempty"`
    ServicePort string            `json:"servicePort,omitempty"`
    Labels      map[string]string `json:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty"`
}

type EchoResponse struct {
    Request    RequestInfo              `json:"request"`
    Server     ServerInfo               `json:"server"`
    Kubernetes *KubernetesInfo          `json:"kubernetes,omitempty"`
    JwtTokens  map[string]JwtInfo       `json:"jwtTokens,omitempty"`
}
```

### JWT Decoding Comparison

#### Current Java (197 lines)

**Characteristics**:
- Jackson ObjectMapper for JSON parsing
- Manual Base64URL decoding with padding calculation
- Configuration via @ConfigProperty
- Comprehensive error handling with logging

#### Golang Equivalent (estimated 80-120 lines)

**Characteristics**:
- Standard library `encoding/json` for parsing
- Standard library `encoding/base64` for decoding (built-in Base64URL support)
- Environment variables or config file for configuration
- Similar error handling with logging

**Key Advantage**: Go's standard library has built-in Base64URL encoding/decoding, while Java requires manual padding handling.

### HTML Generation Comparison

#### Current Java

- Uses StringBuilder with extensive string concatenation
- Manual HTML escaping function
- ~260 lines for complete HTML generation
- Inline CSS embedded in Java code

#### Golang Equivalent

**Option 1: Template-based** (recommended)
```go
// Define template once
var htmlTemplate = template.Must(template.New("echo").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>{{.PageTitle}}</title>
    <style>{{ .CSS }}</style>
</head>
<body>
    <h1>{{.PageTitle}}</h1>
    {{range .Request.Headers}}
        <tr><th>{{.Key}}</th><td>{{.Value}}</td></tr>
    {{end}}
</body>
</html>
`))

// Render with data
htmlTemplate.Execute(w, response)
```

**Pros**: More maintainable, cleaner separation, auto-escaping
**Lines**: ~50-80 lines for template + rendering logic

**Option 2: String builder** (similar to Java)
- Would be similar in complexity to Java
- Less idiomatic in Go

---

## Pros and Cons

### Pros of Golang Rewrite

#### 1. **Significantly Simpler Codebase**
- **40-60% reduction in lines of code**
- No getter/setter boilerplate (Java: 293 lines → Go: ~100 lines for models)
- No complex annotation-based configuration
- Simpler dependency injection (or none needed)
- More straightforward control flow

#### 2. **Single Binary Deployment**
- Go compiles to a single static binary by default
- No JVM, no framework runtime, no dependencies
- Simpler Docker images (can use `FROM scratch` for minimal size)
- Current Java native image: ~50MB, Go binary: ~10-15MB

#### 3. **Faster Compilation**
- Go: Seconds for full build
- Java/Quarkus: Minutes for native compilation
- Better developer experience during iteration

#### 4. **Lower Resource Requirements**
- Go runtime is lighter than JVM (even Quarkus native)
- Expected memory footprint: 10-20MB vs 30-50MB
- Lower CPU usage at idle

#### 5. **Simpler Build Process**
- No Maven/Gradle complexity
- `go build` is the entire build command
- No wrapper scripts needed
- Built-in cross-compilation for different platforms

#### 6. **Standard Library Richness**
- HTTP server in standard library (production-ready)
- JWT, JSON, base64 encoding all in standard library
- No external dependencies required for core functionality

#### 7. **Easier Maintenance**
- Less code = fewer bugs
- Simpler mental model
- Go's explicit error handling (vs Java exceptions)
- No framework version upgrades (Quarkus 3.x → 4.x)

#### 8. **Better IDE Experience**
- Faster IDE startup and response
- Go's tooling is simpler and faster
- Built-in formatter (`go fmt`), linter (`go vet`)

#### 9. **Language Consistency**
- If other services are in Go, maintains consistency
- Kubernetes/Cloud-native ecosystem is Go-centric
- Easier for Go developers to contribute

#### 10. **Built-in Concurrency**
- Goroutines for concurrent request handling
- Simpler than Java's threading model
- Better performance for I/O-bound operations

### Cons of Golang Rewrite

#### 1. **Loss of Quarkus Ecosystem**
- No built-in dev mode with hot reload
- No automatic configuration binding
- No built-in health check standards (MicroProfile)
- Would need to implement or find libraries for these features

#### 2. **Development Effort**
- Complete rewrite: 3-5 days of development + testing
- Need to rewrite all tests
- Documentation updates required
- Risk of introducing new bugs during migration

#### 3. **Native Compilation Already Solved**
- Current solution already achieves fast startup (<100ms)
- Native image size already small (~50MB)
- Go's advantage here is marginal for this specific use case

#### 4. **Quarkus Features Loss**
- Built-in OpenAPI/Swagger generation
- Quarkus extensions ecosystem
- MicroProfile standards compliance
- Built-in fault tolerance patterns
- Integrated testing support (@QuarkusTest)

#### 5. **Java Expertise**
- Team may have more Java experience
- Existing Java tooling and CI/CD pipelines
- Corporate Java standards and compliance

#### 6. **Less Sophisticated Metrics**
- Prometheus client in Go is good but less integrated than Quarkus Micrometer
- Would need manual timer/counter management
- No automatic JVM metrics

#### 7. **Error Handling Verbosity**
- Go's explicit error handling can be verbose
- `if err != nil` checks after every operation
- Java's try-catch can be more concise for some patterns

#### 8. **No Framework Magic**
- Content negotiation must be manual
- No automatic parameter injection
- More boilerplate for request parsing
- However, this is also a Pro (explicit is better than implicit)

#### 9. **Testing Sophistication**
- REST Assured (Java) is more feature-rich than Go's httptest
- Quarkus testing annotations are convenient
- Go tests are simpler but less integrated

#### 10. **Migration Risk**
- Risk of feature parity issues during migration
- Need to retest all functionality
- Potential for subtle bugs in edge cases
- Documentation might miss some details

---

## Feature Parity Analysis

Can all current features be implemented in Go? **Yes, with complete parity.**

| Feature | Current Java/Quarkus | Golang Equivalent | Complexity |
|---------|---------------------|-------------------|------------|
| All HTTP Methods | ✅ JAX-RS annotations | ✅ Standard library or router | Similar/Easier |
| JSON Responses | ✅ Jackson | ✅ encoding/json | Similar |
| HTML Responses | ✅ StringBuilder | ✅ html/template | Easier |
| JWT Decoding | ✅ Manual implementation | ✅ Standard library or jwt-go | Similar |
| Prometheus Metrics | ✅ Micrometer | ✅ prometheus/client_golang | Similar |
| Health Checks | ✅ MicroProfile Health | ✅ Manual endpoints | Simpler |
| Kubernetes Metadata | ✅ Environment variables | ✅ Environment variables | Identical |
| Custom Status Codes | ✅ Header parsing | ✅ Header parsing | Identical |
| Content Negotiation | ✅ JAX-RS @Produces | ✅ Manual header check | Similar |
| Request Info | ✅ HttpServerRequest | ✅ http.Request | Identical |
| HTML Escaping | ✅ Manual function | ✅ html/template (built-in) | Easier |
| Configuration | ✅ MicroProfile Config | ✅ Environment vars / viper | Simpler |
| Native Compilation | ✅ GraalVM (complex) | ✅ Built-in (simple) | Much Easier |
| Docker Build | ✅ Multi-stage | ✅ Multi-stage | Simpler |

**Verdict**: All features have straightforward Go equivalents. Some are simpler in Go (HTML templating, native compilation), others are similar in complexity.

---

## Performance Comparison

### Startup Time

| Metric | Java/Quarkus JVM | Java/Quarkus Native | Golang |
|--------|------------------|---------------------|--------|
| Cold Start | 2-3 seconds | <100ms | 10-50ms |
| Hot Reload | 1-2 seconds | N/A | 1-2 seconds (with air/fresh) |

**Winner**: Golang (slightly faster cold start, but Quarkus native is already excellent)

### Memory Footprint

| Metric | Java/Quarkus JVM | Java/Quarkus Native | Golang |
|--------|------------------|---------------------|--------|
| Idle Memory | 150-200MB | 30-50MB | 10-20MB |
| Under Load | 200-400MB | 50-100MB | 20-40MB |

**Winner**: Golang (2-3x lower memory usage)

### Request Latency

Based on typical benchmarks for similar echo servers:

| Metric | Java/Quarkus JVM | Java/Quarkus Native | Golang (stdlib) | Golang (Fiber) |
|--------|------------------|---------------------|-----------------|----------------|
| p50 Latency | <1ms | <1ms | <1ms | <1ms |
| p99 Latency | 2-5ms | 2-5ms | 1-3ms | 1-2ms |
| Throughput | 50k req/s | 60k req/s | 70k req/s | 100k req/s |

**Winner**: Golang (especially with optimized frameworks like Fiber), but all are excellent for this use case

### Build Time

| Operation | Java/Quarkus | Golang |
|-----------|--------------|--------|
| Full Build (JVM) | 30-60 seconds | 5-10 seconds |
| Full Build (Native) | 3-5 minutes | 5-10 seconds |
| Incremental | 10-20 seconds | 2-5 seconds |

**Winner**: Golang (significantly faster builds)

### Binary Size

| Artifact | Java/Quarkus JVM | Java/Quarkus Native | Golang | Golang (UPX compressed) |
|----------|------------------|---------------------|--------|-------------------------|
| Binary | N/A (needs JVM) | ~50MB | 10-15MB | 5-8MB |
| Docker Image | 200-300MB | 50-80MB | 15-25MB | 10-15MB |

**Winner**: Golang (3-5x smaller artifacts)

**Performance Summary**: Golang has advantages in memory footprint, binary size, and build time. Runtime performance is comparable, with Golang having a slight edge in throughput. However, **for this specific echo server use case, the performance differences are negligible** – all options are more than fast enough.

---

## Ecosystem and Tooling

### Build and Development Tools

| Tool | Java/Quarkus | Golang |
|------|-------------|--------|
| Build Tool | Maven (complex) | go build (simple) |
| Package Manager | Maven (XML) | Go modules (simple) |
| Formatter | Maven plugin | go fmt (built-in) |
| Linter | Checkstyle, SpotBugs | golangci-lint |
| Testing | JUnit, REST Assured | testing package, httptest |
| Coverage | JaCoCo | go test -cover |
| Dev Mode | quarkus:dev (excellent) | Manual or air/fresh |
| Hot Reload | Built-in | Requires tool |

### Dependency Management

**Java/Quarkus**:
- 3 direct dependencies (Quarkus BOM)
- ~50+ transitive dependencies
- Maven dependency:tree for visibility
- Potential for version conflicts

**Golang**:
- 0-3 direct dependencies (can use only stdlib)
- ~10-20 transitive dependencies (if using frameworks)
- go.mod for explicit versioning
- Minimal dependency conflicts (module system)

### Container Ecosystem

**Java/Quarkus**:
- Multi-stage build required for native
- Base images: Mandrel builder (~1GB), Quarkus micro (~30MB)
- Total build time: 5-10 minutes for native

**Golang**:
- Single or multi-stage build
- Base images: golang:alpine (~300MB) or scratch (~0MB)
- Total build time: 1-2 minutes

### Cloud-Native Ecosystem Alignment

**Golang Advantages**:
- Kubernetes is written in Go
- Most CNCF projects are Go-based
- Better integration with K8s client libraries
- Common language for DevOps/SRE teams

**Java Advantages**:
- MicroProfile standards (portable across vendors)
- Enterprise adoption
- More mature observability integrations

---

## Migration Effort Estimate

### Development Tasks

| Task | Estimated Time | Complexity |
|------|---------------|------------|
| Project setup (go.mod, structure) | 1 hour | Low |
| HTTP handlers (echo endpoint) | 4-6 hours | Medium |
| Data models | 2 hours | Low |
| JWT decoding service | 2-3 hours | Low |
| HTML template implementation | 3-4 hours | Medium |
| Kubernetes metadata gathering | 1 hour | Low |
| Metrics integration (Prometheus) | 2-3 hours | Medium |
| Health check endpoints | 1 hour | Low |
| Configuration management | 1-2 hours | Low |
| Unit tests | 4-6 hours | Medium |
| Integration tests | 3-4 hours | Medium |
| Docker build optimization | 2 hours | Low |
| **Total Development** | **26-34 hours** | **~3-5 days** |

### Additional Tasks

| Task | Estimated Time |
|------|---------------|
| Documentation updates (README, USAGE_GUIDE, etc.) | 4-6 hours |
| CI/CD pipeline updates | 2-4 hours |
| Kubernetes manifests (minimal changes) | 1 hour |
| Performance testing and optimization | 4-6 hours |
| Code review and refinement | 4-6 hours |
| **Total Additional** | **15-23 hours** |

### Total Migration Effort

- **Development + Testing**: 26-34 hours (3-5 days)
- **Documentation + Infrastructure**: 15-23 hours (2-3 days)
- **Total**: **41-57 hours (5-8 days)**

### Risk Factors

1. **Learning Curve**: If team is unfamiliar with Go (+20-40% time)
2. **Hidden Complexity**: Edge cases not captured in current tests (+10-20% time)
3. **Performance Tuning**: If specific optimizations needed (+10-15% time)
4. **Documentation Parity**: Ensuring all docs are updated (+10-20% time)

**With Risk Buffer**: 50-85 hours (7-11 days)

---

## Recommendations

### Recommendation: **Conditional Proceed**

The decision to rewrite in Golang should be based on strategic priorities rather than technical necessity.

### ✅ Rewrite in Golang IF:

1. **Simplicity is a priority**
   - You value reduced code complexity over framework features
   - Easier maintenance is more important than advanced features

2. **You're building a Go-based ecosystem**
   - Other services are in Go
   - Team has Go expertise or wants to develop it
   - Consistency across services is valuable

3. **You want minimal dependencies**
   - Current Maven + Quarkus + GraalVM stack feels heavy
   - Simpler build and deployment pipeline desired

4. **You have time for the migration**
   - Can allocate 1-2 weeks for development + testing
   - Can handle the documentation updates

5. **Binary size and memory matter significantly**
   - Running many instances where 2-3x memory savings matter
   - Storage costs for images are a concern

6. **You prefer explicit over implicit**
   - Dislike annotation-driven "magic"
   - Want clearer control flow

### ❌ Stay with Java/Quarkus IF:

1. **Current solution works well**
   - No performance or maintenance issues
   - Team is productive with Java

2. **Quarkus features are valuable**
   - Use dev mode extensively
   - Benefit from MicroProfile standards
   - Plan to add more Quarkus extensions

3. **Java expertise is strong**
   - Team has deep Java knowledge
   - Corporate standards require Java
   - Existing CI/CD optimized for Java

4. **Risk aversion**
   - Can't afford potential bugs during migration
   - Documentation accuracy is critical
   - Don't want to retest everything

5. **Native compilation already solves startup/size**
   - Current performance is satisfactory
   - Marginal improvements don't justify effort

### Hybrid Approach: **Start Fresh Projects in Go**

A compromise strategy:
- Keep current echo-server in Java/Quarkus (it works well)
- Use Go for new services going forward
- Gradually build Go expertise
- Revisit echo-server rewrite in 6-12 months when:
  - Team has Go experience
  - Go codebase patterns are established
  - Clear benefit is demonstrated

---

## Sample Golang Implementation Snippet

To provide a concrete sense of the code, here's what the main handler might look like:

```go
package handlers

import (
    "encoding/json"
    "html/template"
    "net/http"
    "os"
    "strings"
    "time"
    
    "echo-server/models"
    "echo-server/services"
)

var htmlTemplate = template.Must(template.ParseFiles("templates/echo.html"))

func EchoHandler(jwtService *services.JWTService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        startTime := time.Now()
        defer recordMetrics(r.Method, r.URL.Path, time.Since(startTime))
        
        // Build response
        response := models.EchoResponse{
            Request:    buildRequestInfo(r),
            Server:     buildServerInfo(),
            Kubernetes: getKubernetesInfo(),
            JwtTokens:  jwtService.ExtractAndDecodeJWTs(r.Header),
        }
        
        // Get custom status code
        statusCode := http.StatusOK
        if customStatus := r.Header.Get("x-set-response-status-code"); customStatus != "" {
            if code, err := strconv.Atoi(customStatus); err == nil && code >= 200 && code <= 599 {
                statusCode = code
            }
        }
        
        // Content negotiation
        acceptHeader := r.Header.Get("Accept")
        if strings.Contains(acceptHeader, "text/html") {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.WriteHeader(statusCode)
            htmlTemplate.Execute(w, response)
        } else {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(statusCode)
            json.NewEncoder(w).Encode(response)
        }
    }
}

func buildRequestInfo(r *http.Request) models.RequestInfo {
    headers := make(map[string]string)
    for key, values := range r.Header {
        headers[key] = strings.Join(values, ", ")
    }
    
    return models.RequestInfo{
        Method:        r.Method,
        Path:          r.URL.Path,
        Query:         r.URL.RawQuery,
        Headers:       headers,
        RemoteAddress: getRemoteAddress(r),
    }
}

func getKubernetesInfo() *models.KubernetesInfo {
    namespace := os.Getenv("K8S_NAMESPACE")
    podName := os.Getenv("K8S_POD_NAME")
    
    if namespace == "" || podName == "" {
        return nil
    }
    
    return &models.KubernetesInfo{
        Namespace:   namespace,
        PodName:     podName,
        PodIP:       os.Getenv("K8S_POD_IP"),
        NodeName:    os.Getenv("K8S_NODE_NAME"),
        ServiceHost: os.Getenv("KUBERNETES_SERVICE_HOST"),
        ServicePort: os.Getenv("KUBERNETES_SERVICE_PORT"),
        Labels:      getEnvWithPrefix("K8S_LABEL_"),
        Annotations: getEnvWithPrefix("K8S_ANNOTATION_"),
    }
}
```

**Key Observations**:
- Much more concise than Java equivalent
- No annotations, explicit control flow
- Standard library types throughout
- Error handling is explicit but straightforward

---

## Conclusion

### Technical Verdict

Golang would provide a **simpler, lighter, faster-to-build** alternative with comparable runtime performance. The rewrite would reduce code complexity by 40-60% and simplify the build/deployment pipeline significantly.

### Strategic Verdict

The rewrite is **technically feasible and beneficial** but **not urgently necessary**. The current Java/Quarkus implementation is:
- ✅ Performant (native compilation already fast)
- ✅ Feature-complete
- ✅ Well-tested
- ✅ Well-documented
- ✅ Production-ready

### Decision Framework

**Choose Golang** if your priorities are:
1. Code simplicity and maintainability
2. Minimal dependencies and single-binary deployment
3. Ecosystem consistency (if building Go-heavy infrastructure)
4. Lower resource usage across many instances

**Stay with Java/Quarkus** if your priorities are:
1. Leveraging existing team expertise
2. Using Quarkus ecosystem features
3. Maintaining stability and avoiding migration risk
4. Following enterprise Java standards

### Final Recommendation

**Recommended Path**: 
1. **Keep the current implementation** for now – it works well and is mature
2. **Prototype a minimal Go version** (2-3 days) to validate assumptions
3. **Decide based on the prototype** whether full migration is worth the effort
4. If proceeding, **migrate in phases**: core functionality → testing → documentation → deployment

The good news: **You can't go wrong either way.** Both Java/Quarkus and Golang are excellent choices for this type of cloud-native service.

---

## Appendix: Alternative Frameworks Considered

### Go Frameworks

1. **Standard Library (`net/http`)**
   - Pros: Zero dependencies, stable, sufficient for this use case
   - Cons: Manual routing, no middleware out-of-box
   - **Recommendation**: Best choice for simplicity

2. **Gin**
   - Pros: Popular, good performance, rich middleware
   - Cons: Extra dependency, slightly more complex
   - **Recommendation**: Good if you want framework features

3. **Echo**
   - Pros: Very similar name to this project!, good performance
   - Cons: Another dependency
   - **Recommendation**: Could be confusing naming

4. **Fiber**
   - Pros: Best performance benchmarks, Express-like API
   - Cons: Uses fasthttp (not standard library), bleeding edge
   - **Recommendation**: Only if max performance needed

### Verdict on Frameworks

For this project, **standard library `net/http` is recommended**. The routing is simple (wildcard paths), and we don't need complex middleware. Adding a framework would add complexity without significant benefit.

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-20  
**Author**: Research for Echo Server Golang Investigation
