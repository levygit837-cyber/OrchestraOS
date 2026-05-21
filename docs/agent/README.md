# Documentação para Agentes

Este diretorio contem tudo que um agente precisa para se situar e operar corretamente no projeto OrchestraOS.

> **Diferenca de `docs/` geral:** os demais diretorios em `docs/` descrevem *o que o sistema e*. Este diretorio descreve *como o agente deve operar aqui*.

---

## Arquivos Principais

| Arquivo        | Proposito                                                                    |
| -------------- | ---------------------------------------------------------------------------- |
| `PLAYBOOK.md`  | Fluxo obrigatorio que todo agente deve seguir antes de implementar qualquer tarefa |

## Templates de Artefatos

Cada template e um ponto de partida para artefatos que o agente gera durante o ciclo de trabalho.

| Template                | Quando usar                                  | Onde guardar                           |
| ----------------------- | -------------------------------------------- | -------------------------------------- |
| `templates/BRIEFING.md` | Inicio de toda tarefa                        | Worktree temporario ou anexo ao plano  |
| `templates/SPEC.md`     | Tarefas que alteram comportamento observavel | Ao lado do plano ou no PR              |
| `templates/PLAN.md`     | Tarefas complexas (multi-arquivo)            | Worktree ou branch da tarefa           |
| `templates/PROGRESS.md` | Durante a execucao                           | Worktree, atualizado a cada checkpoint |
| `templates/REVIEW.md`   | Antes de abrir PR                            | Ao lado do plano ou no corpo do PR     |

---

## Fluxo Resumido

```text
1. Ler AGENTS.md
2. Ler PLAYBOOK.md
3. Gerar artefatos necessarios (BRIEFING, SPEC, PLAN)
4. Executar com PROGRESS
5. Entregar com REVIEW
6. Commit via ./scripts/git/safe-commit.sh
```

---

## Nivel de Autonomia Vigente

Consulte `AGENTS.md` e o canvas do projeto para o nivel aprovado.
Nenhum agente deve assumir autonomia maior que a definida explicitamente.
