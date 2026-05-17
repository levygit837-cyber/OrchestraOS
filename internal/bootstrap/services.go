package bootstrap

import (
	"context"
	"database/sql"

	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/coordination"
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
	return taskmod.NewTaskService(db,
		func(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error {
			return coordination.CancelTaskDependents(ctx, tx, taskID, input)
		},
	)
}

// taskToDomain converts a local task.Task to domain.Task for cross-module compatibility.
// TODO[ADR-0022]: remove when orchestrator.TaskServiceReader, run.TaskReader, workunit.TaskReader
// and prompt.PrepareAndPersistInput.Task use *task.Task directly.
func taskToDomain(t *taskmod.Task) *domain.Task {
	if t == nil {
		return nil
	}
	return &domain.Task{
		ID:                   t.ID,
		Title:                t.Title,
		Description:          t.Description,
		Status:               domain.TaskStatus(t.Status),
		Priority:             domain.Priority(t.Priority),
		RiskLevel:            domain.RiskLevel(t.RiskLevel),
		CreatedFromMessageID: t.CreatedFromMessageID,
		AcceptanceCriteria:   t.AcceptanceCriteria,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

// taskReaderAdapter wraps task.Repository to return domain.Task for cross-module compatibility.
// TODO: remove when all TaskReader interfaces use *task.Task directly.
type taskReaderAdapter struct {
	repo *taskmod.Repository
}

func (a *taskReaderAdapter) GetByID(id string) (*domain.Task, error) {
	t, err := a.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return taskToDomain(t), nil
}

// agentToDomain converts a local agent.Agent to domain.Agent for cross-module compatibility.
// TODO: remove when agentsession.AgentReader and orchestrator.AgentManager use *agent.Agent directly.
func agentToDomain(a *agentmod.Agent) *domain.Agent {
	if a == nil {
		return nil
	}
	return &domain.Agent{
		ID:                     a.ID,
		Name:                   a.Name,
		Profile:                a.Profile,
		Capabilities:           a.Capabilities,
		AllowedTools:           a.AllowedTools,
		DefaultPromptFragments: a.DefaultPromptFragments,
		RuntimeType:            domain.AgentRuntimeType(a.RuntimeType),
	}
}

// agentReaderAdapter wraps agent.Repository to return domain.Agent for cross-module compatibility.
// TODO: remove when agentsession.AgentReader uses *agent.Agent directly.
type agentReaderAdapter struct {
	repo *agentmod.Repository
}

func (a *agentReaderAdapter) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	agent, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return agentToDomain(agent), nil
}

// agentManagerAdapter wraps agent.AgentService to implement orchestrator.AgentManager with domain.Agent.
// TODO: remove when orchestrator.AgentManager uses agent.RuntimeType and *agent.Agent directly.
type agentManagerAdapter struct {
	svc *agentmod.AgentService
}

func (a *agentManagerAdapter) FindOrCreate(ctx context.Context, profile string, runtimeType domain.AgentRuntimeType) (*domain.Agent, error) {
	agent, err := a.svc.FindOrCreate(ctx, profile, agentmod.RuntimeType(runtimeType))
	if err != nil {
		return nil, err
	}
	return agentToDomain(agent), nil
}

// RunService creates a RunService with standard repository factories.
func RunService(db *sql.DB) *runmod.RunService {
	return runmod.NewRunService(db,
		func(tx *sql.Tx) runmod.TaskReader { return &taskReaderAdapter{repo: taskmod.NewRepository(tx)} },
		func(tx *sql.Tx) runmod.WorkUnitReader { return workunitmod.NewRepository(tx) },
		func(ctx context.Context, tx *sql.Tx, run *domain.Run, target domain.RunStatus, input transition.TransitionInput) error {
			return coordination.TransitionRunWithWorkUnit(ctx, tx, run, target, input)
		},
	)
}

// WorkUnitService creates a WorkUnitService with standard repository factories.
func WorkUnitService(db *sql.DB) *workunitmod.WorkUnitService {
	return workunitmod.NewWorkUnitService(db,
		func(tx *sql.Tx) workunitmod.TaskReader { return &taskReaderAdapter{repo: taskmod.NewRepository(tx)} },
		func(tx *sql.Tx) workunitmod.TaskGraphManager { return taskgraphmod.NewRepository(tx) },
	)
}

// AgentSessionService creates an AgentSessionService with standard dependencies.
func AgentSessionService(db *sql.DB) *agentsessionmod.AgentSessionService {
	return agentsessionmod.NewAgentSessionService(db,
		func(tx *sql.Tx) agentsessionmod.AgentReader { return &agentReaderAdapter{repo: agentmod.NewRepository(tx)} },
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
			return &taskReaderAdapter{repo: taskmod.NewRepository(executor)}
		},
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
		func(executor dbcore.DBTX) triggermod.AgentSessionReader {
			return agentsessionmod.NewRepository(executor)
		},
		func(executor dbcore.DBTX) triggermod.WorkUnitReader { return workunitmod.NewRepository(executor) },
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
		TaskService:         &taskAdapter{db: db, svc: TaskService(db)},
		TaskGraphService:    &taskGraphAdapter{db: db, svc: taskGraphSvc},
		RunService:          &runAdapter{svc: runSvc},
		AgentService:        &agentManagerAdapter{svc: agentSvc},
		AgentSessionService: &sessionAdapter{svc: sessionSvc},
		PromptOrchestrator:  &promptAdapter{orch: coordination.NewPromptOrchestrator(db, promptSvc)},
		ReviewService:       &reviewAdapter{svc: reviewSvc},
		TriggerService:      triggerSvc,
		WorkUnitLister:      &wuListerAdapter{db: db},
		RuntimeEventRelay:   RuntimeEventRelay,
		NewFakeRuntime:      func() orchestratormod.Runtime { return &runtimeAdapter{r: agentmod.NewFakeRuntime()} },
		NewGeminiRuntime:    func() orchestratormod.Runtime { return &runtimeAdapter{r: agentmod.NewGeminiRuntime()} },
	})
}

