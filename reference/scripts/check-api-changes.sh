#!/bin/bash

# Check Defuddle API Changes Script
# Checks for Defuddle API changes to help track modifications that need to be synchronized in the Go version

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
DEFUDDLE_DIR="$PROJECT_ROOT/reference/defuddle"

echo "🔍 Checking Defuddle API changes..."

# Check if submodule exists
if [ ! -e "$DEFUDDLE_DIR/.git" ]; then
    echo "❌ Defuddle submodule not found. Please run: git submodule update --init --recursive"
    exit 1
fi

cd "$DEFUDDLE_DIR"

# Get version information
CURRENT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "no-tag")
CURRENT_COMMIT=$(git rev-parse HEAD)

echo "📍 Current version: $CURRENT_TAG ($CURRENT_COMMIT)"

# Check if comparison version is provided
if [ $# -eq 1 ]; then
    COMPARE_VERSION="$1"
    echo "🔄 Comparing with version: $COMPARE_VERSION"
else
    # Get previous version
    COMPARE_VERSION=$(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || echo "")
    if [ -z "$COMPARE_VERSION" ]; then
        echo "⚠️  No previous version found. Showing current API structure."
        COMPARE_VERSION="HEAD"
    else
        echo "🔄 Comparing with previous version: $COMPARE_VERSION"
    fi
fi

echo ""
echo "==================== API ANALYSIS ===================="

# Analyze main exports
echo "📦 Main Exports:"
echo "----------------------------------------"
grep -n "export" src/index.ts 2>/dev/null || echo "No main exports found"

echo ""
echo "📦 Full Bundle Exports:"
echo "----------------------------------------"
grep -n "export" src/index.full.ts 2>/dev/null || echo "No full bundle exports found"

echo ""
echo "📦 Node Bundle Exports:"
echo "----------------------------------------"
grep -n "export" src/node.ts 2>/dev/null || echo "No node bundle exports found"

echo ""
echo "🏗️  Main Defuddle Class:"
echo "----------------------------------------"
grep -A 10 "export class Defuddle" src/defuddle.ts 2>/dev/null || echo "Defuddle class not found"

echo ""
echo "📋 Type Definitions:"
echo "----------------------------------------"
grep -n "export.*interface\|export.*type" src/types.ts 2>/dev/null || echo "No type exports found"

echo ""
echo "⚙️  Options Interface:"
echo "----------------------------------------"
grep -A 20 "interface.*Options\|type.*Options" src/types.ts 2>/dev/null || echo "Options interface not found"

# If there's a comparison version, show changes
if [ "$COMPARE_VERSION" != "HEAD" ] && git rev-parse "$COMPARE_VERSION" >/dev/null 2>&1; then
    echo ""
    echo "==================== CHANGES ANALYSIS ===================="
    
    echo "📝 Modified Files:"
    echo "----------------------------------------"
    git diff --name-only "$COMPARE_VERSION" HEAD -- src/ | grep -E '\.(ts|js)$' || echo "No source files changed"
    
    echo ""
    echo "🔄 API Changes in Main Files:"
    echo "----------------------------------------"
    
    # Check API changes in main files
    MAIN_FILES=("src/defuddle.ts" "src/types.ts" "src/index.ts" "src/index.full.ts" "src/node.ts")
    
    for file in "${MAIN_FILES[@]}"; do
        if git diff --quiet "$COMPARE_VERSION" HEAD -- "$file"; then
            echo "  ✅ $file - No changes"
        else
            echo "  🔄 $file - Modified"
            echo "     Changes:"
            git diff "$COMPARE_VERSION" HEAD -- "$file" | grep -E '^[+-].*export|^[+-].*interface|^[+-].*type|^[+-].*class|^[+-].*function' | head -10
            echo ""
        fi
    done
    
    echo "📊 Package.json Changes:"
    echo "----------------------------------------"
    if git diff --quiet "$COMPARE_VERSION" HEAD -- package.json; then
        echo "  ✅ No package.json changes"
    else
        echo "  🔄 Package.json modified:"
        git diff "$COMPARE_VERSION" HEAD -- package.json | grep -E '^[+-].*version|^[+-].*dependencies|^[+-].*exports' || echo "  No significant changes"
    fi
fi

echo ""
echo "==================== GO IMPLEMENTATION CHECKLIST ===================="
echo "📋 Items to verify in Go implementation:"
echo "----------------------------------------"
echo "  □ Defuddle struct matches TypeScript class"
echo "  □ Options struct matches TypeScript interface"
echo "  □ All public methods are implemented"
echo "  □ Return types match (DefuddleResult struct)"
echo "  □ Error handling is consistent"
echo "  □ Bundle equivalents are available (core, full, node)"
echo "  □ All exported functions are available"
echo "  □ Type definitions are complete"

echo ""
echo "🔗 Useful commands:"
echo "  View specific file changes: git diff $COMPARE_VERSION HEAD -- <file>"
echo "  View commit history: git log --oneline $COMPARE_VERSION..HEAD"
echo "  Check specific version: $0 <version-tag>"

cd "$PROJECT_ROOT"