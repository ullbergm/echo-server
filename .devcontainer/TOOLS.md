# Dev Container Tools Reference

This document lists all tools pre-installed in the Echo Server dev container.

## 🛠️ Core Development Tools

| Tool | Version | Purpose | Command |
|------|---------|---------|---------|
| **Go** | 1.25.x | Programming language | `go version` |
| **Git** | 2.x | Version control | `git --version` |
| **GitHub CLI** | Latest | GitHub integration | `gh --version` |
| **Docker** | Latest | Container runtime | `docker --version` |
| **Task** | 3.x | Task runner | `task --version` |
| **Node.js** | LTS (20.x) | JavaScript runtime | `node --version` |
| **npm** | Latest | Node package manager | `npm --version` |
| **Python** | 3.11 | Python runtime | `python3 --version` |
| **pip** | Latest | Python package manager | `pip --version` |

## 📋 Code Quality Tools

| Tool | Installation | Purpose | Usage |
|------|-------------|---------|-------|
| **golangci-lint** | Via `go run` | Go linter (40+ linters) | `task lint` |
| **gosec** | Via `go run` | Go security scanner | `task security` |
| **gofmt** | Built-in | Go code formatter | `gofmt -w .` |
| **goimports** | Auto-installed | Import organizer | Auto on save |
| **pre-commit** | pip install | Git hook framework | `pre-commit run` |
| **hadolint** | Binary download | Dockerfile linter | `hadolint Dockerfile` |
| **yamllint** | pip install | YAML linter | `yamllint .` |
| **markdownlint-cli** | npm global | Markdown linter | `markdownlint *.md` |

## ✅ Commit Tools

| Tool | Installation | Purpose | Usage |
|------|-------------|---------|-------|
| **commitlint** | npm global | Commit message linter | `task commit-lint` |
| **conventional-pre-commit** | pre-commit hook | Validates commit format | Auto on commit |

## 🔌 VS Code Extensions

| Extension | ID | Purpose |
|-----------|-----|---------|
| **Go** | golang.go | Go language support |
| **Docker** | ms-azuretools.vscode-docker | Docker support |
| **Kubernetes Tools** | ms-kubernetes-tools.vscode-kubernetes-tools | K8s integration |
| **GitLens** | eamodio.gitlens | Enhanced Git features |
| **EditorConfig** | editorconfig.editorconfig | Editor settings |
| **Code Spell Checker** | streetsidesoftware.code-spell-checker | Spell checking |
| **Markdownlint** | davidanson.vscode-markdownlint | Markdown linting |
| **Task** | task.vscode-task | Task runner integration |
| **YAML** | redhat.vscode-yaml | YAML language support |
| **Hadolint** | exiasr.hadolint | Dockerfile linting |

## 🚀 Quick Verification Commands

Run these commands after container creation to verify everything is installed:

```bash
# Core tools
go version
git --version
gh --version
docker --version
task --version
node --version
npm --version
python3 --version
pip --version

# Code quality tools
pre-commit --version
hadolint --version
yamllint --version
markdownlint --version
npx @commitlint/cli --version

# Go tools (installed on-demand)
go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest version
go run github.com/securego/gosec/v2/cmd/gosec@latest -version

# Pre-commit hooks status
pre-commit run --all-files --verbose
```

## 📦 Tool Installation Methods

### Devcontainer Features (devcontainer.json)

- Node.js LTS
- Python 3.11
- Docker-in-Docker
- Git
- GitHub CLI

### Post-Create Script (post-create.sh)

- Task runner (shell script)
- pre-commit (pip)
- hadolint (binary download)
- yamllint (pip)
- markdownlint-cli (npm)
- commitlint + config-conventional (npm)

### On-Demand (via go run)

- golangci-lint
- gosec
- air (hot reload)

### VS Code Extensions (devcontainer.json)

- All extensions installed automatically

## 🔄 Updating Tools

### Update Node.js packages

```bash
npm update -g markdownlint-cli @commitlint/cli @commitlint/config-conventional
```

### Update Python packages

```bash
pip install --upgrade pre-commit yamllint
```

### Update pre-commit hooks

```bash
task pre-commit-update
# or
pre-commit autoupdate
```

### Update Go tools

Go tools are always latest when using `go run <package>@latest`

### Rebuild container with latest versions

```bash
# In VS Code: F1 → "Dev Containers: Rebuild Container"
```

## 🎯 Environment Variables

The dev container sets these environment variables:

```bash
GO111MODULE=on          # Enable Go modules
CGO_ENABLED=0           # Disable CGO for static binaries
```

## 📝 Additional Resources

- [Dev Container README](.devcontainer/README.md) - Setup and troubleshooting
- [DEVELOPER_SETUP.md](../DEVELOPER_SETUP.md) - Complete development guide
- [Taskfile.yml](../Taskfile.yml) - Available task commands
- [.pre-commit-config.yaml](../.pre-commit-config.yaml) - Pre-commit hook configuration
