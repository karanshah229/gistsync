#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: Manual Gist Mapping"
echo "------------------------------------------------"

setup_test_env
# Initialize
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init

# Create a test file and a remote public gist for it
echo "▶️ Generating remote gist..."
echo "initial-content" > "$TEST_DIR/test_manual.txt"
GIST_ID=$(XDG_CONFIG_HOME="$TEST_ROOT" gh gist create "$TEST_DIR/test_manual.txt" --public | awk '{print $NF}' | xargs basename)

if [ -z "$GIST_ID" ]; then
    echo "❌ ERROR: Failed to create Gist."
    exit 1
fi

assert_gist_exists "$GIST_ID"

# 1. Test clean manual mapping (identical hashes)
echo "▶️ Testing Clean Manual Mapping..."
# Since hashes match, it should say already up to date and create the mapping
$GISTSYNC_BIN sync "$TEST_DIR/test_manual.txt" "$GIST_ID" | grep "already up to date" || { echo "❌ Expected Noop"; exit 1; }
grep -q "$GIST_ID" "$CONFIG_DIR/state.json" || { echo "❌ Mapping not saved to state.json"; exit 1; }
echo "✅ Clean mapping successful."

# 2. Test manual mapping overwrite prompt (default to yes)
echo "▶️ Testing Overwrite Prompt..."
# Piped newline confirms the default empty input (yes)
echo "" | $GISTSYNC_BIN sync "$TEST_DIR/test_manual.txt" "$GIST_ID" | grep "already up to date" || { echo "❌ Expected Noop on overwrite"; exit 1; }
echo "✅ Overwrite prompt successful."

# 3. Test conflict detection on diverge
echo "▶️ Testing Conflict Detection..."
echo "divergent-content" > "$TEST_DIR/test_manual.txt"
# We pipe in a newline (yes) to the confirmation prompt of an existing mapping, and we expect a CONFLICT since local diverged from remote
echo "" | $GISTSYNC_BIN sync "$TEST_DIR/test_manual.txt" "$GIST_ID" | grep "Conflict detected" || { echo "❌ Expected Conflict"; exit 1; }
echo "✅ Conflict properly detected."

# Cleanup
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
GIST_INIT_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | grep -v "$GIST_ID" | cut -d '"' -f 4 | xargs)
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_INIT_ID" --yes || true

echo "✅ Manual Mapping Test Successful!"
