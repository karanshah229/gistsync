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
  "log_level": "info",
  "default_provider": "github",
  "autostart": true
}
EOF
echo '{"version": "0.1.0", "mappings": []}' > "$CONFIG_DIR/state.json"

# 1. Config List
echo "▶️ Testing 'config list'..."
ListOut=$($GISTSYNC_BIN config list)
if echo "$ListOut" | grep -q "watch_interval_seconds:.*60" && echo "$ListOut" | grep -q "default_provider:.*github"; then
    echo "✅ Config values listed correctly."
else
    echo "❌ Config list mismatch."
    echo "Output: $ListOut"
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

# 4. Config Sync
echo "▶️ Testing 'config sync' (Initial)..."
# Using --provider github to avoid prompts, simulating an initial sync
SyncOut=$($GISTSYNC_BIN config sync --provider github 2>&1)
if echo "$SyncOut" | grep -q "initial sync"; then
    echo "✅ Command correctly triggered initial sync."
else
    echo "❌ Command failed to trigger initial sync."
    echo "Output: $SyncOut"
    exit 1
fi

# Verify mapping was created
if grep -q "$CONFIG_DIR" "$CONFIG_DIR/state.json"; then
    echo "✅ Configuration mapping added to state.json."
else
    echo "❌ Mapping missing from state.json."
    exit 1
fi

# Running again to verify stability (Stabilize + NOOP)
echo "▶️ Testing 'config sync' (Stability)..."
$GISTSYNC_BIN config sync > /dev/null 2>&1 # Stabilize
SyncOut2=$($GISTSYNC_BIN config sync 2>&1)
if echo "$SyncOut2" | grep -q "already up to date"; then
    echo "✅ Config sync reached stable NOOP state."
else
    echo "❌ Config sync failed to reach NOOP."
    echo "Output: $SyncOut2"
    exit 1
fi

echo "✅ Config Tests Successful!"
