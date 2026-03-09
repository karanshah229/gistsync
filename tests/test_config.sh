#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: 'gistsync config' commands"
echo "------------------------------------------------"

setup_test_env
mkdir -p "$CONFIG_DIR"
cat <<EOF > "$CONFIG_DIR/config.json"
{
  "watch_interval_seconds": 60,
  "watch_debounce_ms": 500,
  "log_level": "info"
}
EOF
echo '{"version": "0.1.0", "mappings": []}' > "$CONFIG_DIR/state.json"

# 1. Config List
echo "▶️ Testing 'config list'..."
ListOut=$($GISTSYNC_BIN config list)
if echo "$ListOut" | grep -q "watch_interval_seconds: 60"; then
    echo "✅ Config values listed correctly."
else
    echo "❌ Config list mismatch."
    exit 1
fi

# 2. Config Set
echo "▶️ Testing 'config set'..."
$GISTSYNC_BIN config set watch_interval_seconds 120
if grep -q '"watch_interval_seconds": 120' "$CONFIG_DIR/config.json"; then
    echo "✅ Config updated via 'set'."
else
    echo "❌ Config 'set' failed."
    exit 1
fi

# 3. Config Get
echo "▶️ Testing 'config get'..."
GetOut=$($GISTSYNC_BIN config get watch_interval_seconds)
if [ "$GetOut" == "120" ]; then
    echo "✅ Config 'get' returned correct value."
else
    echo "❌ Config 'get' mismatch: got '$GetOut', expected '120'."
    exit 1
fi

echo "✅ Config Tests Successful!"
