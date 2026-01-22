#!/bin/bash
# Install Node.js tools manually if post-create script failed
# Run this if you see "npm: command not found" errors

set -e

echo "🔧 Installing Node.js developer tools..."

# Check if npm is available
if ! command -v npm &> /dev/null; then
    echo "❌ Error: npm is not installed or not in PATH"
    echo ""
    echo "Please ensure Node.js is installed:"
    echo "  • Check devcontainer features in .devcontainer/devcontainer.json"
    echo "  • Or install manually: sudo apt update && sudo apt install -y nodejs npm"
    exit 1
fi

echo "✓ npm is available ($(npm --version))"

# Install markdownlint-cli
echo "📄 Installing markdownlint-cli..."
npm install -g markdownlint-cli

# Install commitlint and config
echo "✅ Installing commitlint..."
npm install -g @commitlint/cli @commitlint/config-conventional

echo ""
echo "✅ All Node.js tools installed successfully!"
echo ""
echo "Installed tools:"
echo "  • markdownlint-cli: $(markdownlint --version)"
echo "  • commitlint: $(npx @commitlint/cli --version)"
echo ""
echo "You can now use:"
echo "  task commit-lint -- 'your commit message'"
echo "  markdownlint *.md"
