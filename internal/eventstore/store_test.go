package eventstore

import (
	"encoding/json"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestValidateOperationalTaskGraphCreatedPayloadRequiresMatchingTask(t *testing.T) {
	taskID := "task-1"
	payload := domain.TaskGraphCreatedPayload{
		TaskID:          taskID,
		GraphID:         "graph-1",
		GraphVersion:    1,
		PlannerStrategy: "local_heuristic_v1",
		Nodes: []domain.TaskGraphNodeInfo{
			{
				ID:                 "work-unit-1",
				Title:              "Work unit",
				Objective:          "Validate graph payload",
				AgentProfile:       "default",
				OwnedPaths:         []string{},
				ReadPaths:          []string{},
				AcceptanceCriteria: []string{"graph payload references the parent task"},
				ValidationPlan:     []string{"go test ./internal/eventstore"},
			},
		},
		Edges: []domain.TaskGraphEdgeInfo{},
	}

	if err := validateOperationalPayload(taskGraphEnvelope(t, taskID, payload)); err != nil {
		t.Fatalf("expected matching task_id to be valid, got %v", err)
	}

	missingTask := payload
	missingTask.TaskID = ""
	if err := validateOperationalPayload(taskGraphEnvelope(t, taskID, missingTask)); err == nil {
		t.Fatal("expected missing payload task_id to be rejected")
	}

	wrongTask := payload
	wrongTask.TaskID = "task-2"
	if err := validateOperationalPayload(taskGraphEnvelope(t, taskID, wrongTask)); err == nil {
		t.Fatal("expected mismatched payload task_id to be rejected")
	}
}

func taskGraphEnvelope(t *testing.T, taskID string, payload domain.TaskGraphCreatedPayload) *domain.EventEnvelope {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return &domain.EventEnvelope{
		Type:    "task.graph_created",
		TaskID:  taskID,
		Payload: payloadBytes,
	}
}
