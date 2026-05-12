package bootstrap

import (
	"context"
	"database/sql"

	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/orchestration"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
	reviewmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/review"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
	triggermod "github.com/levygit837-cyber/OrchestraOS/internal/modules/trigger"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// TaskService creates a TaskService with standard dependencies.
func TaskService(db *sql.DB) *taskmod.TaskService {
	return taskmod.NewTaskService(db,
		func(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error {
			return orchestration.CancelTaskDependents(ctx, tx, taskID, input)
		},
	)
}

// RunService creates a RunService with standard repository factories.
func RunService(db *sql.DB) *runmod.RunService {
	return runmod.NewRunService(db,
		func(tx *sql.Tx) runmod.TaskReader { return taskmod.NewRepository(tx) },
		func(tx *sql.Tx) runmod.WorkUnitReader { return workunitmod.NewRepository(tx) },
		func(ctx context.Context, tx *sql.Tx, run *domain.Run, target domain.RunStatus, input transition.TransitionInput) error {
			return orchestration.TransitionRunWithWorkUnit(ctx, tx, run, target, input)
		},
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
	return agentsessionmod.NewAgentSessionService(db,
		func(tx *sql.Tx) agentsessionmod.AgentReader { return agentmod.NewRepository(tx) },
	)
}

// AgentService creates an AgentService with standard dependencies.
func AgentService(db *sql.DB) *agentmod.AgentService {
	return agentmod.NewAgentService(db)
}

// TaskGraphService creates a TaskGraphService with standard repository factories.
func TaskGraphService(db *sql.DB) *taskgraphmod.TaskGraphService {
	return taskgraphmod.NewTaskGraphService(db,
		func(executor dbcore.DBTX) taskgraphmod.TaskReader { return taskmod.NewRepository(executor) },
		func(executor dbcore.DBTX) taskgraphmod.WorkUnitCreator { return workunitmod.NewRepository(executor) },
		func(executor dbcore.DBTX) taskgraphmod.WorkUnitLister { return workunitmod.NewRepository(executor) },
	)
}

// PromptService creates a PromptService with standard dependencies.
func PromptService(db *sql.DB) *promptmod.PromptService {
	return promptmod.NewPromptService(db)
}

// ReviewService creates a ReviewService with standard dependencies.
func ReviewService(db *sql.DB) *reviewmod.ReviewService {
	return reviewmod.NewReviewService(db)
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

// TriggerService creates a TriggerService with standard repository factories.
func TriggerService(db *sql.DB) *triggermod.TriggerService {
	return triggermod.NewTriggerService(db,
		func(executor dbcore.DBTX) triggermod.RunReader { return runmod.NewRepository(executor) },
		func(executor dbcore.DBTX) triggermod.AgentSessionReader { return agentsessionmod.NewRepository(executor) },
		func(executor dbcore.DBTX) triggermod.WorkUnitReader { return workunitmod.NewRepository(executor) },
	)
}

// RuntimeEventRelay creates a RuntimeEventRelay wired to domain services.
func RuntimeEventRelay(db *sql.DB) *orchestration.RuntimeEventRelay {
	return orchestration.NewRuntimeEventRelay(
		db,
		AgentSessionService(db),
		RunService(db),
	)
}
