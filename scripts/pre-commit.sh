#!/usr/bin/env bash

# pre-commit.sh — install to .git/hooks/pre-commit to run checks before every commit.
# Usage: cp scripts/pre-commit.sh .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit

set -euo pipefail

echo "=== Pre-commit checks ==="

echo "--- go vet ./... ---"
go vet ./...

echo "--- go test ./tests/architecture/... ---"
go test ./tests/architecture/... -count=1

echo "--- ./scripts/verify-contracts.sh ---"
./scripts/verify-contracts.sh

echo "=== Pre-commit checks passed ==="
