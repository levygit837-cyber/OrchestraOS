#!/usr/bin/env bash
set -euo pipefail

# check_module_size.sh warns when a module grows beyond the recommended limit.
# The goal is to keep each module "head-sized" for LLM context efficiency.
#
# Default thresholds:
#   WARNING  = 800 lines  (soft limit — consider refactoring)
#   CRITICAL = 1200 lines (hard limit — must split or simplify)

WARNING_LIMIT=800
CRITICAL_LIMIT=1200

EXIT_CODE=0

echo "=== Module Size Check ==="
echo ""

for mod_dir in internal/modules/*; do
    if [ ! -d "${mod_dir}" ]; then
        continue
    fi

    mod_name=$(basename "${mod_dir}")
    # Count non-test Go files only
    total_lines=$(find "${mod_dir}" -maxdepth 1 -name '*.go' ! -name '*_test.go' -exec wc -l {} + | tail -1 | awk '{print $1}')

    if [ -z "${total_lines}" ] || [ "${total_lines}" -eq 0 ]; then
        total_lines=0
    fi

    status="OK"
    if [ "${total_lines}" -gt "${CRITICAL_LIMIT}" ]; then
        status="CRITICAL"
        EXIT_CODE=1
    elif [ "${total_lines}" -gt "${WARNING_LIMIT}" ]; then
        status="WARNING"
    fi

    printf "  %-20s %5s lines  [%s]\n" "${mod_name}" "${total_lines}" "${status}"
done

echo ""

if [ "${EXIT_CODE}" -ne 0 ]; then
    echo "❌ CRITICAL: one or more modules exceed ${CRITICAL_LIMIT} lines."
    echo "   Consider splitting domain logic or moving reusable code to internal/core/."
    echo "   Thresholds: WARNING=${WARNING_LIMIT}, CRITICAL=${CRITICAL_LIMIT}"
else
    echo "✅ All modules are within acceptable size limits."
fi

exit ${EXIT_CODE}
