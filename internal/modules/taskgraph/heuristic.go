// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package taskgraph

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

const (
	MaxGraphWorkUnits  = 5
	minGraphWorkUnits  = 2
	balanceNumerator   = 3
	balanceDenominator = 2
)

var afterMarkerPattern = regexp.MustCompile(`^\[after:\s*([0-9,\s]+)\]\s*`)

type criterionPlan struct {
	Index  int
	Text   string
	Deps   []int
	Weight int
}

type groupedCriteria struct {
	Criteria []criterionPlan
	Weight   int
}

func BuildLocalHeuristicGraphPlan(task *task.Task) (*GraphPlan, error) {
	criteria, err := parseCriterionPlans(task.AcceptanceCriteria)
	if err != nil {
		return nil, err
	}
	if err := validateCriterionDependencies(criteria); err != nil {
		return nil, err
	}
	groups, err := balancedCriterionGroups(criteria)
	if err != nil {
		return nil, err
	}

	graphID := uuid.New().String()
	workUnits := make([]workunit.WorkUnit, len(groups))
	criterionGroup := map[int]int{}
	for groupIndex, group := range groups {
		for _, criterion := range group.Criteria {
			criterionGroup[criterion.Index] = groupIndex
		}
	}

	for i, group := range groups {
		workUnits[i] = workunit.WorkUnit{
			ID:                   uuid.New().String(),
			TaskID:               task.ID,
			TaskGraphID:          graphID,
			Title:                fmt.Sprintf("%s - Parte %d", task.Title, i+1),
			Objective:            workUnitObjective(group.Criteria),
			AssignedAgentProfile: "default",
			Status:               workunit.StatusCreated,
			OwnedPaths:           []string{},
			ReadPaths:            []string{},
			AcceptanceCriteria:   criterionTexts(group.Criteria),
			ValidationPlan:       []string{"Validar criterios de aceite da work unit e registrar evidencia."},
			DependsOn:            []string{},
		}
	}

	edges := []domain.TaskGraphEdgeInfo{}
	for groupIndex, group := range groups {
		depGroups := map[int]bool{}
		for _, criterion := range group.Criteria {
			for _, depCriterion := range criterion.Deps {
				depGroup := criterionGroup[depCriterion]
				if depGroup != groupIndex {
					depGroups[depGroup] = true
				}
			}
		}
		depIndexes := make([]int, 0, len(depGroups))
		for depGroup := range depGroups {
			depIndexes = append(depIndexes, depGroup)
		}
		sort.Ints(depIndexes)
		for _, depGroup := range depIndexes {
			depID := workUnits[depGroup].ID
			workUnits[groupIndex].DependsOn = append(workUnits[groupIndex].DependsOn, depID)
			edges = append(edges, domain.TaskGraphEdgeInfo{
				From:   depID,
				To:     workUnits[groupIndex].ID,
				Type:   "blocks",
				Reason: "explicit acceptance criterion dependency",
			})
		}
	}

	inputs := make([]graphWorkUnitInput, 0, len(workUnits))
	for _, wu := range workUnits {
		inputs = append(inputs, graphWorkUnitInput{
			ID:        wu.ID,
			DependsOn: wu.DependsOn,
		})
	}
	if err := validateWorkUnitDependencies(inputs); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, "task_graph_service.cycle_detected", err)
	}

	nodes := make([]domain.TaskGraphNodeInfo, 0, len(workUnits))
	for _, wu := range workUnits {
		nodes = append(nodes, domain.TaskGraphNodeInfo{
			ID:                 wu.ID,
			Title:              wu.Title,
			Objective:          wu.Objective,
			AgentProfile:       wu.AssignedAgentProfile,
			OwnedPaths:         wu.OwnedPaths,
			ReadPaths:          wu.ReadPaths,
			AcceptanceCriteria: wu.AcceptanceCriteria,
			ValidationPlan:     wu.ValidationPlan,
		})
	}
	return &GraphPlan{
		GraphID:   graphID,
		WorkUnits: workUnits,
		Nodes:     nodes,
		Edges:     edges,
		Rationale: "Local heuristic decomposition from task acceptance criteria.",
	}, nil
}

func parseCriterionPlans(rawCriteria []string) ([]criterionPlan, error) {
	if len(rawCriteria) < minGraphWorkUnits {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "task_graph_service.insufficient_input", "task decomposition requires at least two acceptance criteria")
	}
	criteria := make([]criterionPlan, 0, len(rawCriteria))
	for i, raw := range rawCriteria {
		text, deps, err := parseAfterMarker(strings.TrimSpace(raw), len(rawCriteria))
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(text) == "" {
			return nil, apperrors.New(apperrors.CodeInvalidInput, "task_graph_service.insufficient_input", "acceptance criteria must not be empty after dependency marker")
		}
		criteria = append(criteria, criterionPlan{
			Index:  i,
			Text:   text,
			Deps:   deps,
			Weight: criterionWeight(text),
		})
	}
	return criteria, nil
}

