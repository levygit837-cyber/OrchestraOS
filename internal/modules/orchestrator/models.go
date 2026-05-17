package orchestrator

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	reviewmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/review"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
	triggermod "github.com/levygit837-cyber/OrchestraOS/internal/modules/trigger"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

type RunTaskOptions struct {
	RuntimeType     string
	PlannerStrategy string
	MaxSteps        int
	TimeoutSeconds  int
}

type RunTaskResult struct {
	TaskID    string
	RunIDs    []string
	Status    string
	ReviewIDs []string
}

type WorkUnitExecutionResult struct {
	WorkUnitID string
	RunID      string
	ReviewID   string
	Success    bool
	Error      string
}

type DecomposeInput struct {
	TaskID          string
	PlannerStrategy string
	CreatedBy       string
}

type DecomposeResult struct {
	Graph     *taskgraphmod.TaskGraph
	WorkUnits []workunitmod.WorkUnit
}

type CreateRunInput struct {
	TaskID     string
	WorkUnitID string
}

type CreateAgentSessionInput struct {
	AgentID    string
	RunID      string
	TaskID     string
	WorkUnitID string
}

type PreparePromptInput struct {
	RunID          string
	AgentSessionID string
}

type PreparedPrompt struct {
	SystemPrompt    string
	TaskPrompt      string
	CombinedPrompt  string
	PromptHash      string
	Toolset         []string
	PromptSnapshot  *domain.PromptSnapshot
	ToolsetSnapshot *domain.ToolsetSnapshot
}

type Runtime interface {
	Start(ctx context.Context, config RuntimeConfig) error
	Stop(ctx context.Context) error
	SendEvent(ctx context.Context, event *domain.EventEnvelope) error
	ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error)
	Status() RuntimeStatus
}

type RuntimeConfig struct {
	RunID             string
	WorkUnitID        string
	TaskID            string
	AgentID           string
	Prompt            string
	SystemPrompt      string
	TaskPrompt        string
	PromptSnapshotID  string
	ToolsetSnapshotID string
	PromptHash        string
	Toolset           []string
	OwnedPaths        []string
	ReadPaths         []string
	MaxSteps          int
	Timeout           int
}

type RuntimeStatus struct {
	State         string
	CurrentStep   int
	LastHeartbeat int64
}

// TODO[ADR-0022]: migrar para *task.Task
type TaskServiceReader interface {
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	Complete(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*domain.Task], error)
	Fail(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*domain.Task], error)
}

type TaskGraphManager interface {
	GetActiveByTask(taskID string) (*taskgraphmod.TaskGraph, error)
	Decompose(ctx context.Context, input DecomposeInput) (*DecomposeResult, error)
}

// TODO[ADR-0022]: migrar para *run.Run quando run module desacoplar de domain.Run
type RunLifecycleManager interface {
	Create(ctx context.Context, input CreateRunInput) (*transition.OperationResult[*domain.Run], error)
	Start(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*domain.Run], error)
}

type AgentManager interface {
	FindOrCreate(ctx context.Context, profile string, runtimeType domain.AgentRuntimeType) (*domain.Agent, error)
}

// TODO[ADR-0022]: migrar retornos para *agentsession.AgentSession.
type SessionManager interface {
	Create(ctx context.Context, input CreateAgentSessionInput) (*transition.OperationResult[*domain.AgentSession], error)
	Connect(ctx context.Context, sessionID, connectionID, sandboxID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error)
	Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error)
}

type PromptPreparer interface {
	PrepareRunPrompt(ctx context.Context, input PreparePromptInput) (*PreparedPrompt, error)
}

type ReviewManager interface {
	Create(ctx context.Context, runID, workUnitID, taskID, agentSessionID string, gateType reviewmod.ValidationGate) (*transition.OperationResult[*reviewmod.Review], error)
}

type TriggerEvaluator interface {
	EvaluateRun(ctx context.Context, runID string) ([]*triggermod.Trigger, error)
}

// TODO[ADR-0022]: migrar para []workunit.WorkUnit
type WorkUnitLister interface {
	ListByTaskGraph(graphID string) ([]domain.WorkUnit, error)
}
