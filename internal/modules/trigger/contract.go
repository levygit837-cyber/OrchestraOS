package trigger

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Detectors are deterministic and side-effect free.
//   2. TriggerService persists triggers and emits events.
//   3. NEVER call run.Service or agentsession.Service methods.
//   4. NEVER write SQL outside queries.go.
//
// For full contracts and invariants, read CONTRACTS.md in this directory.
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
	Name:    "trigger",
	Purpose: "Configurable anomaly detection and threshold triggers for runs, sessions, and work units",
}
