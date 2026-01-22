# Developer Setup Guide

This guide will help you set up your development environment for the echo-server project.

## Quick Start Options

### Option 1: Dev Container (Recommended - Zero Setup!)

**GitHub Codespaces:**

1. Click "Code" → "Create codespace on main" in GitHub
2. Wait for the environment to load (1-2 minutes)
3. All tools are pre-installed and configured!
4. Start coding: `task run`

**VS Code Dev Container:**

1. Install [Docker](https://docs.docker.com/get-docker/) and VS Code with [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. Clone the repository
3. Open in VS Code and press `F1` → "Dev Containers: Reopen in Container"
4. Wait for container setup (3-5 minutes on first run)
5. All tools pre-installed! Start with `task run`

✅ **Pre-installed in dev container:**

- Go 1.25, Node.js LTS, Python 3.11
- Task, pre-commit, commitlint, hadolint, yamllint, markdownlint
- golangci-lint, gosec (via go run)
- All VS Code extensions
- Pre-commit hooks auto-configured

See [.devcontainer/README.md](.devcontainer/README.md) for details.

### Option 2: Local Setup (Manual Installation)

## Prerequisites (Local Setup)

### Required

- **Go 1.25+**: [Download and install Go](https://go.dev/dl/)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **Task**: [Install Task](https://taskfile.dev/installation/)

### Optional but Recommended

- **Docker**: For container testing - [Install Docker](https://docs.docker.com/get-docker/)
- **Python 3.7+**: For pre-commit hooks - [Install Python](https://www.python.org/downloads/)
- **Node.js 18+**: For commitlint - [Install Node.js](https://nodejs.org/)

## Initial Setup

### 1. Clone the Repository

```bash
git clone https://github.com/ullbergm/echo-server.git
cd echo-server
```

### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Verify dependencies
task deps-verify
```

### 3. Install Pre-commit Hooks (Optional but Recommended)

Pre-commit hooks ensure code quality before each commit:

```bash
# Install pre-commit (choose one method)

# On macOS:
brew install pre-commit

# On Linux/Windows with pip:
pip install pre-commit
# or
pip3 install pre-commit

# Verify installation
pre-commit --version

# Install git hooks
task pre-commit-install
```

### 4. Verify Setup

```bash
# Run tests to verify everything works
task test

# Build the application
task build

# Run the server
task run
```

Visit <http://localhost:8080> to verify the server is running.

## Development Workflow

### Day-to-Day Development

```bash
# 1. Start development server with hot reload
task dev

# 2. Make your changes to the code

# 3. Run tests frequently
task test

# 4. Check code quality before committing
task pre-commit-run

# 5. Commit with conventional commit format
git commit -m "feat(handlers): add new endpoint"
```

### Code Quality Checks

The project has several code quality tools:

```bash
# Format code
task fmt

# Run linter
task lint

# Fix linting issues automatically
task lint-fix

# Run security checks
task security

# Run all checks (like CI)
task ci
```

### Testing

```bash
# Run all tests
task test

# Run tests with coverage
task test-coverage

# View coverage report in browser
task coverage-view

# Run tests with race detection
task test-race

# Run benchmarks
task bench
```

### Pre-commit Hooks

If you installed pre-commit hooks, they run automatically on `git commit`:

#### What hooks run

1. **Code formatting** - Formats Go code with `gofmt`
2. **Linting** - Runs `golangci-lint` with auto-fix
3. **Tests** - Runs tests with race detection
4. **Dependencies** - Verifies `go.mod` and runs `go mod tidy`
5. **YAML/Markdown** - Lints configuration and documentation files
6. **Dockerfile** - Lints Dockerfile with hadolint
7. **Commit message** - Validates conventional commit format
8. **Secrets** - Detects potential secrets in code
9. **File checks** - Removes trailing whitespace, fixes line endings

#### Manual usage

```bash
# Run all hooks on all files
task pre-commit-run

# Run specific hook
pre-commit run golangci-lint --all-files

# Skip hooks (not recommended)
git commit --no-verify -m "emergency fix"

# Update hook versions
task pre-commit-update
```

### Commit Messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

#### Format

```
<type>(<scope>): <subject>
```

#### Examples

```bash
feat(handlers): add custom status code support
fix(jwt): handle missing authorization header gracefully
docs: update README with TLS configuration examples
test(services): add body parser edge cases
refactor(models): simplify response structure
perf(handlers): optimize JSON encoding
```

#### Validate commit message

```bash
task commit-lint -- "feat(api): add new endpoint"
```

See [COMMIT_CONVENTION.md](COMMIT_CONVENTION.md) for detailed guidelines.

### Docker Development

```bash
# Build Docker image
task docker-build

# Run container locally
task docker-run

# Build and run in one command
task docker-all

# Access the server
curl http://localhost:8080
```

### Kubernetes Testing

```bash
# Apply Kubernetes manifests
kubectl apply -f deploy/kubernetes/

# Check pod status
kubectl get pods

# View logs
kubectl logs -f deployment/echo-server

# Test the service
kubectl port-forward service/echo-server 8080:80
curl http://localhost:8080
```

## IDE Setup

### VS Code (Recommended)

The project includes VS Code configuration in `.vscode/settings.json`:

**Recommended extensions:**

- Go (golang.go)
- Task (task.vscode-task)
- YAML (redhat.vscode-yaml)
- Docker (ms-azuretools.vscode-docker)
- GitLens (eamodio.gitlens)

**Install extensions:**

```bash
code --install-extension golang.go
code --install-extension task.vscode-task
code --install-extension redhat.vscode-yaml
code --install-extension ms-azuretools.vscode-docker
code --install-extension eamodio.gitlens
```

### GoLand / IntelliJ IDEA

1. Open the project directory
2. GoLand will automatically detect the Go project
3. Enable Go modules: `Settings → Go → Go Modules → Enable Go modules integration`
4. Install Task plugin from marketplace
5. Configure File Watchers for `gofmt` and `goimports`

## Troubleshooting

### Pre-commit hooks fail to install

**Problem:** `task pre-commit-install` fails with "pre-commit: command not found"

**Solution:**

```bash
# Install pre-commit
pip install pre-commit
# or on macOS
brew install pre-commit

# Then retry
task pre-commit-install
```

### Tests fail with "race detector not available"

**Problem:** `task test-race` fails

**Solution:**

```bash
# Ensure CGO is enabled
export CGO_ENABLED=1

# Retry
task test-race
```

### Docker build fails

**Problem:** Docker build fails with network or permission errors

**Solution:**

```bash
# Check Docker is running
docker ps

# Rebuild without cache
docker build --no-cache -t echo-server:latest .
```

### Go module issues

**Problem:** Missing dependencies or module errors

**Solution:**

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Verify dependencies
task deps-verify

# Update dependencies
task deps-update
```

### Linter fails with "unknown configuration option"

**Problem:** golangci-lint fails with configuration errors

**Solution:**

```bash
# Update golangci-lint to latest version
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Or let Task handle it (uses latest automatically)
task lint
```

## Additional Resources

- [Go Documentation](https://go.dev/doc/)
- [Fiber Documentation](https://docs.gofiber.io/)
- [Task Documentation](https://taskfile.dev/)
- [Pre-commit Documentation](https://pre-commit.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Project README](README.md)
- [Commit Convention](COMMIT_CONVENTION.md)

## Getting Help

- Check the [README.md](README.md) for usage information
- Review existing [GitHub Issues](https://github.com/ullbergm/echo-server/issues)
- Create a new issue for bugs or feature requests
- Review [COMMIT_CONVENTION.md](COMMIT_CONVENTION.md) for commit guidelines

## Quick Reference

### Most Common Commands

```bash
task run               # Run the server
task dev               # Run with hot reload
task test              # Run tests
task lint              # Run linter
task build             # Build binary
task pre-commit-run    # Run all pre-commit checks
task docker-build      # Build Docker image
task help              # Show detailed help
```

### Environment Variables

```bash
# Server configuration
PORT=8080              # HTTP port (default: 8080)
TLS_ENABLED=true       # Enable HTTPS (default: false)
TLS_PORT=8443          # HTTPS port (default: 8443)
TLS_CERT_FILE=/path    # Custom TLS certificate
TLS_KEY_FILE=/path     # Custom TLS key
READINESS_DELAY=5s     # Startup delay for K8s readiness

# JWT configuration
JWT_HEADER_NAMES=Authorization,X-JWT-Token  # Headers to check for JWT

# Development
AIR_LOG=1              # Enable air logs in dev mode
```
