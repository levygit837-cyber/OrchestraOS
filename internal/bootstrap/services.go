package bootstrap

import (
	"database/sql"

	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/event"
	promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// TaskService creates a TaskService with standard dependencies.
func TaskService(db *sql.DB) *taskmod.TaskService {
	return taskmod.NewTaskService(db)
}

// RunService creates a RunService with standard repository factories.
func RunService(db *sql.DB) *runmod.RunService {
	return runmod.NewRunService(db,
		func(tx *sql.Tx) runmod.TaskReader { return taskmod.NewRepository(tx) },
		func(tx *sql.Tx) runmod.WorkUnitReader { return workunitmod.NewRepository(tx) },
	)
}

// WorkUnitService creates a WorkUnitService with standard repository factories.
func WorkUnitService(db *sql.DB) *workunitmod.WorkUnitService {
	return workunitmod.NewWorkUnitService(db,
		func(tx *sql.Tx) workunitmod.TaskReader { return taskmod.NewRepository(tx) },
		func(tx *sql.Tx) workunitmod.TaskGraphManager { return taskgraphmod.NewRepository(tx) },
	)
}

// AgentSessionService creates an AgentSessionService with standard dependencies.
func AgentSessionService(db *sql.DB) *agentsessionmod.AgentSessionService {
	return agentsessionmod.NewAgentSessionService(db)
}

// TaskGraphService creates a TaskGraphService with standard dependencies.
func TaskGraphService(db *sql.DB) *taskgraphmod.TaskGraphService {
	return taskgraphmod.NewTaskGraphService(db)
}

// PromptService creates a PromptService with standard dependencies.
func PromptService(db *sql.DB) *promptmod.PromptService {
	return promptmod.NewPromptService(db)
}

// EventService creates an EventService with standard dependencies.
func EventService(executor dbcore.DBTX) *eventmod.Service {
	return eventmod.NewService(executor)
}

// GeminiPlanner creates a new GeminiPlanner instance.
func GeminiPlanner() (*taskgraphmod.GeminiPlanner, error) {
	return taskgraphmod.NewGeminiPlanner()
}

// PlannerPrompt renders a planner prompt for a task.
func PlannerPrompt(task *domain.Task) (string, error) {
	return taskgraphmod.PlannerPrompt(task)
}

// ValidateGraphPlan validates a graph plan.
func ValidateGraphPlan(plan *taskgraphmod.GraphPlan) error {
	return taskgraphmod.ValidateGraphPlan(plan)
}
