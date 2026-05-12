# Plans — Estrutura de Planos de Execução

Este diretório contém planos de execução para agentes, organizados por categoria e fase.

## Estrutura

```
plans/
├── README.md                          # Este arquivo
├── active/                            # Planos em execução
│   └── {fase}/
│       └── {ID-do-agente}-{tarefa}/
│           ├── plan.md                # Plano individual do agente
│           └── checklist.md           # Checklist de execução (Ralph Loop)
├── archive/                           # Planos concluídos
│   └── {fase}/
│       └── {ID}-{tarefa}/
│           ├── plan.md
│           └── checklist-completed.md
└── templates/                         # Templates reutilizáveis
    └── modulo-go-completo.md
```

## Convenção de Nomenclatura

```
ORCH-{FASE}-{RODADA}-{AGENTE}-{tarefa}

Exemplo: ORCH-F05-R01-A01-agentservice
  → Orquestrador, Fase 5, Rodada 1, Agente 1, Tarefa: agentservice
```

## Regras

1. **1 Plano = 1 Agente** — Nunca coloque prompts de múltiplos agentes no mesmo arquivo
2. **Checklist acompanha o plano** — Todo plano tem seu `-checklist.md` no mesmo diretório
3. **Mova para archive quando concluído** — Renomeie checklist para `-completed.md` e mova a pasta
4. **Não edite planos em execução** — Se precisar alterar, crie uma nova rodada (R02, R03...)

## Como Usar

### Para o Orquestrador:
```
Crie plano em: plans/active/{fase}/{ID}-{tarefa}/
  → plan.md (prompt completo para o agente)
  → checklist.md (itens não marcados)
```

### Para o Agente Executor:
```
Use a skill execute.
Leia o plano: plans/active/{fase}/{ID}-{tarefa}/plan.md
Siga o Ralph Loop atualizando: plans/active/{fase}/{ID}-{tarefa}/checklist.md
```

## Fases Ativas

| Fase | Diretório | Status | Agentes |
|------|-----------|--------|---------|
| Fase 5 — Orquestração | `active/fase-05-orquestracao/` | Em execução | 3 |

## Fases Arquivadas

Nenhuma.
