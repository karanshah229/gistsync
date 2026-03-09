#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: 'gistsync init' flow"
echo "------------------------------------------------"

setup_test_env

# 1. Fresh Init (No Restore, with Backup)
echo "▶️ Testing Fresh Init..."
# Responses: Restore? No (n), Default Provider? ENTER, 3 config fields? ENTER, Backup? ENTER
printf "n\n\n\n\n\n" | $GISTSYNC_BIN init

if [ -f "$CONFIG_DIR/config.json" ] && [ -f "$CONFIG_DIR/state.json" ]; then
    echo "✅ Local config and state created."
else
    echo "❌ Local config or state missing."
    exit 1
fi

# 2. Verify Backup Gist
GIST_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
if [ -n "$GIST_ID" ]; then
    assert_gist_exists "$GIST_ID"
    echo "✅ Backup Gist ID found: $GIST_ID"
else
    echo "❌ Backup Gist ID NOT found in state.json."
    exit 1
fi

# 3. Restore Flow (Fresh Machine Simulation)
echo "▶️ Testing Restore Flow..."
setup_test_env
# Restore with piped input (Restore? y, Provider? ENTER, Backup? ENTER, Sync? ENTER)
printf "y\n\n\n\n" | $GISTSYNC_BIN init

if [ -f "$CONFIG_DIR/config.json" ] && [ -f "$CONFIG_DIR/state.json" ]; then
    echo "✅ Local config and state restored."
else
    echo "❌ Local config or state restoration failed."
    exit 1
fi

# Check if PENDING was replaced
if grep -q "PENDING" "$CONFIG_DIR/state.json"; then
    echo "❌ ERROR: state.json still contains PENDING remote_id."
    exit 1
else
    echo "✅ PENDING remote_id was correctly replaced."
fi

# Cleanup
XDG_CONFIG_HOME="$TEST_ROOT" XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
echo "✅ Test Init Successful!"
