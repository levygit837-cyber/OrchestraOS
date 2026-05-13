package orchestrator

import (
	_ "embed"
)

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. OrchestratorService is the ONLY component that may coordinate cross-module operations.
//   2. NEVER import repositories directly; use only domain services injected via Dependencies.
//   3. RunTask must be deterministic and fully auditable via events.
//   4. Work units are executed sequentially in the first cut.
//
// For full contracts, read CONTRACTS.md in this directory.
// For purpose and dependencies, read README.md in this directory.

//go:embed README.md
var _readme string

//go:embed CONTRACTS.md
var _contracts string

// ModuleContract marks this file as the entry point for LLM agents.
var ModuleContract = struct {
	Name    string
	Purpose string
}{
	Name:    "orchestrator",
	Purpose: "Central orchestration service that automates end-to-end task execution across all domain services",
}

// Event types emitted by the orchestrator.
const (
	EventTaskStarted        = "orchestrator.task_started"
	EventWorkUnitStarted    = "orchestrator.work_unit_started"
	EventWorkUnitCompleted  = "orchestrator.work_unit_completed"
	EventWorkUnitFailed     = "orchestrator.work_unit_failed"
	EventTaskCompleted      = "orchestrator.task_completed"
	EventTaskFailed         = "orchestrator.task_failed"
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
