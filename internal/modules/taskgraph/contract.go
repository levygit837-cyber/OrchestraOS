package taskgraph

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. A Task can have at most ONE active TaskGraph at any time.
//   2. Graph plans MUST be validated before persistence (cycle detection, node count).
//   3. NEVER call task.Service or workunit.Service directly — use DI interfaces only.
//   4. NEVER put business logic in repository.go — pure CRUD only.
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
	Name:    "taskgraph",
	Purpose: "Decompose tasks into directed acyclic graphs (DAGs) of WorkUnits",
}
