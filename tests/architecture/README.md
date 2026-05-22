# Architecture Tests

This directory contains architectural guard tests that enforce the rules defined
in [ADR-0030: Simplified Modular Architecture](../../docs/adr/0030-simplified-modular-architecture.md).

## Test Suite

| Test File | Rule Enforced | Description |
|---|---|---|
| `module_boundaries_test.go` | **Pilar 2** — Zero cross-module imports | Verifies that no module under `internal/modules/` imports another module. Only `orchestrator/` and `bootstrap/` may import multiple modules. |
| `repository_purity_test.go` | **Pilar 3** — CRUD-only repositories | Detects business logic in `repository.go` files: status-based branching, deduplication, hardcoded status strings, field validation (`Sequence == 0`), ON CONFLICT, and non-CRUD method names. |
| `domain_import_integrity_test.go` | **Pilar 1** — Domain-centralized types | Verifies that all shared entity types are defined in `internal/domain/` and that modules import `internal/domain/` to use them. |
| `code_anomalies_test.go` | **CODING_STANDARDS.md** | Detects anti-patterns: `panic()`, `fmt.Println`, inline SQL, ignored errors/values/tuples without documentation, and ignored errors in `defer` blocks. |
| `service_decomposition_test.go` | **CODING_STANDARDS.md** | Verifies that `service_<sub>.go` only exists when `service.go` has > 300 lines. |
| `cmd_bootstrap_di_test.go` | **ADR-0030** | Verifies that `cmd/` does not import modules directly (must use `internal/bootstrap/` for DI). |

## Running the Tests

```bash
# Run all architecture tests
go test ./tests/architecture/... -v

# Run a specific test
go test ./tests/architecture/... -run TestModuleBoundaries -v
go test ./tests/architecture/... -run TestRepositoryPurity -v
go test ./tests/architecture/... -run TestDomainImportIntegrity -v
go test ./tests/architecture/... -run TestCodeAnomalies -v
go test ./tests/architecture/... -run TestServiceDecomposition -v
go test ./tests/architecture/... -run TestCmdBootstrapDI -v
```

## Heuristics and Exceptions

### Module Boundaries (`TestModuleBoundaries`)
- **Scope:** All directories under `internal/modules/`
- **Exclusions:** `orchestrator/`, `bootstrap/`
- **Detection:** Any import matching `github.com/levygit837-cyber/OrchestraOS/internal/modules/<other>` is flagged.

### Repository Purity (`TestRepositoryPurity` + `TestRepositoryMethodNames`)
- **Scope:** All `repository.go` files under `internal/modules/` and `internal/core/`
- **Flagged patterns:**
  - `if` statements comparing with `Status*` constants
  - `if existing != nil` (deduplication pattern)
  - `return ... != ... || ... > 1` (reference/upsert detection)
  - Hardcoded status strings: `"active"`, `"inactive"`, `"running"`, `"completed"`, `"failed"`, `"cancelled"`, `"pending"`, `"validated"`, `"ready"`
  - Field validation: `Sequence == 0` (beyond simple nil check)
  - SQL strings containing `ON CONFLICT`
- **Allowed patterns:**
  - `time.Now()` / `time.Now().UTC()` for timestamps (common practice)
  - `scan*` helpers for database row scanning
  - Method names starting with: `Create`, `Get`, `List`, `Update`, `Delete`, `Count`, `Exists`, `Save`, `Insert`, `Find`, `Fetch`, `Remove`, `Query`, `scan`

### Domain Import Integrity (`TestDomainImportIntegrity`)
- **Scope:** `internal/domain/*.go` and `internal/modules/*`
- **Expected types in domain:** `Task`, `Run`, `WorkUnit`, `Agent`, `AgentSession`, `TaskGraph`, `PromptFragment`, `PromptSnapshot`, `ToolsetSnapshot`, `ComposedPrompt`, `Review`, `Trigger` and their associated status/enums.
- **Note:** This test will fail until Task T5 (code refactor) migrates types to `internal/domain/`.

### Code Anomalies (`TestCodeAnomalies`)
- **Scope:** `internal/modules/` and `internal/core/`
- **Flagged patterns:**
  - `panic()` calls
  - `fmt.Println` / `fmt.Printf`
  - SQL strings outside `queries.go`
  - `_ = call()` without documented reason comment
  - `_ = variable` without documented reason comment
  - `_, _ = call()` without documented reason comment
  - `_ = call()` inside `defer func() { ... }()` without documented reason
- **Safe-to-ignore calls:** `Close`, `SetDeadline`, `SetReadDeadline`, `SetWriteDeadline`, `Write`, `Sync`
- **Ignore comments:** Any comment containing `ignore`, `nolint`, `safe to ignore`, `cannot fail`, or `intentionally`

### Service Decomposition (`TestServiceDecomposition`)
- **Scope:** All modules under `internal/modules/`
- **Rule:** `service_<sub>.go` is only permitted when `service.go` has > 300 lines.
- **Detection:** If `service.go` has ≤ 300 lines and `service_*.go` exists, fail.

### Cmd Bootstrap DI (`TestCmdBootstrapDI`)
- **Scope:** All files under `cmd/`
- **Rule:** `cmd/` must not import `internal/modules/*` directly.
- **Detection:** Any import of `github.com/levygit837-cyber/OrchestraOS/internal/modules/<name>` in `cmd/` is flagged.

## Adding Exceptions

If a test flags legitimate code, add an inline comment:

```go
// Safe to ignore: <reason>
_ = someCall()

// nolint:repository-purity — <reason>
if status == StatusRunning {
    // ...
}
```

## Expected State

These tests are designed to **fail with current violations**. The failures prove
the tests detect real problems. Tasks T4 (mapping) and T5 (refactor) will fix
the violations.

| Test | Violations Detected | Fix Task |
|---|---|---|
| `TestModuleBoundaries` | 17 cross-module imports (7 unique pairs) | T5 |
| `TestRepositoryPurity` | 13 business logic occurrences (status, dedup, hardcoded, validation) | T5 |
| `TestRepositoryMethodNames` | 5 non-CRUD methods | T5 |
| `TestDomainImportIntegrity` | 26 types missing from `internal/domain/` | T5 |
| `TestCodeAnomalies` | 18 anomalies (ignored values, inline SQL) | T5 |
| `TestServiceDecomposition` | 1 module with unjustified `service_<sub>.go` | T5 |
| `TestCmdBootstrapDI` | 11 cmd files importing modules directly | T5 |
