package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// FakeRuntime simulates an agent runtime for testing
type FakeRuntime struct {
	config         RuntimeConfig
	status         RuntimeStatus
	eventChan      chan *domain.EventEnvelope
	stopChan       chan struct{}
	started        bool
	heartbeatCount int
}

// NewFakeRuntime creates a new fake agent runtime
func NewFakeRuntime() *FakeRuntime {
	return &FakeRuntime{
		eventChan: make(chan *domain.EventEnvelope, 100),
		stopChan:  make(chan struct{}),
	}
}

// Start starts the fake runtime simulation
func (f *FakeRuntime) Start(ctx context.Context, config RuntimeConfig) error {
	if f.started {
		return fmt.Errorf("runtime already started")
	}

	f.config = config
	f.status = RuntimeStatus{
		State:       "starting",
		CurrentStep: 0,
	}
	f.started = true

	// Emit agent.connected event
	f.emitEvent("agent.connected", "v1", map[string]interface{}{
		"agent_id": config.AgentID,
		"run_id":   config.RunID,
		"status":   "connected",
	})

	// Emit agent.started event
	f.status.State = "running"
	f.emitEvent("agent.started", "v1", map[string]interface{}{
		"agent_id":            config.AgentID,
		"run_id":              config.RunID,
		"work_unit":           config.WorkUnitID,
		"prompt_hash":         config.PromptHash,
		"prompt_snapshot_id":  config.PromptSnapshotID,
		"toolset_snapshot_id": config.ToolsetSnapshotID,
		"toolset":             config.Toolset,
	})

	f.emitHeartbeat()

	// Start background goroutines
	go f.heartbeatLoop()
	go f.simulationLoop()

	return nil
}

// Stop stops the fake runtime
func (f *FakeRuntime) Stop(ctx context.Context) error {
	if !f.started {
		return nil
	}

	close(f.stopChan)
	f.status.State = "stopped"
	f.started = false

	f.emitEvent("agent.stopped", "v1", map[string]interface{}{
		"agent_id": f.config.AgentID,
		"run_id":   f.config.RunID,
		"reason":   "requested",
	})

	return nil
}

// SendEvent sends an event to the runtime (e.g., tool approval)
func (f *FakeRuntime) SendEvent(ctx context.Context, event *domain.EventEnvelope) error {
	// Handle incoming events (like tool approvals)
	switch event.Type {
	case "tool.approved":
		// Continue with tool execution
		f.emitEvent("agent.tool_executed", "v1", map[string]interface{}{
			"agent_id": f.config.AgentID,
			"run_id":   f.config.RunID,
			"tool":     "simulated_tool",
		})
	case "tool.denied":
		f.emitEvent("agent.tool_denied", "v1", map[string]interface{}{
			"agent_id": f.config.AgentID,
			"run_id":   f.config.RunID,
			"reason":   "tool denied by policy",
		})
	}
	return nil
}

// ReceiveEvent receives events from the runtime
func (f *FakeRuntime) ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error) {
	select {
	case event := <-f.eventChan:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.stopChan:
		return nil, fmt.Errorf("runtime stopped")
	}
}

// Status returns the current runtime status
func (f *FakeRuntime) Status() RuntimeStatus {
	return f.status
}

func (f *FakeRuntime) heartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !f.started {
				return
			}
			f.emitHeartbeat()
		case <-f.stopChan:
			return
		}
	}
}

func (f *FakeRuntime) emitHeartbeat() {
	f.heartbeatCount++
	f.status.LastHeartbeat = time.Now().Unix()
	f.emitEvent("agent.heartbeat", "v1", map[string]interface{}{
		"agent_id": f.config.AgentID,
		"run_id":   f.config.RunID,
		"count":    f.heartbeatCount,
	})
}

func (f *FakeRuntime) simulationLoop() {
	// Simulate agent work with delays
	time.Sleep(1 * time.Second)

	if !f.started {
		return
	}

	// Emit checkpoint
	f.status.CurrentStep = 1
	f.emitEvent("agent.checkpoint_reached", "v1", map[string]interface{}{
		"checkpoint_id":   uuid.New().String(),
		"current_goal":    "initial analysis",
		"minimal_summary": "Task analyzed successfully",
		"ledger": map[string]interface{}{
			"current_goal":    "initial analysis",
			"completed_goals": []string{},
			"pending_todos":   []string{"implementation"},
			"blockers":        []string{},
			"risks":           []string{},
		},
		"files_read":     []string{},
		"files_modified": []string{},
		"evidence_refs":  []string{"fake-runtime:analysis"},
	})

	// Simulate tool request
	time.Sleep(500 * time.Millisecond)
	if !f.started {
		return
	}

	f.emitEvent("agent.tool_requested", "v1", map[string]interface{}{
		"agent_id": f.config.AgentID,
		"run_id":   f.config.RunID,
		"tool":     "read_file",
		"input":    map[string]string{"path": "README.md"},
		"reason":   "Need to understand project structure",
	})

	// Wait a bit and simulate completion
	time.Sleep(2 * time.Second)
	if !f.started {
		return
	}

	f.status.CurrentStep = 2
	f.emitEvent("agent.checkpoint_reached", "v1", map[string]interface{}{
		"checkpoint_id":   uuid.New().String(),
		"current_goal":    "implementation",
		"minimal_summary": "Changes made successfully",
		"ledger": map[string]interface{}{
			"current_goal":    "implementation",
			"completed_goals": []string{"initial analysis", "implementation"},
			"pending_todos":   []string{},
			"blockers":        []string{},
			"risks":           []string{},
		},
		"files_read":     []string{"README.md"},
		"files_modified": []string{"main.go", "utils.go"},
		"evidence_refs":  []string{"fake-runtime:implementation"},
	})

	// Simulate successful completion
	time.Sleep(1 * time.Second)
	if !f.started {
		return
	}

	result := map[string]interface{}{
		"status":    "completed",
		"summary":   "Task completed successfully",
		"artifacts": []string{"main.go", "utils.go", "test.go"},
		"metrics": map[string]int{
			"files_created": 2,
			"files_updated": 1,
			"tests_passed":  3,
		},
	}

	f.emitEvent("agent.completed", "v1", result)
	f.status.State = "completed"
}

func (f *FakeRuntime) emitEvent(eventType, version string, payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)

	event := &domain.EventEnvelope{
		ID:          uuid.New().String(),
		Type:        eventType,
		Version:     version,
		TaskID:      f.config.TaskID,
		RunID:       f.config.RunID,
		WorkUnitID:  f.config.WorkUnitID,
		AgentID:     f.config.AgentID,
		Sequence:    0, // Will be set by event store
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		CreatedAt:   time.Now(),
		Payload:     payloadBytes,
	}

	// Non-blocking send
	select {
	case f.eventChan <- event:
	default:
		// Channel full, drop event (for simulation purposes)
	}
}

// Random failure simulation
func (f *FakeRuntime) maybeFail() bool {
	// 10% chance of simulated failure
	return rand.Float32() < 0.1
}
