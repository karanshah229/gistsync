#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: Subcommands (config, provider, status, remove)"
echo "------------------------------------------------"

setup_test_env
# Initialize for command testing
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init

# 1. Config Commands
echo "▶️ Testing 'config' subcommands..."
$GISTSYNC_BIN config set watch_interval_seconds 120
$GISTSYNC_BIN config get watch_interval_seconds | grep "120"
$GISTSYNC_BIN config list | grep "watch_interval_seconds: 120"
echo "✅ Configuration GET/SET successful."

# 2. Provider Commands
echo "▶️ Testing 'provider' subcommands..."
$GISTSYNC_BIN provider info | grep "GitHub"
$GISTSYNC_BIN provider github test | grep "Success"
echo "✅ Provider info and test successful."

# 3. Status and Remove (Mocked Tracking)
echo "▶️ Testing 'status' and 'remove'..."
touch "$TEST_DIR/testfile.txt"
$GISTSYNC_BIN sync "$TEST_DIR/testfile.txt"
$GISTSYNC_BIN status "$TEST_DIR/testfile.txt" | grep "NOOP"
$GISTSYNC_BIN remove "$TEST_DIR/testfile.txt"
$GISTSYNC_BIN status "$TEST_DIR/testfile.txt" | grep "UNTRACKED"
echo "✅ Status, sync, and remove successful."

# Cleanup Gists
GIST_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
echo "✅ Subcommands Test Successful!"
