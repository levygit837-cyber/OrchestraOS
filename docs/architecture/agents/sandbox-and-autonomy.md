# Sandbox e Autonomia

> Este documento detalha a implementacao operacional do sandbox.
> Para a decisao arquitetural e politica de autonomia, consulte [ADR 0004: Sandbox e Autonomia Inicial](/docs/adr/0004-sandbox-and-autonomy.md).

## Objetivo

Detalhes de execucao do isolamento por task e das restricoes do container.

## Worktree

Diretorio por task:

- Local: `~/.local/share/orchestraos/worktrees/{repo_id}/{task_id}`
- Servidor: `/var/lib/orchestraos/worktrees/{repo_id}/{task_id}`

Prefixo de branch recomendado para runtime Codex/CLI:

```text
codex/task-{task_id}-{slug}
```

## Container

Regras operacionais do Docker para agentes:

- Sem `--privileged`.
- Sem montagem do Docker socket.
- Sem montagem da home do usuario.
- Montar apenas o worktree da task e diretorios explicitamente aprovados.
- Usuario nao-root quando possivel (`--user 1000:1000`).
- Limites de CPU (`--cpus`), memoria (`--memory`), processos (`--pids-limit`) e tempo (`--timeout`).
- Rede bloqueada por padrao (`--network none`) ou liberada por allowlist.
- Segredos injetados via variaveis de ambiente temporarias, nao persistidas em imagem ou volume.

## Rede

- Default: `--network none`.
- Quando necessario: criar rede bridge dedicada por task com allowlist de destinos.
- Registrar todo trafego liberado no Event Store.

## Segredos

- Nunca copiar para worktrees, logs, prompts ou artefatos.
- Injetar via `docker run -e` ou mounts temporarios em `/run/secrets`.
- Registrar qual politica permitiu o acesso no Event Store.

## Ferramentas

Classificacao de risco operacional:

| Risco    | Exemplos                                                | Aprovacao                         |
|----------|---------------------------------------------------------|-----------------------------------|
| `safe`   | Ler arquivo, listar diretorio, formatar codigo          | Auto-aprovada pelo Go             |
| `low`    | Executar testes locais, compilar                        | Auto-aprovada no worktree da task |
| `medium` | Acessar rede restrita, instalar dependencia             | Agente Inteligente (ADR 0023)     |
| `high`   | Escrever fora do worktree, push, PR, comando destrutivo | Escalonar para humano             |

## Checkpoints Operacionais

O agente deve registrar checkpoints antes de:

- Mudar de foco de implementacao.
- Solicitar aprovacao de ferramenta.
- Produzir diff relevante.
- Encerrar uma work unit.

Cada checkpoint deve ser persistido via `AgentSessionService.Checkpoint()` (ADR 0011).

## Encerramento e Limpeza

Fluxo de finalizacao de task:

1. Coletar diff, logs e evidencias.
2. Registrar validacoes no Event Store.
3. Encerrar processo do agente.
4. Parar e remover container.
5. Manter ou remover worktree conforme politica de retencao.
6. Enviar status à CLI/GitHub.

## Evolucao

- gVisor ou Firecracker como evolucao de sandbox quando o risco aumentar.
- Politica de autonomia evolui conforme ADR 0004.
