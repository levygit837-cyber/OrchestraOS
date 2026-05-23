package decomposer

const (
	EventDecompositionStarted   = "decomposer.decomposition_started"
	EventDecompositionCompleted = "decomposer.decomposition_completed"
	EventDecompositionFailed    = "decomposer.decomposition_failed"
	EventGraphBuildStarted      = "decomposer.graph_build_started"
	EventGraphValidated         = "decomposer.graph_validated"
	EventGraphRejected          = "decomposer.graph_rejected"
	EventWorkUnitsCreated       = "decomposer.work_units_created"
	EventRetryAttempted         = "decomposer.retry_attempted"
)
