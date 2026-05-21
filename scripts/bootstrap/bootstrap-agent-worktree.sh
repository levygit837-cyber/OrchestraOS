#!/bin/bash
# Bootstrap Agent Worktree
# Cria branch e worktree isolada para um agente executor
#
# Uso: ./scripts/bootstrap/bootstrap-agent-worktree.sh <AGENTE_ID> <NOME_TAREFA>
# Exemplo: ./scripts/bootstrap/bootstrap-agent-worktree.sh A01 agentservice

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_NAME="$(basename "$REPO_ROOT")"

AGENT_ID="${1:-}"
TASK_NAME="${2:-}"

if [ -z "$AGENT_ID" ] || [ -z "$TASK_NAME" ]; then
    echo "Uso: $0 <AGENTE_ID> <NOME_TAREFA>"
    echo "Exemplo: $0 A01 agentservice"
    exit 1
fi

BRANCH_NAME="agent-${AGENT_ID}/${TASK_NAME}"
WORKTREE_NAME="${REPO_NAME}-${AGENT_ID}"
WORKTREE_PATH="$(dirname "$REPO_ROOT")/${WORKTREE_NAME}"

echo "=== Bootstrap Agent Worktree ==="
echo "Agente ID: $AGENT_ID"
echo "Tarefa: $TASK_NAME"
echo "Branch: $BRANCH_NAME"
echo "Worktree: $WORKTREE_PATH"
echo ""

cd "$REPO_ROOT"

# Verificar se já existe branch
if git show-ref --verify --quiet "refs/heads/${BRANCH_NAME}"; then
    echo "⚠️  Branch '${BRANCH_NAME}' já existe."
    read -p "Deseja recriar? (s/N): " RECREATE
    if [[ "$RECREATE" =~ ^[Ss]$ ]]; then
        git branch -D "$BRANCH_NAME" || true
    else
        echo "Usando branch existente."
    fi
fi

# Verificar se já existe worktree
if git worktree list | grep -q "$WORKTREE_PATH"; then
    echo "⚠️  Worktree '${WORKTREE_PATH}' já existe."
    read -p "Deseja recriar? (s/N): " RECREATE
    if [[ "$RECREATE" =~ ^[Ss]$ ]]; then
        git worktree remove "$WORKTREE_PATH" --force || true
        rm -rf "$WORKTREE_PATH"
    else
        echo "Usando worktree existente."
        echo ""
        echo "✅ Agente pronto:"
        echo "   Branch: $BRANCH_NAME"
        echo "   Worktree: $WORKTREE_PATH"
        exit 0
    fi
fi

# Detectar branch principal
MAIN_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "📌 Branch principal detectada: ${MAIN_BRANCH}"

# Criar branch a partir da principal atual
echo "🌿 Criando branch '${BRANCH_NAME}'..."
git checkout -b "$BRANCH_NAME"

# Voltar para principal
git checkout "$MAIN_BRANCH"

# Criar worktree
echo "📁 Criando worktree em '${WORKTREE_PATH}'..."
git worktree add "$WORKTREE_PATH" "$BRANCH_NAME"

echo ""
echo "✅ Agente ${AGENT_ID} pronto:"
echo "   Branch: $BRANCH_NAME"
echo "   Worktree: $WORKTREE_PATH"
echo ""
echo "Próximos passos:"
echo "   1. Abra o workspace do agente em: ${WORKTREE_PATH}"
echo "   2. Chame o agente com o plano correspondente"
echo ""
echo "Para remover depois:"
echo "   git worktree remove ${WORKTREE_PATH}"
echo "   git branch -D ${BRANCH_NAME}"
