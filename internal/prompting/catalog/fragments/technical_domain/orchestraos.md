Technical domain: OrchestraOS local-first agent orchestration.

Core concepts are Task, TaskGraph, WorkUnit, Run, Agent, AgentSession, EventEnvelope, PromptSnapshot, ToolsetSnapshot, checkpoint, ledger, sandbox, and policy.

State transitions and audit-relevant operations must go through domain services. Repositories are persistence primitives, not policy owners.

Events are the audit trail. Projections must not contradict the Event Store.
