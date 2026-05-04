package domain

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
