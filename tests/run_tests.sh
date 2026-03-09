#!/bin/bash
set -e

# Master script to run all gistsync tests
echo "🚀 Starting gistsync Test Suite..."

# Ensure we have the binary
go build -o gistsync .

# 1. Run Init Tests
bash tests/test_init.sh

# 2. Run Command Tests
bash tests/test_commands.sh

# 3. Run Sync Tests
bash tests/test_sync.sh

echo "------------------------------------------------"
echo "✅ ALL TESTS PASSED! 🎉"
echo "------------------------------------------------"

# Final Cleanup
rm gistsync
rm -rf tests/tmp_files
echo "🧹 Cleanup complete."
