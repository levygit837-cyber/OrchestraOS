package workunit

type Status string

const (
	StatusCreated         Status = "created"
	StatusPlanned         Status = "planned"
	StatusScheduled       Status = "scheduled"
	StatusBlocked         Status = "blocked"
	StatusRunning         Status = "running"
	StatusWaitingApproval Status = "waiting_approval"
	StatusPaused          Status = "paused"
	StatusValidating      Status = "validating"
	StatusCompleted       Status = "completed"
	StatusFailed          Status = "failed"
	StatusCancelled       Status = "cancelled"
)

type WorkUnit struct {
	ID                   string   `json:"id"`
	TaskID               string   `json:"task_id"`
	TaskGraphID          string   `json:"task_graph_id"`
	Title                string   `json:"title"`
	Objective            string   `json:"objective"`
	AssignedAgentProfile string   `json:"assigned_agent_profile"`
	Status               Status   `json:"status"`
	OwnedPaths           []string `json:"owned_paths"`
	ReadPaths            []string `json:"read_paths"`
	AcceptanceCriteria   []string `json:"acceptance_criteria"`
	ValidationPlan       []string `json:"validation_plan"`
	DependsOn            []string `json:"depends_on"`
}
