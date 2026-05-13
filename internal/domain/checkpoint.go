package domain

import "time"

// CheckpointTrigger identifies the reason a checkpoint was requested.
type CheckpointTrigger string

const (
	CheckpointTriggerRuntimeCheckpoint CheckpointTrigger = "runtime_checkpoint"
	CheckpointTriggerGoalCompleted     CheckpointTrigger = "goal_completed"
	CheckpointTriggerFocusChange       CheckpointTrigger = "focus_change"
	CheckpointTriggerBeforeValidation  CheckpointTrigger = "before_validation"
	CheckpointTriggerDiffProduced      CheckpointTrigger = "diff_produced"
	CheckpointTriggerBeforeCompletion  CheckpointTrigger = "before_completion"
	CheckpointTriggerToolRequest       CheckpointTrigger = "tool_request"
	CheckpointTriggerToolExecuted      CheckpointTrigger = "tool_executed"
	CheckpointTriggerTimeout           CheckpointTrigger = "timeout"
	CheckpointTriggerHeartbeat         CheckpointTrigger = "heartbeat"
	CheckpointTriggerManualDebug       CheckpointTrigger = "manual_debug"
)

// HeartbeatInput carries payload for a session heartbeat.
type HeartbeatInput struct {
	EventID string
	Payload map[string]interface{}
}

// CheckpointInput carries parameters for a checkpoint.
type CheckpointInput struct {
	EventID        string
	CheckpointID   string
	CurrentGoal    string
	MinimalSummary string
	Ledger         map[string]interface{}
	EvidenceRefs   []string
	OccurredAt     time.Time
	Source         string
	Extra          map[string]interface{}
}

// CheckpointSuggestion is the result of evaluating whether to checkpoint.
type CheckpointSuggestion struct {
	ShouldCheckpoint bool
	Reason           string
	Trigger          CheckpointTrigger
	Input            CheckpointInput
}

// AutoCheckpointInput carries parameters for an automatic checkpoint.
type AutoCheckpointInput struct {
	EventID        string
	Trigger        CheckpointTrigger
	CurrentGoal    string
	MinimalSummary string
	Ledger         map[string]interface{}
	EvidenceRefs   []string
	OccurredAt     time.Time
	SourceEventID  string
	Runtime        string
	FilesRead      []string
	FilesModified  []string
	CompletedGoals []string
	PendingTodos   []string
	Blockers       []string
	Risks          []string
	Decisions      []string
	NextGoal       string
	Force          bool
	Extra          map[string]interface{}
}
