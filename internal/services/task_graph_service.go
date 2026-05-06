package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
)

const (
	localHeuristicPlanner = "local_heuristic_v1"
	maxGraphWorkUnits     = 5
	minGraphWorkUnits     = 2
	balanceNumerator      = 3
	balanceDenominator    = 2
)

var afterMarkerPattern = regexp.MustCompile(`^\[after:\s*([0-9,\s]+)\]\s*`)

type TaskGraphService struct {
	db *sql.DB
}

type DecomposeTaskGraphInput struct {
	TaskID        string
	EventID       string
	ReplaceActive bool
	CreatedBy     string
}

type TaskGraphDecomposeResult struct {
	Graph     *domain.TaskGraph
	WorkUnits []domain.WorkUnit
	Event     *domain.EventEnvelope
	Duplicate bool
}

type graphPlan struct {
	GraphID   string
	WorkUnits []domain.WorkUnit
	Nodes     []domain.TaskGraphNodeInfo
	Edges     []domain.TaskGraphEdgeInfo
	Rationale string
}

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

func NewTaskGraphService(database *sql.DB) *TaskGraphService {
	return &TaskGraphService{db: database}
}

func (s *TaskGraphService) Decompose(ctx context.Context, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, error) {
	op := "task_graph_service.decompose"
	if err := validateRequiredUUID(input.TaskID, "task_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	if input.CreatedBy == "" {
		input.CreatedBy = "task_graph_service"
	}
	if input.EventID != "" {
		if result, duplicate, err := s.duplicateResult(ctx, input); err != nil {
			return nil, err
		} else if duplicate {
			return result, nil
		}
	}

	task, err := repository.NewTaskRepository(s.db).GetByID(input.TaskID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_task", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "task_graph_service.get_task", "task not found")
	}

	plan, err := buildLocalHeuristicGraphPlan(task)
	if err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "task_graph_service.begin_decompose")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	if err := acquireAdvisoryTxLock(ctx, tx, "task_graph:"+task.ID, "task_graph_service.task_graph_lock"); err != nil {
		return nil, err
	}
	if input.EventID != "" {
		if result, duplicate, err := s.duplicateResultInTx(ctx, tx, input); err != nil {
			return nil, err
		} else if duplicate {
			if err := commitTx(tx, "task_graph_service.commit_duplicate"); err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	task, err = getTask(ctx, tx, input.TaskID)
	if err != nil {
		return nil, err
	}
	graphRepo := repository.NewTaskGraphRepository(tx)
	active, err := graphRepo.GetActiveByTask(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_active_graph", err)
	}
	if active != nil && !input.ReplaceActive {
		return nil, apperrors.New(apperrors.CodeConflict, "task_graph_service.active_graph_exists", "active task graph already exists")
	}
	version, err := graphRepo.NextVersion(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.next_version", err)
	}
	now := time.Now().UTC()
	if active != nil {
		if err := graphRepo.SupersedeActiveByTask(task.ID, now); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.supersede_active", err)
		}
	}

	graph := &domain.TaskGraph{
		ID:              plan.GraphID,
		TaskID:          task.ID,
		Version:         version,
		Status:          domain.TaskGraphStatusActive,
		PlannerStrategy: localHeuristicPlanner,
		Rationale:       plan.Rationale,
		CreatedBy:       input.CreatedBy,
		NodeCount:       len(plan.WorkUnits),
		EdgeCount:       len(plan.Edges),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := graphRepo.Create(graph); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.persistence", err)
	}

	payload, err := marshalPayload("task_graph_service.graph_payload", map[string]interface{}{
		"task_id":          graph.TaskID,
		"graph_id":         graph.ID,
		"graph_version":    graph.Version,
		"planner_strategy": graph.PlannerStrategy,
		"rationale":        graph.Rationale,
		"created_by":       graph.CreatedBy,
		"nodes":            plan.Nodes,
		"edges":            plan.Edges,
	})
	if err != nil {
		return nil, err
	}
	graphAppend, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "task.graph_created",
		Version:     eventVersionV1,
		TaskID:      task.ID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if graphAppend.Duplicate {
		return nil, apperrors.New(apperrors.CodeConflict, "task_graph_service.idempotency", "event_id became duplicate during graph creation")
	}

	workUnitRepo := repository.NewWorkUnitRepository(tx)
	for i := range plan.WorkUnits {
		wu := plan.WorkUnits[i]
		if err := workUnitRepo.Create(&wu); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.persistence", err)
		}
		wuPayload, err := marshalPayload("task_graph_service.work_unit_payload", map[string]interface{}{
			"work_unit_id":           wu.ID,
			"task_id":                wu.TaskID,
			"task_graph_id":          wu.TaskGraphID,
			"title":                  wu.Title,
			"objective":              wu.Objective,
			"assigned_agent_profile": wu.AssignedAgentProfile,
			"status":                 wu.Status,
			"owned_paths":            wu.OwnedPaths,
			"read_paths":             wu.ReadPaths,
			"acceptance_criteria":    wu.AcceptanceCriteria,
			"validation_plan":        wu.ValidationPlan,
			"depends_on":             wu.DependsOn,
		})
		if err != nil {
			return nil, err
		}
		if _, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
			Type:        "work_unit.created",
			Version:     eventVersionV1,
			TaskID:      wu.TaskID,
			WorkUnitID:  wu.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     wuPayload,
		}); err != nil {
			return nil, err
		}
		plan.WorkUnits[i] = wu
	}

	if err := commitTx(tx, "task_graph_service.commit_decompose"); err != nil {
		return nil, err
	}
	return &TaskGraphDecomposeResult{
		Graph:     graph,
		WorkUnits: plan.WorkUnits,
		Event:     &graphAppend.Event,
	}, nil
}

