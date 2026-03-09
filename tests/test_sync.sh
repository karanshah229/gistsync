#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: Core Syncing (Nested Paths, Conflicts)"
echo "------------------------------------------------"

setup_test_env
# Initialize
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init

# 1. Nested Directory Sync (Path Flattening)
echo "▶️ Testing Nested Directory Sync (Flattening)..."
mkdir -p "$TEST_DIR/nested/deep"
echo "hello" > "$TEST_DIR/nested/deep/file.txt"
$GISTSYNC_BIN sync "$TEST_DIR/nested"

GIST_ID=$(grep "nested" "$CONFIG_DIR/state.json" | grep -v "gistsync" | cut -d '"' -f 4)
if [ -z "$GIST_ID" ]; then
    echo "❌ ERROR: Failed to get Gist ID for 'nested'."
    exit 1
fi

# Verify Gist content via gh (Look for flattened separator '---')
XDG_CONFIG_HOME="$TEST_ROOT" gh api "gists/$GIST_ID" | grep "deep---file.txt"
echo "✅ Path flattening (---) verified in Gist."

# 2. 2-way Sync (Push & Pull)
echo "▶️ Testing 2-way Sync..."
# Push: Modified local
echo "modified" > "$TEST_DIR/nested/deep/file.txt"
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep "Sync successful"

# Pull: Modified remote
gh_set_file_in_gist "$GIST_ID" "deep---file.txt" "remote-change" # This is a helper I'll add to common.sh
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep "Sync successful"
grep -q "remote-change" "$TEST_DIR/nested/deep/file.txt"
echo "✅ Local Push and Remote Pull successful."

# 3. Conflicts
echo "▶️ Testing Conflicts..."
echo "local-conflict" > "$TEST_DIR/nested/deep/file.txt"
gh_set_file_in_gist "$GIST_ID" "deep---file.txt" "remote-conflict"
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep "CONFLICT"
echo "✅ Conflict detection successful."

# Cleanup
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
GIST_ID_INIT=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID_INIT" --yes || true
echo "✅ Sync Test Successful!"
