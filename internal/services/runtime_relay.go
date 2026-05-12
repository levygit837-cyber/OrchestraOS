package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/orchestration"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
)

// RuntimeEventRelay consumes events from a runtime and routes them to the
// appropriate domain services. It blocks until the runtime completes or fails.
type RuntimeEventRelay struct {
	db             *sql.DB
	sessionService *agentsessionmod.AgentSessionService
	runService     *runmod.RunService
}

// RelayConfig holds the identifiers needed to route runtime events.
type RelayConfig struct {
	SessionID   string
	RunID       string
	RuntimeType string
	AgentID     string
	OnEvent     func(*domain.EventEnvelope) // optional progress callback
}

// NewRuntimeEventRelay creates a relay wired to the given domain services.
func NewRuntimeEventRelay(
	db *sql.DB,
	sessionService *agentsessionmod.AgentSessionService,
	runService *runmod.RunService,
) *RuntimeEventRelay {
	return &RuntimeEventRelay{
		db:             db,
		sessionService: sessionService,
		runService:     runService,
	}
}

// Run blocks, consuming runtime events until the runtime completes, fails, or
// the context is cancelled. It returns the final run status and any error.
func (r *RuntimeEventRelay) Run(ctx context.Context, runtime agent.Runtime, config RelayConfig) (domain.RunStatus, error) {
	for {
		event, err := runtime.ReceiveEvent(ctx)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
				lastErr := r.handleTimeout(ctx, config)
				return domain.RunStatusFailed, lastErr
			}
			lastErr := fmt.Errorf("runtime event receive error: %w", err)
			_ = r.handleRuntimeError(ctx, config, lastErr)
			return domain.RunStatusFailed, lastErr
		}

		if config.OnEvent != nil {
			config.OnEvent(event)
		}

		var processErr error
		switch event.Type {
		case "agent.heartbeat":
			processErr = r.handleHeartbeat(ctx, config, event)
		case "agent.checkpoint_reached":
			processErr = r.handleCheckpoint(ctx, config, event)
		case "agent.completed":
			processErr = r.handleCompleted(ctx, config, event)
		case "agent.failed":
			processErr = r.handleFailed(ctx, config, event)
			if processErr != nil {
				_ = r.handleRuntimeError(ctx, config, processErr)
			}
			return domain.RunStatusFailed, processErr
		case "agent.tool_requested":
			processErr = r.handleToolRequested(ctx, config, event)
		default:
			processErr = r.appendEvent(ctx, event)
		}

		if processErr != nil {
			_ = r.handleRuntimeError(ctx, config, processErr)
			return domain.RunStatusFailed, processErr
		}

		// Auto-checkpoint is evaluated for all event types; only matching
		// triggers (tool_request, tool_executed, completed) produce a checkpoint.
		if err := r.maybeAutoCheckpoint(ctx, config, event); err != nil {
			_ = r.handleRuntimeError(ctx, config, err)
			return domain.RunStatusFailed, err
		}

		if event.Type == "agent.completed" {
			break
		}
	}

	// Ensure session is stopped and run is validated+completed.
	if _, err := r.sessionService.Stop(ctx, config.SessionID, orchestration.TransitionInput{
		Runtime: config.RuntimeType,
		AgentID: config.AgentID,
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return domain.RunStatusFailed, fmt.Errorf("failed to stop session after completion: %w", err)
		}
	}

	if _, err := r.runService.Validate(ctx, config.RunID, orchestration.TransitionInput{
		Runtime: config.RuntimeType,
		AgentID: config.AgentID,
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return domain.RunStatusFailed, fmt.Errorf("failed to validate run after completion: %w", err)
		}
	}

	evidenceRef := fmt.Sprintf("%s-runtime:agent.completed", config.RuntimeType)
	justification := fmt.Sprintf("%s runtime completed with agent.completed event", config.RuntimeType)
	if _, err := r.runService.Complete(ctx, config.RunID, orchestration.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		EvidenceRefs:  []string{evidenceRef},
		Justification: justification,
	}); err != nil {
		lastErr := fmt.Errorf("failed to complete run: %w", err)
		_ = r.handleRuntimeError(ctx, config, lastErr)
		return domain.RunStatusFailed, lastErr
	}

	return domain.RunStatusCompleted, nil
}

func (r *RuntimeEventRelay) handleHeartbeat(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	payload, err := decodePayloadMap(event)
	if err != nil {
		return err
	}
	_, err = r.sessionService.Heartbeat(ctx, config.SessionID, agentsessionmod.HeartbeatInput{
		EventID: event.ID,
		Payload: payload,
	})
	return err
}

func (r *RuntimeEventRelay) handleCheckpoint(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	_, err := r.sessionService.CheckpointFromEvent(ctx, config.SessionID, event)
	if err != nil {
		return fmt.Errorf("failed to persist checkpoint: %w", err)
	}
	return nil
}

func (r *RuntimeEventRelay) handleCompleted(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	if err := r.appendEvent(ctx, event); err != nil {
		return err
	}
	return nil
}

