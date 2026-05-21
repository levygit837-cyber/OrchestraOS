#!/usr/bin/env bash
#
# new-domain.sh — Registra um novo domínio/contexto para agentes
#
# Uso:
#   ./scripts/new-domain.sh <nome-do-dominio>
#
# Cria:
#   docs/agent/domains/<nome>/
#   ├── README.md      # Contexto persistente do módulo
#   └── CONTRACTS.md   # Regras de fronteira (opcional)
#
# O nome deve corresponder a um módulo em internal/modules/.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ─── Validações ───
if [[ $# -lt 1 ]]; then
  echo "Erro: nome do domínio é obrigatório." >&2
  echo "Uso: ./scripts/new-domain.sh <nome>" >&2
  exit 1
fi

DOMAIN="$1"

if [[ ! "$DOMAIN" =~ ^[a-z][a-z0-9_-]*$ ]]; then
  echo "Erro: nome do domínio deve ser kebab-case, começando com letra." >&2
  exit 1
fi

MODULE_DIR="${PROJECT_ROOT}/internal/modules/${DOMAIN}"
DOMAIN_DIR="${PROJECT_ROOT}/docs/agent/domains/${DOMAIN}"

# ─── Verifica se módulo existe ───
if [[ ! -d "$MODULE_DIR" ]]; then
  echo "Aviso: módulo '${DOMAIN}' não existe em internal/modules/." >&2
  read -r -p "Deseja criar o domínio mesmo assim? [y/N] " CONFIRM
  if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
    echo "Cancelado." >&2
    exit 0
  fi
fi

# ─── Cria estrutura ───
mkdir -p "$DOMAIN_DIR"

# ─── Descobre arquivos do módulo ───
FILE_LIST=""
if [[ -d "$MODULE_DIR" ]]; then
  FILE_LIST="$(find "$MODULE_DIR" -maxdepth 1 -type f -name '*.go' | sort | sed 's|^.*/||' | sed 's/^/  - /')"
fi

# ─── Gera README.md ───
cat > "${DOMAIN_DIR}/README.md" <<EOF
# Domain: ${DOMAIN}

## Responsabilidade

Descreva em uma frase o que este domínio faz no OrchestraOS.

## Módulo

- **Caminho:** \`internal/modules/${DOMAIN}/\`
- **Tipo:** domínio | coordenação | infra

## Arquivos Principais

${FILE_LIST:-  (módulo ainda não existe)}

## Dependências

Liste pacotes \`core/*\` e outros módulos que este domínio utiliza.

## Regras Específicas

Documente regras que agentes devem seguir ao modificar este domínio.

## Tarefas Recentes

| Data | Task | Status |
|------|------|--------|

## Referências

- ADR:
- Documentação:
EOF

# ─── Gera CONTRACTS.md (opcional) ───
cat > "${DOMAIN_DIR}/CONTRACTS.md" <<EOF
# Contracts: ${DOMAIN}

## Invariantes

- Invariante 1

## State Machine

\`\`\`text
[state] --event--> [next_state]
\`\`\`

## Boundary Rules

- Regra de fronteira 1
EOF

# ─── Output ───
echo ""
echo "✅ Domain registrado: ${DOMAIN}"
echo "   Diretorio: ${DOMAIN_DIR}/"
echo ""
echo "   Arquivos criados:"
ls -1 "${DOMAIN_DIR}" | sed 's/^/      - /'
echo ""

if [[ -d "$MODULE_DIR" ]]; then
  echo "   Módulo existente: ${MODULE_DIR}/"
else
  echo "   Para criar o módulo, execute: ./scripts/new-module.sh ${DOMAIN}"
fi
