# Playbook de Execucao para Agentes

Este documento define o fluxo obrigatorio que todo agente deve seguir antes e durante a implementacao de qualquer tarefa no OrchestraOS.

---

## 0. Pre-requisitos (sempre)

Antes de iniciar qualquer trabalho:

1. **Ler `AGENTS.md`** — regras de autonomia, commits e padroes de codigo.
2. **Ler `docs/canvas/project-canvas.md`** — entender a visao macro e a proxima fronteira do projeto.
3. **Consultar `docs/adr/`** — verificar se existe decisao arquitetural relevante para a tarefa.

---

## 1. Analise e Definicao

### 1.1 Contexto (sempre)

- Gerar um `BRIEFING.md` (use o template em `docs/agent/templates/BRIEFING.md`).
- Incluir: motivacao da mudanca, escopo explicito (dentro e fora), estado atual do sistema e arquivos relevantes.

### 1.2 Especificacao (quando altera comportamento)

- Se a tarefa muda comportamento observavel ou interfaces externas, gerar `SPEC.md`.
- Definir: entradas, saidas esperadas, edge cases e criterios de aceitacao.

### 1.3 Plano (quando a tarefa e complexa)

- Se a tarefa toca multiplos arquivos ou modulos, gerar `PLAN.md`.
- Escolher o tipo apropriado entre os 4 definidos em `docs/development/plan-types.md`:
  - **Faseado** — dependencias temporais claras
  - **Por Dominio** — modulos independentes
  - **Arvore de Decisoes** — problema aberto, multiplos caminhos
  - **Cenario-Based** — feature user-facing com fluxos variados

---

## 2. Execucao

### 2.1 Checkpoints

- Manter um `PROGRESS.md` atualizado durante a execucao.
- Registrar: passo concluido, falhas encontradas, decisoes tomadas no caminho e arquivos alterados.

### 2.2 Validacoes

- Rodar `go vet` em todo codigo Go alterado.
- Rodar `go test ./...` para os pacotes tocados.
- Rodar `./scripts/verify-contracts.sh` se houver mudanca em contratos.
- Rodar `./scripts/lint.sh` antes de considerar pronto.
- **Nunca commitar se alguma validacao falhar.**

### 2.3 Escopo Minimo

- Implementar a menor mudanca suficiente para atender ao briefing/spec.
- Evitar refatoracoes colaterais nao solicitadas.

---

## 3. Entrega

### 3.1 Revisao Pre-commit

- Gerar `REVIEW.md` com: testes executados, ADRs impactados, riscos residuais e notas para revisor humano.

### 3.2 Commit e PR

- Usar obrigatoriamente: `./scripts/safe-commit.sh "mensagem descritiva"`.
- O script cria feature branch automaticamente se voce estiver na `main`.
- Push da branch e abertura de Pull Request.
- Aguardar CI passar antes de solicitar merge.

### 3.3 Pos-entrega

- Se a mudanca alterar comportamento, arquitetura ou processo, atualizar a documentacao relevante em `docs/`.
- Se for uma decisao arquitetural nova, criar ADR em `docs/adr/`.

---

## Checklist Rapido

```
[ ] AGENTS.md lido
[ ] Canvas consultado
[ ] ADRs relevantes verificados
[ ] BRIEFING.md gerado
[ ] SPEC.md gerado (se aplica)
[ ] PLAN.md gerado (se aplica)
[ ] PROGRESS.md atualizado durante execucao
[ ] Validacoes passaram (go vet, tests, contracts, lint)
[ ] REVIEW.md gerado
[ ] Commit via safe-commit.sh
[ ] PR aberto e CI passou
```
