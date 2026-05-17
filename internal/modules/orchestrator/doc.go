// Package orchestrator provides the Task Execution Workflow Engine.
//
// ⚠️ FUTURE RENAME: This module will be renamed to runner/ or taskflow/.
// The name "orchestrator" will be reserved for a future Agent Orchestrator module
// (director/). See docs/adr/0027-orchestrator-module-naming.md.
//
// The OrchestratorService is NOT an "Agent Orchestrator". It does not decide
// which task to run, allocate resources, or prioritize work. It is a deterministic
// workflow engine that executes a single task from start to finish.
//
// What it does:
//   - Decomposes tasks into work units via TaskGraphService
//   - Executes work units sequentially respecting DAG dependencies
//   - Manages agent lifecycle via AgentService and AgentSessionService
//   - Prepares prompts via PromptOrchestrator
//   - Executes runtimes (Fake, Gemini) with event relay
//   - Creates reviews for validation gates
//   - Evaluates triggers for anomaly detection
//
// What it does NOT do:
//   - Decide which task to execute or when (future director/ module)
//   - Perform low-level transaction coordination (belongs to core/coordination/)
//
// This service follows the architectural guidance from ADR 0020 (Orchestrator Service),
// ADR 0021 (Agent Service), and ADR 0023 (Hybrid Intelligent Orchestrator).
//
// The first implementation is sequential (1 work unit at a time) with deterministic
// Go logic. Future iterations may add parallelism.
package orchestrator
