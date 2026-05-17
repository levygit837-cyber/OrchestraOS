# Plans — Estrutura de Planos de Execução

Este diretório contém planos de execução para agentes. Os planos são **decompostos** em micro-tarefas independentes, cada uma com escopo pequeno, verificável e de fácil execução. Toda a serialização segue uma convenção padronizada que permite rastreabilidade, paralelização e reexecução controlada.

## Estrutura

```text
plans/
├── README.md                  # Este arquivo
├── active/                    # Planos em execução
│   └── {fase}/
│       └── {ID-do-plano}/
│           ├── plan.md        # Prompt completo com contexto, escopo e regras
│           └── checklist.md   # Checklist de execução (Ralph Loop)
├── completed/                 # Planos finalizados
│   └── {fase}/
│       └── {ID-do-plano}/
│           ├── plan.md
│           └── checklist.md   # Checklist completo com todos os itens marcados
└── templates/                 # Templates reutilizáveis
    └── modulo-go-completo.md
```

## Convenção de Serialização

Todo plano recebe um identificador único e imutável.

Formato.

```text
ORCH-{FASE}-{RODADA}-{AGENTE}-{tarefa}
```

Exemplo.

```text
ORCH-F05-R01-A01-agentservice
```

| Segmento | Significado | Exemplo |
| --- | --- | --- |
| `ORCH` | Prefixo fixo (Orquestrador) | — |
| `{FASE}` | Identificador da fase (ex: F05, F28) | `F05` = Fase 5 |
| `{RODADA}` | Iteração dentro da fase (R01, R02...) | `R01` = primeira rodada |
| `{AGENTE}` | Identificador do agente executor | `A01` = Agente 1 |
| `{tarefa}` | Descrição curta do escopo (kebab-case) | `agentservice` |

**Exemplo completo:** `ORCH-F05-R01-A01-agentservice` significa Orquestrador, Fase 5, Rodada 1, Agente 1, Tarefa: implementar AgentService.

## Princípio — Planos Decompostos

Cada plano representa **uma única unidade de trabalho** que um único agente pode executar de ponta a ponta sem precisar de coordenação externa. Não agrupamos múltiplas responsabilidades em um único plano.

### Características de um plano válido

- **Escopo pequeno** — Pode ser completado em poucas iterações do Ralph Loop
- **Independente** — Não bloqueia outros planos (ou declara dependências explícitas)
- **Verificável** — Tem critérios de aceite claros e testáveis
- **Isolado** — Define fronteiras de código que pode e não pode tocar
- **Serializado** — ID único permite rastreamento em canvas, ADRs e commits

## Ciclo de Vida de um Plano

### 1. Criação (Orquestrador)

```text
plans/active/{fase}/{ID-do-plano}/
  → plan.md      (contexto, escopo, regras, critérios de aceite)
  → checklist.md (itens pendentes, não marcados)
```

### 2. Execução (Agente Executor)

```text
Use a skill execute.
Leia o plano:  plans/active/{fase}/{ID-do-plano}/plan.md
Siga o Ralph Loop atualizando: plans/active/{fase}/{ID-do-plano}/checklist.md
```

A cada iteração:

1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (testes passam? build verde?)
4. **ATUALIZAR** o checklist marcando o item como concluído (`[x]`)
5. **CONTINUAR** para o próximo item

### 3. Conclusão

Quando todos os itens do checklist estiverem marcados:

- Mova o diretório de `plans/active/{fase}/` para `plans/completed/{fase}/`
- O checklist permanece no estado final (tudo `[x]`) — **não renomeie** para `-completed.md`

## Regras

1. **1 Plano = 1 Agente = 1 Micro-tarefa** — Nunca coloque prompts de múltiplos agentes no mesmo arquivo
2. **Checklist acompanha o plano** — Todo plano tem seu `checklist.md` no mesmo diretório
3. **Mova para completed quando concluído** — Transfira o diretório inteiro para `plans/completed/{fase}/`
4. **Não edite planos em execução** — Se precisar alterar, crie uma nova rodada (R02, R03...)
5. **Commits via safe-commit** — Use `./scripts/safe-commit.sh` ao final de cada ciclo significativo

## Fases Ativas

| Fase | Diretório | Status | Planos |
| --- | --- | --- | --- |
| Fase 28 — Refinamento | `active/f28-r01/` | Em execução | 6 |

## Fases Completadas

| Fase | Diretório | Status | Planos |
| --- | --- | --- | --- |
| Fase 5 — Orquestração | `completed/fase-05-orquestracao/` | Concluído | 9 |
