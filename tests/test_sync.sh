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

GIST_ID=$(grep "nested" "$CONFIG_DIR/state.json" | grep -v "gistsync" | cut -d '"' -f 4 | head -n 1)
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
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep "pushed to remote"

# 2.1 NOOP Check
echo "▶️ Testing NOOP check..."
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep -i "already up to date"
echo "✅ NOOP check successful."

# Pull: Modified remote
gh_set_file_in_gist "$GIST_ID" "deep---file.txt" "remote-change"
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep "pulled from remote"
grep -q "remote-change" "$TEST_DIR/nested/deep/file.txt"
echo "✅ Local Push and Remote Pull successful."

# 3. Conflicts
echo "▶️ Testing Conflicts..."
echo "local-conflict" > "$TEST_DIR/nested/deep/file.txt"
gh_set_file_in_gist "$GIST_ID" "deep---file.txt" "remote-conflict"
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep "CONFLICT"
echo "✅ Conflict detection successful."

# 4. Virtual State Projection (Backups)
echo "▶️ Testing Virtual State Projection (Backup)..."
echo "backup-test" >> "$CONFIG_DIR/config.json"
# Syncing the config directory should inject state.json into the backup
$GISTSYNC_BIN sync "$CONFIG_DIR" | grep -i "pushed"
GIST_ID_BACKUP=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4 | head -n 1)
XDG_CONFIG_HOME="$TEST_ROOT" gh api "gists/$GIST_ID_BACKUP" | grep "state.json"
echo "✅ Virtual State Projection verified (state.json found in backup Gist)."

# 5. Recovery from Remote Deletion
echo "▶️ Testing Recovery from Remote Deletion..."
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes
$GISTSYNC_BIN sync "$TEST_DIR/nested" | grep -i "creating new gist"
echo "✅ Successfully re-mapped after remote Gist deletion."

# Cleanup
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
GIST_ID_INIT=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4 | head -n 1)
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID_INIT" --yes || true
echo "✅ Sync Test Successful!"
