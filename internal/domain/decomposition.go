package domain

import "time"

// ============================================================================
// Decomposition Domain
// ============================================================================

type DecompositionRequest struct {
	TaskID    string      `json:"task_id"`
	RawInput  string      `json:"raw_input"`
	Context   TaskContext `json:"context"`
	CreatedAt time.Time   `json:"created_at"`
}

type DecompositionResult struct {
	RequestID string    `json:"request_id"`
	TaskID    string    `json:"task_id"`
	Graph     DAGGraph  `json:"graph"`
	WorkUnits []WUSpec  `json:"work_units"`
	Rationale string    `json:"rationale"`
	Strategy  string    `json:"strategy"`
	CreatedAt time.Time `json:"created_at"`
}

type WUSpec struct {
	NodeID             string   `json:"node_id"`
	Title              string   `json:"title"`
	Objective          string   `json:"objective"`
	Context            WUContext `json:"context"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	DependsOn          []string `json:"depends_on"`
	SuggestedAgent     string   `json:"suggested_agent"`
}
