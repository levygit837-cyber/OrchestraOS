package domain

// ============================================================================
// Task & WorkUnit Context Domain
// ============================================================================

type TaskContext struct {
	TaskID      string   `json:"task_id"`
	Intent      string   `json:"intent"`
	RawInput    string   `json:"raw_input"`
	Domains     []string `json:"domains"`
	Constraints []string `json:"constraints"`
}

type WUContext struct {
	Domain      string   `json:"domain"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
}
