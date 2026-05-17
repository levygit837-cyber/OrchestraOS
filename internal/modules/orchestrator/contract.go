package orchestrator

import (
	_ "embed"
)

// GLOBAL RULES (apply to ALL modules — do NOT remove):
//   1. NEVER import internal/modules/* for services, repositories, or business logic.
//      ALLOWED: import types (structs, enums) from another module ONLY for DI
//      interface return types. See ADR-0026 for full policy.
//   2. NEVER import internal/domain for entity structs or entity enums.
//      ALLOWED: EventEnvelope, EventPriority, checkpoint types, generic payloads.
//   3. NEVER write SQL outside queries.go.
//   4. NEVER call panic() — return apperrors.Error.
//   5. NEVER put business logic in repository.go.
//   6. ALWAYS emit a domain event on mutation.
//   7. ALWAYS validate inputs with core/validation on boundaries.
//
// MODULE-TYPE RULES (orchestrator is NOT a domain module — it coordinates them):
//   - OrchestratorService is the ONLY component that may coordinate cross-module operations.
//   - NEVER import repositories directly; use only domain services injected via Dependencies.
//
// MODULE-SPECIFIC RULES (orchestrator only):
//   - RunTask must be deterministic and fully auditable via events.
//   - Work units are executed sequentially in the first cut.
//   - This is the ONLY module allowed to import core/coordination.
//
// ALLOWED core/* imports:
//   - core/apperrors, core/db, core/validation, core/event
//   - core/statemachine, core/transition, core/serialization
//   - core/coordination (exclusive permission)
// FORBIDDEN core/* imports:
//   - None (orchestrator has broad permissions, but still must not abuse them)
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
	Name:    "orchestrator",
	Purpose: "Central orchestration service that automates end-to-end task execution across all domain services",
}

// Event types emitted by the orchestrator.
const (
	EventTaskStarted       = "orchestrator.task_started"
	EventWorkUnitStarted   = "orchestrator.work_unit_started"
	EventWorkUnitCompleted = "orchestrator.work_unit_completed"
	EventWorkUnitFailed    = "orchestrator.work_unit_failed"
	EventTaskCompleted     = "orchestrator.task_completed"
	EventTaskFailed        = "orchestrator.task_failed"
)

// Valid runtime types.
const (
	RuntimeTypeFake     = "fake"
	RuntimeTypeGemini   = "gemini"
	RuntimeTypeCodexCLI = "codex_cli"
)

// Valid planner strategies.
const (
	PlannerStrategyLocalHeuristic = "local_heuristic_v1"
	PlannerStrategyLLMGemini      = "llm_gemini_v1"
)

// Default configuration values.
const (
	DefaultMaxSteps       = 10
	DefaultTimeoutSeconds = 300
)
