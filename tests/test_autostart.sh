#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: 'gistsync autostart' commands"
echo "------------------------------------------------"

setup_test_env
# Create a dummy config so we can test autostart (which needs config)
mkdir -p "$CONFIG_DIR"
echo '{"autostart": false}' > "$CONFIG_DIR/config.json"
echo '{"version": "0.1.0", "mappings": []}' > "$CONFIG_DIR/state.json"

# 1. Enable Autostart
echo "▶️ Testing 'autostart enable'..."
$GISTSYNC_BIN autostart enable
if grep -q '"autostart": true' "$CONFIG_DIR/config.json"; then
    echo "✅ Config updated to autostart: true."
else
    echo "❌ Config NOT updated."
    exit 1
fi

# 2. Status check
echo "▶️ Testing 'autostart status'..."
StatusOut=$($GISTSYNC_BIN autostart status)
if echo "$StatusOut" | grep -q "ENABLED"; then
    echo "✅ Status correctly reported as ENABLED."
else
    echo "❌ Status mismatch."
    exit 1
fi

# 3. Disable Autostart
echo "▶️ Testing 'autostart disable'..."
$GISTSYNC_BIN autostart disable
if grep -q '"autostart": false' "$CONFIG_DIR/config.json"; then
    echo "✅ Config updated to autostart: false."
else
    echo "❌ Config NOT updated."
    exit 1
fi

echo "✅ Autostart Tests Successful!"
