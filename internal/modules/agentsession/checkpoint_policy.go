package agentsession

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type CheckpointTrigger string

const (
	CheckpointTriggerRuntimeCheckpoint CheckpointTrigger = "runtime_checkpoint"
	CheckpointTriggerGoalCompleted     CheckpointTrigger = "goal_completed"
	CheckpointTriggerFocusChange       CheckpointTrigger = "focus_change"
	CheckpointTriggerBeforeValidation  CheckpointTrigger = "before_validation"
	CheckpointTriggerDiffProduced      CheckpointTrigger = "diff_produced"
	CheckpointTriggerBeforeCompletion  CheckpointTrigger = "before_completion"
	CheckpointTriggerToolRequest       CheckpointTrigger = "tool_request"
	CheckpointTriggerToolExecuted      CheckpointTrigger = "tool_executed"
	CheckpointTriggerTimeout           CheckpointTrigger = "timeout"
	CheckpointTriggerHeartbeat         CheckpointTrigger = "heartbeat"
	CheckpointTriggerManualDebug       CheckpointTrigger = "manual_debug"
)

type AutoCheckpointInput struct {
	EventID        string
	Trigger        CheckpointTrigger
	CurrentGoal    string
	MinimalSummary string
	Ledger         map[string]interface{}
	EvidenceRefs   []string
	OccurredAt     time.Time
	SourceEventID  string
	Runtime        string
	FilesRead      []string
	FilesModified  []string
	CompletedGoals []string
	PendingTodos   []string
	Blockers       []string
	Risks          []string
	Decisions      []string
	NextGoal       string
	Force          bool
	Extra          map[string]interface{}
}

type CheckpointRecord struct {
	Event          domain.EventEnvelope
	AgentSessionID string
	CheckpointID   string
	CurrentGoal    string
	MinimalSummary string
	Ledger         map[string]interface{}
	EvidenceRefs   []string
	OccurredAt     time.Time
	Source         string
}

type RecoverableCheckpointState struct {
	Session          AgentSession
	LastCheckpoint   *CheckpointRecord
	RecoverableState json.RawMessage
}

func (s *AgentSessionService) SuggestCheckpoint(ctx context.Context, sessionID string, input domain.AutoCheckpointInput) (*domain.CheckpointSuggestion, error) {
	op := "agent_session_service.suggest_checkpoint"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validateCheckpointTrigger(input.Trigger, op); err != nil {
		return nil, err
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_suggest_checkpoint")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)
	if _, err := RequireByID(ctx, tx, sessionID); err != nil {
		return nil, err
	}

	should, reason := shouldCheckpoint(input)
	checkpointInput, err := buildCheckpointInput(sessionID, input, should)
	if err != nil {
		return nil, err
	}
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_suggest_checkpoint"); err != nil {
		return nil, err
	}
	return &domain.CheckpointSuggestion{
		ShouldCheckpoint: should,
		Reason:           reason,
		Trigger:          input.Trigger,
		Input:            checkpointInput,
	}, nil
}

func (s *AgentSessionService) AutomaticCheckpoint(ctx context.Context, sessionID string, input domain.AutoCheckpointInput) (*transition.OperationResult[*AgentSession], *domain.CheckpointSuggestion, error) {
	suggestion, err := s.SuggestCheckpoint(ctx, sessionID, input)
	if err != nil {
		return nil, nil, err
	}
	if !suggestion.ShouldCheckpoint {
		return nil, suggestion, nil
	}
	result, err := s.Checkpoint(ctx, sessionID, suggestion.Input)
	if err != nil {
		return nil, suggestion, err
	}
	return result, suggestion, nil
}

func (s *AgentSessionService) CheckpointFromEvent(ctx context.Context, sessionID string, event *domain.EventEnvelope) (*transition.OperationResult[*AgentSession], error) {
	op := "agent_session_service.checkpoint_from_event"
	if event == nil {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "event envelope is required")
	}
	if event.Type != "agent.checkpoint_reached" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "event type must be agent.checkpoint_reached")
	}
	input, err := checkpointInputFromEvent(event)
	if err != nil {
		return nil, err
	}
	return s.Checkpoint(ctx, sessionID, input)
}

