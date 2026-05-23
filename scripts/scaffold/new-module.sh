#!/usr/bin/env bash
set -euo pipefail

# new-module.sh creates a new module from the template in docs/templates/module/.
# Follows ADR-0019 (Simplified Modular Architecture) structure.
# Usage: ./scripts/scaffold/new-module.sh <module-name> [--with-optional]
# Example: ./scripts/scaffold/new-module.sh billing
# Example: ./scripts/scaffold/new-module.sh billing --with-optional

WITH_OPTIONAL=false
MODULE_NAME=""

for arg in "$@"; do
    case "$arg" in
        --with-optional)
            WITH_OPTIONAL=true
            ;;
        -*)
            echo "Unknown flag: $arg"
            echo "Usage: $0 <module-name> [--with-optional]"
            exit 1
            ;;
        *)
            if [ -z "$MODULE_NAME" ]; then
                MODULE_NAME="$arg"
            else
                echo "Usage: $0 <module-name> [--with-optional]"
                exit 1
            fi
            ;;
    esac
done

if [ -z "$MODULE_NAME" ]; then
    echo "Usage: $0 <module-name> [--with-optional]"
    exit 1
fi

MODULE_PATH="internal/modules/${MODULE_NAME}"
TEMPLATE_DIR="docs/templates/module"

if [ -d "${MODULE_PATH}" ]; then
    echo "Error: module '${MODULE_NAME}' already exists at ${MODULE_PATH}"
    exit 1
fi

mkdir -p "${MODULE_PATH}"

# ADR-0019: 5 mandatory files for every module
MANDATORY_FILES="doc.go README.md models.go repository.go service.go"
OPTIONAL_FILES="contract.go CONTRACTS.md events.go queries.go validation.go"

for file in ${MANDATORY_FILES}; do
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

if [ "$WITH_OPTIONAL" = true ]; then
    for file in ${OPTIONAL_FILES}; do
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
    echo "Following ADR-0019 (Simplified Modular Architecture) with 5 mandatory + 5 optional files."
else
    echo "Created module '${MODULE_NAME}' at ${MODULE_PATH}"
    echo "Following ADR-0019 (Simplified Modular Architecture) with 5 mandatory files."
fi

echo "Next steps:"
echo "  1. Edit ${MODULE_PATH}/README.md and fill in responsibilities"
echo "  2. Implement models.go, repository.go, service.go"
echo "  3. Add the service factory to internal/bootstrap/services.go"
echo "  4. Run: go test ./${MODULE_PATH}"
