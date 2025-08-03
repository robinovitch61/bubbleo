#!/bin/bash

# Install Go 1.23.3 for the bubbleo project
set -e

GO_VERSION="1.23.3"
GO_OS="linux"

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        GO_ARCH="amd64"
        ;;
    aarch64)
        GO_ARCH="arm64"
        ;;
    armv6l)
        GO_ARCH="armv6l"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

GO_TARBALL="go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz"
GO_URL="https://golang.org/dl/${GO_TARBALL}"

echo "Installing Go ${GO_VERSION} for ${GO_OS}/${GO_ARCH}..."

# Download Go tarball
echo "Downloading ${GO_URL}..."
curl -LO "$GO_URL"

# Remove existing Go installation and extract new one
echo "Installing to /usr/local/go..."
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "$GO_TARBALL"

# Clean up tarball
rm "$GO_TARBALL"

# Add Go to PATH if not already present
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

if ! grep -q "/usr/local/go/bin" ~/.profile; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
fi

echo "Go ${GO_VERSION} installed successfully!"
echo "Please run 'source ~/.bashrc' or restart your terminal to update PATH."
echo "Verify installation with: go version"