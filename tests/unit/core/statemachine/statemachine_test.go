package statemachine_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestRunTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.RunStatus
		to      domain.RunStatus
		ctx     statemachine.TransitionContext
		wantErr bool
	}{
		{
			name: "created to running",
			from: domain.RunStatusCreated,
			to:   domain.RunStatusRunning,
		},
		{
			name:    "running cannot complete without validating",
			from:    domain.RunStatusRunning,
			to:      domain.RunStatusCompleted,
			ctx:     statemachine.TransitionContext{Justification: "runtime finished"},
			wantErr: true,
		},
		{
			name: "validating completes with evidence",
			from: domain.RunStatusValidating,
			to:   domain.RunStatusCompleted,
			ctx:  statemachine.TransitionContext{EvidenceRefs: []string{"validation.completed:test"}},
		},
		{
			name:    "validating cannot complete without evidence",
			from:    domain.RunStatusValidating,
			to:      domain.RunStatusCompleted,
			wantErr: true,
		},
		{
			name:    "terminal cannot resume",
			from:    domain.RunStatusCompleted,
			to:      domain.RunStatusRunning,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateRun, string(tt.from), string(tt.to), tt.ctx)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestTaskTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.TaskStatus
		to      domain.TaskStatus
		ctx     statemachine.TransitionContext
		wantErr bool
	}{
		{name: "created to triaged", from: domain.TaskStatusCreated, to: domain.TaskStatusTriaged},
		{name: "triaged to planned", from: domain.TaskStatusTriaged, to: domain.TaskStatusPlanned},
		{name: "planned cannot jump to running", from: domain.TaskStatusPlanned, to: domain.TaskStatusRunning, wantErr: true},
		{name: "validating completes with justification", from: domain.TaskStatusValidating, to: domain.TaskStatusCompleted, ctx: statemachine.TransitionContext{Justification: "manual validation accepted"}},
		{name: "completed terminal", from: domain.TaskStatusCompleted, to: domain.TaskStatusRunning, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateTask, string(tt.from), string(tt.to), tt.ctx)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestWorkUnitTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.WorkUnitStatus
		to      domain.WorkUnitStatus
		ctx     statemachine.TransitionContext
		wantErr bool
	}{
		{name: "created to running allowed for current MVP CLI", from: domain.WorkUnitStatusCreated, to: domain.WorkUnitStatusRunning},
		{name: "running to validating", from: domain.WorkUnitStatusRunning, to: domain.WorkUnitStatusValidating},
		{name: "validating completes with evidence", from: domain.WorkUnitStatusValidating, to: domain.WorkUnitStatusCompleted, ctx: statemachine.TransitionContext{EvidenceRefs: []string{"artifact:diff"}}},
		{name: "running cannot complete directly", from: domain.WorkUnitStatusRunning, to: domain.WorkUnitStatusCompleted, ctx: statemachine.TransitionContext{EvidenceRefs: []string{"artifact:diff"}}, wantErr: true},
		{name: "failed terminal", from: domain.WorkUnitStatusFailed, to: domain.WorkUnitStatusRunning, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(tt.from), string(tt.to), tt.ctx)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestAgentSessionTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    domain.AgentSessionStatus
		to      domain.AgentSessionStatus
		wantErr bool
	}{
		{name: "starting to running", from: domain.AgentSessionStatusStarting, to: domain.AgentSessionStatusRunning},
		{name: "running to stopping", from: domain.AgentSessionStatusRunning, to: domain.AgentSessionStatusStopping},
		{name: "stopping to stopped", from: domain.AgentSessionStatusStopping, to: domain.AgentSessionStatusStopped},
		{name: "running directly to stopped blocked", from: domain.AgentSessionStatusRunning, to: domain.AgentSessionStatusStopped, wantErr: true},
		{name: "stopped terminal", from: domain.AgentSessionStatusStopped, to: domain.AgentSessionStatusRunning, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateAgentSession, string(tt.from), string(tt.to), statemachine.TransitionContext{})
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestReplayProjection(t *testing.T) {
	taskID := uuid.New().String()
	workUnitID := uuid.New().String()
	runID := uuid.New().String()
	agentID := "agent-test"

	events := []domain.EventEnvelope{
		event("task.created", taskID, "", "", "", nil),
		event("work_unit.created", taskID, "", workUnitID, "", nil),
		event("run.started", taskID, runID, workUnitID, "", nil),
		event("agent.session_running", taskID, runID, workUnitID, agentID, nil),
		event("agent.checkpoint_reached", taskID, runID, workUnitID, agentID, map[string]interface{}{
			"checkpoint_id":   "checkpoint-1",
			"current_goal":    "implementation",
			"ledger":          map[string]interface{}{"pending_todos": []string{}},
			"minimal_summary": "ready",
		}),
		event("run.validating", taskID, runID, workUnitID, "", nil),
		event("run.completed", taskID, runID, workUnitID, "", nil),
	}

	state := statemachine.Project(events)
	if state.TaskStatus != domain.TaskStatusCreated {
		t.Fatalf("expected task status created, got %s", state.TaskStatus)
	}
	if state.WorkUnitStatuses[workUnitID] != domain.WorkUnitStatusCreated {
		t.Fatalf("expected work unit status created, got %s", state.WorkUnitStatuses[workUnitID])
	}
	if state.RunStatuses[runID] != domain.RunStatusCompleted {
		t.Fatalf("expected run status completed, got %s", state.RunStatuses[runID])
	}
	if state.AgentSessionStatus[agentID] != domain.AgentSessionStatusRunning {
		t.Fatalf("expected agent session status running, got %s", state.AgentSessionStatus[agentID])
	}
	if state.LastCheckpoint == nil || state.LastCheckpoint.Type != "agent.checkpoint_reached" {
		t.Fatalf("expected latest checkpoint, got %+v", state.LastCheckpoint)
	}
}

func TestReplayProjectionRejectsInvalidTransitions(t *testing.T) {
	taskID := uuid.New().String()
	runID := uuid.New().String()

	events := []domain.EventEnvelope{
		event("run.started", taskID, runID, "", "", nil),
		event("run.completed", taskID, runID, "", "", map[string]interface{}{
			"evidence_refs": []string{"validation.completed:test"},
		}),
	}

	if _, err := statemachine.ProjectStrict(events); err == nil {
		t.Fatal("expected replay to reject running to completed transition")
	}
}

func TestReplayProjectionAcceptsCompletedWithEvidence(t *testing.T) {
	taskID := uuid.New().String()
	runID := uuid.New().String()

	events := []domain.EventEnvelope{
		event("run.started", taskID, runID, "", "", nil),
		event("run.validating", taskID, runID, "", "", nil),
		event("run.completed", taskID, runID, "", "", map[string]interface{}{
			"evidence_refs": []string{"validation.completed:test"},
		}),
	}

	state, err := statemachine.ProjectStrict(events)
	if err != nil {
		t.Fatalf("expected replay to accept valid completion, got %v", err)
	}
	if state.RunStatuses[runID] != domain.RunStatusCompleted {
		t.Fatalf("expected run completed, got %s", state.RunStatuses[runID])
	}
}

func event(eventType, taskID, runID, workUnitID, agentID string, payload map[string]interface{}) domain.EventEnvelope {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payloadBytes, _ := json.Marshal(payload)
	return domain.EventEnvelope{
		Type:       eventType,
		TaskID:     taskID,
		RunID:      runID,
		WorkUnitID: workUnitID,
		AgentID:    agentID,
		Payload:    payloadBytes,
	}
}
