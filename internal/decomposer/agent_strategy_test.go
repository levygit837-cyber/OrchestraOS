package decomposer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/decomposer"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type fakeAgentRuntime struct {
	response string
	err      error
}

func (f *fakeAgentRuntime) Execute(_ context.Context, _ *domain.Prompt) (string, error) {
	return f.response, f.err
}

func TestAgentStrategy_Success(t *testing.T) {
	t.Parallel()
	rt := &fakeAgentRuntime{
		response: `{
			"rationale": "Split by auth and api domains",
			"work_units": [
				{
					"title": "Auth Module",
					"objective": "Implement authentication",
					"domain": "auth",
					"description": "Handle user login",
					"acceptance_criteria": ["users can log in"],
					"depends_on_indices": [],
					"suggested_agent": "code_worker"
				},
				{
					"title": "API Endpoints",
					"objective": "Build REST API",
					"domain": "api",
					"description": "Create endpoints",
					"acceptance_criteria": ["endpoints respond"],
					"depends_on_indices": [0],
					"suggested_agent": "code_worker"
				}
			]
		}`,
	}

	s := decomposer.NewAgentStrategy(rt)
	req := &domain.DecompositionRequest{
		TaskID:   "t1",
		RawInput: "Add login with API",
		Context:  domain.TaskContext{TaskID: "t1", Intent: "login"},
	}

	result, err := s.Decompose(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.WorkUnits) != 2 {
		t.Fatalf("expected 2 specs, got %d", len(result.WorkUnits))
	}
	if result.WorkUnits[0].Context.Domain != "auth" {
		t.Errorf("expected domain auth, got %s", result.WorkUnits[0].Context.Domain)
	}
	if len(result.WorkUnits[1].DependsOn) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(result.WorkUnits[1].DependsOn))
	}
}

func TestAgentStrategy_RuntimeError(t *testing.T) {
	t.Parallel()
	rt := &fakeAgentRuntime{err: errors.New("LLM unavailable")}
	s := decomposer.NewAgentStrategy(rt)
	req := &domain.DecompositionRequest{TaskID: "t1", RawInput: "test"}

	_, err := s.Decompose(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAgentStrategy_InvalidJSON(t *testing.T) {
	t.Parallel()
	rt := &fakeAgentRuntime{response: "not json"}
	s := decomposer.NewAgentStrategy(rt)
	req := &domain.DecompositionRequest{TaskID: "t1", RawInput: "test"}

	_, err := s.Decompose(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAgentStrategy_EmptyWorkUnits(t *testing.T) {
	t.Parallel()
	rt := &fakeAgentRuntime{response: `{"rationale":"empty","work_units":[]}`}
	s := decomposer.NewAgentStrategy(rt)
	req := &domain.DecompositionRequest{TaskID: "t1", RawInput: "test"}

	_, err := s.Decompose(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty work units")
	}
}

func TestAgentStrategy_Name(t *testing.T) {
	t.Parallel()
	s := decomposer.NewAgentStrategy(nil)
	if s.Name() != "agent_llm_v1" {
		t.Errorf("expected agent_llm_v1, got %s", s.Name())
	}
}
