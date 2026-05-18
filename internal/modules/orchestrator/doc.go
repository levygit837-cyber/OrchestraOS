// Package orchestrator provides the Task Execution Workflow Engine.
//
// The OrchestratorService is the cross-module coordination layer that executes
// tasks end-to-end. A future director/ module may handle higher-level agent
// orchestration (resource allocation, prioritization), but orchestrator/ remains
// the workflow engine.
//
// The OrchestratorService is NOT an "Agent Orchestrator". It does not decide
// which task to run, allocate resources, or prioritize work. It is a deterministic
// workflow engine that executes a single task from start to finish.
//
// What it does:
//   - Decomposes tasks into work units via TaskGraphService
//   - Executes work units sequentially respecting DAG dependencies
//   - Manages agent lifecycle via AgentService and AgentSessionService
//   - Prepares prompts via PromptService
//   - Executes runtimes (Fake, Gemini) with event relay
//   - Creates reviews for validation gates
//   - Evaluates triggers for anomaly detection
//
// What it does NOT do:
//   - Decide which task to execute or when (future director/ module)
//   - Perform low-level transaction coordination (belongs to owner modules via DI interfaces)
//
// This service follows the architectural guidance from ADR 0020 (Orchestrator Service)
// and ADR 0023 (Hybrid Intelligent Orchestrator).
//
// The first implementation is sequential (1 work unit at a time) with deterministic
// Go logic. Future iterations may add parallelism.
package orchestrator
