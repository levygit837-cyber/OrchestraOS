package taskgraph

import (
	"fmt"
	"strings"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

const (
	minPlannerWorkUnits = 1
	maxPlannerWorkUnits = 10
)

// ValidateGraphPlan performs deterministic validation on a planner-generated graph plan.
func ValidateGraphPlan(plan *GraphPlan) error {
	op := "planner_validator.validate"

	if plan == nil {
		return apperrors.New(apperrors.CodeValidation, op, "plan is nil")
	}

	// Work unit count bounds
	wuCount := len(plan.WorkUnits)
	if wuCount < minPlannerWorkUnits {
		return apperrors.New(apperrors.CodeValidation, op,
			fmt.Sprintf("plan must have at least %d work unit(s), got %d", minPlannerWorkUnits, wuCount))
	}
	if wuCount > maxPlannerWorkUnits {
		return apperrors.New(apperrors.CodeValidation, op,
			fmt.Sprintf("plan must have at most %d work unit(s), got %d", maxPlannerWorkUnits, wuCount))
	}

	// Build lookup map for IDs
	idSet := make(map[string]bool, wuCount)
	for _, wu := range plan.WorkUnits {
		if strings.TrimSpace(wu.ID) == "" {
			return apperrors.New(apperrors.CodeValidation, op, "work unit has empty id")
		}
		idSet[wu.ID] = true
	}

	// Validate each work unit and build dependency graph
	graph := make(map[string][]string, wuCount)
	for i, wu := range plan.WorkUnits {
		prefix := fmt.Sprintf("work_unit[%d]", i)

		if strings.TrimSpace(wu.Title) == "" {
			return apperrors.New(apperrors.CodeValidation, op, prefix+": title is required")
		}
		if strings.TrimSpace(wu.Objective) == "" {
			return apperrors.New(apperrors.CodeValidation, op, prefix+": objective is required")
		}
		if strings.TrimSpace(wu.AssignedAgentProfile) == "" {
			return apperrors.New(apperrors.CodeValidation, op, prefix+": assigned_agent_profile is required")
		}
		if err := ValidatePlannerProfile(wu.AssignedAgentProfile); err != nil {
			return apperrors.Wrap(apperrors.CodeValidation, op+":"+prefix+":profile", err)
		}
		if len(wu.AcceptanceCriteria) == 0 {
			return apperrors.New(apperrors.CodeValidation, op, prefix+": acceptance_criteria must have at least one item")
		}
		for j, ac := range wu.AcceptanceCriteria {
			if strings.TrimSpace(ac) == "" {
				return apperrors.New(apperrors.CodeValidation, op,
					fmt.Sprintf("%s: acceptance_criteria[%d] must not be empty", prefix, j))
			}
		}
		if len(wu.ValidationPlan) == 0 {
			return apperrors.New(apperrors.CodeValidation, op, prefix+": validation_plan must have at least one item")
		}
		for j, vp := range wu.ValidationPlan {
			if strings.TrimSpace(vp) == "" {
				return apperrors.New(apperrors.CodeValidation, op,
					fmt.Sprintf("%s: validation_plan[%d] must not be empty", prefix, j))
			}
		}

		// Validate dependencies point to existing work units
		for j, depID := range wu.DependsOn {
			if !idSet[depID] {
				return apperrors.New(apperrors.CodeValidation, op,
					fmt.Sprintf("%s: depends_on[%d] references unknown work unit %q", prefix, j, depID))
			}
		}

		graph[wu.ID] = append([]string{}, wu.DependsOn...)
	}

	// Cycle detection (DFS)
	visiting := make(map[string]bool, wuCount)
	visited := make(map[string]bool, wuCount)

	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return apperrors.New(apperrors.CodeValidation, op, "work unit dependencies form a cycle")
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
