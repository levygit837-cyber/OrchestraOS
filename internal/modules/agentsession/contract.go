package agentsession

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Session status transitions are atomic and emit exactly one domain event.
//   2. Terminal statuses (stopped, failed) are immutable.
//   3. Checkpoint persists recoverable_state, ledger, and evidence_refs atomically.
//   4. Timeout must pause the associated Run in the SAME transaction.
//   5. NEVER call run.Service methods — use run.NewRepository(tx) for Run pause.
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
	Name:    "agentsession",
	Purpose: "Agent session lifecycle, heartbeat, checkpoint, and timeout detection",
}
