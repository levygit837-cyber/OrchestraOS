package domain

import "time"

// ============================================================================
// DAG Graph Domain
// ============================================================================

type DAGNodeStatus string

const (
	DAGNodeStatusPending  DAGNodeStatus = "pending"
	DAGNodeStatusValid    DAGNodeStatus = "valid"
	DAGNodeStatusInvalid  DAGNodeStatus = "invalid"
	DAGNodeStatusRetrying DAGNodeStatus = "retrying"
)

type DAGNode struct {
	ID        string        `json:"id"`
	GraphID   string        `json:"graph_id"`
	Label     string        `json:"label"`
	Context   WUContext     `json:"context"`
	DependsOn []string      `json:"depends_on"`
	Status    DAGNodeStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type DAGEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type DAGGraph struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	Nodes     []DAGNode `json:"nodes"`
	Edges     []DAGEdge `json:"edges"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// NodeByID returns the node with the given ID, or nil if not found.
func (g *DAGGraph) NodeByID(id string) *DAGNode {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}
