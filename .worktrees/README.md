# Worktrees Directory

Diretório para Git worktrees de agentes paralelos.

## Regras Absolutas

1. Uma worktree = uma branch = uma task
2. NUNCA rode `git reset --hard` ou `git clean -fd` em uma worktree
3. Sempre commit antes de trocar de aba no Windsurf
4. Merge só via PR, nunca merge local entre worktrees

## Worktrees Ativas

Nenhuma worktree ativa no momento.

## Comandos

```bash
./scripts/worktree-agent.sh create <task-slug> [branch-base]
./scripts/worktree-agent.sh list
./scripts/worktree-agent.sh status
./scripts/worktree-agent.sh remove <task-slug>
```
