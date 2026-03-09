#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: Subcommands (config, provider, status, remove)"
echo "------------------------------------------------"

setup_test_env
# Initialize for command testing
printf "n\n\n\n\n\n" | ./gistsync init

# 1. Config Commands
echo "▶️ Testing 'config' subcommands..."
./gistsync config set watch_interval_seconds 120
./gistsync config get watch_interval_seconds | grep "120"
./gistsync config list | grep "watch_interval_seconds: 120"
echo "✅ Configuration GET/SET successful."

# 2. Provider Commands
echo "▶️ Testing 'provider' subcommands..."
./gistsync provider info | grep "GitHub"
./gistsync provider github test | grep "Success"
echo "✅ Provider info and test successful."

# 3. Status and Remove (Mocked Tracking)
echo "▶️ Testing 'status' and 'remove'..."
touch "$TEST_DIR/testfile.txt"
./gistsync sync "$TEST_DIR/testfile.txt"
./gistsync status "$TEST_DIR/testfile.txt" | grep "NOOP"
./gistsync remove "$TEST_DIR/testfile.txt"
./gistsync status "$TEST_DIR/testfile.txt" | grep "UNTRACKED"
echo "✅ Status, sync, and remove successful."

# Cleanup Gists
GIST_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
gh gist delete "$GIST_ID" --yes || true
echo "✅ Subcommands Test Successful!"
