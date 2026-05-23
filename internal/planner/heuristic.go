package planner

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

const (
	maxGraphWorkUnits  = 5
	minGraphWorkUnits  = 2
	balanceNumerator   = 3
	balanceDenominator = 2
)

var afterMarkerPattern = regexp.MustCompile(`^\[after:\s*([0-9,\s]+)\]\s*`)

// Heuristic is a deterministic planner that decomposes a task based on
// its acceptance criteria, grouping them into balanced work units with
// explicit dependency edges.
type Heuristic struct{}

func NewHeuristic() *Heuristic { return &Heuristic{} }

func (h *Heuristic) Plan(_ context.Context, task *domain.Task) (*Plan, error) {
	criteria, err := parseCriteria(task.AcceptanceCriteria)
	if err != nil {
		return nil, err
	}
	if err := validateCriteriaDeps(criteria); err != nil {
		return nil, err
	}
	groups, err := balancedGroups(criteria)
	if err != nil {
		return nil, err
	}

	graphID := uuid.New().String()
	wus := buildWorkUnits(task, graphID, groups)
	resolveGroupDeps(wus, groups)

	if err := validateDAG(wus); err != nil {
		return nil, err
	}

	return &Plan{
		GraphID:   graphID,
		WorkUnits: wus,
		Rationale: "Local heuristic decomposition from task acceptance criteria.",
	}, nil
}

func buildWorkUnits(task *domain.Task, graphID string, groups []criteriaGroup) []domain.WorkUnit {
	wus := make([]domain.WorkUnit, len(groups))
	for i, g := range groups {
		wus[i] = domain.WorkUnit{
			ID:                   uuid.New().String(),
			TaskID:               task.ID,
			TaskGraphID:          graphID,
			Title:                fmt.Sprintf("%s - Part %d", task.Title, i+1),
			Objective:            "Atender criterios de aceite: " + strings.Join(criteriaTexts(g.criteria), "; "),
			AssignedAgentProfile: "default",
			Status:               domain.WorkUnitStatusCreated,
			OwnedPaths:           []string{},
			ReadPaths:            []string{},
			AcceptanceCriteria:   criteriaTexts(g.criteria),
			ValidationPlan:       []string{"Validar criterios de aceite."},
			DependsOn:            []string{},
		}
	}
	return wus
}

func resolveGroupDeps(wus []domain.WorkUnit, groups []criteriaGroup) {
	criterionGroup := map[int]int{}
	for gi, g := range groups {
		for _, c := range g.criteria {
			criterionGroup[c.index] = gi
		}
	}

	for gi, g := range groups {
		depGroups := map[int]bool{}
		for _, c := range g.criteria {
			for _, dep := range c.deps {
				dg := criterionGroup[dep]
				if dg != gi {
					depGroups[dg] = true
				}
			}
		}
		depIdxs := make([]int, 0, len(depGroups))
		for dg := range depGroups {
			depIdxs = append(depIdxs, dg)
		}
		sort.Ints(depIdxs)
		for _, dg := range depIdxs {
			wus[gi].DependsOn = append(wus[gi].DependsOn, wus[dg].ID)
		}
	}
}

type criterion struct {
	index  int
	text   string
	deps   []int
	weight int
}

type criteriaGroup struct {
	criteria []criterion
	weight   int
}

func parseCriteria(raw []string) ([]criterion, error) {
	if len(raw) < minGraphWorkUnits {
		return nil, apperrors.New(apperrors.KindValidation, "planner.parse", "task decomposition requires at least two acceptance criteria")
	}
	result := make([]criterion, 0, len(raw))
	for i, r := range raw {
		text, deps, err := parseAfterMarker(strings.TrimSpace(r), len(raw))
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(text) == "" {
			return nil, apperrors.New(apperrors.KindValidation, "planner.parse", "acceptance criteria must not be empty after dependency marker")
		}
		w := len(strings.Fields(text))
		if w == 0 {
			w = 1
		}
		result = append(result, criterion{index: i, text: text, deps: deps, weight: w})
	}
	return result, nil
}

