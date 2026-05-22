#!/usr/bin/env bash
#
# worktree-agent.sh — Gerenciamento de Git Worktrees para Agentes Paralelos
#
# Uso:
#   ./scripts/worktree-agent.sh create <task-slug> [branch-base]
#   ./scripts/worktree-agent.sh list
#   ./scripts/worktree-agent.sh remove <task-slug>
#   ./scripts/worktree-agent.sh status
#
# Regras:
#   - Cada worktree = uma branch exclusiva
#   - Branch base default: master (ou main)
#   - Worktrees ficam em .worktrees/<task-slug>/
#   - NUNCA use git reset --hard ou git clean -fd em worktrees
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WORKTREES_DIR="${PROJECT_ROOT}/.worktrees"

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

log_info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
log_ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

detect_base_branch() {
    cd "$PROJECT_ROOT"
    if git show-ref --verify --quiet refs/heads/master; then
        echo "master"
    elif git show-ref --verify --quiet refs/heads/main; then
        echo "main"
    else
        echo "master"
    fi
}

# ---------------------------------------------------------------------------
# Comando: create
# ---------------------------------------------------------------------------

cmd_create() {
    local task_slug="${1:-}"
    local branch_base="${2:-$(detect_base_branch)}"

    if [[ -z "$task_slug" ]]; then
        log_error "Uso: $0 create <task-slug> [branch-base]"
        exit 1
    fi

    local branch_name="feature/${task_slug}"
    local worktree_path="${WORKTREES_DIR}/${task_slug}"

    cd "$PROJECT_ROOT"

    # Verificar se branch já existe
    if git show-ref --verify --quiet "refs/heads/${branch_name}"; then
        log_warn "Branch '${branch_name}' já existe."
        read -p "Deseja usar a branch existente? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operação cancelada."
            exit 0
        fi
    else
        log_info "Criando branch '${branch_name}' a partir de '${branch_base}'..."
        git branch "$branch_name" "$branch_base"
        log_ok "Branch criada."
    fi

    # Verificar se worktree já existe
    if [[ -d "$worktree_path" ]]; then
        log_warn "Worktree '${worktree_path}' já existe."
        log_info "Abrindo: $worktree_path"
    else
        log_info "Criando worktree em '${worktree_path}'..."
        git worktree add "$worktree_path" "$branch_name"
        log_ok "Worktree criado."
    fi

    # Copiar arquivos de ambiente necessários
    if [[ -f "$PROJECT_ROOT/.env" && ! -f "$worktree_path/.env" ]]; then
        cp "$PROJECT_ROOT/.env" "$worktree_path/.env"
        log_info ".env copiado para worktree."
    fi

    echo ""
    log_ok "Worktree pronto!"
    echo ""
    echo "  Path:     $worktree_path"
    echo "  Branch:   $branch_name"
    echo "  Base:     $branch_base"
    echo ""
    echo "  Próximo passo no Windsurf:"
    echo "    File > Open Folder > $worktree_path"
    echo ""
    echo "  ⚠️  REGRAS:"
    echo "     • NUNCA 'git reset --hard' neste worktree"
    echo "     • NUNCA 'git clean -fd' neste worktree"
    echo "     • SEMPRE commit antes de trocar de aba"
    echo "     • Merge apenas via Pull Request"
}

# ---------------------------------------------------------------------------
# Comando: list
# ---------------------------------------------------------------------------

cmd_list() {
    cd "$PROJECT_ROOT"

    echo ""
    echo "Git Worktrees:"
    echo "--------------"
    git worktree list --porcelain | awk '
        /^worktree/ { path=$2 }
        /^branch/ { branch=$2 }
        /^HEAD/ { head=$2 }
        /^$/ {
            if (path && branch) {
                gsub(/^refs\/heads\//, "", branch)
                printf "  %-40s  %-30s  %s\n", path, branch, substr(head,1,8)
            }
            path=""; branch=""; head=""
        }
    '
    echo ""
}

# ---------------------------------------------------------------------------
# Comando: remove
# ---------------------------------------------------------------------------

cmd_remove() {
    local task_slug="${1:-}"

    if [[ -z "$task_slug" ]]; then
        log_error "Uso: $0 remove <task-slug>"
        exit 1
    fi

    local worktree_path="${WORKTREES_DIR}/${task_slug}"

    if [[ ! -d "$worktree_path" ]]; then
        log_error "Worktree '${worktree_path}' não encontrado."
        exit 1
    fi

    cd "$PROJECT_ROOT"

    log_warn "Removendo worktree '${worktree_path}'..."
    log_warn "Certifique-se de que todos os commits foram pushados!"
    read -p "Continuar? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Operação cancelada."
        exit 0
    fi

    git worktree remove "$worktree_path" 2>/dev/null || {
        log_warn "Worktree tem modificações não commitadas. Forçando remoção..."
        git worktree remove --force "$worktree_path"
    }

    # Remover diretório se sobrou algo
    if [[ -d "$worktree_path" ]]; then
        rm -rf "$worktree_path"
    fi

    log_ok "Worktree removido."
}

# ---------------------------------------------------------------------------
# Comando: status
# ---------------------------------------------------------------------------

cmd_status() {
    cd "$PROJECT_ROOT"

    echo ""
    echo "Status dos Worktrees:"
    echo "---------------------"

    for wt_path in "$WORKTREES_DIR"/*/; do
        [[ -d "$wt_path" ]] || continue
        local slug
        slug=$(basename "$wt_path")

        cd "$wt_path"
        local branch
        branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "?")
        local dirty=""
        if [[ -n $(git status --porcelain 2>/dev/null) ]]; then
            dirty="${YELLOW}(dirty)${NC}"
        else
            dirty="${GREEN}(clean)${NC}"
        fi
        local ahead_behind
        ahead_behind=$(git rev-list --left-right --count origin/${branch}...HEAD 2>/dev/null | awk '{print "behind "$1", ahead "$2}' || echo "no remote")

        echo -e "  ${slug}:"
        echo -e "    branch:  ${branch}"
        echo -e "    status:  ${dirty}"
        echo -e "    remote:  ${ahead_behind}"
    done

    echo ""
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
    local cmd="${1:-help}"

    case "$cmd" in
        create)
            shift
            cmd_create "$@"
            ;;
        list)
            cmd_list
            ;;
        remove)
            shift
            cmd_remove "$@"
            ;;
        status)
            cmd_status
            ;;
        help|--help|-h)
            echo "Uso: $0 {create|list|remove|status} [args...]"
            echo ""
            echo "  create <task-slug> [branch-base]  Cria worktree + branch para uma task"
            echo "  list                              Lista todos os worktrees"
            echo "  remove <task-slug>                Remove um worktree (após merge)"
            echo "  status                            Mostra status detalhado de cada worktree"
            echo ""
            ;;
        *)
            log_error "Comando desconhecido: $cmd"
            echo "Use: $0 help"
            exit 1
            ;;
    esac
}

main "$@"
