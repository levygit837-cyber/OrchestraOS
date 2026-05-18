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
		from    string
		to      string
		ctx     statemachine.TransitionContext
		wantErr bool
	}{
		{
			name: "created to running",
			from: "created",
			to:   "running",
		},
		{
			name:    "running cannot complete without validating",
			from:    "running",
			to:      "completed",
			ctx:     statemachine.TransitionContext{Justification: "runtime finished"},
			wantErr: true,
		},
		{
			name: "validating completes with evidence",
			from: "validating",
			to:   "completed",
			ctx:  statemachine.TransitionContext{EvidenceRefs: []string{"validation.completed:test"}},
		},
		{
			name:    "validating cannot complete without evidence",
			from:    "validating",
			to:      "completed",
			wantErr: true,
		},
		{
			name:    "terminal cannot resume",
			from:    "completed",
			to:      "running",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateRun, tt.from, tt.to, tt.ctx)
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
		from    string
		to      string
		ctx     statemachine.TransitionContext
		wantErr bool
	}{
		{name: "created to triaged", from: "created", to: "triaged"},
		{name: "triaged to planned", from: "triaged", to: "planned"},
		{name: "planned cannot jump to running", from: "planned", to: "running", wantErr: true},
		{name: "validating completes with justification", from: "validating", to: "completed", ctx: statemachine.TransitionContext{Justification: "manual validation accepted"}},
		{name: "completed terminal", from: "completed", to: "running", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateTask, tt.from, tt.to, tt.ctx)
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
		from    string
		to      string
		ctx     statemachine.TransitionContext
		wantErr bool
	}{
		{name: "created to running allowed for current MVP CLI", from: "created", to: "running"},
		{name: "running to validating", from: "running", to: "validating"},
		{name: "validating completes with evidence", from: "validating", to: "completed", ctx: statemachine.TransitionContext{EvidenceRefs: []string{"artifact:diff"}}},
		{name: "running cannot complete directly", from: "running", to: "completed", ctx: statemachine.TransitionContext{EvidenceRefs: []string{"artifact:diff"}}, wantErr: true},
		{name: "failed terminal", from: "failed", to: "running", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateWorkUnit, tt.from, tt.to, tt.ctx)
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
		from    string
		to      string
		wantErr bool
	}{
		{name: "starting to running", from: "starting", to: "running"},
		{name: "running to stopping", from: "running", to: "stopping"},
		{name: "stopping to stopped", from: "stopping", to: "stopped"},
		{name: "running directly to stopped blocked", from: "running", to: "stopped", wantErr: true},
		{name: "stopped terminal", from: "stopped", to: "running", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := statemachine.CanTransition(statemachine.AggregateAgentSession, tt.from, tt.to, statemachine.TransitionContext{})
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
	if state.TaskStatus != "created" {
		t.Fatalf("expected task status created, got %s", state.TaskStatus)
	}
	if state.WorkUnitStatuses[workUnitID] != "created" {
		t.Fatalf("expected work unit status created, got %s", state.WorkUnitStatuses[workUnitID])
	}
	if state.RunStatuses[runID] != "completed" {
		t.Fatalf("expected run status completed, got %s", state.RunStatuses[runID])
	}
	if state.AgentSessionStatus[agentID] != "running" {
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
	if state.RunStatuses[runID] != "completed" {
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
