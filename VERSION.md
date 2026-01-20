# Echo Server Version Management

The Echo Server version is dynamically injected at build time using Go's `-ldflags` feature. This avoids hardcoding the version in the source code.

## How It Works

The version is stored in a variable in `main.go`:

```go
// Version is set via ldflags at build time
var Version = "dev"
```

At build time, the value is replaced using:
```bash
go build -ldflags "-X main.Version=YOUR_VERSION"
```

## Build Methods

### 1. Using Make (Recommended)

The Makefile automatically extracts the version from git:

```bash
# Build with version from git
make build

# Run with version from git
make run

# See what version would be used
make version
```

**Version resolution priority:**
1. Latest git tag + commit (e.g., `v1.2.3-5-gabcdef`)
2. Current commit hash (if no tags exist)
3. "dev" (if not in a git repository)

### 2. Manual Build

Specify any version manually:

```bash
# With specific version
go build -ldflags "-X main.Version=1.0.0" -o echo-server

# With git tag
VERSION=$(git describe --tags --always --dirty)
go build -ldflags "-X main.Version=$VERSION" -o echo-server

# Development build (uses default "dev")
go build -o echo-server
```

### 3. Docker Build

The Dockerfile automatically injects the git version:

```bash
docker build -t echo-server:latest .
```

## Tagging Releases

To create a versioned release:

```bash
# Create and push a tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# Build with the new version
make build
./echo-server
```

## Version Display

The version appears in:
- Application startup logs
- Fiber app name: `Echo Server {VERSION}`
- HTML template footer
- Response metadata (via `Version` field)

## CI/CD Integration

For automated builds, inject the version from your CI system:

```bash
# GitHub Actions
go build -ldflags "-X main.Version=${{ github.ref_name }}"

# GitLab CI
go build -ldflags "-X main.Version=$CI_COMMIT_TAG"

# Generic CI with version file
VERSION=$(cat VERSION)
go build -ldflags "-X main.Version=$VERSION"
```
