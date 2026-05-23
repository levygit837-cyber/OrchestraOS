# 0021. Agent-Based DAG Generation Pipeline

**Status:** Accepted
**Data:** 2026-05-23
**Extends:** ADR-0020 (Thin Orchestrator Pipeline)

---

## 1. Contexto

O Thin Orchestrator (ADR-0020) usa um planner heurístico que decompõe tasks a partir de `acceptance_criteria`. Embora funcional e determinístico, essa abordagem não entende semântica, contexto ou intenção. Tasks complexas que envolvem múltiplos domínios (auth, runtime, database, API) são divididas apenas por peso textual, não por separação contextual.

### 1.1 Problemas identificados

| Problema | Evidência |
|----------|-----------|
| Sem separação contextual | WUs misturam domínios diferentes |
| Planner fixo | Não há estratégia alternativa ao heurístico |
| Sem retry de geração | Falha no planner aborta todo o pipeline |
| Agentes fixos | Agente é atribuído uma vez, sem troca em runtime |

---

## 2. Decisão

Implementar um **pipeline de geração de DAG baseado em agente** com 3 novos packages:

```
internal/
  daggen/        # Construção e validação de DAG graphs
  decomposer/    # Pipeline de decomposição com estratégias (agent/heuristic)
  assignment/    # Acoplamento de agentes a WUs em runtime
```

### 2.1 Domain Types

Novos tipos em `domain/` (pure, zero deps):

- `DAGNode`, `DAGEdge`, `DAGGraph` — estrutura do grafo
- `TaskContext`, `WUContext` — contexto separado por domínio
- `DecompositionRequest`, `DecompositionResult`, `WUSpec` — pipeline I/O
- `AgentAssignment` — acoplamento agente↔WU com lifecycle

### 2.2 Pipeline

```
Task → DecompositionRequest
     → Strategy.Decompose() → DecompositionResult (WUSpecs + DAGGraph)
     → daggen.BuildGraph() → validated DAGGraph
     → daggen.BuildWorkUnits() → []WorkUnit
     → assignment.Assign() → AgentAssignment per WU
```

### 2.3 Regras

1. **Separação contextual** — cada WU tem um único domínio de contexto.
2. **Validação obrigatória** — grafos não-acíclicos são rejeitados.
3. **Retry com preservação** — falhas na geração usam exponential backoff via `retry/`.
4. **WU referencia Task-mãe** — `WUSpec.NodeID` → `DAGNode.ID`; `WorkUnit.TaskID` → `Task.ID`.
5. **Agentes acoplados em runtime** — podem ser trocados via `assignment.Replace()`.

### 2.4 Estratégias de Decomposição

| Estratégia | Implementação | Descrição |
|---|---|---|
| `agent_llm_v1` | `decomposer/agent_strategy.go` | Usa LLM para decomposição semântica com separação por domínio |
| Futura: `heuristic_v2` | — | Wrapper do planner heurístico existente como Strategy |

### 2.5 Error Kinds

Novos em `apperrors/`:

- `KindGraphGeneration` — falha na geração do grafo (retryable)
- `KindGraphValidation` — grafo inválido (não retryable)
- `KindDecomposition` — falha na decomposição
- `KindWorkUnitInvalid` — WU referencia nó inexistente

---

## 3. Consequências

### Positivas

- Tasks são decompostas com entendimento semântico e contextual
- DAGs são validados antes de gerar WorkUnits
- Retry exponencial para falhas de geração
- Agentes podem ser trocados em runtime
- Novos packages seguem as 3 regras do Thin Orchestrator
- Testes de arquitetura atualizados para os novos packages

### Negativas

- `decomposer/` depende de `daggen/` e `retry/` (3 deps internas)
- Agent strategy requer LLM runtime configurado

---

## 4. Testes

| Package | Testes | Cobertura |
|---------|--------|-----------|
| `daggen` | 13 testes (validação + construção) | Ciclos, duplicatas, refs inválidas, campos vazios |
| `decomposer` | 7 testes (pipeline + agent strategy) | Success, errors, retry, parsing |
| `assignment` | 8 testes (assign, replace, remove) | Lifecycle completo |
| `architecture` | Atualizado com 3 novos packages | Deps, size, purity |
