#!/usr/bin/env bash

# pre-commit.sh — install to .git/hooks/pre-commit to run checks before every commit.
# Usage: make setup

set -euo pipefail

echo "=== Pre-commit checks ==="

echo "--- go vet ./... ---"
go vet ./...

echo "--- go test ./tests/architecture/... ---"
go test ./tests/architecture/... -count=1

echo "=== Pre-commit checks passed ==="
