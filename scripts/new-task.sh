#!/usr/bin/env bash
#
# new-task.sh — Cria estrutura de artefatos para uma nova task
#
# Uso:
#   ./scripts/new-task.sh --title "Adicionar timeout em sessão de agente" --domain agentsession --type complex
#   ./scripts/new-task.sh --title "Fix typo no README" --type simple
#
# Opções:
#   --title    Título descritivo da task (obrigatório)
#   --domain   Domínio do OrchestraOS (opcional; omita para tasks transversais)
#   --type     Tipo: simple | complex (padrão: simple)
#   --help     Mostra esta ajuda

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ─── Defaults ───
TYPE="simple"
DOMAIN=""
TITLE=""

# ─── Parse Args ───
while [[ $# -gt 0 ]]; do
  case "$1" in
    --title)
      TITLE="$2"
      shift 2
      ;;
    --domain)
      DOMAIN="$2"
      shift 2
      ;;
    --type)
      TYPE="$2"
      shift 2
      ;;
    --help|-h)
      sed -n '2,14p' "$0"
      exit 0
      ;;
    *)
      echo "Erro: opção desconhecida: $1" >&2
      echo "Use --help para ver o uso." >&2
      exit 1
      ;;
  esac
done

# ─── Validações ───
if [[ -z "$TITLE" ]]; then
  echo "Erro: --title é obrigatório." >&2
  exit 1
fi

if [[ "$TYPE" != "simple" && "$TYPE" != "complex" ]]; then
  echo "Erro: --type deve ser 'simple' ou 'complex'." >&2
  exit 1
fi

if [[ -n "$DOMAIN" && ! -d "${PROJECT_ROOT}/internal/modules/${DOMAIN}" ]]; then
  echo "Aviso: módulo '${DOMAIN}' não encontrado em internal/modules/." >&2
  echo "O domain será criado em docs/agent/domains/ mesmo assim." >&2
fi

# ─── Gera identificadores ───
DATE="$(date +%Y-%m-%d)"
SLUG="$(echo "$TITLE" | tr '[:upper:]' '[:lower:]' | sed 'y/àáâãäçèéêëìíîïñòóôõöùúûü/aaaaaceeeeiiiinooooouuuu/' | sed 's/[^a-z0-9]/-/g' | sed 's/-\+/-/g' | sed 's/^-//;s/-$//' | cut -c1-60)"
TASK_ID="${DATE}_${SLUG}"
BRANCH_NAME="feature/${TASK_ID}"

# ─── Determina diretório destino ───
if [[ -n "$DOMAIN" ]]; then
  DEST_DIR="${PROJECT_ROOT}/docs/agent/domains/${DOMAIN}/${TASK_ID}"
  mkdir -p "${DEST_DIR}"
else
  DEST_DIR="${PROJECT_ROOT}/docs/agent/tasks/${TASK_ID}"
  mkdir -p "${DEST_DIR}"
fi

# ─── Copia templates ───
cp "${PROJECT_ROOT}/docs/agent/templates/BRIEFING.md" "${DEST_DIR}/briefing.md"

if [[ "$TYPE" == "complex" ]]; then
  cp "${PROJECT_ROOT}/docs/agent/templates/SPEC.md" "${DEST_DIR}/spec.md"
  cp "${PROJECT_ROOT}/docs/agent/templates/PLAN.md" "${DEST_DIR}/plan.md"
fi

cp "${PROJECT_ROOT}/docs/agent/templates/REVIEW.md" "${DEST_DIR}/review.md"

# ─── Preenche front matter do briefing ───
cat > "${DEST_DIR}/briefing.md" <<EOF
---
tipo: briefing
task-id: ${TASK_ID}
domain: ${DOMAIN:-transversal}
origem: decisao humana
branch: ${BRANCH_NAME}
status: em-andamento
---

# Briefing: ${TITLE}

## Contexto

Descreva o estado atual do sistema e por que esta mudanca e necessaria.
Inclua links para arquivos, issues ou ADRs relevantes.

## Motivacao

- Problema que esta sendo resolvido
- Oportunidade que esta sendo aproveitada
- Custo de nao fazer

## Escopo

### Dentro do escopo

- Item 1
- Item 2

### Fora do escopo

- Item 1
- Item 2

## Arquivos Relevantes

Liste arquivos, pacotes ou modulos que provavelmente serao tocados.
EOF

# ─── Preenche front matter do review ───
cat > "${DEST_DIR}/review.md" <<EOF
---
tipo: review
task-id: ${TASK_ID}
domain: ${DOMAIN:-transversal}
status: pendente
---

# Review: ${TITLE}

## Resumo da Mudanca

Breve descricao do que foi implementado.

## Testes Executados

- [ ] \`go vet\`
- [ ] \`go test ./...\`
- [ ] \`./scripts/verify-contracts.sh\`
- [ ] \`./scripts/lint.sh\`

## ADRs Impactados

- [ ] Nenhum
- [ ] \`docs/adr/XXXX-...\` (atualizado / novo)

## Documentacao Atualizada

- [ ] Nenhuma
- [ ] \`docs/...\` (qual)

## Riscos Residuais

## Notas para Revisor
EOF

# ─── Preenche spec/plan se complexo ───
if [[ "$TYPE" == "complex" ]]; then
  cat > "${DEST_DIR}/spec.md" <<EOF
---
tipo: spec
task-id: ${TASK_ID}
domain: ${DOMAIN:-transversal}
---

# Spec: ${TITLE}

## Resumo

## Entradas

## Saidas Esperadas

## Fluxo Principal

## Edge Cases

## Criterios de Aceitacao

- [ ] Criterio 1

## Notas Tecnicas
EOF

  cat > "${DEST_DIR}/plan.md" <<EOF
---
tipo: faseado
task-id: ${TASK_ID}
domain: ${DOMAIN:-transversal}
---

# Plan: ${TITLE}

## Fase 1: Setup

- [ ] Tarefa

## Fase 2: Core Implementation

- [ ] Tarefa

## Fase 3: Integracao

- [ ] Tarefa

## Fase 4: Testes e Validacao

- [ ] Tarefa
EOF
fi

# ─── Output ───
echo ""
echo "✅ Task criada: ${TASK_ID}"
echo "   Branch sugerida: ${BRANCH_NAME}"
echo "   Diretorio: ${DEST_DIR}"
echo ""
echo "   Artefatos gerados:"
ls -1 "${DEST_DIR}" | sed 's/^/      - /'
echo ""

if [[ -n "$DOMAIN" ]]; then
  echo "   Domain: ${DOMAIN}"
  echo "   Modulo: internal/modules/${DOMAIN}/"
fi
