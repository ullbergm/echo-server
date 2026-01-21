#!/bin/bash
set -e

# Copy SSH keys from read-only mount to user directory with correct permissions
mkdir -p /home/vscode/.ssh
if [ -d /.ssh-mount ] && [ "$(ls -A /.ssh-mount 2>/dev/null)" ]; then
    cp -r /.ssh-mount/. /home/vscode/.ssh/
    chmod 700 /home/vscode/.ssh
    # Set permissions for public keys
    find /home/vscode/.ssh -type f -name '*.pub' -exec chmod 644 {} \;
    # Set permissions for private keys
    find /home/vscode/.ssh -type f ! -name '*.pub' ! -name 'known_hosts*' ! -name 'config' ! -name 'authorized_keys' -exec chmod 600 {} \;
fi

# Verify Go dependencies
make deps-verify

# Configure Git to use SSH signing
git config --global user.signingkey /home/vscode/.ssh/id_rsa
git config --global gpg.format ssh
git config --global commit.gpgsign true
git config --global gpg.ssh.program ssh-keygen
