#!/bin/bash

# Update Defuddle Reference Script
# Updates the Defuddle reference code and checks for changes

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEFUDDLE_DIR="$PROJECT_ROOT/reference/defuddle"

echo "🔄 Updating Defuddle reference..."

# Check if submodule exists
if [ ! -e "$DEFUDDLE_DIR/.git" ]; then
    echo "❌ Defuddle submodule not found. Please run: git submodule update --init --recursive"
    exit 1
fi

cd "$DEFUDDLE_DIR"

# Get current version information
CURRENT_COMMIT=$(git rev-parse HEAD)
CURRENT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "no-tag")

echo "📍 Current version: $CURRENT_TAG ($CURRENT_COMMIT)"

# Fetch latest changes
echo "📥 Fetching latest changes..."
git fetch origin

# Check if there are updates
LATEST_COMMIT=$(git rev-parse origin/main)

if [ "$CURRENT_COMMIT" = "$LATEST_COMMIT" ]; then
    echo "✅ Already up to date!"
    exit 0
fi

echo "🆕 New changes available!"

# Show change summary
echo "📋 Changes since last update:"
git log --oneline "$CURRENT_COMMIT..$LATEST_COMMIT"

# Ask whether to update
read -p "🤔 Do you want to update to the latest version? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Update cancelled."
    exit 0
fi

# Update to latest version
echo "⬆️ Updating to latest version..."
git pull origin main

# Get new version information
NEW_COMMIT=$(git rev-parse HEAD)
NEW_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "no-tag")

echo "✅ Updated to: $NEW_TAG ($NEW_COMMIT)"

# Show detailed changes
echo "📝 Detailed changes:"
git diff "$CURRENT_COMMIT" "$NEW_COMMIT" --stat

# Check changes in important files
IMPORTANT_FILES=(
    "src/defuddle.ts"
    "src/types.ts"
    "src/metadata.ts"
    "src/scoring.ts"
    "src/standardize.ts"
    "src/constants.ts"
    "package.json"
)

echo "🔍 Checking important files for changes..."
for file in "${IMPORTANT_FILES[@]}"; do
    if git diff --quiet "$CURRENT_COMMIT" "$NEW_COMMIT" -- "$file"; then
        echo "  ✅ $file - No changes"
    else
        echo "  🔄 $file - Modified"
    fi
done

# Return to project root
cd "$PROJECT_ROOT"

# Commit submodule update
echo "💾 Committing submodule update..."
git add reference/defuddle

# Generate commit message
if [ "$NEW_TAG" != "no-tag" ]; then
    COMMIT_MSG="chore: update defuddle reference to $NEW_TAG"
else
    COMMIT_MSG="chore: update defuddle reference to latest commit"
fi

git commit -m "$COMMIT_MSG

Updated from $CURRENT_COMMIT to $NEW_COMMIT

Changes:
$(cd "$DEFUDDLE_DIR" && git log --oneline "$CURRENT_COMMIT..$NEW_COMMIT")"

echo "🎉 Defuddle reference updated successfully!"
echo "📚 Next steps:"
echo "  1. Review the changes in reference/defuddle"
echo "  2. Update Go implementation to match new features/fixes"
echo "  3. Update tests to maintain compatibility"
echo "  4. Update version mapping in reference/README.md" 