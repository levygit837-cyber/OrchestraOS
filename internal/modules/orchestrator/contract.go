package orchestrator

// Event types emitted by the orchestrator.
const (
	EventTaskStarted        = "orchestrator.task_started"
	EventWorkUnitStarted    = "orchestrator.work_unit_started"
	EventWorkUnitCompleted  = "orchestrator.work_unit_completed"
	EventWorkUnitFailed     = "orchestrator.work_unit_failed"
	EventTaskCompleted      = "orchestrator.task_completed"
	EventTaskFailed         = "orchestrator.task_failed"
)

// Valid runtime types.
const (
	RuntimeTypeFake     = "fake"
	RuntimeTypeGemini   = "gemini"
	RuntimeTypeCodexCLI = "codex_cli"
)

// Valid planner strategies.
const (
	PlannerStrategyLocalHeuristic = "local_heuristic_v1"
	PlannerStrategyLLMGemini      = "llm_gemini_v1"
)

// Default configuration values.
const (
	DefaultMaxSteps       = 10
	DefaultTimeoutSeconds = 300
)
