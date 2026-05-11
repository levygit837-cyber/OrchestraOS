#!/usr/bin/env bash
set -euo pipefail

# verify-contracts.sh checks that module documentation stays in sync with code.
# It runs the architecture test suite which validates:
#   - Every module has README.md, CONTRACTS.md, doc.go, and queries.go
#   - State machines in CONTRACTS.md mention all status constants from models.go
#   - README.md Allowed Dependencies reflect actual imports
#   - Module boundary rules are respected

cd "$(dirname "$0")/.."

echo "=== Running architecture and contract verification tests ==="
go test ./tests/architecture/... -count=1

echo "=== Contract verification passed ==="
