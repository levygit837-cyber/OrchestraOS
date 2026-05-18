package statemachine

import (
	"encoding/json"
	"strings"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type ReplayState struct {
	TaskStatus         string
	WorkUnitStatuses   map[string]string
	RunStatuses        map[string]string
	AgentSessionStatus map[string]string
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
		WorkUnitStatuses:   map[string]string{},
		RunStatuses:        map[string]string{},
		AgentSessionStatus: map[string]string{},
	}

	for i := range events {
		event := events[i]
		if status, ok := reduceTask(event); ok {
			if strict {
				if err := validateReplayTransition(AggregateTask, state.TaskStatus, status, event); err != nil {
					return state, err
				}
			}
			state.TaskStatus = status
		}
		if event.WorkUnitID != "" {
			if status, ok := reduceWorkUnit(event); ok {
				if strict {
					if err := validateReplayTransition(AggregateWorkUnit, state.WorkUnitStatuses[event.WorkUnitID], status, event); err != nil {
						return state, err
					}
				}
				state.WorkUnitStatuses[event.WorkUnitID] = status
			}
		}
		if event.RunID != "" {
			if status, ok := reduceRun(event); ok {
				if strict {
					if err := validateReplayTransition(AggregateRun, state.RunStatuses[event.RunID], status, event); err != nil {
						return state, err
					}
				}
				state.RunStatuses[event.RunID] = status
			}
		}
		if event.AgentID != "" && event.RunID != "" {
			if status, ok := reduceAgentSession(event); ok {
				if strict {
					if err := validateReplayTransition(AggregateAgentSession, state.AgentSessionStatus[event.AgentID], status, event); err != nil {
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

func reduceTask(event domain.EventEnvelope) (string, bool) {
	switch event.Type {
	case "task.created":
		return "created", true
	case "task.triaged":
		return "triaged", true
	case "task.planned":
		return "planned", true
	case "task.scheduled":
		return "scheduled", true
	case "task.sandbox_preparing":
		return "sandbox_preparing", true
	case "task.started":
		return "running", true
	case "task.waiting_approval":
		return "waiting_approval", true
	case "task.paused":
		return "paused", true
	case "task.validating":
		return "validating", true
	case "task.completed":
		return "completed", true
	case "task.failed":
		return "failed", true
	case "task.cancelled":
		return "cancelled", true
	default:
		return "", false
	}
}

func reduceWorkUnit(event domain.EventEnvelope) (string, bool) {
	switch event.Type {
	case "work_unit.created":
		return "created", true
	case "work_unit.planned":
		return "planned", true
	case "work_unit.scheduled":
		return "scheduled", true
	case "work_unit.blocked":
		return "blocked", true
	case "work_unit.started":
		return "running", true
	case "work_unit.waiting_approval":
		return "waiting_approval", true
	case "work_unit.paused":
		return "paused", true
	case "work_unit.validating":
		return "validating", true
	case "work_unit.completed":
		return "completed", true
	case "work_unit.failed":
		return "failed", true
	case "work_unit.cancelled":
		return "cancelled", true
	default:
		return "", false
	}
}

func reduceRun(event domain.EventEnvelope) (string, bool) {
	switch event.Type {
	case "run.created":
		return "created", true
	case "run.started":
		return "running", true
	case "run.waiting_approval":
		return "waiting_approval", true
	case "run.paused":
		return "paused", true
	case "run.resumed":
		return "running", true
	case "run.validating":
		return "validating", true
	case "run.completed":
		return "completed", true
	case "run.failed":
		return "failed", true
	case "run.cancelled":
		return "cancelled", true
	default:
		return "", false
	}
}

func reduceAgentSession(event domain.EventEnvelope) (string, bool) {
	switch {
	case event.Type == "agent.session_starting":
		return "starting", true
	case event.Type == "agent.session_running":
		return "running", true
	case event.Type == "agent.session_waiting_approval":
		return "waiting_approval", true
	case event.Type == "agent.session_paused":
		return "paused", true
	case event.Type == "agent.session_disconnected":
		return "disconnected", true
	case event.Type == "agent.session_stopping":
		return "stopping", true
	case event.Type == "agent.session_stopped":
		return "stopped", true
	case event.Type == "agent.session_failed":
		return "failed", true
	case event.Type == "agent.connected" || event.Type == "agent.started":
		return "running", true
	case strings.HasPrefix(event.Type, "agent."):
		return "", false
	default:
		return "", false
	}
}
