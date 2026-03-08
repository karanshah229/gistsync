#!/bin/bash

# Gistsync uninstaller script
# Usage: /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/karanshah229/gistsync/main/scripts/uninstall.sh)"

set -e

BINARY_NAME="gistsync"
INSTALL_DIR="/usr/local/bin"

echo "🗑️ Removing $BINARY_NAME..."

# 1. Remove binary
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "Stopping and removing binary from $INSTALL_DIR (may require sudo)..."
    if [ -w "$INSTALL_DIR" ]; then
        rm "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo rm "$INSTALL_DIR/$BINARY_NAME"
    fi
else
    echo "ℹ️ Binary not found in $INSTALL_DIR. Skipping."
fi

# 2. Remove configuration and state
CONFIG_DIR="$HOME/.config/gistsync"
# Handle XDG_CONFIG_HOME if set
if [ -n "$XDG_CONFIG_HOME" ]; then
    CONFIG_DIR="$XDG_CONFIG_HOME/gistsync"
fi

if [ -d "$CONFIG_DIR" ]; then
    echo "Removing configuration and state from $CONFIG_DIR..."
    rm -rf "$CONFIG_DIR"
else
    echo "ℹ️ Configuration directory not found. Skipping."
fi

echo "✨ Successfully uninstalled $BINARY_NAME. See you again! 🏔️"