func (s *TaskGraphService) ListByTask(ctx context.Context, taskID string) ([]domain.TaskGraph, error) {
	_ = ctx
	if err := validateRequiredUUID(taskID, "task_id", "task_graph_service.list_by_task"); err != nil {
		return nil, err
	}
	return repository.NewTaskGraphRepository(s.db).ListByTask(taskID)
}

func (s *TaskGraphService) duplicateResult(ctx context.Context, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, bool, error) {
	return s.duplicateResultWithExecutor(ctx, s.db, input)
}

func (s *TaskGraphService) duplicateResultInTx(ctx context.Context, tx *sql.Tx, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, bool, error) {
	return s.duplicateResultWithExecutor(ctx, tx, input)
}

func (s *TaskGraphService) duplicateResultWithExecutor(ctx context.Context, executor repository.DBTX, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, bool, error) {
	_ = ctx
	store, err := newEventStore(executor)
	if err != nil {
		return nil, false, err
	}
	existing, err := store.Get(input.EventID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_existing_event", err)
	}
	if existing == nil {
		return nil, false, nil
	}
	if existing.Type != "task.graph_created" || existing.TaskID != input.TaskID {
		return nil, false, apperrors.New(apperrors.CodeConflict, "task_graph_service.idempotency", "event_id already exists for a different operation")
	}
	var payload domain.TaskGraphCreatedPayload
	if err := json.Unmarshal(existing.Payload, &payload); err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodeValidation, "task_graph_service.existing_payload", err)
	}
	graph, err := repository.NewTaskGraphRepository(executor).GetByID(payload.GraphID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_existing_graph", err)
	}
	if graph == nil {
		return nil, false, apperrors.New(apperrors.CodePersistence, "task_graph_service.get_existing_graph", "existing graph event has no graph projection")
	}
	workUnits, err := repository.NewWorkUnitRepository(executor).ListByTaskGraph(graph.ID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.list_existing_work_units", err)
	}
	return &TaskGraphDecomposeResult{
		Graph:     graph,
		WorkUnits: workUnits,
		Event:     existing,
		Duplicate: true,
	}, true, nil
}

func buildLocalHeuristicGraphPlan(task *domain.Task) (*graphPlan, error) {
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
	workUnits := make([]domain.WorkUnit, len(groups))
	criterionGroup := map[int]int{}
	for groupIndex, group := range groups {
		for _, criterion := range group.Criteria {
			criterionGroup[criterion.Index] = groupIndex
		}
	}

	for i, group := range groups {
		workUnits[i] = domain.WorkUnit{
			ID:                   uuid.New().String(),
			TaskID:               task.ID,
			TaskGraphID:          graphID,
			Title:                fmt.Sprintf("%s - Parte %d", task.Title, i+1),
			Objective:            workUnitObjective(group.Criteria),
			AssignedAgentProfile: "default",
			Status:               domain.WorkUnitStatusCreated,
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

	inputs := make([]CreateWorkUnitInput, 0, len(workUnits))
	for _, wu := range workUnits {
		inputs = append(inputs, CreateWorkUnitInput{
			ID:                   wu.ID,
			TaskID:               wu.TaskID,
			TaskGraphID:          wu.TaskGraphID,
			Title:                wu.Title,
			Objective:            wu.Objective,
			AssignedAgentProfile: wu.AssignedAgentProfile,
			OwnedPaths:           wu.OwnedPaths,
			ReadPaths:            wu.ReadPaths,
			AcceptanceCriteria:   wu.AcceptanceCriteria,
			ValidationPlan:       wu.ValidationPlan,
			DependsOn:            wu.DependsOn,
		})
	}
	if err := validateWorkUnitDependencies(inputs, nil); err != nil {
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
	return &graphPlan{
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
	if maxGroups > maxGraphWorkUnits {
		maxGroups = maxGraphWorkUnits
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
