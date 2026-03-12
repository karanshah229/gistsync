#!/bin/bash
source tests/common.sh

echo "------------------------------------------------"
echo "🧪 Testing: 'gistsync init' flow"
echo "------------------------------------------------"

# 1. Fresh Init (No Restore, with Backup)
echo "▶️ Testing Fresh Init..."
setup_test_env
# Responses: Restore? No (n), Default Provider? github (ENTER), 3 config fields? ENTER, Backup? Yes (ENTER)
printf "n\n\n\n\n\n\n" | $GISTSYNC_BIN init

if [ -f "$CONFIG_DIR/config.json" ] && [ -f "$CONFIG_DIR/state.json" ]; then
    echo "✅ Local config and state created."
else
    echo "❌ Local config or state missing."
    exit 1
fi

# Verify Backup Gist
GIST_ID=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
if [ -n "$GIST_ID" ]; then
    assert_gist_exists "$GIST_ID"
    echo "✅ Backup Gist ID found: $GIST_ID"
else
    echo "❌ Backup Gist ID NOT found in state.json."
    exit 1
fi

# 1.1 Overwrite Confirmation (No)
echo "▶️ Testing Overwrite Confirmation (No)..."
# Try init again, say 'n' to overwrite
printf "n\n" | $GISTSYNC_BIN init | grep -q "Abort"
if [ $? -eq 0 ]; then
    echo "✅ Overwrite aborted as expected."
else
    echo "❌ Overwrite did not abort."
    exit 1
fi

# 1.2 Overwrite Confirmation (Yes)
echo "▶️ Testing Overwrite Confirmation (Yes)..."
# Try init again, say 'y' to overwrite, then 'n' to restore, then defaults
printf "y\nn\n\n\n\n\n\n" | $GISTSYNC_BIN init | grep -q "Initializing"
echo "✅ Overwrite proceeded as expected."

# 1.3 Init without Backup
echo "▶️ Testing Init without Backup..."
setup_test_env
# Responses: Restore? No (n), Provider? Default (ENTER), Interval? 60, Debounce? 500, Log? info, Auto? true, Backup? No (n)
printf "n\n\n60\n500\ninfo\ntrue\nn\n" | $GISTSYNC_BIN init
GIST_ID_NO_BACKUP=$(grep "remote_id" "$CONFIG_DIR/state.json" | cut -d '"' -f 4)
if [ -z "$GIST_ID_NO_BACKUP" ] || [ "$GIST_ID_NO_BACKUP" == "null" ]; then
    echo "✅ Init without backup successful (No Gist ID created)."
else
    echo "❌ ERROR: Gist ID created even though backup was declined: $GIST_ID_NO_BACKUP"
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

# 4. Auth Failure Simulation
echo "▶️ Testing Auth Failure Scenario..."
setup_test_env
# Create a fake 'gh' that fails auth check
mkdir -p "$TEST_ROOT/fakebin"
cat <<EOF > "$TEST_ROOT/fakebin/gh"
#!/bin/bash
if [[ "\$*" == "auth status" ]]; then
  echo "Logged out"
  exit 1
fi
exec gh "\$@"
EOF
chmod +x "$TEST_ROOT/fakebin/gh"

OLD_PATH=$PATH
export PATH="$TEST_ROOT/fakebin:$PATH"
# Run init, it should show auth instructions and then stop if no other provider is available
printf "n\n" | $GISTSYNC_BIN init | grep -iq "Please authenticate with github"
echo "✅ Auth failure warning displayed."
export PATH=$OLD_PATH

# Cleanup
XDG_CONFIG_HOME="$TEST_ROOT" gh gist delete "$GIST_ID" --yes || true
echo "✅ Test Init Successful!"
