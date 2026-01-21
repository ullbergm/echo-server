# Dev Container Setup

This devcontainer provides a complete development environment for the Echo Server project.

## Features

- **Go 1.25** on Debian Bookworm
- **Docker-in-Docker** for building container images
- **Git** and **GitHub CLI** pre-installed
- **SSH keys mounted** from host machine (read-only)
- **VS Code extensions** for Go development, Docker, Kubernetes, and more
- **Auto-formatting** on save with gofmt
- **Port forwarding** for port 8080 (Echo Server)

## Included VS Code Extensions

- Go language support with IntelliSense, formatting, and linting
- Docker and Kubernetes tools
- GitLens for enhanced git integration
- EditorConfig support
- Spell checker and Markdown linting

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

2. **Verify Setup**:
   ```bash
   go version        # Should show Go 1.25.x
   make test         # Run tests
   make run          # Run the application
   ```

3. **Development Workflow**:
   ```bash
   # Install air for hot reload (already done in postCreateCommand)
   air               # Run with hot reload
   
   # Or use make commands
   make build        # Build binary
   make test         # Run tests
   make lint         # Run linters
   ```

## Port Forwarding

Port 8080 is automatically forwarded when the application runs. You can access it at:
- http://localhost:8080

## Environment Variables

The container sets:
- `GO111MODULE=on` - Enable Go modules
- `CGO_ENABLED=0` - Disable CGO for static binaries

## Customization

To add more tools or modify the environment:
1. Edit `.devcontainer/devcontainer.json`
2. Add features, extensions, or modify settings
3. Rebuild the container: `F1` → "Dev Containers: Rebuild Container"

## Troubleshooting

### SSH Keys Not Working

Ensure your SSH keys exist at:
- Linux/Mac: `~/.ssh/`
- Windows: `%USERPROFILE%\.ssh\`

The container mounts these as read-only, so they won't be modified.

### Port Already in Use

If port 8080 is already in use on your host:
1. Stop any running echo-server instances
2. Or modify `forwardPorts` in `devcontainer.json`

### Go Tools Not Installing

Run manually inside the container:
```bash
go install golang.org/x/tools/gopls@latest
go install github.com/air-verse/air@latest
```
