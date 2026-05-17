#!/usr/bin/env bash
set -euo pipefail

# new-plan.sh — Scaffold a new plan + checklist from templates
# Usage: ./scripts/new-plan.sh <plan-id> <task-name> [agent-id]
# Example: ./scripts/new-plan.sh ORCH-F05-R01-A01 agentservice

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PLANS_DIR="${REPO_ROOT}/plans/active"
TEMPLATES_DIR="${REPO_ROOT}/plans/templates"

PLAN_ID="${1:-}"
TASK_NAME="${2:-}"
AGENT_ID="${3:-agent}"

if [[ -z "$PLAN_ID" || -z "$TASK_NAME" ]]; then
    echo "Usage: $0 <plan-id> <task-name> [agent-id]"
    echo "Example: $0 ORCH-F05-R01-A01 agentservice"
    exit 1
fi

# Flat structure: plans/active/{ID}-{task-name}/
DEST_DIR="${PLANS_DIR}/${PLAN_ID}-${TASK_NAME}"

mkdir -p "$DEST_DIR"

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Generate plan.md from template
if [[ -f "${TEMPLATES_DIR}/plan.md" ]]; then
    sed \
        -e "s/{TASK_NAME}/${TASK_NAME}/g" \
        -e "s/{PLAN_ID}/${PLAN_ID}/g" \
        -e "s/{AGENT_ID}/${AGENT_ID}/g" \
        -e "s/{ISO_TIMESTAMP}/${TIMESTAMP}/g" \
        "${TEMPLATES_DIR}/plan.md" > "${DEST_DIR}/plan.md"
    echo "✅ Created: ${DEST_DIR}/plan.md"
else
    echo "⚠️ Template not found: ${TEMPLATES_DIR}/plan.md"
fi

# Generate checklist.md from template
if [[ -f "${TEMPLATES_DIR}/checklist.md" ]]; then
    sed \
        -e "s/{TASK_NAME}/${TASK_NAME}/g" \
        -e "s/{PLAN_ID}/${PLAN_ID}/g" \
        -e "s/{AGENT_ID}/${AGENT_ID}/g" \
        -e "s/{ISO_TIMESTAMP}/${TIMESTAMP}/g" \
        "${TEMPLATES_DIR}/checklist.md" > "${DEST_DIR}/checklist.md"
    echo "✅ Created: ${DEST_DIR}/checklist.md"
else
    echo "⚠️ Template not found: ${TEMPLATES_DIR}/checklist.md"
fi

echo ""
echo "Plan scaffolded at: ${DEST_DIR}"
echo "Next steps:"
echo "  1. Edit ${DEST_DIR}/plan.md with task-specific details"
echo "  2. Customize ${DEST_DIR}/checklist.md with specific items"
echo "  3. Commit: git add ${DEST_DIR} && git commit -m 'plan: add ${PLAN_ID} ${TASK_NAME}'"
