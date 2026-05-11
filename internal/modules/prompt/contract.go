package prompt

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Prompt snapshots are deduplicated by composition_hash (UPSERT semantics).
//   2. MaxAutonomyLevel = 2 is the highest allowed autonomy level for any prompt.
//   3. All RequiredCategories must be present in a composed prompt.
//   4. Toolset snapshots are immutable once created.
//   5. NEVER call service methods from other modules.
//   6. NEVER write SQL outside queries.go.
//
// For full contracts and state machine, read CONTRACTS.md in this directory.
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
	Name:    "prompt",
	Purpose: "Prompt composition, snapshot deduplication, and toolset management for agent runs",
}
