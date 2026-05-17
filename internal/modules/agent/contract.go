package agent

import (
	"context"
	_ "embed"
)

// GLOBAL RULES (apply to ALL modules — do NOT remove):
//   1. NEVER import internal/modules/* directly.
//   2. NEVER import internal/domain for entity structs.
//   3. NEVER write SQL outside queries.go.
//   4. NEVER call panic() — return apperrors.Error.
//   5. NEVER put business logic in repository.go.
//   6. ALWAYS emit a domain event on mutation.
//   7. ALWAYS validate inputs with core/validation on boundaries.
//
// MODULE-TYPE RULES (apply to ALL domain modules):
//   1. Status transitions are atomic and emit exactly one domain event.
//   2. Terminal statuses are immutable.
//   3. ALWAYS call core/statemachine.CanTransition before mutating state.
//   4. NEVER call another module's Service methods — use DI interfaces.
//
// MODULE-SPECIFIC RULES (agent only):
//   - Every runtime must implement the Runtime interface completely.
//   - FakeRuntime responses must be deterministic for the same input.
//   - GeminiPlanner returns either a fully valid GraphPlan or an error — no partial plans.
//   - NEVER mutate tasks, work_units, runs, or agent_sessions tables.
//
// ALLOWED core/* imports:
//   - core/apperrors, core/db, core/validation, core/event
//   - core/statemachine, core/transition, core/serialization
// FORBIDDEN core/* imports:
//   - core/coordination (reserved for orchestrator module only)
//
// For full contracts, read CONTRACTS.md in this directory.

//go:embed README.md
//nolint:unused // embed placeholder for architecture test
var _readme string

//go:embed CONTRACTS.md
//nolint:unused // embed placeholder for architecture test
var _contracts string

var ModuleContract = struct {
	Name    string
	Purpose string
}{
	Name:    "agent",
	Purpose: "Define agent runtime interfaces and implementations (fake, gemini, codex-cli, external)",
}

// AgentReader abstracts agent reads to avoid cyclic imports.
type AgentReader interface {
	GetByID(ctx context.Context, id string) (*Agent, error)
}
