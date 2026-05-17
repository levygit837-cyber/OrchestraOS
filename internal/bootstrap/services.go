package bootstrap

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/coordination"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	orchestratormod "github.com/levygit837-cyber/OrchestraOS/internal/modules/orchestrator"
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
	return taskmod.NewTaskService(db, coordination.CancelTaskDependents)
}

// planWorkUnitToDomain converts a taskgraph PlanWorkUnit to a workunit.WorkUnit.
func planWorkUnitToDomain(pwu *taskgraphmod.PlanWorkUnit) *workunitmod.WorkUnit {
	if pwu == nil {
		return nil
	}
	return &workunitmod.WorkUnit{
		ID:                   pwu.ID,
		TaskID:               pwu.TaskID,
		TaskGraphID:          pwu.TaskGraphID,
		Title:                pwu.Title,
		Objective:            pwu.Objective,
		AssignedAgentProfile: pwu.AssignedAgentProfile,
		OwnedPaths:           pwu.OwnedPaths,
		ReadPaths:            pwu.ReadPaths,
		AcceptanceCriteria:   pwu.AcceptanceCriteria,
		ValidationPlan:       pwu.ValidationPlan,
		DependsOn:            pwu.DependsOn,
	}
}

// workUnitToPlan converts a workunit.WorkUnit to a taskgraph PlanWorkUnit.
func workUnitToPlan(wu *workunitmod.WorkUnit) taskgraphmod.PlanWorkUnit {
	if wu == nil {
		return taskgraphmod.PlanWorkUnit{}
	}
	return taskgraphmod.PlanWorkUnit{
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
	}
}

// agentReaderAdapter wraps agent.Repository to implement agentsession.AgentReader.
type agentReaderAdapter struct {
	repo *agentmod.Repository
}

func (a *agentReaderAdapter) GetByID(ctx context.Context, id string) (*agentmod.Agent, error) {
	return a.repo.GetByID(ctx, id)
}

// agentSessionReaderAdapter wraps agentsession.Repository to return *agentsession.AgentSession for trigger module.
type agentSessionReaderAdapter struct {
	repo *agentsessionmod.Repository
}

func (a *agentSessionReaderAdapter) GetByID(id string) (*agentsessionmod.AgentSession, error) {
	return a.repo.GetByID(id)
}

// runReaderAdapter wraps run.Repository to return *run.Run for trigger module.
type runReaderAdapter struct {
	repo *runmod.Repository
}

func (a *runReaderAdapter) GetByID(id string) (*runmod.Run, error) {
	return a.repo.GetByID(id)
}

// RunService creates a RunService with standard repository factories.
func RunService(db *sql.DB) *runmod.RunService {
	return runmod.NewRunService(db,
		func(tx *sql.Tx) runmod.TaskReader { return taskmod.NewRepository(tx) },
		func(tx *sql.Tx) runmod.WorkUnitReader {
			return workunitmod.NewRepository(tx)
		},
		coordination.TransitionRunWithWorkUnit,
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
		func(tx *sql.Tx) agentsessionmod.AgentReader {
			return &agentReaderAdapter{repo: agentmod.NewRepository(tx)}
		},
	)
}

// AgentService creates an AgentService with standard dependencies.
func AgentService(db *sql.DB) *agentmod.AgentService {
	return agentmod.NewAgentService(db)
}

// TaskGraphService creates a TaskGraphService with standard repository factories.
func TaskGraphService(db *sql.DB) *taskgraphmod.TaskGraphService {
	return taskgraphmod.NewTaskGraphService(db,
		func(executor dbcore.DBTX) taskgraphmod.TaskReader {
			return taskmod.NewRepository(executor)
		},
		func(executor dbcore.DBTX) taskgraphmod.WorkUnitCreator {
			return &workUnitCreatorAdapter{repo: workunitmod.NewRepository(executor)}
		},
		func(executor dbcore.DBTX) taskgraphmod.WorkUnitLister {
			return &workUnitListerAdapter{repo: workunitmod.NewRepository(executor)}
		},
	)
}

// workUnitCreatorAdapter bridges workunit.Repository to taskgraph.WorkUnitCreator.
// workUnitCreatorAdapter bridges workunit.Repository to taskgraph.WorkUnitCreator.
type workUnitCreatorAdapter struct {
	repo *workunitmod.Repository
}

func (a *workUnitCreatorAdapter) Create(wu *taskgraphmod.PlanWorkUnit) error {
	return a.repo.Create(planWorkUnitToDomain(wu))
}

// workUnitListerAdapter bridges workunit.Repository to taskgraph.WorkUnitLister.
// workUnitListerAdapter bridges workunit.Repository to taskgraph.WorkUnitLister.
type workUnitListerAdapter struct {
	repo *workunitmod.Repository
}

