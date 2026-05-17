package orchestrator

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/coordination"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Dependencies holds all services required by the OrchestratorService.
type Dependencies struct {
	DB                  *sql.DB
	TaskService         TaskServiceReader
	TaskGraphService    TaskGraphManager
	RunService          RunLifecycleManager
	AgentService        AgentManager
	AgentSessionService SessionManager
	PromptOrchestrator  PromptPreparer
	ReviewService       ReviewManager
	TriggerService      TriggerEvaluator
	WorkUnitLister      WorkUnitLister
	RuntimeEventRelay   func(db *sql.DB) *coordination.RuntimeEventRelay
	NewFakeRuntime      func() Runtime
	NewGeminiRuntime    func() Runtime
}

// Service is the orchestrator that coordinates end-to-end task execution.
type Service struct {
	deps Dependencies
}

// NewService creates a new OrchestratorService.
func NewService(deps Dependencies) *Service {
	return &Service{deps: deps}
}

// RunTask executes a task from start to finish.
func (s *Service) RunTask(ctx context.Context, taskID string, options RunTaskOptions) (*RunTaskResult, error) {
	const op = "orchestrator.run_task"

	// Validate input
	if err := ValidateTaskID(taskID); err != nil {
		return nil, err
	}
	if err := ValidateRunTaskOptions(options); err != nil {
		return nil, err
	}

	// Get task
	task, err := s.deps.TaskService.GetByID(ctx, taskID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, op, "task not found")
	}

	// Emit task started event
	if err := s.emitOrchestratorEvent(ctx, EventTaskStarted, taskID, "", "", map[string]interface{}{
		"runtime_type":     options.RuntimeType,
		"planner_strategy": options.PlannerStrategy,
		"max_steps":        options.MaxSteps,
		"timeout_seconds":  options.TimeoutSeconds,
	}); err != nil {
		return nil, err
	}

	// Decompose task if no active graph exists
	graph, err := s.deps.TaskGraphService.GetActiveByTask(taskID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if graph == nil {
		decomposeResult, err := s.deps.TaskGraphService.Decompose(ctx, DecomposeInput{
			TaskID:          taskID,
			PlannerStrategy: options.PlannerStrategy,
			CreatedBy:       "orchestrator",
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
		}
		graph = decomposeResult.Graph
	}

	// List work units for the graph
	workUnits, err := s.listWorkUnitsByGraph(ctx, graph.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}

	// Sort work units topologically
	sortedWUs, err := s.topologicalSort(workUnits)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op, err)
	}

	// Execute work units sequentially
	result := &RunTaskResult{
		TaskID:    taskID,
		RunIDs:    make([]string, 0, len(sortedWUs)),
		ReviewIDs: make([]string, 0),
		Status:    "completed",
	}

	var completedCount, failedCount int
	for _, wu := range sortedWUs {
		wuResult, err := s.executeWorkUnit(ctx, &wu, task, options)
		if err != nil {
			result.Status = "failed"
			return result, apperrors.Wrap(apperrors.CodeInternal, op, err)
		}

		result.RunIDs = append(result.RunIDs, wuResult.RunID)
		if wuResult.ReviewID != "" {
			result.ReviewIDs = append(result.ReviewIDs, wuResult.ReviewID)
		}

		if wuResult.Success {
			completedCount++
		} else {
			failedCount++
			result.Status = "partial"
		}
	}

	// Update task status based on results
	if failedCount > 0 && completedCount == 0 {
		result.Status = "failed"
		if _, err := s.deps.TaskService.Fail(ctx, taskID, transition.TransitionInput{
			Runtime:       options.RuntimeType,
			FailureReason: fmt.Sprintf("%d work units failed", failedCount),
			Justification: "all work units failed",
		}); err != nil {
			return result, apperrors.Wrap(apperrors.CodePersistence, op, err)
		}
		if err := s.emitOrchestratorEvent(ctx, EventTaskFailed, taskID, "", "", map[string]interface{}{
			"completed": completedCount,
			"failed":    failedCount,
		}); err != nil {
			return result, err
		}
	} else if completedCount == len(sortedWUs) {
		if _, err := s.deps.TaskService.Complete(ctx, taskID, transition.TransitionInput{
			Runtime:       options.RuntimeType,
			Justification: "all work units completed successfully",
		}); err != nil {
			return result, apperrors.Wrap(apperrors.CodePersistence, op, err)
		}
		if err := s.emitOrchestratorEvent(ctx, EventTaskCompleted, taskID, "", "", map[string]interface{}{
			"completed": completedCount,
		}); err != nil {
			return result, err
		}
	}

	return result, nil
}

