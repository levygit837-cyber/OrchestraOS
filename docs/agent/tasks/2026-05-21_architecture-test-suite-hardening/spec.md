---
tipo: spec
task-id: 2026-05-21_architecture-test-suite-hardening
domain: transversal
status: em-andamento
---

# Spec: Architecture Test Suite Simplification

## Testes a Implementar

### 1. TestModuleBoundaries (simplificado)

**Lógica:**
1. Para cada módulo em `internal/modules/*`, parsear todos os arquivos `.go` (exceto `_test.go`).
2. Extrair imports.
3. Se qualquer import for de `github.com/levygit837-cyber/OrchestraOS/internal/modules/OUTRO_MODULO`, falhar.
4. **Exceções:** `orchestrator/` e `bootstrap/` podem importar múltiplos módulos.

**Mensagem de erro:**
```
module "run" imports "task" — modules must NOT import other modules (ADR-0030).
Only orchestrator/ and bootstrap/ may import multiple modules.
```

### 2. TestRepositoryPurity

**Heurísticas de detecção:**
1. **Status-based logic:** Procurar `if` statements em `repository.go` que comparem variáveis com constants `Status*`.
2. **Time-based side effects:** Procurar `time.Now()` ou `time.Now().UTC()` em `repository.go`.
3. **Upsert/dedup logic:** Procurar `ON CONFLICT` em strings SQL dentro de `repository.go`.
4. **Validation logic:** Procurar validações de campos (`if id == ""`, `if sequence < 0`).

**Escopo:** `internal/modules/*/repository.go` + `internal/core/*/repository.go`

### 3. TestDomainImportIntegrity

**Lógica:**
1. Parsear `internal/domain/*.go`.
2. Verificar que todos os entity types compartilhados estão definidos lá.
3. Verificar que `internal/modules/*` importam `internal/domain` para usar esses tipos (não definem seus próprios).

**Entity types que devem estar em domain/:**
- Task, TaskStatus, TaskPriority, TaskRiskLevel
- Run, RunStatus, RunResult
- WorkUnit, WorkUnitStatus
- Agent, AgentRuntimeType, AgentStatus
- AgentSession, AgentSessionStatus
- TaskGraph, TaskGraphStatus
- PromptFragment, PromptSnapshot, ToolsetSnapshot, ComposedPrompt
- Review, ReviewStatus, ReviewValidationGate
- Trigger, TriggerStatus, TriggerType

### 4. TestCodeAnomalies

**Detecções:**
1. `_ = variável` (não apenas `_ = call()`)
2. `_ = call()` dentro de `defer func() { ... }()`
3. SQL sem `FROM` (inline SQL strings)
4. `panic()` em qualquer lugar
5. `fmt.Println` / `fmt.Printf` em qualquer lugar

## Edge Cases
- Build tags que excluem a plataforma atual devem ser skipados.
- Pacotes com `//go:build ignore` devem ser ignorados.
- `orchestrator` e `bootstrap` são excluídos de `TestModuleBoundaries`.

## Critérios de Aceitação
- [ ] Todos os 4 testes compilam e rodam
- [ ] Testes falham com as violações atuais (prova de detecção)
- [ ] Mensagens de erro são claras e acionáveis
