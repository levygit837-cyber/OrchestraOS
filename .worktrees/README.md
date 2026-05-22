# Worktrees Directory

Diretório para Git worktrees de agentes paralelos.
CADA WORKTREE É UMA BRANCH SEPARADA. Nunca duas worktrees na mesma branch.

## Regras Absolutas

1. Uma worktree = uma branch = uma task
2. NUNCA rode `git reset --hard` ou `git clean -fd` em uma worktree
3. Sempre commit antes de trocar de aba no Windsurf
4. Merge só via PR, nunca merge local entre worktrees

## Worktrees Ativas

| Worktree | Branch | Task | Agente | Status |
|----------|--------|------|--------|--------|
| main | master | - | - | base |
| t4-pattern-mapping | feature/t4-pattern-mapping | T4: Mapeamento de padrões e tipos para domain/ | Kimi K2.6 (Aba 1) | 🟢 Criado |
| t5-code-refactor | feature/t5-code-refactor | T5: Refatoração de código (migração de tipos) | Kimi K2.6 (Aba 2) | 🟢 Criado |

## Como Abrir no Windsurf

### Aba 1 — Task T4 (Mapeamento)
```
File > Open Folder > /home/levybonito/Documentos/OrchestraOS/.worktrees/t4-pattern-mapping
```
**Prompt inicial para o agente:**
> Você está no worktree `t4-pattern-mapping` (branch `feature/t4-pattern-mapping`).
> Sua missão é executar a Task T4: Mapear todos os tipos compartilhados que precisam migrar de `internal/modules/*/models.go` para `internal/domain/`.
> Leia o briefing em `docs/agent/tasks/2026-05-21_architecture-patterns-and-refactor-mapping/briefing.md`.
> REGRAS: NUNCA `git reset --hard` ou `git clean -fd`. SEMPRE commit antes de encerrar.

### Aba 2 — Task T5 (Refatoração)
```
File > Open Folder > /home/levybonito/Documentos/OrchestraOS/.worktrees/t5-code-refactor
```
**Prompt inicial para o agente:**
> Você está no worktree `t5-code-refactor` (branch `feature/t5-code-refactor`).
> Sua missão é executar a Task T5: Refatorar o código movendo tipos para `internal/domain/` e atualizando imports.
> Leia o briefing em `docs/agent/tasks/2026-05-21_code-refactor-for-architecture-violations/briefing.md`.
> REGRAS: NUNCA `git reset --hard` ou `git clean -fd`. SEMPRE commit antes de encerrar.

## Comandos

### Criar worktree para nova task
```bash
./scripts/worktree-agent.sh create <task-slug> <branch-base>
```

### Listar worktrees
```bash
./scripts/worktree-agent.sh list
```

### Ver status detalhado
```bash
./scripts/worktree-agent.sh status
```

### Remover worktree (após merge)
```bash
./scripts/worktree-agent.sh remove <task-slug>
```
