#!/bin/bash
set -e

echo "🚀 Setting up Echo Server development environment..."

# Copy SSH keys from read-only mount to user directory with correct permissions
mkdir -p /home/vscode/.ssh
if [ -d /.ssh-mount ] && [ "$(ls -A /.ssh-mount 2>/dev/null)" ]; then
    echo "📋 Copying SSH keys..."
    cp -r /.ssh-mount/. /home/vscode/.ssh/
    chmod 700 /home/vscode/.ssh
    # Set permissions for public keys
    find /home/vscode/.ssh -type f -name '*.pub' -exec chmod 644 {} \;
    # Set permissions for private keys
    find /home/vscode/.ssh -type f ! -name '*.pub' ! -name 'known_hosts*' ! -name 'config' ! -name 'authorized_keys' -exec chmod 600 {} \;
fi

# Configure Git to use SSH signing
echo "🔐 Configuring Git signing..."
git config --global user.signingkey /home/vscode/.ssh/id_rsa
git config --global gpg.format ssh
git config --global commit.gpgsign true
git config --global gpg.ssh.program ssh-keygen

# Install Task if not present
if ! command -v task &> /dev/null; then
    echo "📦 Installing Task runner..."
    sudo sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
fi

# Install pre-commit
echo "🎣 Installing pre-commit..."
pip install --user pre-commit

# Install hadolint for Dockerfile linting
echo "🐳 Installing hadolint..."
HADOLINT_VERSION="2.12.0"
sudo wget -O /usr/local/bin/hadolint "https://github.com/hadolint/hadolint/releases/download/v${HADOLINT_VERSION}/hadolint-Linux-x86_64"
sudo chmod +x /usr/local/bin/hadolint

# Install yamllint
echo "📝 Installing yamllint..."
pip install --user yamllint

# Wait for Node.js/npm to be available (devcontainer features may still be installing)
echo "⏳ Waiting for Node.js/npm..."
for i in {1..30}; do
    if command -v npm &> /dev/null; then
        echo "✓ Node.js/npm is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "⚠️  Warning: npm not found after 30 seconds. Skipping npm packages."
        echo "   You can install them manually later with:"
        echo "   npm install -g markdownlint-cli @commitlint/cli @commitlint/config-conventional"
        NPM_AVAILABLE=false
    fi
    sleep 1
done

# Install markdownlint-cli
if [ "${NPM_AVAILABLE}" != "false" ]; then
    echo "📄 Installing markdownlint-cli..."
    npm install -g markdownlint-cli

    # Install commitlint
    echo "✅ Installing commitlint..."
    npm install -g @commitlint/cli @commitlint/config-conventional
fi

# Verify Go dependencies
echo "🔍 Verifying Go dependencies..."
if [ -f "Taskfile.yml" ]; then
    task deps-verify || true
elif [ -f "Makefile" ]; then
    make deps-verify || true
else
    go mod verify || true
fi

# Install pre-commit hooks
echo "🪝 Installing pre-commit hooks..."
if [ -f ".pre-commit-config.yaml" ]; then
    # Add pre-commit to PATH for this session
    export PATH="$HOME/.local/bin:$PATH"

    pre-commit install
    pre-commit install --hook-type commit-msg
    echo "✓ Pre-commit hooks installed"
else
    echo "⚠️  No .pre-commit-config.yaml found, skipping hook installation"
fi

# Download Go dependencies
echo "📥 Downloading Go dependencies..."
go mod download

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  task --list          - List all available tasks"
echo "  task run             - Run the server"
echo "  task dev             - Run with hot reload"
echo "  task test            - Run tests"
echo "  task pre-commit-run  - Run pre-commit checks"
echo ""
echo "Happy coding! 🎉"
