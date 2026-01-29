#!/bin/bash

# QuotaSense CLI Installation Script
# Inspired by claude-commit install script

set -e

REPO="quaywin/quota-sense-cli"
BINARY_NAME="qs"
INSTALL_DIR=""
USE_SUDO=""

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

# Determine INSTALL_DIR
if [[ ":$PATH:" == *":$HOME/.local/bin:"* ]] && [ -d "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=""
else
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO="sudo"
fi

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
    $USE_SUDO mv $BINARY_NAME $INSTALL_DIR/
else
    echo -e "${BLUE}Downloading from GitHub...${NC}"
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    if curl -L "$URL" -o "$TMP_DIR/package.tar.gz"; then
        tar -xzf "$TMP_DIR/package.tar.gz" -C "$TMP_DIR"
        if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
            $USE_SUDO mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
        else
            # Try to find the binary if it's in a subdirectory
            BINARY_PATH=$(find "$TMP_DIR" -type f -name "$BINARY_NAME" | head -n 1)
            if [ -n "$BINARY_PATH" ]; then
                $USE_SUDO mv "$BINARY_PATH" "$INSTALL_DIR/"
            else
                echo -e "Error: Binary not found in the downloaded archive."
                exit 1
            fi
        fi
    else
        echo -e "Error: Failed to download from $URL"
        echo -e "Note: If the release hasn't been created yet, use './install.sh --source' to build locally."
        exit 1
    fi
fi

echo -e "${GREEN}QuotaSense CLI installed successfully!${NC}"
echo -e "Run '${BLUE}qs${NC}' to get started."