// --- Adapters (bridge module types → orchestrator-local interfaces per ADR 0022) ---

// taskAdapter wraps task.TaskService to implement orchestrator.TaskServiceReader with domain.Task.
// TODO[ADR-0022]: remove when orchestrator.TaskServiceReader uses *task.Task directly.
type taskAdapter struct {
	db  *sql.DB
	svc *taskmod.TaskService
}

func (a *taskAdapter) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	_ = ctx
	t, err := taskmod.NewRepository(a.db).GetByID(id)
	if err != nil {
		return nil, err
	}
	return taskToDomain(t), nil
}
func (a *taskAdapter) Complete(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*domain.Task], error) {
	res, err := a.svc.Complete(ctx, taskID, input)
	if err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Task]{
		Value:     taskToDomain(res.Value),
		Event:     res.Event,
		Duplicate: res.Duplicate,
	}, nil
}
func (a *taskAdapter) Fail(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*domain.Task], error) {
	res, err := a.svc.Fail(ctx, taskID, input)
	if err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Task]{
		Value:     taskToDomain(res.Value),
		Event:     res.Event,
		Duplicate: res.Duplicate,
	}, nil
}

type taskGraphAdapter struct {
	db  *sql.DB
	svc *taskgraphmod.TaskGraphService
}

func (a *taskGraphAdapter) GetActiveByTask(taskID string) (*domain.TaskGraph, error) {
	return taskgraphmod.NewRepository(a.db).GetActiveByTask(taskID)
}
func (a *taskGraphAdapter) Decompose(ctx context.Context, input orchestratormod.DecomposeInput) (*orchestratormod.DecomposeResult, error) {
	res, err := a.svc.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID: input.TaskID, PlannerStrategy: input.PlannerStrategy, CreatedBy: input.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return &orchestratormod.DecomposeResult{Graph: res.Graph, WorkUnits: res.WorkUnits}, nil
}

type runAdapter struct{ svc *runmod.RunService }

func (a *runAdapter) Create(ctx context.Context, input orchestratormod.CreateRunInput) (*transition.OperationResult[*domain.Run], error) {
	return a.svc.Create(ctx, runmod.CreateRunInput{TaskID: input.TaskID, WorkUnitID: input.WorkUnitID})
}
func (a *runAdapter) Start(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*domain.Run], error) {
	return a.svc.Start(ctx, runID, input)
}

type sessionAdapter struct {
	svc *agentsessionmod.AgentSessionService
}

func (a *sessionAdapter) Create(ctx context.Context, input orchestratormod.CreateAgentSessionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	return a.svc.Create(ctx, agentsessionmod.CreateAgentSessionInput{
		AgentID: input.AgentID, RunID: input.RunID, TaskID: input.TaskID, WorkUnitID: input.WorkUnitID,
	})
}
func (a *sessionAdapter) Connect(ctx context.Context, sessionID, connectionID, sandboxID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	return a.svc.Connect(ctx, sessionID, connectionID, sandboxID, input)
}
func (a *sessionAdapter) Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	return a.svc.Stop(ctx, sessionID, input)
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

type reviewAdapter struct{ svc *reviewmod.ReviewService }

func (a *reviewAdapter) Create(ctx context.Context, runID, workUnitID, taskID, agentSessionID string, gateType domain.ValidationGate) (*transition.OperationResult[*domain.Review], error) {
	return a.svc.Create(ctx, reviewmod.CreateReviewInput{
		RunID: runID, WorkUnitID: workUnitID, TaskID: taskID,
		AgentSessionID: agentSessionID, GateType: gateType,
	})
}

type wuListerAdapter struct{ db *sql.DB }

func (a *wuListerAdapter) ListByTaskGraph(graphID string) ([]domain.WorkUnit, error) {
	return workunitmod.NewRepository(a.db).ListByTaskGraph(graphID)
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
