package task

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Task status transitions are atomic and emit exactly one domain event.
//   2. Terminal statuses (completed, failed, cancelled) are immutable.
//   3. Cancellation cascades to dependent WorkUnits and Runs in the SAME transaction.
//   4. NEVER call run.Service or workunit.Service methods — use repositories only.
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
	Name:    "task",
	Purpose: "Task lifecycle from creation through completion with state-machine enforcement",
}
