# Coding Standards

## Purpose

Rules for every contributor (human or LLM). Violations caught by CI (`golangci-lint`, architecture tests) block merge. Violations not caught by tools are blockers during review.

---

## Package Structure

```
internal/
  domain/       # Pure types — zero internal imports
  planner/      # Task → DAG decomposition
  executor/     # DAG → topological execution
  runtime/      # Agent execution interface
  store/        # Unified persistence (interface + implementations)
  event/        # Event emitter
  apperrors/    # Standardized errors
```

### 3 Golden Rules

1. **domain/ is pure** — zero imports from other internal packages.
2. **Dependencies flow down** — packages depend on domain; never on siblings.
3. **SQL confined to store/** — no SQL patterns in any other package.

---

## Function Complexity

- No function body may exceed **40 lines** (enforced by `TestMaxFunctionComplexity`).
- Extract helpers when a function grows beyond the limit.

---

## Package Size

- No package may exceed **800 lines** of non-test Go code (enforced by `TestPackageSizeLimit`).

---

## Error Handling

1. All errors MUST be `apperrors.Error`.
2. Raw errors MUST be wrapped with `apperrors.Wrap`.
3. Use `apperrors.KindNotFound` for missing entities.
4. Use `apperrors.KindValidation` for input validation failures.
5. Use `apperrors.KindConflict` for concurrency violations.
6. Never ignore an error (`_ = someCall()`). If it truly cannot fail, document why.

---

## Forbidden Patterns

- `panic()` — always return errors
- `fmt.Println` / `fmt.Printf` — use structured logging or return errors
- `helpers.go` or `utils.go` — put reusable code in the right package
- Inline SQL outside `store/`
- Global mutable variables (maps, slices, channels, pointers)

---

## Naming Conventions

| Construct | Convention | Example |
|-----------|-----------|---------|
| Interface | `*er` verb | `Planner`, `Runtime` |
| Constructor | `New*` | `NewHeuristic()` |
| Result struct | `*Result` | `executor.Result` |
| Status constant | `*Status*` | `TaskStatusCreated` |
| Error operation | `package.function` | `planner.parse` |

---

## Testing

1. Unit tests next to the file they test (`*_test.go`).
2. Architecture tests in `tests/architecture/`.
3. Run `make check` before committing.

---

## Adding a New Package

1. Create directory under `internal/`.
2. Add to `allowedImports` in `tests/architecture/architecture_test.go`.
3. Only import `domain/` and declared dependencies.
4. Run `make arch` to verify.
