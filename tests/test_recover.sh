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

# 5. HWM (High Water Mark) and Interruption Warning
echo "▶️ Testing HWM and Interruption Warning..."
# 5.1 Manual Log Injection (Interruption)
# Ensure timestamps are later than the last recovery complete entry
sleep 1
LOG_FILE=$(ls -t "$CONFIG_DIR/logs/"*.log | head -n 1)
# Add a SYNC_SUCCESS without CHECKPOINT
cat <<EOF >> "$LOG_FILE"
{"time":"$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")","level":"INFO","msg":"sync success","type":"SYNC_SUCCESS","pid":9999,"local_path":"$TEST_DIR/interrupted_file.txt","remote_id":"gist123","hash":"hash123","provider":"github"}
EOF

$GISTSYNC_BIN recover 2>&1 | grep -i "potential interruption"
if grep -q "interrupted_file.txt" "$CONFIG_DIR/state.json"; then
    echo "✅ Interrupted sync recovered with warning."
else
    echo "❌ Interrupted sync NOT recovered."
    # Debug: show state and logs
    # cat "$CONFIG_DIR/state.json"
    # tail -n 5 "$LOG_FILE"
    exit 1
fi

# 5.2 HWM Check (Recover should be up to date)
echo "▶️ Testing HWM (Up to date check)..."
$GISTSYNC_BIN recover | grep -i "already up to date"
echo "✅ HWM correctly prevents redundant replays."

# 6. Cleanup
GIST_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4 | head -n 1)
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
echo "✅ Recovery Test Successful!"