func parseAfterMarker(raw string, criteriaCount int) (string, []int, error) {
	matches := afterMarkerPattern.FindStringSubmatch(raw)
	if len(matches) == 0 {
		return raw, nil, nil
	}
	deps := []int{}
	parts := strings.Split(matches[1], ",")
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		index, err := strconv.Atoi(value)
		if err != nil || index < 1 || index > criteriaCount {
			return "", nil, apperrors.New(apperrors.CodeInvalidInput, "task_graph_service.unknown_dependency", "dependency marker references an unknown acceptance criterion")
		}
		deps = append(deps, index-1)
	}
	text := strings.TrimSpace(raw[len(matches[0]):])
	return text, deps, nil
}

func validateCriterionDependencies(criteria []criterionPlan) error {
	graph := map[int][]int{}
	for _, criterion := range criteria {
		graph[criterion.Index] = append([]int{}, criterion.Deps...)
		for _, dep := range criterion.Deps {
			if dep < 0 || dep >= len(criteria) {
				return apperrors.New(apperrors.CodeInvalidInput, "task_graph_service.unknown_dependency", "dependency marker references an unknown acceptance criterion")
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
			return apperrors.New(apperrors.CodeInvalidInput, "task_graph_service.cycle_detected", "acceptance criterion dependencies must be acyclic")
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

func balancedCriterionGroups(criteria []criterionPlan) ([]groupedCriteria, error) {
	maxGroups := len(criteria)
	if maxGroups > MaxGraphWorkUnits {
		maxGroups = MaxGraphWorkUnits
	}
	for groupCount := maxGroups; groupCount >= minGraphWorkUnits; groupCount-- {
		groups := assignCriteriaToGroups(criteria, groupCount)
		if groupsBalanced(groups) {
			return groups, nil
		}
	}
	return nil, apperrors.New(apperrors.CodeInvalidInput, "task_graph_service.unbalanced_work_units", "acceptance criteria cannot be split into similarly sized work units")
}

func assignCriteriaToGroups(criteria []criterionPlan, groupCount int) []groupedCriteria {
	ordered := append([]criterionPlan{}, criteria...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Weight == ordered[j].Weight {
			return ordered[i].Index < ordered[j].Index
		}
		return ordered[i].Weight > ordered[j].Weight
	})
	groups := make([]groupedCriteria, groupCount)
	for _, criterion := range ordered {
		target := 0
		for i := 1; i < len(groups); i++ {
			if groups[i].Weight < groups[target].Weight || (groups[i].Weight == groups[target].Weight && len(groups[i].Criteria) < len(groups[target].Criteria)) {
				target = i
			}
		}
		groups[target].Criteria = append(groups[target].Criteria, criterion)
		groups[target].Weight += criterion.Weight
	}
	for i := range groups {
		sort.Slice(groups[i].Criteria, func(a, b int) bool {
			return groups[i].Criteria[a].Index < groups[i].Criteria[b].Index
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Criteria[0].Index < groups[j].Criteria[0].Index
	})
	return groups
}

func groupsBalanced(groups []groupedCriteria) bool {
	minWeight := 0
	maxWeight := 0
	for i, group := range groups {
		if len(group.Criteria) == 0 || group.Weight == 0 {
			return false
		}
		if i == 0 || group.Weight < minWeight {
			minWeight = group.Weight
		}
		if group.Weight > maxWeight {
			maxWeight = group.Weight
		}
	}
	return maxWeight*balanceDenominator <= minWeight*balanceNumerator
}

func criterionWeight(text string) int {
	weight := len(strings.Fields(text))
	if weight == 0 {
		return 1
	}
	return weight
}

func criterionTexts(criteria []criterionPlan) []string {
	texts := make([]string, 0, len(criteria))
	for _, criterion := range criteria {
		texts = append(texts, criterion.Text)
	}
	return texts
}

func workUnitObjective(criteria []criterionPlan) string {
	return "Atender criterios de aceite: " + strings.Join(criterionTexts(criteria), "; ")
}

type graphWorkUnitInput struct {
	ID        string
	DependsOn []string
}

func validateWorkUnitDependencies(inputs []graphWorkUnitInput) error {
	op := "taskgraph.validate_dependencies"
	known := map[string]bool{}
	graph := map[string][]string{}
	for _, wu := range inputs {
		known[wu.ID] = true
		graph[wu.ID] = append([]string{}, wu.DependsOn...)
	}
	for _, wu := range inputs {
		for _, dep := range wu.DependsOn {
			if !known[dep] {
				return apperrors.New(apperrors.CodeInvalidInput, op, "dependency does not exist in task graph")
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
			return apperrors.New(apperrors.CodeInvalidInput, op, "work unit dependencies must be acyclic")
		}
		visiting[id] = true
		for _, dep := range graph[id] {
			if _, ok := graph[dep]; ok {
				if err := visit(dep); err != nil {
					return err
				}
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

// TaskForGraphTest creates a minimal task for testing graph decomposition.
func TaskForGraphTest(criteria []string) *task.Task {
	return &task.Task{
		ID:                 uuid.New().String(),
		Title:              "Task graph test",
		AcceptanceCriteria: criteria,
	}
}
