package taskgraph

import "time"

type Status string

const (
	StatusActive     Status = "active"
	StatusSuperseded Status = "superseded"
)

type TaskGraph struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	Version         int       `json:"version"`
	Status          Status    `json:"status"`
	PlannerStrategy string    `json:"planner_strategy"`
	Rationale       string    `json:"rationale"`
	CreatedBy       string    `json:"created_by"`
	NodeCount       int       `json:"node_count"`
	EdgeCount       int       `json:"edge_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
// It mirrors workunit.WorkUnit but lives in the taskgraph package to avoid import cycles.
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
