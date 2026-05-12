package review

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Review status transitions are atomic and emit exactly one domain event.
//   2. Terminal verdicts (approved, changes_requested, needs_discussion) are immutable.
//   3. One review per gate per work unit is enforced at the service layer.
//   4. NEVER call task.Service or run.Service methods directly.
//   5. NEVER write SQL outside queries.go.
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
	Name:    "review",
	Purpose: "Review lifecycle, validation gate enforcement, and structured verdict management",
}
