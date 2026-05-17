// Package review provides the Review domain module.
//
// Responsibilities:
//   - Creating reviews linked to runs, work units, or tasks.
//   - Managing review lifecycle: pending → in_progress → verdict.
//   - Enforcing verdict immutability once submitted.
//   - Preventing duplicate active reviews for the same scope (work_unit, run, task) + gate.
//   - Emitting domain events for every state change.
//
// Entry points:
//   - ReviewService for coordination-level operations.
//   - Repository for persistence (context-aware).
//
// For contracts and boundaries, read CONTRACTS.md.
package review
