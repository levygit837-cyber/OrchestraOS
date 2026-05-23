package daggen

import (
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// BuildGraph constructs a DAGGraph from a DecompositionResult.
// It validates the graph and returns an error if invalid.
func BuildGraph(result *domain.DecompositionResult) (*domain.DAGGraph, error) {
	if result == nil {
		return nil, apperrors.New(apperrors.KindGraphGeneration, "daggen.build", "nil decomposition result")
	}
	if len(result.WorkUnits) == 0 {
		return nil, apperrors.New(apperrors.KindGraphGeneration, "daggen.build", "decomposition produced zero work units")
	}

	g := &domain.DAGGraph{
		ID:        result.Graph.ID,
		TaskID:    result.TaskID,
		Status:    "pending_validation",
		CreatedAt: time.Now(),
	}
	g.Nodes, g.Edges = buildNodesAndEdges(g.ID, result.WorkUnits)

	if err := Validate(g); err != nil {
		g.Status = "rejected"
		return g, err
	}
	g.Status = "validated"
	return g, nil
}

func buildNodesAndEdges(graphID string, specs []domain.WUSpec) ([]domain.DAGNode, []domain.DAGEdge) {
	nodes := make([]domain.DAGNode, len(specs))
	now := time.Now()
	for i, s := range specs {
		nodeID := s.NodeID
		if nodeID == "" {
			nodeID = uuid.New().String()
		}
		nodes[i] = domain.DAGNode{
			ID:        nodeID,
			GraphID:   graphID,
			Label:     s.Title,
			Context:   s.Context,
			DependsOn: normalizeDeps(s.DependsOn),
			Status:    domain.DAGNodeStatusPending,
			CreatedAt: now,
		}
	}
	edges := deriveEdges(specs)
	return nodes, edges
}

func deriveEdges(specs []domain.WUSpec) []domain.DAGEdge {
	var edges []domain.DAGEdge
	for _, s := range specs {
		for _, dep := range s.DependsOn {
			edges = append(edges, domain.DAGEdge{From: dep, To: s.NodeID})
		}
	}
	return edges
}

func normalizeDeps(deps []string) []string {
	if deps == nil {
		return []string{}
	}
	return deps
}

// BuildWorkUnits converts validated DAGGraph nodes into domain WorkUnits.
func BuildWorkUnits(task *domain.Task, graph *domain.DAGGraph, specs []domain.WUSpec) ([]domain.WorkUnit, error) {
	if graph.Status != "validated" {
		return nil, apperrors.New(apperrors.KindGraphValidation, "daggen.build_wu", "cannot build work units from non-validated graph")
	}

	nodeMap := make(map[string]*domain.DAGNode, len(graph.Nodes))
	for i := range graph.Nodes {
		nodeMap[graph.Nodes[i].ID] = &graph.Nodes[i]
	}

	wus := make([]domain.WorkUnit, 0, len(specs))
	for _, s := range specs {
		node := nodeMap[s.NodeID]
		if node == nil {
			return nil, apperrors.New(apperrors.KindWorkUnitInvalid, "daggen.build_wu", "spec references unknown node: "+s.NodeID)
		}
		wus = append(wus, specToWorkUnit(task, graph.ID, s, node))
	}
	return wus, nil
}

func specToWorkUnit(task *domain.Task, graphID string, s domain.WUSpec, node *domain.DAGNode) domain.WorkUnit {
	return domain.WorkUnit{
		ID:                   uuid.New().String(),
		TaskID:               task.ID,
		TaskGraphID:          graphID,
		Title:                s.Title,
		Objective:            s.Objective,
		AssignedAgentProfile: agentOrDefault(s.SuggestedAgent),
		Status:               domain.WorkUnitStatusCreated,
		OwnedPaths:           []string{},
		ReadPaths:            []string{},
		AcceptanceCriteria:   s.AcceptanceCriteria,
		ValidationPlan:       []string{"Validate acceptance criteria."},
		DependsOn:            node.DependsOn,
	}
}

func agentOrDefault(agent string) string {
	if agent == "" {
		return "default"
	}
	return agent
}
