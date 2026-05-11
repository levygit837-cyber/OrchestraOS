package agent

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Every runtime must implement the Runtime interface completely.
//   2. FakeRuntime responses must be deterministic for the same input.
//   3. GeminiPlanner returns either a fully valid GraphPlan or an error — no partial plans.
//   4. NEVER import internal/modules/* or internal/core/orchestration.
//   5. NEVER mutate tasks, work_units, runs, or agent_sessions tables.
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
	Name:    "agent",
	Purpose: "Define agent runtime interfaces and implementations (fake, gemini, codex-cli, external)",
}