// executeWorkUnit executes a single work unit.
func (s *Service) executeWorkUnit(ctx context.Context, wu *domain.WorkUnit, task *domain.Task, options RunTaskOptions) (*WorkUnitExecutionResult, error) {
	const op = "orchestrator.execute_work_unit"

	result := &WorkUnitExecutionResult{
		WorkUnitID: wu.ID,
	}

	// Emit work unit started event
	if err := s.emitOrchestratorEvent(ctx, EventWorkUnitStarted, task.ID, wu.ID, "", map[string]interface{}{
		"title":   wu.Title,
		"profile": wu.AssignedAgentProfile,
	}); err != nil {
		return result, err
	}

	// Create run
	runCreateResult, err := s.deps.RunService.Create(ctx, CreateRunInput{
		TaskID:     task.ID,
		WorkUnitID: wu.ID,
	})
	if err != nil {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, "", map[string]interface{}{
			"reason": "failed to create run",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
		result.Error = fmt.Sprintf("failed to create run: %v", err)
		return result, nil
	}
	result.RunID = runCreateResult.Value.ID

	// Start run
	if _, err := s.deps.RunService.Start(ctx, result.RunID, transition.TransitionInput{
		Runtime:       options.RuntimeType,
		Justification: "starting work unit execution",
	}); err != nil {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "failed to start run",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
		result.Error = fmt.Sprintf("failed to start run: %v", err)
		return result, nil
	}

	// Determine profile (default to code_worker if not set)
	profile := wu.AssignedAgentProfile
	if profile == "" {
		profile = "code_worker"
	}

	// Find or create agent
	agent, err := s.deps.AgentService.FindOrCreate(ctx, profile, ConvertRuntimeType(options.RuntimeType))
	if err != nil {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "failed to find or create agent",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
		result.Error = fmt.Sprintf("failed to find or create agent: %v", err)
		return result, nil
	}

	// Create agent session
	sessionCreateResult, err := s.deps.AgentSessionService.Create(ctx, CreateAgentSessionInput{
		AgentID:    agent.ID,
		RunID:      result.RunID,
		TaskID:     task.ID,
		WorkUnitID: wu.ID,
	})
	if err != nil {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "failed to create agent session",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
		result.Error = fmt.Sprintf("failed to create agent session: %v", err)
		return result, nil
	}

	// Connect session
	if _, err := s.deps.AgentSessionService.Connect(ctx, sessionCreateResult.Value.ID, uuid.New().String(), "", transition.TransitionInput{
		Runtime:       options.RuntimeType,
		AgentID:       agent.ID,
		Justification: "connecting agent session for execution",
	}); err != nil {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "failed to connect agent session",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
		result.Error = fmt.Sprintf("failed to connect agent session: %v", err)
		return result, nil
	}

	// Prepare prompt
	preparedPrompt, err := s.deps.PromptOrchestrator.PrepareRunPrompt(ctx, PreparePromptInput{
		RunID:          result.RunID,
		AgentSessionID: sessionCreateResult.Value.ID,
	})
	if err != nil {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "failed to prepare prompt",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
		result.Error = fmt.Sprintf("failed to prepare prompt: %v", err)
		return result, nil
	}

	// Instantiate runtime
	var runtime Runtime
	switch options.RuntimeType {
	case RuntimeTypeFake:
		runtime = s.deps.NewFakeRuntime()
	case RuntimeTypeGemini:
		runtime = s.deps.NewGeminiRuntime()
	default:
		runtime = s.deps.NewFakeRuntime()
	}

	config := RuntimeConfig{
		RunID:             result.RunID,
		WorkUnitID:        wu.ID,
		TaskID:            task.ID,
		AgentID:           agent.ID,
		Prompt:            preparedPrompt.CombinedPrompt,
		SystemPrompt:      preparedPrompt.SystemPrompt,
		TaskPrompt:        preparedPrompt.TaskPrompt,
		PromptSnapshotID:  preparedPrompt.PromptSnapshot.ID,
		ToolsetSnapshotID: preparedPrompt.ToolsetSnapshot.ID,
		PromptHash:        preparedPrompt.PromptHash,
		Toolset:           preparedPrompt.Toolset,
		OwnedPaths:        wu.OwnedPaths,
		ReadPaths:         wu.ReadPaths,
		MaxSteps:          options.MaxSteps,
		Timeout:           options.TimeoutSeconds,
	}

	// Set timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(options.TimeoutSeconds)*time.Second)
	defer cancel()

	// Start runtime in goroutine
	runtimeErr := make(chan error, 1)
	routineDone := make(chan struct{})
	go func() {
		defer close(routineDone)
		if err := runtime.Start(timeoutCtx, config); err != nil {
			runtimeErr <- err
		} else {
			runtimeErr <- nil
		}
	}()

	// Start event relay
	relay := s.deps.RuntimeEventRelay(s.deps.DB)
	finalStatus, relayErr := relay.Run(timeoutCtx, runtime, coordination.RelayConfig{
		SessionID:   sessionCreateResult.Value.ID,
		RunID:       result.RunID,
		RuntimeType: options.RuntimeType,
		AgentID:     agent.ID,
	})

	// Wait for runtime to complete with timeout safety
	select {
	case err := <-runtimeErr:
		if err != nil {
			if relayErr != nil {
				result.Error = fmt.Sprintf("runtime error: %v, relay error: %v", err, relayErr)
			} else {
				result.Error = fmt.Sprintf("runtime error: %v", err)
			}
			if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
				"reason": "runtime execution failed",
				"error":  result.Error,
			}); err != nil {
				return result, err
			}
		}
	case <-timeoutCtx.Done():
		result.Error = "runtime timed out"
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "runtime timeout",
			"error":  result.Error,
		}); err != nil {
			return result, err
		}
	}

	if relayErr != nil {
		if result.Error == "" {
			result.Error = fmt.Sprintf("relay error: %v", relayErr)
		}
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "runtime relay failed",
			"error":  result.Error,
		}); err != nil {
			return result, err
		}
	}

	// Ensure runtime is stopped
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	_ = runtime.Stop(stopCtx)

	// Ensure agent session is stopped (best effort cleanup)
	if _, err := s.deps.AgentSessionService.Stop(ctx, sessionCreateResult.Value.ID, transition.TransitionInput{
		Runtime:       options.RuntimeType,
		Justification: "work unit execution completed",
	}); err != nil {
		// Log but don't fail - best effort cleanup
	}

	// Wait for goroutine to complete to prevent leaks
	<-routineDone

	// TODO[ADR-0022]: usar run.StatusCompleted quando orchestrator consumir *run.Run
	result.Success = finalStatus == domain.RunStatusCompleted

	// Note: Review creation for validation gates is deferred to future iteration
	// as WorkUnit does not currently have a ValidationGate field.

	// Evaluate triggers for anomalies
	if _, err := s.deps.TriggerService.EvaluateRun(ctx, result.RunID); err != nil {
		// Log error but don't fail the work unit
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitFailed, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"reason": "failed to evaluate triggers (non-blocking)",
			"error":  err.Error(),
		}); err != nil {
			return result, err
		}
	}

	// Emit completion event
	if result.Success {
		if err := s.emitOrchestratorEvent(ctx, EventWorkUnitCompleted, task.ID, wu.ID, result.RunID, map[string]interface{}{
			"review_id": result.ReviewID,
		}); err != nil {
			return result, err
		}
	}

	return result, nil
}

