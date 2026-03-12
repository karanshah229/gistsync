#!/bin/bash
set -e

# Master script to run all gistsync tests
echo "🚀 Starting gistsync Test Suite..."

# Ensure we have the binary
go build -o gistsync .

# 1. Run Unit Tests
echo "🧪 Running Go Unit Tests..."
go test ./core/... ./internal/... ./pkg/...

# 2. Run UI Integration Tests
bash tests/test_ui.sh

# 3. Run Autostart Tests
bash tests/test_autostart.sh

# 4. Run Config Tests
bash tests/test_config.sh

# 5. Run Init Tests
bash tests/test_init.sh

# 6. Run Command Tests (status, remove, provider)
bash tests/test_commands.sh

# 7. Run Sync Tests
bash tests/test_sync.sh

# 8. Run Logging and WAL Recovery Tests
bash tests/test_logging.sh
bash tests/test_recover.sh

# 9. Run Manual Mapping Tests
bash tests/test_manual_mapping.sh

echo "------------------------------------------------"
echo "✅ ALL TESTS PASSED! 🎉"
echo "------------------------------------------------"

# Final Cleanup
rm gistsync
rm -rf tests/tmp_files
echo "🧹 Cleanup complete."
