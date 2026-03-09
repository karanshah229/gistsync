#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: CLI UI Elements"
echo "------------------------------------------------"

setup_test_env

# 1. Test Help Output (Header/Icons)
echo "▶️ Testing Help output structure..."
HelpOut=$($GISTSYNC_BIN --help)
if echo "$HelpOut" | grep -q "A fast and efficient CLI tool"; then
    echo "✅ App description found in help."
else
    echo "❌ App description NOT found in help."
    exit 1
fi

# 2. Test Error UI (Running sync without config)
echo "▶️ Testing Error UI (no config)..."
SyncOut=$($GISTSYNC_BIN sync 2>&1 || true)
if echo "$SyncOut" | grep -q "❌ configuration or state is missing"; then
    echo "✅ Error icon and message found."
else
    echo "❌ Error UI mismatch."
    echo "Actual output: $SyncOut"
    exit 1
fi

# 3. Test Info/Header UI (Provider Info)
echo "▶️ Testing Info/Header UI (provider info)..."
InfoOut=$($GISTSYNC_BIN provider info)
if echo "$InfoOut" | grep -q "Provider Setup Information" && echo "$InfoOut" | grep -q "GitHub"; then
    echo "✅ Header and Info strings found."
else
    echo "❌ Provider info UI mismatch."
    exit 1
fi

echo "✅ UI Integration Tests Successful!"
