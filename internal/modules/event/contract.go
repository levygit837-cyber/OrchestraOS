package event

import _ "embed"

// CRITICAL RULES — read these before editing ANY file in this package:
//   1. Event append is idempotent: same ID + same content = no-op, returning existing envelope.
//   2. Event ID collision with different content returns CodeConflict.
//   3. Every envelope passes JSON-Schema validation before storage.
//   4. NEVER import internal/modules/* (event is a leaf dependency).
//   5. NEVER write inline SQL — all persistence goes through core/eventstore.
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
	Name:    "event",
	Purpose: "Idempotent event appending with schema validation and deduplication",
}
