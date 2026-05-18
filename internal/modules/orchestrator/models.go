package orchestrator

import (
	"context"

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

type PreparedPrompt struct {
	SystemPrompt    string
	TaskPrompt      string
	CombinedPrompt  string
	PromptHash      string
	Toolset         []string
	PromptSnapshot  *promptmod.PromptSnapshot
	ToolsetSnapshot *promptmod.ToolsetSnapshot
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

type TaskServiceReader interface {
	GetByID(ctx context.Context, id string) (*taskmod.Task, error)
	Complete(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*taskmod.Task], error)
	Fail(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*taskmod.Task], error)
}

type TaskGraphManager interface {
	GetActiveByTask(taskID string) (*taskgraphmod.TaskGraph, error)
	Decompose(ctx context.Context, input DecomposeInput) (*DecomposeResult, error)
}

type RunLifecycleManager interface {
	Create(ctx context.Context, input CreateRunInput) (*transition.OperationResult[*runmod.Run], error)
	Start(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*runmod.Run], error)
}

type AgentManager interface {
	FindOrCreate(ctx context.Context, profile string, runtimeType agentmod.RuntimeType) (*agentmod.Agent, error)
}

type SessionManager interface {
	Create(ctx context.Context, input CreateAgentSessionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	Connect(ctx context.Context, sessionID, connectionID, sandboxID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
}

type PromptPersistence interface {
	PersistComposedPrompt(ctx context.Context, composed *promptmod.ComposedPrompt, metadata promptmod.PersistMetadata) (*promptmod.PreparedRunPrompt, error)
}

type ReviewManager interface {
	Create(ctx context.Context, runID, workUnitID, taskID, agentSessionID string, gateType reviewmod.ValidationGate) (*transition.OperationResult[*reviewmod.Review], error)
}

type TriggerEvaluator interface {
	EvaluateRun(ctx context.Context, runID string) ([]*triggermod.Trigger, error)
}

type WorkUnitLister interface {
	ListByTaskGraph(graphID string) ([]workunitmod.WorkUnit, error)
}
