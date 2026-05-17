// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package workunit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func validateCreateWorkUnitInput(input CreateWorkUnitInput) error {
	op := "work_unit_service.validate_create"
	if err := validation.RequiredUUID(input.ID, "work_unit_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.TaskGraphID, "task_graph_id", op); err != nil {
		return err
	}
	if err := validation.RequiredText(input.Title, "title", op); err != nil {
		return err
	}
	if err := validation.RequiredText(input.Objective, "objective", op); err != nil {
		return err
	}
	if err := validation.RequiredText(input.AssignedAgentProfile, "assigned_agent_profile", op); err != nil {
		return err
	}
	if err := validation.StringList(input.OwnedPaths, "owned_paths", op, false); err != nil {
		return err
	}
	if err := validation.StringList(input.ReadPaths, "read_paths", op, false); err != nil {
		return err
	}
	if err := validation.StringList(input.AcceptanceCriteria, "acceptance_criteria", op, false); err != nil {
		return err
	}
	if err := validation.StringList(input.ValidationPlan, "validation_plan", op, false); err != nil {
		return err
	}
	for _, dep := range input.DependsOn {
		if err := validation.RequiredUUID(dep, "depends_on", op); err != nil {
			return err
		}
	}
	return nil
}

func ValidateWorkUnitDependencies(inputs []CreateWorkUnitInput, existing []domain.WorkUnit) error {
	op := "work_unit_service.validate_dependencies"
	known := map[string]bool{}
	for _, wu := range existing {
		known[wu.ID] = true
	}
	graph := map[string][]string{}
	for _, input := range inputs {
		known[input.ID] = true
		graph[input.ID] = append([]string{}, input.DependsOn...)
	}
	for _, input := range inputs {
		for _, dep := range input.DependsOn {
			if !known[dep] {
				return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("dependency %s does not exist in task graph", dep))
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

func ValidateDependenciesCompleted(ctx context.Context, tx *sql.Tx, wu *domain.WorkUnit) error {
	_ = ctx
	if len(wu.DependsOn) == 0 {
		return nil
	}
	repo := NewRepository(tx)
	for _, depID := range wu.DependsOn {
		dep, err := repo.GetByID(depID)
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.get_dependency", err)
		}
		if dep == nil {
			return apperrors.New(apperrors.CodeInvalidInput, "work_unit_service.get_dependency", "dependency not found")
		}
		if dep.TaskID != wu.TaskID || dep.TaskGraphID != wu.TaskGraphID {
			return apperrors.New(apperrors.CodeInvalidInput, "work_unit_service.get_dependency", "dependency belongs to a different task graph")
		}
		if dep.Status != domain.WorkUnitStatusCompleted {
			return apperrors.New(apperrors.CodeInvalidTransition, "work_unit_service.dependencies", "dependencies must be completed before scheduling or starting")
		}
	}
	return nil
}

func ValidateOwnedPathAvailability(ctx context.Context, tx *sql.Tx, wu *domain.WorkUnit) error {
	_ = ctx
	if len(wu.OwnedPaths) == 0 {
		return nil
	}
	all, err := NewRepository(tx).ListByTask(wu.TaskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.path_conflict", err)
	}
	for _, other := range all {
		if other.ID == wu.ID {
			continue
		}
		if other.Status != domain.WorkUnitStatusScheduled && other.Status != domain.WorkUnitStatusRunning && other.Status != domain.WorkUnitStatusValidating {
			continue
		}
		if pathsOverlap(wu.OwnedPaths, other.OwnedPaths) {
			return apperrors.New(apperrors.CodeConflict, "work_unit_service.path_conflict", "owned_paths conflict with another active work unit")
		}
	}
	return nil
}

func pathsOverlap(left, right []string) bool {
	for _, l := range left {
		for _, r := range right {
			if l == r {
				return true
			}
		}
	}
	return false
}
