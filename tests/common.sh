#!/bin/bash

# Common utilities for gistsync tests
TEST_ROOT="$(pwd)/tests/tmp_files"
GISTSYNC_BIN="env XDG_CONFIG_HOME=$TEST_ROOT ./gistsync"
TEST_DIR="$TEST_ROOT/test_data"
CONFIG_DIR="$TEST_ROOT/gistsync"

# Helper for standard yes/no confirmation (ENTER = yes)
function confirm_yes {
    echo "" # Just ENTER
}

# Helper to ensure we are running from project root
if [ ! -f "main.go" ]; then
    echo "Error: Run tests from the project root."
    exit 1
fi

# Build once for all tests
echo "🔨 Building gistsync..."
go build -o gistsync .

function setup_test_env {
    echo "🧹 Cleaning local environment..."
    rm -rf "$CONFIG_DIR"
    rm -rf "$TEST_DIR"
    mkdir -p "$TEST_DIR"
    mkdir -p "$CONFIG_DIR"

    # Link gh config so provider tests work (gh inherits XDG_CONFIG_HOME)
    if [ -d "$HOME/.config/gh" ] && [ ! -d "$TEST_ROOT/gh" ]; then
        mkdir -p "$TEST_ROOT"
        ln -s "$HOME/.config/gh" "$TEST_ROOT/gh"
    fi
}

function cleanup_test_gists {
    echo "🧹 Cleaning up remote test gists..."
    # List gists with a description containing 'gistsync-test' or similar
    # For now, we'll manually track or just delete all gists created in tests
    # A better way is to delete gists by ID if we captured them
    :
}

function gh_set_file_in_gist {
    local gid=$1
    local filename=$2
    local content=$3
    
    # Construct JSON payload for PATCH
    # {"files": {"filename": {"content": "content"}}}
    local payload=$(printf '{"files": {"%s": {"content": "%s"}}}' "$filename" "$content")
    
    XDG_CONFIG_HOME="$TEST_ROOT" gh api -X PATCH "gists/$gid" --input - <<< "$payload" > /dev/null
}

function assert_gist_exists {
    local gid=$1
    if XDG_CONFIG_HOME="$TEST_ROOT" gh gist view "$gid" > /dev/null 2>&1; then
        echo "✅ Gist $gid exists."
    else
        echo "❌ Gist $gid NOT found."
        exit 1
    fi
}
