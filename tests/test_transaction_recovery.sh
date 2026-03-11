#!/bin/bash
set -e

# Import common test helpers
source "$(dirname "$0")/common.sh"

setup_test_env

echo "🧪 Testing Transactional Recovery (Interrupted Local Commit)..."

# Export vars so they are available to subshells and tools
export XDG_CONFIG_HOME="$TEST_ROOT"

# 0. Initialize
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init

# 1. Perform a successful sync to generate some logs
# We use a dummy file
echo "initial content" > "$TEST_DIR/transaction_test.txt"
$GISTSYNC_BIN sync "$TEST_DIR/transaction_test.txt"

# 2. Get the latest log file and its PID
LATEST_LOG=$(ls -t "$CONFIG_DIR/logs"/*.log | head -n 1)
PID=$(grep -o '"pid":[0-9]*' "$LATEST_LOG" | head -n 1 | cut -d: -f2)

# 3. Simulate a crash: Inject a SYNC_SUCCESS without a CHECKPOINT
# We'll create a new log entry in the same file with a new "sync" but no commit
echo "updated content" > "$TEST_DIR/transaction_test.txt"
NEW_HASH=$(shasum -a 256 "$TEST_DIR/transaction_test.txt" | cut -d' ' -f1)
TIME=$(date -u +"%Y-%m-%dT%H:%M:%S.%NZ")

# Manually append a SYNC_SUCCESS to the log
cat >> "$LATEST_LOG" <<EOF
{"time":"$TIME","level":"INFO","msg":"Sync success","pid":$PID,"type":"SYNC_SUCCESS","local_path":"$TEST_DIR/transaction_test.txt","remote_id":"manual-test-id","hash":"$NEW_HASH","is_folder":false,"provider":"github","public":false}
EOF

# 4. Delete state.json
rm "$CONFIG_DIR/state.json"

# 5. Run recover
echo "🔍 Running recover (should detect interrupted commit)..."
$GISTSYNC_BIN recover > recover_output_1.txt 2>&1
cat recover_output_1.txt

# 5.5 Run recover again (should be silent)
echo "🔍 Running recover again (should be silent/up-to-date)..."
$GISTSYNC_BIN recover > recover_output_2.txt 2>&1
cat recover_output_2.txt

# 6. Verify Detection Message in first run
if grep -q "Detected interrupted local commit" recover_output_1.txt; then
    echo "✅ Interrupted commit detected correctly in first run"
else
    echo "❌ Failed to detect interrupted commit in first run"
    exit 1
fi

# 6.5 Verify NO Detection Message in second run
if grep -q "Detected interrupted local commit" recover_output_2.txt; then
    echo "❌ Failed: redundant warning in second run"
    exit 1
fi

if grep -q "State is already up to date" recover_output_2.txt; then
    echo "✅ Redundant run suppressed correctly"
else
    echo "❌ Failed to suppress redundant run"
    exit 1
fi

# 7. Verify Recovery of the uncommitted hash
if grep -q "$NEW_HASH" "$CONFIG_DIR/state.json"; then
    echo "✅ Successfully recovered uncommitted hash"
else
    echo "❌ Failed to recover uncommitted hash"
    exit 1
fi

echo "🎉 Transactional Recovery test passed!"
# rm -rf "$TEST_ROOT" # Cleanup happens in next setup_test_env or manually
