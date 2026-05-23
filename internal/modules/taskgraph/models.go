package taskgraph

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0030.

type Status = domain.TaskGraphStatus
type TaskGraph = domain.TaskGraph

const (
	StatusActive     = domain.TaskGraphStatusActive
	StatusSuperseded = domain.TaskGraphStatusSuperseded
)

// TaskGraphNodeInfo is a local type describing a node in a task graph.
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

// TaskGraphEdgeInfo is a local type describing an edge in a task graph.
type TaskGraphEdgeInfo struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"`
	Reason string `json:"reason,omitempty"`
}

// TaskGraphCreatedPayload is the payload for a task graph created event.
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

// PlanWorkUnit represents a work unit within a decomposition plan.
// Deprecated: Use domain.WorkUnit directly once all cross-module imports are resolved.
type PlanWorkUnit struct {
	ID                   string   `json:"id"`
	TaskID               string   `json:"task_id"`
	TaskGraphID          string   `json:"task_graph_id"`
	Title                string   `json:"title"`
	Objective            string   `json:"objective"`
	AssignedAgentProfile string   `json:"assigned_agent_profile"`
	OwnedPaths           []string `json:"owned_paths"`
	ReadPaths            []string `json:"read_paths"`
	AcceptanceCriteria   []string `json:"acceptance_criteria"`
	ValidationPlan       []string `json:"validation_plan"`
	DependsOn            []string `json:"depends_on"`
}
