// Package orchestrator provides the OrchestratorService, which coordinates
// the end-to-end execution of tasks across all domain services.
//
// The OrchestratorService is the control plane that:
// - Decomposes tasks into work units via TaskGraphService
// - Executes work units sequentially respecting DAG dependencies
// - Manages agent lifecycle via AgentService and AgentSessionService
// - Prepares prompts via PromptOrchestrator
// - Executes runtimes (Fake, Gemini) with event relay
// - Creates reviews for validation gates
// - Evaluates triggers for anomaly detection
//
// This service follows the architectural guidance from ADR 0020 (Orchestrator Service),
// ADR 0021 (Agent Service), and ADR 0023 (Hybrid Intelligent Orchestrator).
//
// The first implementation is sequential (1 work unit at a time) with deterministic
// Go logic. Future iterations may add parallelism and LLM-based strategic decisions.
package orchestrator
