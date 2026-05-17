# Contracts: {{MODULE_NAME}}

## Invariants

1. [TODO: invariant 1]
2. [TODO: invariant 2]
3. [TODO: invariant 3]

---

## State Machine

### States

```
[TODO: list all possible statuses]
```

### Valid Transitions

```
[TODO: add valid transitions, e.g.]
created → running → completed
      ↘ failed
```

### Invalid Transitions

```
[TODO: list explicitly forbidden transitions]
terminal → any (terminal statuses are immutable)
```

---

## Boundary Rules

1. NEVER call `Service` methods from other modules directly.
2. NEVER mutate tables belonging to other modules.
3. NEVER write SQL outside `queries.go`.
4. ALWAYS validate inputs at module boundaries using `core/validation`.
5. ALWAYS wrap database errors with `apperrors.Wrap`.

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Invalid input syntax |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | Entity does not exist |
| `CodeInvalidTransition` | State machine violation |
| `CodeConflict` | Idempotency / concurrency violation |
| `CodePersistence` | Database errors |

---

## File Decomposition

[TODO: document any service_*.go files here]

### service_[sub].go
- **Reason:** [why was service.go decomposed?]
- **Rules:** [any specific rules for this sub-file]

---

## Related ADRs

- ADR-0022: Vertical Slice Architecture
- ADR-0025: Module Standardization
