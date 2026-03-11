#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: WAL Recovery Logic"
echo "------------------------------------------------"

setup_test_env

# 1. Setup Initial State and Generate Logs
echo "▶️ Generating sync events in logs..."
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init > /dev/null
mkdir -p "$TEST_DIR/recover_test"
echo "val1" > "$TEST_DIR/recover_test/file1.txt"
$GISTSYNC_BIN sync "$TEST_DIR/recover_test/file1.txt" > /dev/null

GIST_ID=$(grep "recover_test/file1.txt" "$CONFIG_DIR/state.json" | cut -d '"' -f 4 | head -n 1)
HASH=$(grep -A 5 "recover_test/file1.txt" "$CONFIG_DIR/state.json" | grep "last_synced_hash" | cut -d '"' -f 4 | head -n 1)

# 2. "Disaster": Delete state.json
echo "▶️ Simulating disaster: deleting state.json..."
rm "$CONFIG_DIR/state.json"

# 3. Recover State
echo "▶️ Running gistsync recover..."
$GISTSYNC_BIN recover > /dev/null

# 4. Verify Reconstruction
if [ ! -f "$CONFIG_DIR/state.json" ]; then
    echo "❌ ERROR: state.json was not reconstructed."
    exit 1
fi

RECOVERED_HASH=$(grep -A 5 "recover_test/file1.txt" "$CONFIG_DIR/state.json" | grep "last_synced_hash" | cut -d '"' -f 4 | head -n 1)
if [ "$HASH" != "$RECOVERED_HASH" ]; then
    echo "❌ ERROR: Recovered hash mismatch. Expected $HASH, got $RECOVERED_HASH"
    # Show the file content for debugging
    cat "$CONFIG_DIR/state.json"
    exit 1
fi
echo "✅ state.json successfully reconstructed with correct hash."

# 5. Cleanup
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
echo "✅ Recovery Test Successful!"
