// Package prompt implements prompt-engineering and snapshot management.
//
// # Responsibility
// Prepares, composes and stores prompts and toolsets for agent runs.
// Handles prompt snapshots (deduplicated by composition hash), toolset
// snapshots, fragment catalog assembly, and variable substitution.
//
// # Key Types
//   - PromptService: domain service for prompt operations
//   - PrepareRunPromptInput: input to prepare a prompt for a run
//   - PreparedRunPrompt: resulting prompt with hashes and snapshots
//   - Composer: assembles fragments into system + task prompts
//   - Toolset: manages available tools for a session
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/orchestration: OperationResult
//   - core/validation: input validation
//   - domain: PromptSnapshot, ToolsetSnapshot
//
// # Related Packages
//   - run/: prompts are prepared for runs
//   - agentsession/: sessions receive prompt snapshots
//   - task/: task data feeds into prompt composition
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Prompt snapshots are deduplicated by composition_hash (UPSERT semantics).
//   - MaxAutonomyLevel = 2 is the highest allowed autonomy level for any prompt.
//   - All RequiredCategories must be present in a composed prompt.
//   - Toolset snapshots are immutable once created.
//   - NEVER call service methods from other modules.
//   - NEVER write SQL outside queries.go.
//
// For full contracts, invariants, and boundary rules:
//   READ: README.md  → purpose, dependencies, file map
//   READ: CONTRACTS.md → invariants, execution rules, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package prompt
