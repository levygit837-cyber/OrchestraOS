#!/usr/bin/env bash
set -euo pipefail

# new-module.sh creates a new module from the template in docs/templates/module/.
# Follows ADR-0025 (Module Standardization) structure.
# Usage: ./scripts/scaffold/new-module.sh <module-name>
# Example: ./scripts/scaffold/new-module.sh billing

if [ $# -ne 1 ]; then
    echo "Usage: $0 <module-name>"
    exit 1
fi

MODULE_NAME="$1"
MODULE_PATH="internal/modules/${MODULE_NAME}"
TEMPLATE_DIR="docs/templates/module"

if [ -d "${MODULE_PATH}" ]; then
    echo "Error: module '${MODULE_NAME}' already exists at ${MODULE_PATH}"
    exit 1
fi

mkdir -p "${MODULE_PATH}"

# ADR-0025: 10 mandatory files for every module
for file in doc.go contract.go README.md CONTRACTS.md models.go events.go queries.go repository.go service.go validation.go; do
    src="${TEMPLATE_DIR}/${file}.tmpl"
    if [ ! -f "${src}" ]; then
        src="${TEMPLATE_DIR}/${file}"
    fi
    dst="${MODULE_PATH}/${file}"
    if [ -f "${src}" ]; then
        sed \
            -e "s/{{MODULE_NAME}}/${MODULE_NAME}/g" \
            -e "s/{{PACKAGE}}/${MODULE_NAME}/g" \
            -e "s/{{RESPONSIBILITY}}/TODO: define responsibility/g" \
            -e "s/{{NON_RESPONSIBILITY}}/TODO: define non-responsibility/g" \
            -e "s/{{INVARIANT}}/TODO: define invariant/g" \
            -e "s/{{ALLOWED_MODULE}}/TODO: define allowed module/g" \
            -e "s/{{VALID_TRANSITIONS}}/TODO: define valid transitions/g" \
            -e "s/{{INVALID_TRANSITIONS}}/TODO: define invalid transitions/g" \
            "${src}" > "${dst}"
    fi
done

echo "Created module '${MODULE_NAME}' at ${MODULE_PATH}"
echo "Following ADR-0025 (Module Standardization) with 10 mandatory files."
echo "Next steps:"
echo "  1. Edit ${MODULE_PATH}/README.md and fill in responsibilities"
echo "  2. Edit ${MODULE_PATH}/CONTRACTS.md and define invariants"
echo "  3. Implement models.go, queries.go, repository.go, service.go, validation.go"
echo "  4. Add the service factory to internal/bootstrap/services.go"
echo "  5. Run: go test ./${MODULE_PATH}"
