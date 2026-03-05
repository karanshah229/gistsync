#!/bin/bash

# Gistsync Pre-commit Hook
# Handles interactive version bumping and changelogging.

set -e

# Path to files
VERSION_FILE="VERSION"
VERSION_CMD_FILE="cmd/VERSION"
CHANGELOG_FILE="CHANGELOG.md"

# Colors for terminal
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if VERSION file is already staged
if git diff --cached --name-only | grep -E -q "^(${VERSION_FILE}|${VERSION_CMD_FILE})$"; then
    echo -e "${GREEN}✓ Version is already bumped and staged.${NC}"
    exit 0
fi

echo -e "${YELLOW}⚠ Version not bumped in this commit.${NC}"
read -p "Do you want to bump the version and add a changelog entry? [y/N] " -r < /dev/tty
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Skipping version bump.${NC}"
    exit 0
fi

# Get current version
CURRENT_VERSION=$(cat "$VERSION_FILE" | xargs)
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

echo -e "Current version: ${BLUE}$CURRENT_VERSION${NC}"
echo "Select bump type:"
echo "1) Patch (${MAJOR}.${MINOR}.$((PATCH + 1)))"
echo "2) Minor (${MAJOR}.$((MINOR + 1)).0)"
echo "3) Major ($((MAJOR + 1)).0.0)"
read -p "Choice [1-3]: " -r BUMP_TYPE < /dev/tty

case $BUMP_TYPE in
    1) PATCH=$((PATCH + 1)) ;;
    2) MINOR=$((MINOR + 1)); PATCH=0 ;;
    3) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
    *) echo "Invalid choice. Aborting."; exit 1 ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo -e "New version: ${GREEN}$NEW_VERSION${NC}"

# Ask for changelog message
read -p "Enter changelog message (short summary): " -r MESSAGE < /dev/tty
if [ -z "$MESSAGE" ]; then
    echo "Message cannot be empty. Aborting."
    exit 1
fi

# Update VERSION file
echo "$NEW_VERSION" > "$VERSION_FILE"

# Update CHANGELOG.md
DATE=$(date +%Y-%m-%d)
TEMP_CHANGELOG=$(mktemp)

# Create file if it doesn't exist
if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "# Changelog" > "$CHANGELOG_FILE"
    echo "" >> "$CHANGELOG_FILE"
fi

# Insert new entry at the top (after the header)
{
    head -n 2 "$CHANGELOG_FILE"
    echo "## [$NEW_VERSION] - $DATE"
    echo "- $MESSAGE"
    echo ""
    tail -n +3 "$CHANGELOG_FILE"
} > "$TEMP_CHANGELOG"

mv "$TEMP_CHANGELOG" "$CHANGELOG_FILE"

# Auto-stage the changes
git add "$VERSION_FILE" "$CHANGELOG_FILE"

echo -e "${GREEN}✓ Version bumped to $NEW_VERSION and staged with changelog!${NC}"
