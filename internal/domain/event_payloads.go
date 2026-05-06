package domain

type TaskGraphCreatedPayload struct {
	TaskID          string              `json:"task_id"`
	GraphID         string              `json:"graph_id"`
	GraphVersion    int                 `json:"graph_version"`
	PlannerStrategy string              `json:"planner_strategy"`
	Rationale       string              `json:"rationale,omitempty"`
	CreatedBy       string              `json:"created_by,omitempty"`
	Nodes           []TaskGraphNodeInfo `json:"nodes"`
	Edges           []TaskGraphEdgeInfo `json:"edges"`
}

type TaskGraphNodeInfo struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Objective          string   `json:"objective"`
	AgentProfile       string   `json:"agent_profile"`
	OwnedPaths         []string `json:"owned_paths"`
	ReadPaths          []string `json:"read_paths"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	ValidationPlan     []string `json:"validation_plan"`
}

type TaskGraphEdgeInfo struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"`
	Reason string `json:"reason,omitempty"`
}

type AgentLedgerUpdatedPayload struct {
	Ledger map[string]interface{} `json:"ledger"`
}

type AgentCheckpointReachedPayload struct {
	CheckpointID   string                 `json:"checkpoint_id"`
	CurrentGoal    string                 `json:"current_goal"`
	Ledger         map[string]interface{} `json:"ledger"`
	MinimalSummary string                 `json:"minimal_summary"`
}

type ArtifactCreatedPayload struct {
	ArtifactID string `json:"artifact_id"`
	Kind       string `json:"kind"`
	URI        string `json:"uri"`
}

type ValidationCompletedPayload struct {
	ValidationID string `json:"validation_id"`
	Status       string `json:"status"`
}

type PromptSnapshotCreatedPayload struct {
	PromptSnapshotID string `json:"prompt_snapshot_id"`
	Hash             string `json:"hash"`
}

type ToolsetSnapshotCreatedPayload struct {
	ToolsetSnapshotID string `json:"toolset_snapshot_id"`
	AgentSessionID    string `json:"agent_session_id"`
}
