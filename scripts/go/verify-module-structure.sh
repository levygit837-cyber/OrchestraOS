#!/usr/bin/env bash
set -euo pipefail

# verify-module-structure.sh
# Validates that all modules in internal/modules/* have the mandatory files
# per ADR-0019 (Simplified Modular Architecture).
#
# Usage: ./scripts/go/verify-module-structure.sh
# Exit code: 0 if all modules are valid, 1 otherwise.

MODULES_DIR="internal/modules"

# ADR-0019: 5 mandatory files for every module
MANDATORY_FILES=(
    doc.go
    README.md
    models.go
    repository.go
    service.go
)

# Modules that are coordination-only and do not own a database table
NO_TABLE_MODULES=(
    orchestrator
)

EXIT_CODE=0
MODULE_COUNT=0
MISSING_COUNT=0

if [ ! -d "$MODULES_DIR" ]; then
    echo "ERROR: Modules directory '$MODULES_DIR' not found"
    exit 1
fi

is_no_table_module() {
    local mod="$1"
    for nt in "${NO_TABLE_MODULES[@]}"; do
        if [ "$nt" = "$mod" ]; then
            return 0
        fi
    done
    return 1
}

for module_path in "$MODULES_DIR"/*/; do
    module_name=$(basename "$module_path")
    MODULE_COUNT=$((MODULE_COUNT + 1))

    for file in "${MANDATORY_FILES[@]}"; do
        # orchestrator is allowed to skip repository.go and queries.go
        if [ "$file" = "repository.go" ] && is_no_table_module "$module_name"; then
            continue
        fi

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
