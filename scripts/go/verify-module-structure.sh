#!/usr/bin/env bash
set -euo pipefail

# verify-module-structure.sh
# Validates that all modules in internal/modules/* have the mandatory files
# per ADR-0022 (Module Standardization).
#
# Usage: ./scripts/go/verify-module-structure.sh
# Exit code: 0 if all modules are valid, 1 otherwise.

MODULES_DIR="internal/modules"
MANDATORY_FILES=(
    doc.go
    contract.go
    README.md
    CONTRACTS.md
    models.go
    events.go
    queries.go
    repository.go
    service.go
    validation.go
)

EXIT_CODE=0
MODULE_COUNT=0
MISSING_COUNT=0

if [ ! -d "$MODULES_DIR" ]; then
    echo "ERROR: Modules directory '$MODULES_DIR' not found"
    exit 1
fi

for module_path in "$MODULES_DIR"/*/; do
    module_name=$(basename "$module_path")
    MODULE_COUNT=$((MODULE_COUNT + 1))

    for file in "${MANDATORY_FILES[@]}"; do
        if [ ! -f "$module_path/$file" ]; then
            echo "ERROR: $module_name missing mandatory file: $file"
            EXIT_CODE=1
            MISSING_COUNT=$((MISSING_COUNT + 1))
        fi
    done
done

echo "---"
echo "Modules checked: $MODULE_COUNT"
if [ $EXIT_CODE -eq 0 ]; then
    echo "Result: All modules have mandatory files ✅"
else
    echo "Result: $MISSING_COUNT mandatory file(s) missing ❌"
fi

exit $EXIT_CODE
