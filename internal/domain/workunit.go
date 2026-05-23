package domain

// ============================================================================
// Workunit Domain
// ============================================================================

type WorkUnitStatus string

const (
	WorkUnitStatusCreated         WorkUnitStatus = "created"
	WorkUnitStatusPlanned         WorkUnitStatus = "planned"
	WorkUnitStatusScheduled       WorkUnitStatus = "scheduled"
	WorkUnitStatusBlocked         WorkUnitStatus = "blocked"
	WorkUnitStatusRunning         WorkUnitStatus = "running"
	WorkUnitStatusWaitingApproval WorkUnitStatus = "waiting_approval"
	WorkUnitStatusPaused          WorkUnitStatus = "paused"
	WorkUnitStatusValidating      WorkUnitStatus = "validating"
	WorkUnitStatusCompleted       WorkUnitStatus = "completed"
	WorkUnitStatusFailed          WorkUnitStatus = "failed"
	WorkUnitStatusCancelled       WorkUnitStatus = "cancelled"
)

type WorkUnit struct {
	ID                   string         `json:"id"`
	TaskID               string         `json:"task_id"`
	TaskGraphID          string         `json:"task_graph_id"`
	Title                string         `json:"title"`
	Objective            string         `json:"objective"`
	AssignedAgentProfile string         `json:"assigned_agent_profile"`
	Status               WorkUnitStatus `json:"status"`
	OwnedPaths           []string       `json:"owned_paths"`
	ReadPaths            []string       `json:"read_paths"`
	AcceptanceCriteria   []string       `json:"acceptance_criteria"`
	ValidationPlan       []string       `json:"validation_plan"`
	DependsOn            []string       `json:"depends_on"`
}
