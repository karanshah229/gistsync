#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: Logging and JSON WAL Generation"
echo "------------------------------------------------"

setup_test_env

# 1. Initialization and Log File Creation
echo "▶️ Testing Log directory and file creation..."
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init > /dev/null
if [ ! -d "$CONFIG_DIR/logs" ]; then
    echo "❌ ERROR: Log directory not created."
    exit 1
fi

LOG_FILE=$(ls "$CONFIG_DIR/logs" | head -n 1)
if [ -z "$LOG_FILE" ]; then
    echo "❌ ERROR: No log file generated."
    exit 1
fi
echo "✅ Log file created: $LOG_FILE"

# 2. Verify JSON format and Common Fields
echo "▶️ Testing JSON structure and fields..."
FIRST_LINE=$(head -n 1 "$CONFIG_DIR/logs/$LOG_FILE")

# Check for required fields
echo "$FIRST_LINE" | grep -q '"time":' || { echo "❌ ERROR: 'time' field missing in logs"; exit 1; }
echo "$FIRST_LINE" | grep -q '"level":' || { echo "❌ ERROR: 'level' field missing in logs"; exit 1; }
echo "$FIRST_LINE" | grep -q '"msg":' || { echo "❌ ERROR: 'msg' field missing in logs"; exit 1; }
echo "$FIRST_LINE" | grep -q '"pid":' || { echo "❌ ERROR: 'pid' field missing in logs"; exit 1; }
echo "✅ JSON structure verified."

# 3. Verify Sync Events in Logs
echo "▶️ Testing Sync event logging..."
mkdir -p "$TEST_DIR/logs_test"
echo "log test" > "$TEST_DIR/logs_test/test.txt"
$GISTSYNC_BIN sync "$TEST_DIR/logs_test/test.txt" > /dev/null

# Find the sync success event in ANY log file
cat "$CONFIG_DIR/logs"/*.log | grep -q '"type":"SYNC_START"' || { echo "❌ ERROR: SYNC_START not found in logs"; exit 1; }
cat "$CONFIG_DIR/logs"/*.log | grep -q '"type":"SYNC_SUCCESS"' || { echo "❌ ERROR: SYNC_SUCCESS not found in logs"; exit 1; }
cat "$CONFIG_DIR/logs"/*.log | grep -q "/logs_test/test.txt" || { echo "❌ ERROR: Local path not found in sync logs"; exit 1; }
echo "✅ Sync events correctly logged."

# 4. Cleanup
GIST_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
echo "✅ Logging Test Successful!"