func (r *RuntimeEventRelay) handleFailed(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	if err := r.appendEvent(ctx, event); err != nil {
		return err
	}

	payload, err := decodePayloadMap(event)
	if err != nil {
		return err
	}

	failureReason := "runtime failed"
	if r, ok := payload["reason"].(string); ok && r != "" {
		failureReason = r
	}

	if _, err := r.sessionService.Fail(ctx, config.SessionID, orchestration.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		FailureReason: failureReason,
		Justification: "runtime reported failure via agent.failed event",
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return fmt.Errorf("failed to fail session: %w", err)
		}
	}

	if _, err := r.runService.Fail(ctx, config.RunID, orchestration.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		FailureReason: failureReason,
		Justification: "runtime reported failure via agent.failed event",
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return fmt.Errorf("failed to fail run: %w", err)
		}
	}

	return nil
}

func (r *RuntimeEventRelay) handleToolRequested(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	if err := r.appendEvent(ctx, event); err != nil {
		return err
	}
	return nil
}

func (r *RuntimeEventRelay) handleTimeout(ctx context.Context, config RelayConfig) error {
	recoverableState, _ := json.Marshal(map[string]interface{}{
		"runtime":      config.RuntimeType,
		"reason":       "runtime timed out",
		"timed_out_at": time.Now().UTC().Format(time.RFC3339Nano),
	})

	if _, err := r.sessionService.Timeout(ctx, config.SessionID, recoverableState, orchestration.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		FailureReason: "runtime timed out",
		Justification: "context deadline exceeded before runtime completed",
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return fmt.Errorf("session timeout failed: %w", err)
		}
	}

	if _, err := r.runService.Timeout(ctx, config.RunID, orchestration.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		FailureReason: "runtime timed out",
		Justification: "context deadline exceeded before runtime completed",
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return fmt.Errorf("run timeout failed: %w", err)
		}
	}

	return nil
}

func (r *RuntimeEventRelay) handleRuntimeError(ctx context.Context, config RelayConfig, cause error) error {
	input := orchestration.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		FailureReason: cause.Error(),
		Justification: "runtime failed after operational state was started",
	}

	if config.SessionID != "" {
		if _, err := r.sessionService.Fail(ctx, config.SessionID, input); err != nil {
			var appErr *apperrors.Error
			if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
				return fmt.Errorf("session fail compensation failed: %w", err)
			}
		}
	}
	if config.RunID != "" {
		if _, err := r.runService.Fail(ctx, config.RunID, input); err != nil {
			var appErr *apperrors.Error
			if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
				return fmt.Errorf("run fail compensation failed: %w", err)
			}
		}
	}
	return nil
}

func (r *RuntimeEventRelay) appendEvent(ctx context.Context, event *domain.EventEnvelope) error {
	eventService := eventmod.NewService(r.db)
	if _, err := eventService.Append(ctx, event); err != nil {
		return fmt.Errorf("failed to append event %q: %w", event.Type, err)
	}
	return nil
}

func (r *RuntimeEventRelay) maybeAutoCheckpoint(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	trigger, ok := checkpointTriggerForRuntimeEvent(event.Type)
	if !ok {
		return nil
	}
	payload, err := decodePayloadMap(event)
	if err != nil {
		return err
	}
	currentGoal := event.Type
	if value, ok := payload["current_goal"].(string); ok && value != "" {
		currentGoal = value
	}
	summary := fmt.Sprintf("automatic checkpoint for runtime event %s", event.Type)
	if value, ok := payload["summary"].(string); ok && value != "" {
		summary = value
	} else if value, ok := payload["reason"].(string); ok && value != "" {
		summary = value
	}
	ledger := map[string]interface{}{
		"runtime_event_type": event.Type,
		"runtime_event_id":   event.ID,
		"runtime_payload":    payload,
		"pending_todos":      []interface{}{},
	}
	if value, ok := payload["ledger"].(map[string]interface{}); ok {
		ledger = value
	}
	_, _, err = r.sessionService.AutomaticCheckpoint(ctx, config.SessionID, agentsessionmod.AutoCheckpointInput{
		EventID: uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:auto_checkpoint:"+config.SessionID+":"+event.ID+":"+string(trigger))).String(),
		Trigger: trigger,
		CurrentGoal:    currentGoal,
		MinimalSummary: summary,
		Ledger:         ledger,
		EvidenceRefs:   []string{"event:" + event.ID},
		OccurredAt:     event.CreatedAt,
		SourceEventID:  event.ID,
		Runtime:        config.RuntimeType,
		Extra: map[string]interface{}{
			"runtime_event_type": event.Type,
		},
	})
	return err
}

func decodePayloadMap(event *domain.EventEnvelope) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if len(event.Payload) == 0 {
		return map[string]interface{}{}, nil
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode event payload %s: %w", event.ID, err)
	}
	return payload, nil
}

func checkpointTriggerForRuntimeEvent(eventType string) (agentsessionmod.CheckpointTrigger, bool) {
	switch eventType {
	case "agent.tool_requested":
		return agentsessionmod.CheckpointTriggerToolRequest, true
	case "agent.tool_executed", "tool.completed":
		return agentsessionmod.CheckpointTriggerToolExecuted, true
	case "agent.completed":
		return agentsessionmod.CheckpointTriggerBeforeCompletion, true
	default:
		return "", false
	}
}
