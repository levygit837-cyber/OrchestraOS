#!/usr/bin/env bash
set -euo pipefail

# lint.sh runs all linters and static analysis checks.
# Must pass before opening a PR.

cd "$(dirname "$0")/../.."

echo "=== go vet ./... ==="
go vet ./...

echo "=== go test ./tests/architecture/... ==="
go test ./tests/architecture/... -count=1

echo "=== ./scripts/go/verify-module-structure.sh ==="
./scripts/go/verify-module-structure.sh

echo "=== golangci-lint run ./... ==="
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
else
    echo "WARNING: golangci-lint not found. Run ./scripts/go/install-tools.sh first."
    echo "Skipping golangci-lint (go vet already passed)."
fi

echo "=== Lint passed ==="
