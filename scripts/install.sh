#!/bin/bash

# Gistsync installer script (Homebrew-style)
# Usage: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/karanshah229/gistsync/main/scripts/install.sh)"

set -e

REPO="karanshah229/gistsync"
BINARY_NAME="gistsync"
INSTALL_DIR="/usr/local/bin"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Detected OS: $OS, ARCH: $ARCH"

# Determine latest version from GitHub
LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo "Could not fetch latest version. Please check your connection or the repository."
    exit 1
fi

echo "Installing $BINARY_NAME $LATEST_VERSION..."

# Construct download URL
# Example format: gistsync_Darwin_arm64.tar.gz
# Note: This should match GoReleaser naming convention
# We'll use capitalized OS name for the URL as per GoReleaser default
OS_CAPITALIZED="$(tr '[:lower:]' '[:upper:]' <<< ${OS:0:1})${OS:1}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/${BINARY_NAME}_${OS_CAPITALIZED}_${ARCH}.tar.gz"

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading from $DOWNLOAD_URL..."
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_NAME.tar.gz"

# Extract and install
tar -xzf "$TMP_DIR/$BINARY_NAME.tar.gz" -C "$TMP_DIR"
chmod +x "$TMP_DIR/$BINARY_NAME"

echo "Moving binary to $INSTALL_DIR (may require sudo)..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
fi

echo "Successfully installed $BINARY_NAME $LATEST_VERSION to $INSTALL_DIR"
gistsync --help
