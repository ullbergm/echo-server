# Dev Container Setup

This devcontainer provides a complete development environment for the Echo Server project with all tools pre-installed.

## Features

- **Go 1.25** on Debian Bookworm
- **Docker-in-Docker** for building container images
- **Git** and **GitHub CLI** pre-installed
- **Node.js LTS** for JavaScript tooling (commitlint, markdownlint)
- **Python 3.11** with pip for pre-commit hooks
- **Task runner** for task automation
- **Pre-commit hooks** with automatic installation
- **SSH keys mounted** from host machine (read-only)
- **VS Code extensions** for Go development, Docker, Kubernetes, and more
- **Auto-formatting** on save with gofmt
- **Port forwarding** for port 8080 (Echo Server)

## Pre-installed Developer Tools

### Core Development

- **Go 1.25.x** - Go programming language
- **Git 2.x** - Version control
- **GitHub CLI (gh)** - GitHub integration
- **Docker** - Container runtime (Docker-in-Docker)
- **Task** - Task runner for automation

### Code Quality & Linting

- **golangci-lint** - Go linter (installed via go run)
- **gosec** - Go security checker (installed via go run)
- **pre-commit** - Git hook framework
- **hadolint** - Dockerfile linter
- **yamllint** - YAML linter
- **markdownlint-cli** - Markdown linter

### Commit Tools

- **commitlint** - Commit message linter
- **conventional-pre-commit** - Conventional commits validation

## Included VS Code Extensions

- **golang.go** - Go language support with IntelliSense, formatting, and linting
- **ms-azuretools.vscode-docker** - Docker support
- **ms-kubernetes-tools.vscode-kubernetes-tools** - Kubernetes tools
- **eamodio.gitlens** - Enhanced git integration
- **editorconfig.editorconfig** - EditorConfig support
- **streetsidesoftware.code-spell-checker** - Spell checker
- **davidanson.vscode-markdownlint** - Markdown linting
- **task.vscode-task** - Task runner integration
- **redhat.vscode-yaml** - YAML language support
- **exiasr.hadolint** - Dockerfile linting

## SSH Key Mounting

Your SSH keys from `~/.ssh` (or `%USERPROFILE%\.ssh` on Windows) are automatically mounted into the container at `/home/vscode/.ssh` as read-only. This allows you to:

- Clone private repositories
- Push to GitHub/GitLab without re-configuring SSH
- Use SSH-based authentication seamlessly

## Getting Started

1. **Open in Dev Container**:
   - Open the project in VS Code
   - Press `F1` and select "Dev Containers: Reopen in Container"
   - Wait for the container to build and initialize
   - The post-create script will automatically install all tools and pre-commit hooks

2. **Verify Setup**:

   ```bash
   go version           # Should show Go 1.25.x
   node --version       # Should show Node LTS
   python3 --version    # Should show Python 3.11.x
   pre-commit --version # Should show pre-commit version
   task --version       # Should show Task 3.x
   
   # Run tests
   task test
   
   # Check pre-commit hooks
   pre-commit run --all-files
   ```

3. **Development Workflow**:

   ```bash
   # Run with hot reload
   task dev
   
   # Or use standard commands
   task build        # Build binary
   task test         # Run tests
   task lint         # Run linters
   task pre-commit-run  # Run all pre-commit checks
   
   # Make a commit (pre-commit hooks run automatically)
   git commit -m "feat(api): add new endpoint"
   ```

## Port Forwarding

Port 8080 is automatically forwarded when the application runs. You can access it at:

- <http://localhost:8080>

## Pre-commit Hooks

Pre-commit hooks are automatically installed during container setup. They run on every commit to ensure code quality:

### What hooks run

- ✅ Code formatting (gofmt)
- ✅ Linting (golangci-lint)
- ✅ Tests with race detection
- ✅ Go module tidiness
- ✅ YAML/Markdown linting
- ✅ Dockerfile linting
- ✅ Commit message validation
- ✅ Secret detection
- ✅ File hygiene checks

### Manual usage

```bash
# Run all hooks manually
task pre-commit-run

# Run specific hook
pre-commit run golangci-lint --all-files

# Skip hooks (not recommended)
git commit --no-verify -m "message"
```

## Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>
```

**Examples:**

```bash
feat(handlers): add custom status code support
fix(jwt): handle missing authorization header
docs: update README with TLS examples
```

Commit messages are automatically validated by pre-commit hooks. See [COMMIT_CONVENTION.md](../COMMIT_CONVENTION.md) for details.

## Environment Variables

The container sets:

- `GO111MODULE=on` - Enable Go modules
- `CGO_ENABLED=0` - Disable CGO for static binaries

## Customization

To add more tools or modify the environment:

1. Edit `.devcontainer/devcontainer.json` to add features or VS Code extensions
2. Edit `.devcontainer/post-create.sh` to add tools or scripts
3. Rebuild the container: `F1` → "Dev Containers: Rebuild Container"

## Troubleshooting

### Pre-commit Hooks Not Working

If pre-commit hooks aren't running:

```bash
# Reinstall hooks
task pre-commit-install

# Or manually
pre-commit install
pre-commit install --hook-type commit-msg
```

### SSH Keys Not Working

Ensure your SSH keys exist at:

- Linux/Mac: `~/.ssh/`
- Windows: `%USERPROFILE%\.ssh\`

The container mounts these as read-only, so they won't be modified.

### Port Already in Use

If port 8080 is already in use on your host:

1. Stop any running echo-server instances
2. Or modify `forwardPorts` in `devcontainer.json`

### Node.js/npm Not Found

If Node.js tools aren't available after container creation:

```bash
# Verify Node.js installation
node --version
npm --version

# If npm is available but tools weren't installed, run:
bash .devcontainer/install-npm-tools.sh

# If Node.js itself is missing, rebuild the container:
# F1 → "Dev Containers: Rebuild Container"
```

### Python/pip Not Found

If Python tools aren't available:

```bash
# Verify Python installation
python3 --version
pip --version

# Install pip if missing
curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
python3 get-pip.py --user
```

### Go Tools Not Installing

Run manually inside the container:

```bash
go install golang.org/x/tools/gopls@latest
go install github.com/air-verse/air@latest
```
