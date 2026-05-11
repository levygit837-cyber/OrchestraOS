#!/usr/bin/env bash
set -euo pipefail

# safe-commit.sh — wrapper that validates code, creates a feature branch, and commits.
# Usage: ./scripts/safe-commit.sh "commit message"
#
# This script ensures you never commit to main accidentally.

if [ $# -lt 1 ]; then
    echo "Usage: $0 \"commit message\""
    exit 1
fi

MSG="$1"
CURRENT_BRANCH=$(git branch --show-current)

if [ "$CURRENT_BRANCH" = "main" ]; then
    echo "❌ You are on 'main'. Creating a feature branch instead..."
    TIMESTAMP=$(date +%s)
    BRANCH_NAME="auto/change-${TIMESTAMP}"
    git checkout -b "$BRANCH_NAME"
    echo "✅ Created and switched to branch: $BRANCH_NAME"
fi

echo "=== Running pre-commit validations ==="
./scripts/pre-commit.sh

echo "=== Committing ==="
git add -A
git commit -m "$MSG"

echo ""
echo "✅ Committed successfully."
echo "Next step: git push origin $(git branch --show-current)"
