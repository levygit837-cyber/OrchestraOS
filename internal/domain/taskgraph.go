package domain

import "time"

// ============================================================================
// Taskgraph Domain
// ============================================================================

type TaskGraphStatus string

const (
	TaskGraphStatusActive     TaskGraphStatus = "active"
	TaskGraphStatusSuperseded TaskGraphStatus = "superseded"
)

type TaskGraph struct {
	ID              string          `json:"id"`
	TaskID          string          `json:"task_id"`
	Version         int             `json:"version"`
	Status          TaskGraphStatus `json:"status"`
	PlannerStrategy string          `json:"planner_strategy"`
	Rationale       string          `json:"rationale"`
	CreatedBy       string          `json:"created_by"`
	NodeCount       int             `json:"node_count"`
	EdgeCount       int             `json:"edge_count"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
