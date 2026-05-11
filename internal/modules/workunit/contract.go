package workunit

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. WorkUnit status transitions are atomic and emit exactly one domain event.
//   2. Terminal statuses (completed, failed, cancelled) are immutable.
//   3. ValidateWorkUnitDependencies ensures the dependency graph is ACYCLIC.
//   4. ValidateOwnedPathAvailability prevents path collisions between active work units.
//   5. NEVER call task.Service or run.Service methods.
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
	Name:    "workunit",
	Purpose: "WorkUnit lifecycle, dependency validation, owned-path availability checks",
}
