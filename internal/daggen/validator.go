package daggen

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Validate checks structural correctness of a DAGGraph:
// - at least one node
// - no duplicate node IDs
// - edges reference existing nodes
// - graph is acyclic
// - every node has non-empty label and context domain
func Validate(g *domain.DAGGraph) error {
	if len(g.Nodes) == 0 {
		return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "graph must have at least one node")
	}
	if err := validateNodeUniqueness(g.Nodes); err != nil {
		return err
	}
	if err := validateEdgeRefs(g); err != nil {
		return err
	}
	if err := validateAcyclic(g); err != nil {
		return err
	}
	return validateNodeFields(g.Nodes)
}

func validateNodeUniqueness(nodes []domain.DAGNode) error {
	seen := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if seen[n.ID] {
			return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "duplicate node ID: "+n.ID)
		}
		seen[n.ID] = true
	}
	return nil
}

func validateEdgeRefs(g *domain.DAGGraph) error {
	known := make(map[string]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		known[n.ID] = true
	}
	for _, e := range g.Edges {
		if !known[e.From] {
			return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "edge references unknown node: "+e.From)
		}
		if !known[e.To] {
			return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "edge references unknown node: "+e.To)
		}
		if e.From == e.To {
			return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "self-referencing edge: "+e.From)
		}
	}
	return nil
}

func validateAcyclic(g *domain.DAGGraph) error {
	adj := buildAdjList(g)
	visiting := make(map[string]bool, len(g.Nodes))
	visited := make(map[string]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		if err := visitNode(n.ID, adj, visiting, visited); err != nil {
			return err
		}
	}
	return nil
}

func visitNode(id string, adj map[string][]string, visiting, visited map[string]bool) error {
	if visited[id] {
		return nil
	}
	if visiting[id] {
		return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "cycle detected involving node: "+id)
	}
	visiting[id] = true
	for _, dep := range adj[id] {
		if err := visitNode(dep, adj, visiting, visited); err != nil {
			return err
		}
	}
	visiting[id] = false
	visited[id] = true
	return nil
}

func buildAdjList(g *domain.DAGGraph) map[string][]string {
	adj := make(map[string][]string, len(g.Nodes))
	for _, n := range g.Nodes {
		adj[n.ID] = append(adj[n.ID], n.DependsOn...)
	}
	for _, e := range g.Edges {
		adj[e.To] = append(adj[e.To], e.From)
	}
	return adj
}

func validateNodeFields(nodes []domain.DAGNode) error {
	for _, n := range nodes {
		if n.Label == "" {
			return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "node missing label: "+n.ID)
		}
		if n.Context.Domain == "" {
			return apperrors.New(apperrors.KindGraphValidation, "daggen.validate", "node missing context domain: "+n.ID)
		}
	}
	return nil
}
