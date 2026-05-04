package statemachine

import (
	"encoding/json"
	"strings"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type ReplayState struct {
	TaskStatus         domain.TaskStatus
	WorkUnitStatuses   map[string]domain.WorkUnitStatus
	RunStatuses        map[string]domain.RunStatus
	AgentSessionStatus map[string]domain.AgentSessionStatus
	LastCheckpoint     *domain.EventEnvelope
}

func Project(events []domain.EventEnvelope) ReplayState {
	state, _ := project(events, false)
	return state
}

func ProjectStrict(events []domain.EventEnvelope) (ReplayState, error) {
	return project(events, true)
}

func project(events []domain.EventEnvelope, strict bool) (ReplayState, error) {
	state := ReplayState{
		WorkUnitStatuses:   map[string]domain.WorkUnitStatus{},
		RunStatuses:        map[string]domain.RunStatus{},
		AgentSessionStatus: map[string]domain.AgentSessionStatus{},
	}

	for i := range events {
		event := events[i]
		if status, ok := reduceTask(event); ok {
			if strict {
				if err := validateReplayTransition(AggregateTask, string(state.TaskStatus), string(status), event); err != nil {
					return state, err
				}
			}
			state.TaskStatus = status
		}
		if event.WorkUnitID != "" {
			if status, ok := reduceWorkUnit(event); ok {
				if strict {
					if err := validateReplayTransition(AggregateWorkUnit, string(state.WorkUnitStatuses[event.WorkUnitID]), string(status), event); err != nil {
						return state, err
					}
				}
				state.WorkUnitStatuses[event.WorkUnitID] = status
			}
		}
		if event.RunID != "" {
			if status, ok := reduceRun(event); ok {
				if strict {
					if err := validateReplayTransition(AggregateRun, string(state.RunStatuses[event.RunID]), string(status), event); err != nil {
						return state, err
					}
				}
				state.RunStatuses[event.RunID] = status
			}
		}
		if event.AgentID != "" && event.RunID != "" {
			if status, ok := reduceAgentSession(event); ok {
				if strict {
					if err := validateReplayTransition(AggregateAgentSession, string(state.AgentSessionStatus[event.AgentID]), string(status), event); err != nil {
						return state, err
					}
				}
				state.AgentSessionStatus[event.AgentID] = status
			}
		}
		if event.Type == "agent.checkpoint_reached" {
			copy := event
			state.LastCheckpoint = &copy
		}
	}

	return state, nil
}

func validateReplayTransition(aggregate Aggregate, from, to string, event domain.EventEnvelope) error {
	if from == "" || from == to {
		return nil
	}
	if err := CanTransition(aggregate, from, to, TransitionContext{
		EvidenceRefs:      evidenceRefsFromEvent(event),
		ValidationEventID: validationEventIDFromEvent(event),
		Justification:     justificationFromEvent(event),
	}); err != nil {
		return apperrors.Wrap(apperrors.CodeInvalidTransition, "statemachine.replay", err)
	}
	return nil
}

func evidenceRefsFromEvent(event domain.EventEnvelope) []string {
	var payload struct {
		EvidenceRefs []string `json:"evidence_refs"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil
	}
	return payload.EvidenceRefs
}

func validationEventIDFromEvent(event domain.EventEnvelope) string {
	var payload struct {
		ValidationEventID string `json:"validation_event_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ""
	}
	return payload.ValidationEventID
}

func justificationFromEvent(event domain.EventEnvelope) string {
	var payload struct {
		Justification string `json:"justification"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return ""
	}
	return payload.Justification
}

func reduceTask(event domain.EventEnvelope) (domain.TaskStatus, bool) {
	switch event.Type {
	case "task.created":
		return domain.TaskStatusCreated, true
	case "task.triaged":
		return domain.TaskStatusTriaged, true
	case "task.planned":
		return domain.TaskStatusPlanned, true
	case "task.scheduled":
		return domain.TaskStatusScheduled, true
	case "task.sandbox_preparing":
		return domain.TaskStatusSandboxPreparing, true
	case "task.started":
		return domain.TaskStatusRunning, true
	case "task.waiting_approval":
		return domain.TaskStatusWaitingApproval, true
	case "task.paused":
		return domain.TaskStatusPaused, true
	case "task.validating":
		return domain.TaskStatusValidating, true
	case "task.completed":
		return domain.TaskStatusCompleted, true
	case "task.failed":
		return domain.TaskStatusFailed, true
	case "task.cancelled":
		return domain.TaskStatusCancelled, true
	default:
		return "", false
	}
}

func reduceWorkUnit(event domain.EventEnvelope) (domain.WorkUnitStatus, bool) {
	switch event.Type {
	case "work_unit.created":
		return domain.WorkUnitStatusCreated, true
	case "work_unit.planned":
		return domain.WorkUnitStatusPlanned, true
	case "work_unit.scheduled":
		return domain.WorkUnitStatusScheduled, true
	case "work_unit.blocked":
		return domain.WorkUnitStatusBlocked, true
	case "work_unit.started":
		return domain.WorkUnitStatusRunning, true
	case "work_unit.waiting_approval":
		return domain.WorkUnitStatusWaitingApproval, true
	case "work_unit.paused":
		return domain.WorkUnitStatusPaused, true
	case "work_unit.validating":
		return domain.WorkUnitStatusValidating, true
	case "work_unit.completed":
		return domain.WorkUnitStatusCompleted, true
	case "work_unit.failed":
		return domain.WorkUnitStatusFailed, true
	case "work_unit.cancelled":
		return domain.WorkUnitStatusCancelled, true
	default:
		return "", false
	}
}

func reduceRun(event domain.EventEnvelope) (domain.RunStatus, bool) {
	switch event.Type {
	case "run.created":
		return domain.RunStatusCreated, true
	case "run.started":
		return domain.RunStatusRunning, true
	case "run.waiting_approval":
		return domain.RunStatusWaitingApproval, true
	case "run.paused":
		return domain.RunStatusPaused, true
	case "run.resumed":
		return domain.RunStatusRunning, true
	case "run.validating":
		return domain.RunStatusValidating, true
	case "run.completed":
		return domain.RunStatusCompleted, true
	case "run.failed":
		return domain.RunStatusFailed, true
	case "run.cancelled":
		return domain.RunStatusCancelled, true
	default:
		return "", false
	}
}

func reduceAgentSession(event domain.EventEnvelope) (domain.AgentSessionStatus, bool) {
	switch {
	case event.Type == "agent.session_starting":
		return domain.AgentSessionStatusStarting, true
	case event.Type == "agent.session_running":
		return domain.AgentSessionStatusRunning, true
	case event.Type == "agent.session_waiting_approval":
		return domain.AgentSessionStatusWaitingApproval, true
	case event.Type == "agent.session_paused":
		return domain.AgentSessionStatusPaused, true
	case event.Type == "agent.session_disconnected":
		return domain.AgentSessionStatusDisconnected, true
	case event.Type == "agent.session_stopping":
		return domain.AgentSessionStatusStopping, true
	case event.Type == "agent.session_stopped":
		return domain.AgentSessionStatusStopped, true
	case event.Type == "agent.session_failed":
		return domain.AgentSessionStatusFailed, true
	case event.Type == "agent.connected" || event.Type == "agent.started":
		return domain.AgentSessionStatusRunning, true
	case strings.HasPrefix(event.Type, "agent."):
		return "", false
	default:
		return "", false
	}
}