func (a *workUnitListerAdapter) ListByTaskGraph(graphID string) ([]taskgraphmod.PlanWorkUnit, error) {
	wus, err := a.repo.ListByTaskGraph(graphID)
	if err != nil {
		return nil, err
	}
	result := make([]taskgraphmod.PlanWorkUnit, len(wus))
	for i, wu := range wus {
		result[i] = workUnitToPlan(&wu)
	}
	return result, nil
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
func PlannerPrompt(task *taskmod.Task) (string, error) {
	return taskgraphmod.PlannerPrompt(task)
}

// ValidateGraphPlan validates a graph plan.
func ValidateGraphPlan(plan *taskgraphmod.GraphPlan) error {
	return taskgraphmod.ValidateGraphPlan(plan)
}

// TriggerService creates a TriggerService with standard repository factories.
func TriggerService(db *sql.DB) *triggermod.TriggerService {
	return triggermod.NewTriggerService(db,
		func(executor dbcore.DBTX) triggermod.RunReader {
			return &runReaderAdapter{repo: runmod.NewRepository(executor)}
		},
		func(executor dbcore.DBTX) triggermod.AgentSessionReader {
			return &agentSessionReaderAdapter{repo: agentsessionmod.NewRepository(executor)}
		},
		func(executor dbcore.DBTX) triggermod.WorkUnitReader {
			return &workUnitReaderAdapter{repo: workunitmod.NewRepository(executor)}
		},
	)
}

// RuntimeEventRelay creates a RuntimeEventRelay wired to domain services.
func RuntimeEventRelay(db *sql.DB) *coordination.RuntimeEventRelay {
	return coordination.NewRuntimeEventRelay(
		db,
		AgentSessionService(db),
		RunService(db),
	)
}

// OrchestratorService creates an OrchestratorService with all dependencies wired.
// Adapters bridge module-specific types to orchestrator-local interfaces per ADR 0022.
func OrchestratorService(db *sql.DB) *orchestratormod.Service {
	taskGraphSvc := TaskGraphService(db)
	runSvc := RunService(db)
	agentSvc := AgentService(db)
	sessionSvc := AgentSessionService(db)
	promptSvc := PromptService(db)
	reviewSvc := ReviewService(db)
	triggerSvc := TriggerService(db)

	return orchestratormod.NewService(orchestratormod.Dependencies{
		DB:                  db,
		TaskService:         &taskServiceAdapter{db: db, svc: TaskService(db)},
		TaskGraphService:    &taskGraphAdapter{db: db, svc: taskGraphSvc},
		RunService:          &runAdapter{svc: runSvc},
		AgentService:        agentSvc,
		AgentSessionService: &sessionAdapter{svc: sessionSvc},
		PromptOrchestrator:  &promptAdapter{orch: coordination.NewPromptOrchestrator(db, promptSvc)},
		ReviewService:       &reviewAdapter{svc: reviewSvc},
		TriggerService:      triggerSvc,
		WorkUnitLister:      workunitmod.NewRepository(db),
		RuntimeEventRelay:   RuntimeEventRelay,
		NewFakeRuntime:      func() orchestratormod.Runtime { return &runtimeAdapter{r: agentmod.NewFakeRuntime()} },
		NewGeminiRuntime:    func() orchestratormod.Runtime { return &runtimeAdapter{r: agentmod.NewGeminiRuntime()} },
	})
}

// --- Adapters (bridge input types between module services and orchestrator interfaces) ---

// taskServiceAdapter bridges task.TaskService to orchestrator.TaskServiceReader.
// Converts input/output structs; entity types are already native.
type taskServiceAdapter struct {
	db  *sql.DB
	svc *taskmod.TaskService
}

func (a *taskServiceAdapter) GetByID(ctx context.Context, id string) (*taskmod.Task, error) {
	return taskmod.NewRepository(a.db).GetByID(id)
}
func (a *taskServiceAdapter) Complete(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*taskmod.Task], error) {
	return a.svc.Complete(ctx, taskID, input)
}
func (a *taskServiceAdapter) Fail(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*taskmod.Task], error) {
	return a.svc.Fail(ctx, taskID, input)
}

// runAdapter bridges run.RunService to orchestrator.RunLifecycleManager.
// Converts orchestrator.CreateRunInput to run.CreateRunInput; returns *run.Run natively.
type runAdapter struct{ svc *runmod.RunService }

func (a *runAdapter) Create(ctx context.Context, input orchestratormod.CreateRunInput) (*transition.OperationResult[*runmod.Run], error) {
	return a.svc.Create(ctx, runmod.CreateRunInput{TaskID: input.TaskID, WorkUnitID: input.WorkUnitID})
}
func (a *runAdapter) Start(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*runmod.Run], error) {
	return a.svc.Start(ctx, runID, input)
}

type taskGraphAdapter struct {
	db  *sql.DB
	svc *taskgraphmod.TaskGraphService
}

