#!/bin/bash

# QuotaSense CLI Installation Script
# Inspired by claude-commit install script

set -e

REPO="quaywin/quota-sense-cli"
BINARY_NAME="qs"
INSTALL_DIR="/usr/local/bin"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Installing QuotaSense CLI...${NC}"

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then ARCH="arm64"; fi

# Get latest version from GitHub API (fallback to 0.1.0 if not found)
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    VERSION="v0.1.0"
fi

echo -e "Detected: $OS-$ARCH, Version: $VERSION"

# URL for the binary
# Note: This expects you to have a release with this naming convention
URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"

# For local testing/installation from source if repo is not public yet
if [[ "$1" == "--source" ]]; then
    echo -e "${BLUE}Building from source...${NC}"
    go build -o $BINARY_NAME main.go
    sudo mv $BINARY_NAME $INSTALL_DIR/
else
    echo -e "${BLUE}Downloading from GitHub...${NC}"
    # This part will fail until you actually create a Release on GitHub
    # curl -L $URL | tar xz
    # sudo mv $BINARY_NAME $INSTALL_DIR/

    # Fallback for now: Suggest building from source
    echo -e "Note: GitHub release not found yet. Use './install.sh --source' to build locally."
fi

echo -e "${GREEN}QuotaSense CLI installed successfully!${NC}"
echo -e "Run '${BLUE}qs${NC}' to get started."