func parseAfterMarker(raw string, count int) (string, []int, error) {
	matches := afterMarkerPattern.FindStringSubmatch(raw)
	if len(matches) == 0 {
		return raw, nil, nil
	}
	var deps []int
	for _, part := range strings.Split(matches[1], ",") {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		idx, err := strconv.Atoi(v)
		if err != nil || idx < 1 || idx > count {
			return "", nil, apperrors.New(apperrors.KindValidation, "planner.parse", "dependency marker references unknown criterion")
		}
		deps = append(deps, idx-1)
	}
	return strings.TrimSpace(raw[len(matches[0]):]), deps, nil
}

func validateCriteriaDeps(criteria []criterion) error {
	graph := map[int][]int{}
	for _, c := range criteria {
		graph[c.index] = append([]int{}, c.deps...)
		for _, d := range c.deps {
			if d < 0 || d >= len(criteria) {
				return apperrors.New(apperrors.KindValidation, "planner.validate", "dependency references unknown criterion")
			}
		}
	}
	visiting := map[int]bool{}
	visited := map[int]bool{}
	var visit func(int) error
	visit = func(id int) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return apperrors.New(apperrors.KindValidation, "planner.validate", "acceptance criterion dependencies must be acyclic")
		}
		visiting[id] = true
		for _, dep := range graph[id] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for id := range graph {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func balancedGroups(criteria []criterion) ([]criteriaGroup, error) {
	maxG := len(criteria)
	if maxG > maxGraphWorkUnits {
		maxG = maxGraphWorkUnits
	}
	for n := maxG; n >= minGraphWorkUnits; n-- {
		groups := assignGroups(criteria, n)
		if balanced(groups) {
			return groups, nil
		}
	}
	return nil, apperrors.New(apperrors.KindValidation, "planner.balance", "criteria cannot be split into balanced work units")
}

func assignGroups(criteria []criterion, n int) []criteriaGroup {
	ordered := append([]criterion{}, criteria...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].weight == ordered[j].weight {
			return ordered[i].index < ordered[j].index
		}
		return ordered[i].weight > ordered[j].weight
	})
	groups := make([]criteriaGroup, n)
	for _, c := range ordered {
		target := 0
		for i := 1; i < len(groups); i++ {
			if groups[i].weight < groups[target].weight || (groups[i].weight == groups[target].weight && len(groups[i].criteria) < len(groups[target].criteria)) {
				target = i
			}
		}
		groups[target].criteria = append(groups[target].criteria, c)
		groups[target].weight += c.weight
	}
	for i := range groups {
		sort.Slice(groups[i].criteria, func(a, b int) bool {
			return groups[i].criteria[a].index < groups[i].criteria[b].index
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].criteria[0].index < groups[j].criteria[0].index
	})
	return groups
}

func balanced(groups []criteriaGroup) bool {
	var minW, maxW int
	for i, g := range groups {
		if len(g.criteria) == 0 || g.weight == 0 {
			return false
		}
		if i == 0 || g.weight < minW {
			minW = g.weight
		}
		if g.weight > maxW {
			maxW = g.weight
		}
	}
	return maxW*balanceDenominator <= minW*balanceNumerator
}

func criteriaTexts(cs []criterion) []string {
	texts := make([]string, len(cs))
	for i, c := range cs {
		texts[i] = c.text
	}
	return texts
}

func validateDAG(wus []domain.WorkUnit) error {
	known := map[string]bool{}
	graph := map[string][]string{}
	for _, wu := range wus {
		known[wu.ID] = true
		graph[wu.ID] = append([]string{}, wu.DependsOn...)
	}
	for _, wu := range wus {
		for _, dep := range wu.DependsOn {
			if !known[dep] {
				return apperrors.New(apperrors.KindValidation, "planner.dag", "dependency does not exist in graph")
			}
		}
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return apperrors.New(apperrors.KindValidation, "planner.dag", "work unit dependencies must be acyclic")
		}
		visiting[id] = true
		for _, dep := range graph[id] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for id := range graph {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
