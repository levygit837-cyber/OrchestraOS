package run

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
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
)

// EventSource abstracts a runtime that produces events.
type EventSource interface {
	ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error)
}

// RelaySessionService abstracts agent-session operations needed by the relay.
type RelaySessionService interface {
	Heartbeat(ctx context.Context, sessionID string, input domain.HeartbeatInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	CheckpointFromEvent(ctx context.Context, sessionID string, event *domain.EventEnvelope) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	Fail(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	Timeout(ctx context.Context, sessionID string, recoverableState json.RawMessage, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
	AutomaticCheckpoint(ctx context.Context, sessionID string, input domain.AutoCheckpointInput) (*transition.OperationResult[*agentsessionmod.AgentSession], *domain.CheckpointSuggestion, error)
}

// RelayRunService abstracts run operations needed by the relay.
type RelayRunService interface {
	Validate(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*Run], error)
	Complete(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*Run], error)
	Fail(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*Run], error)
	Timeout(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*Run], error)
}

// RuntimeEventRelay consumes events from a runtime and routes them to the
// appropriate domain services. It blocks until the runtime completes or fails.
type RuntimeEventRelay struct {
	db             *sql.DB
	sessionService RelaySessionService
	runService     RelayRunService
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
	sessionService RelaySessionService,
	runService RelayRunService,
) *RuntimeEventRelay {
	return &RuntimeEventRelay{
		db:             db,
		sessionService: sessionService,
		runService:     runService,
	}
}

// Run blocks, consuming runtime events until the runtime completes, fails, or
// the context is cancelled. It returns the final run status and any error.
func (r *RuntimeEventRelay) Run(ctx context.Context, runtime EventSource, config RelayConfig) (Status, error) {
	for {
		event, err := runtime.ReceiveEvent(ctx)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) {
				lastErr := r.handleTimeout(ctx, config)
				return StatusFailed, lastErr
			}
			lastErr := fmt.Errorf("runtime event receive error: %w", err)
			_ = r.handleRuntimeError(ctx, config, lastErr)
			return StatusFailed, lastErr
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
			return StatusFailed, processErr
		case "agent.tool_requested":
			processErr = r.handleToolRequested(ctx, config, event)
		default:
			processErr = r.appendEvent(ctx, event)
		}

		if processErr != nil {
			_ = r.handleRuntimeError(ctx, config, processErr)
			return StatusFailed, processErr
		}

		// Auto-checkpoint is evaluated for all event types; only matching
		// triggers (tool_request, tool_executed, completed) produce a checkpoint.
		if err := r.maybeAutoCheckpoint(ctx, config, event); err != nil {
			_ = r.handleRuntimeError(ctx, config, err)
			return StatusFailed, err
		}

		if event.Type == "agent.completed" {
			break
		}
	}

	// Ensure session is stopped and run is validated+completed.
	if _, err := r.sessionService.Stop(ctx, config.SessionID, transition.TransitionInput{
		Runtime: config.RuntimeType,
		AgentID: config.AgentID,
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return StatusFailed, fmt.Errorf("failed to stop session after completion: %w", err)
		}
	}

	if _, err := r.runService.Validate(ctx, config.RunID, transition.TransitionInput{
		Runtime: config.RuntimeType,
		AgentID: config.AgentID,
	}); err != nil {
		var appErr *apperrors.Error
		if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeInvalidTransition {
			return StatusFailed, fmt.Errorf("failed to validate run after completion: %w", err)
		}
	}

	evidenceRef := fmt.Sprintf("%s-runtime:agent.completed", config.RuntimeType)
	justification := fmt.Sprintf("%s runtime completed with agent.completed event", config.RuntimeType)
	if _, err := r.runService.Complete(ctx, config.RunID, transition.TransitionInput{
		Runtime:       config.RuntimeType,
		AgentID:       config.AgentID,
		EvidenceRefs:  []string{evidenceRef},
		Justification: justification,
	}); err != nil {
		lastErr := fmt.Errorf("failed to complete run: %w", err)
		_ = r.handleRuntimeError(ctx, config, lastErr)
		return StatusFailed, lastErr
	}

	return StatusCompleted, nil
}

func (r *RuntimeEventRelay) handleHeartbeat(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	payload, err := decodePayloadMap(event)
	if err != nil {
		return err
	}
	_, err = r.sessionService.Heartbeat(ctx, config.SessionID, domain.HeartbeatInput{
		EventID: event.ID,
		Payload: payload,
	})
	return err
}

func (r *RuntimeEventRelay) handleCheckpoint(ctx context.Context, config RelayConfig, event *domain.EventEnvelope) error {
	_, err := r.sessionService.CheckpointFromEvent(ctx, config.SessionID, event)
	return err
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
	if reason, ok := payload["reason"].(string); ok && reason != "" {
		failureReason = reason
	}

	if _, err := r.sessionService.Fail(ctx, config.SessionID, transition.TransitionInput{
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

	if _, err := r.runService.Fail(ctx, config.RunID, transition.TransitionInput{
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

	if _, err := r.sessionService.Timeout(ctx, config.SessionID, recoverableState, transition.TransitionInput{
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

	if _, err := r.runService.Timeout(ctx, config.RunID, transition.TransitionInput{
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
	input := transition.TransitionInput{
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
	trigger, ok := CheckpointTriggerForRuntimeEvent(event.Type)
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
	_, _, err = r.sessionService.AutomaticCheckpoint(ctx, config.SessionID, domain.AutoCheckpointInput{
		EventID:        uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:auto_checkpoint:"+config.SessionID+":"+event.ID+":"+string(trigger))).String(),
		Trigger:        trigger,
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

// CheckpointTriggerForRuntimeEvent maps a runtime event type to the checkpoint
// trigger that should be used for automatic checkpointing.
func CheckpointTriggerForRuntimeEvent(eventType string) (domain.CheckpointTrigger, bool) {
	switch eventType {
	case "agent.tool_requested":
		return domain.CheckpointTriggerToolRequest, true
	case "agent.tool_executed", "tool.completed":
		return domain.CheckpointTriggerToolExecuted, true
	case "agent.completed":
		return domain.CheckpointTriggerBeforeCompletion, true
	default:
		return "", false
	}
}
