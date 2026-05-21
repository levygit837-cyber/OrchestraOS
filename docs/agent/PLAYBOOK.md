# Playbook de Execucao para Agentes

Este documento define o fluxo obrigatorio que todo agente deve seguir antes e durante a implementacao de qualquer tarefa no OrchestraOS.

---

## 0. Pre-requisitos (sempre)

Antes de iniciar qualquer trabalho:

1. **Ler `AGENTS.md`** — regras de autonomia, commits e padroes de codigo.
2. **Ler `docs/canvas/project-canvas.md`** — entender a visao macro e a proxima fronteira do projeto.
3. **Consultar `docs/adr/`** — verificar se existe decisao arquitetural relevante para a tarefa.
4. **Ler `docs/agent/ARTIFACT_ORGANIZATION.md`** — entender onde e como guardar artefatos.

---

## 1. Analise e Definicao

### 1.1 Contexto (sempre)

- Identificar se a task pertence a um domain (`internal/modules/`) ou e transversal.
- Usar `./scripts/scaffold/new-task.sh --title "..." --domain <domain> --type <simple|complex>` para criar a estrutura.
- Preencher `briefing.md` com: motivacao, escopo (dentro e fora), estado atual e arquivos relevantes.

### 1.2 Especificacao (quando altera comportamento)

- Se a tarefa muda comportamento observavel ou interfaces externas, preencher `spec.md`.
- Definir: entradas, saidas esperadas, edge cases e criterios de aceitacao.

### 1.3 Plano (quando a tarefa e complexa)

- Se a tarefa toca multiplos arquivos ou modulos, preencher `plan.md`.
- Escolher o tipo apropriado entre os 4 definidos em `docs/development/plan-types.md`:
  - **Faseado** — dependencias temporais claras
  - **Por Dominio** — modulos independentes
  - **Arvore de Decisoes** — problema aberto, multiplos caminhos
  - **Cenario-Based** — feature user-facing com fluxos variados

### 1.4 Front Matter Obrigatorio

Todo artefato deve incluir no topo:

```yaml
---
tipo: briefing | spec | plan | progress | review
task-id: 2026-05-21_slug-da-task
domain: <modulo> | transversal
origem: issue #42 | comando CLI | decisao humana
branch: feature/2026-05-21_slug-da-task
status: em-andamento | concluido | cancelado
---
```

---

## 2. Execucao

### 2.1 Checkpoints

- Manter `progress.md` atualizado no worktree (`.orchestraos/artifacts/<task-id>/`) durante a execucao.
- Registrar: passo concluido, falhas encontradas, decisoes tomadas e arquivos alterados.
- O progress.md e efemero — nao precisa ser movido para o repo principal apos entrega.

### 2.2 Validacoes

- Rodar `go vet` em todo codigo Go alterado.
- Rodar `go test ./...` para os pacotes tocados.
- Rodar `./scripts/go/verify-contracts.sh` se houver mudanca em contratos.
- Rodar `./scripts/go/lint.sh` antes de considerar pronto.
- **Nunca commitar se alguma validacao falhar.**

### 2.3 Escopo Minimo

- Implementar a menor mudanca suficiente para atender ao briefing/spec.
- Evitar refatoracoes colaterais nao solicitadas.

---

## 3. Entrega

### 3.1 Revisao Pre-commit

- Gerar `REVIEW.md` com: testes executados, ADRs impactados, riscos residuais e notas para revisor humano.

### 3.2 Commit e PR

- Usar obrigatoriamente: `./scripts/git/safe-commit.sh "mensagem descritiva"`.
- O script cria feature branch automaticamente se voce estiver na `main`.
- Push da branch e abertura de Pull Request.
- Aguardar CI passar antes de solicitar merge.

### 3.3 Pos-entrega

- Mover artefatos relevantes (`briefing.md`, `spec.md`, `plan.md`, `review.md`) do worktree para o repo principal:
  - Task de um domain: `docs/agent/domains/<domain>/<task-id>/`
  - Task transversal: `docs/agent/tasks/<task-id>/`
- Se a mudanca alterar comportamento, arquitetura ou processo, atualizar a documentacao relevante em `docs/`.
- Se for uma decisao arquitetural nova, criar ADR em `docs/adr/`.

---

## Checklist Rapido

```
[ ] AGENTS.md lido
[ ] Canvas consultado
[ ] ADRs relevantes verificados
[ ] Task criada via `./scripts/scaffold/new-task.sh`
[ ] BRIEFING.md preenchido
[ ] SPEC.md preenchido (se aplica)
[ ] PLAN.md preenchido (se aplica)
[ ] PROGRESS.md atualizado durante execucao
[ ] Validacoes passaram (go vet, tests, contracts, lint)
[ ] REVIEW.md preenchido
[ ] Artefatos movidos para docs/agent/
[ ] Commit via safe-commit.sh
[ ] PR aberto e CI passou
```