func (a *taskGraphAdapter) GetActiveByTask(taskID string) (*taskgraphmod.TaskGraph, error) {
	return taskgraphmod.NewRepository(a.db).GetActiveByTask(taskID)
}
func (a *taskGraphAdapter) Decompose(ctx context.Context, input orchestratormod.DecomposeInput) (*orchestratormod.DecomposeResult, error) {
	res, err := a.svc.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID: input.TaskID, PlannerStrategy: input.PlannerStrategy, CreatedBy: input.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	workUnits := make([]workunitmod.WorkUnit, len(res.WorkUnits))
	for i, wu := range res.WorkUnits {
		workUnits[i] = *planWorkUnitToDomain(&wu)
	}
	return &orchestratormod.DecomposeResult{Graph: res.Graph, WorkUnits: workUnits}, nil
}

type sessionAdapter struct {
	svc *agentsessionmod.AgentSessionService
}

func (a *sessionAdapter) Create(ctx context.Context, input orchestratormod.CreateAgentSessionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error) {
	return a.svc.Create(ctx, agentsessionmod.CreateAgentSessionInput{
		AgentID: input.AgentID, RunID: input.RunID, TaskID: input.TaskID, WorkUnitID: input.WorkUnitID,
	})
}
func (a *sessionAdapter) Connect(ctx context.Context, sessionID, connectionID, sandboxID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error) {
	return a.svc.Connect(ctx, sessionID, connectionID, sandboxID, input)
}
func (a *sessionAdapter) Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error) {
	return a.svc.Stop(ctx, sessionID, input)
}

// reviewAdapter bridges review.ReviewService to orchestrator.ReviewManager.
type reviewAdapter struct{ svc *reviewmod.ReviewService }

func (a *reviewAdapter) Create(ctx context.Context, runID, workUnitID, taskID, agentSessionID string, gateType reviewmod.ValidationGate) (*transition.OperationResult[*reviewmod.Review], error) {
	return a.svc.Create(ctx, reviewmod.CreateReviewInput{
		RunID: runID, WorkUnitID: workUnitID, TaskID: taskID,
		AgentSessionID: agentSessionID, GateType: gateType,
	})
}

type promptAdapter struct {
	orch *coordination.PromptOrchestrator
}

func (a *promptAdapter) PrepareRunPrompt(ctx context.Context, input orchestratormod.PreparePromptInput) (*orchestratormod.PreparedPrompt, error) {
	res, err := a.orch.PrepareRunPrompt(ctx, promptmod.PrepareRunPromptInput{
		RunID: input.RunID, AgentSessionID: input.AgentSessionID,
	})
	if err != nil {
		return nil, err
	}
	return &orchestratormod.PreparedPrompt{
		SystemPrompt: res.SystemPrompt, TaskPrompt: res.TaskPrompt, CombinedPrompt: res.CombinedPrompt,
		PromptHash: res.PromptHash, Toolset: res.Toolset,
		PromptSnapshot: res.PromptSnapshot, ToolsetSnapshot: res.ToolsetSnapshot,
	}, nil
}

// workUnitReaderAdapter bridges trigger.WorkUnitReader to workunit.Repository.
type workUnitReaderAdapter struct{ repo *workunitmod.Repository }

func (a *workUnitReaderAdapter) GetByID(id string) (*workunitmod.WorkUnit, error) {
	return a.repo.GetByID(id)
}

type runtimeAdapter struct{ r agentmod.Runtime }

func (a *runtimeAdapter) Start(ctx context.Context, config orchestratormod.RuntimeConfig) error {
	return a.r.Start(ctx, agentmod.RuntimeConfig{
		RunID: config.RunID, WorkUnitID: config.WorkUnitID, TaskID: config.TaskID,
		AgentID: config.AgentID, Prompt: config.Prompt, SystemPrompt: config.SystemPrompt,
		TaskPrompt: config.TaskPrompt, PromptSnapshotID: config.PromptSnapshotID,
		ToolsetSnapshotID: config.ToolsetSnapshotID, PromptHash: config.PromptHash,
		Toolset: config.Toolset, OwnedPaths: config.OwnedPaths, ReadPaths: config.ReadPaths,
		MaxSteps: config.MaxSteps, Timeout: config.Timeout,
	})
}
func (a *runtimeAdapter) Stop(ctx context.Context) error { return a.r.Stop(ctx) }
func (a *runtimeAdapter) SendEvent(ctx context.Context, e *domain.EventEnvelope) error {
	return a.r.SendEvent(ctx, e)
}
func (a *runtimeAdapter) ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error) {
	return a.r.ReceiveEvent(ctx)
}
func (a *runtimeAdapter) Status() orchestratormod.RuntimeStatus {
	s := a.r.Status()
	return orchestratormod.RuntimeStatus{State: s.State, CurrentStep: s.CurrentStep, LastHeartbeat: s.LastHeartbeat}
}
