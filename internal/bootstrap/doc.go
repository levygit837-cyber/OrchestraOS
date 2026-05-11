// Package bootstrap provides pure dependency-injection wiring.
//
// # Responsibility
// Instantiates every domain service with the correct concrete factories
// so that modules never import each other's service logic directly.
// This is the ONLY package allowed to wire cross-module dependencies.
//
// # Key Types
//   - TaskService, RunService, WorkUnitService, AgentSessionService,
//     TaskGraphService, PromptService, EventService: factory functions
//   - GeminiPlanner, PlannerPrompt, ValidateGraphPlan: planner wiring
//
// # Dependencies
//   - core/db: DBTX interface
//   - All modules/: instantiated here
//
// # Related Packages
//   - cmd/orchestraos/cmd: consumes bootstrap to build the CLI
//   - tests/integration: uses bootstrap to set up test services
package bootstrap