// topologicalSort sorts work units respecting their dependencies.
func (s *Service) topologicalSort(workUnits []domain.WorkUnit) ([]domain.WorkUnit, error) {
	// Build adjacency list and in-degree count
	wuMap := make(map[string]*domain.WorkUnit)
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for i := range workUnits {
		wu := &workUnits[i]
		wuMap[wu.ID] = wu
		inDegree[wu.ID] = 0
		adj[wu.ID] = []string{}
	}

	for i := range workUnits {
		wu := &workUnits[i]
		for _, depID := range wu.DependsOn {
			if _, exists := wuMap[depID]; exists {
				adj[depID] = append(adj[depID], wu.ID)
				inDegree[wu.ID]++
			}
		}
	}

	// Kahn's algorithm
	queue := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]domain.WorkUnit, 0)
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		result = append(result, *wuMap[currentID])

		for _, neighborID := range adj[currentID] {
			inDegree[neighborID]--
			if inDegree[neighborID] == 0 {
				queue = append(queue, neighborID)
			}
		}
	}

	if len(result) != len(workUnits) {
		return nil, fmt.Errorf("cycle detected in work unit dependencies")
	}

	return result, nil
}

// listWorkUnitsByGraph lists all work units for a task graph.
func (s *Service) listWorkUnitsByGraph(ctx context.Context, graphID string) ([]domain.WorkUnit, error) {
	return s.deps.WorkUnitLister.ListByTaskGraph(graphID)
}

// emitOrchestratorEvent emits an orchestrator event.
func (s *Service) emitOrchestratorEvent(ctx context.Context, eventType, taskID, workUnitID, runID string, payload map[string]interface{}) error {
	tx, err := dbcore.BeginTx(ctx, s.deps.DB, "orchestrator.emit_event")
	if err != nil {
		return err
	}
	defer dbcore.RollbackTx(tx)

	marshaled, err := serialization.MarshalPayload("orchestrator_event", payload)
	if err != nil {
		return err
	}

	if _, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		Type:        eventType,
		Version:     transition.EventVersionV1,
		TaskID:      taskID,
		WorkUnitID:  workUnitID,
		RunID:       runID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     marshaled,
	}); err != nil {
		return err
	}

	return dbcore.CommitTx(tx, "orchestrator.emit_event")
}