func (s *AgentSessionService) ListCheckpoints(ctx context.Context, sessionID string) ([]CheckpointRecord, error) {
	op := "agent_session_service.list_checkpoints"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_list_checkpoints")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := RequireByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	events, err := eventmod.NewService(tx).ListByRun(ctx, session.RunID)
	if err != nil {
		return nil, err
	}
	records := make([]CheckpointRecord, 0)
	for _, event := range events {
		if event.Type != "agent.checkpoint_reached" {
			continue
		}
		record, err := checkpointRecordFromEvent(event)
		if err != nil {
			return nil, err
		}
		if record.AgentSessionID != "" {
			if record.AgentSessionID != session.ID {
				continue
			}
		} else if event.AgentID != "" && event.AgentID != session.AgentID {
			continue
		}
		records = append(records, record)
	}
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_list_checkpoints"); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *AgentSessionService) RecoverableCheckpoint(ctx context.Context, sessionID string) (*RecoverableCheckpointState, error) {
	op := "agent_session_service.recoverable_checkpoint"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_recoverable_checkpoint")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := RequireByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	events, err := eventmod.NewService(tx).ListByRun(ctx, session.RunID)
	if err != nil {
		return nil, err
	}
	var last *CheckpointRecord
	for _, event := range events {
		if event.Type != "agent.checkpoint_reached" {
			continue
		}
		record, err := checkpointRecordFromEvent(event)
		if err != nil {
			return nil, err
		}
		if record.AgentSessionID != "" {
			if record.AgentSessionID != session.ID {
				continue
			}
		} else if event.AgentID != "" && event.AgentID != session.AgentID {
			continue
		}
		copy := record
		last = &copy
	}
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_recoverable_checkpoint"); err != nil {
		return nil, err
	}
	return &RecoverableCheckpointState{
		Session:          *session,
		LastCheckpoint:   last,
		RecoverableState: session.RecoverableState,
	}, nil
}

func validateCheckpointTrigger(trigger domain.CheckpointTrigger, op string) error {
	switch trigger {
	case domain.CheckpointTriggerRuntimeCheckpoint,
		domain.CheckpointTriggerGoalCompleted,
		domain.CheckpointTriggerFocusChange,
		domain.CheckpointTriggerBeforeValidation,
		domain.CheckpointTriggerDiffProduced,
		domain.CheckpointTriggerBeforeCompletion,
		domain.CheckpointTriggerToolRequest,
		domain.CheckpointTriggerToolExecuted,
		domain.CheckpointTriggerTimeout,
		domain.CheckpointTriggerHeartbeat,
		domain.CheckpointTriggerManualDebug:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid checkpoint trigger %q", trigger))
	}
}

func shouldCheckpoint(input domain.AutoCheckpointInput) (bool, string) {
	if input.Force {
		return true, "forced checkpoint"
	}
	switch input.Trigger {
	case domain.CheckpointTriggerRuntimeCheckpoint:
		return true, "runtime emitted checkpoint event"
	case domain.CheckpointTriggerGoalCompleted:
		return true, "goal completed"
	case domain.CheckpointTriggerFocusChange:
		return true, "agent is changing focus"
	case domain.CheckpointTriggerBeforeValidation:
		return true, "validation is about to start"
	case domain.CheckpointTriggerDiffProduced:
		return true, "relevant diff was produced"
	case domain.CheckpointTriggerBeforeCompletion:
		return true, "work unit is about to complete"
	case domain.CheckpointTriggerToolRequest:
		return true, "tool request is a safe recovery boundary"
	case domain.CheckpointTriggerToolExecuted:
		return true, "tool execution is a safe recovery boundary"
	case domain.CheckpointTriggerTimeout:
		return true, "timeout requires recoverable state"
	case domain.CheckpointTriggerManualDebug:
		return true, "manual debug checkpoint requested"
	default:
		return false, "trigger does not require checkpoint"
	}
}

