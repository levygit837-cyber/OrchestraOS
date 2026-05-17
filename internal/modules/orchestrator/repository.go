// Package orchestrator is a coordination module, not a domain module.
// It does not own a database table; all persistence is delegated to
// domain services injected via Dependencies.
//
// This file is a placeholder to satisfy ADR-0025 module structure.
// If the orchestrator gains its own tables in the future (e.g. execution_plans),
// CRUD logic belongs here.
package orchestrator