func buildCheckpointInput(sessionID string, input domain.AutoCheckpointInput, required bool) (domain.CheckpointInput, error) {
	ledger := copyMap(input.Ledger)
	if ledger == nil {
		ledger = map[string]interface{}{}
	}
	putIfNotEmpty(ledger, "current_goal", input.CurrentGoal)
	putIfNotEmpty(ledger, "next_goal", input.NextGoal)
	putSliceIfNotNil(ledger, "completed_goals", input.CompletedGoals)
	putSliceIfNotNil(ledger, "pending_todos", input.PendingTodos)
	putSliceIfNotNil(ledger, "blockers", input.Blockers)
	putSliceIfNotNil(ledger, "risks", input.Risks)
	putSliceIfNotNil(ledger, "decisions", input.Decisions)
	putIfNotEmpty(ledger, "runtime", input.Runtime)
	putIfNotEmpty(ledger, "source_event_id", input.SourceEventID)
	ledger["checkpoint_trigger"] = input.Trigger

	currentGoal := input.CurrentGoal
	if currentGoal == "" {
		if value, ok := ledger["current_goal"].(string); ok {
			currentGoal = value
		}
	}
	if required && currentGoal == "" {
		return domain.CheckpointInput{}, apperrors.New(apperrors.CodeInvalidInput, "agent_session_service.build_checkpoint", "current_goal is required for automatic checkpoint")
	}
	minimalSummary := input.MinimalSummary
	if minimalSummary == "" {
		minimalSummary = fmt.Sprintf("automatic checkpoint for %s", input.Trigger)
	}
	if required && len(ledger) == 0 {
		return domain.CheckpointInput{}, apperrors.New(apperrors.CodeInvalidInput, "agent_session_service.build_checkpoint", "ledger is required for automatic checkpoint")
	}

	extra := copyMap(input.Extra)
	if extra == nil {
		extra = map[string]interface{}{}
	}
	extra["automatic"] = true
	extra["checkpoint_trigger"] = input.Trigger
	putIfNotEmpty(extra, "source_event_id", input.SourceEventID)
	putIfNotEmpty(extra, "runtime", input.Runtime)
	putIfNotEmpty(extra, "next_goal", input.NextGoal)
	putSliceIfNotNil(extra, "files_read", input.FilesRead)
	putSliceIfNotNil(extra, "files_modified", input.FilesModified)

	eventID := input.EventID
	if eventID == "" && input.SourceEventID != "" {
		eventID = uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:auto_checkpoint:"+sessionID+":"+input.SourceEventID+":"+string(input.Trigger))).String()
	}
	if eventID == "" {
		eventID = uuid.New().String()
	}

	checkpointID := uuid.New().String()
	if input.SourceEventID != "" {
		checkpointID = uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:checkpoint:"+sessionID+":"+input.SourceEventID+":"+string(input.Trigger))).String()
	}

	return domain.CheckpointInput{
		EventID:        eventID,
		CheckpointID:   checkpointID,
		CurrentGoal:    currentGoal,
		MinimalSummary: minimalSummary,
		Ledger:         ledger,
		EvidenceRefs:   input.EvidenceRefs,
		OccurredAt:     input.OccurredAt,
		Source:         "automatic",
		Extra:          extra,
	}, nil
}

func checkpointInputFromEvent(event *domain.EventEnvelope) (domain.CheckpointInput, error) {
	record, err := checkpointRecordFromEvent(*event)
	if err != nil {
		return domain.CheckpointInput{}, err
	}
	extra := map[string]interface{}{
		"source_event_id": event.ID,
	}
	var payload map[string]interface{}
	if len(event.Payload) > 0 {
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return domain.CheckpointInput{}, apperrors.Wrap(apperrors.CodeValidation, "agent_session_service.checkpoint_event_payload", err)
		}
		for key, value := range payload {
			switch key {
			case "agent_session_id", "checkpoint_id", "current_goal", "minimal_summary", "ledger", "evidence_refs", "occurred_at", "source":
				continue
			default:
				extra[key] = value
			}
		}
	}
	return domain.CheckpointInput{
		EventID:        event.ID,
		CheckpointID:   record.CheckpointID,
		CurrentGoal:    record.CurrentGoal,
		MinimalSummary: record.MinimalSummary,
		Ledger:         record.Ledger,
		EvidenceRefs:   record.EvidenceRefs,
		OccurredAt:     event.CreatedAt,
		Source:         "runtime_event",
		Extra:          extra,
	}, nil
}

func checkpointRecordFromEvent(event domain.EventEnvelope) (CheckpointRecord, error) {
	var payload struct {
		AgentSessionID string                 `json:"agent_session_id"`
		CheckpointID   string                 `json:"checkpoint_id"`
		CurrentGoal    string                 `json:"current_goal"`
		MinimalSummary string                 `json:"minimal_summary"`
		Ledger         map[string]interface{} `json:"ledger"`
		EvidenceRefs   []string               `json:"evidence_refs"`
		OccurredAt     string                 `json:"occurred_at"`
		Source         string                 `json:"source"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return CheckpointRecord{}, apperrors.Wrap(apperrors.CodeValidation, "agent_session_service.decode_checkpoint", err)
	}
	occurredAt := event.CreatedAt
	if payload.OccurredAt != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, payload.OccurredAt); err == nil {
			occurredAt = parsed
		}
	}
	return CheckpointRecord{
		Event:          event,
		AgentSessionID: payload.AgentSessionID,
		CheckpointID:   payload.CheckpointID,
		CurrentGoal:    payload.CurrentGoal,
		MinimalSummary: payload.MinimalSummary,
		Ledger:         payload.Ledger,
		EvidenceRefs:   payload.EvidenceRefs,
		OccurredAt:     occurredAt,
		Source:         payload.Source,
	}, nil
}

func copyMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	output := make(map[string]interface{}, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func putIfNotEmpty(target map[string]interface{}, key, value string) {
	if value != "" {
		target[key] = value
	}
}

func putSliceIfNotNil(target map[string]interface{}, key string, values []string) {
	if values != nil {
		target[key] = values
	}
}
