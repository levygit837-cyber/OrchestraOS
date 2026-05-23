package assignment_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/assignment"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestAssign_Success(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	a, err := m.Assign("wu1", "code_worker", "initial assignment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.WorkUnitID != "wu1" {
		t.Errorf("expected wu1, got %s", a.WorkUnitID)
	}
	if a.Status != domain.AssignmentStatusActive {
		t.Errorf("expected active, got %s", a.Status)
	}
}

func TestAssign_EmptyFields(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	_, err := m.Assign("", "agent", "reason")
	if err == nil {
		t.Fatal("expected error for empty WU ID")
	}
	_, err = m.Assign("wu1", "", "reason")
	if err == nil {
		t.Fatal("expected error for empty agent")
	}
}

func TestAssign_Conflict(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	_, _ = m.Assign("wu1", "agent_a", "first")
	_, err := m.Assign("wu1", "agent_b", "second")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestReplace_Success(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	_, _ = m.Assign("wu1", "agent_a", "first")
	a, err := m.Replace("wu1", "agent_b", "upgrade")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.AgentProfile != "agent_b" {
		t.Errorf("expected agent_b, got %s", a.AgentProfile)
	}
	if a.Status != domain.AssignmentStatusActive {
		t.Errorf("expected active, got %s", a.Status)
	}
}

func TestReplace_NoExisting(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	_, err := m.Replace("wu1", "agent_b", "upgrade")
	if err == nil {
		t.Fatal("expected error when no existing assignment")
	}
}

func TestRemove_Success(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	_, _ = m.Assign("wu1", "agent_a", "first")
	if err := m.Remove("wu1", "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := m.ActiveByWorkUnit("wu1")
	if a != nil {
		t.Error("expected nil after removal")
	}
}

func TestRemove_NotFound(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	if err := m.Remove("wu1", "done"); err == nil {
		t.Fatal("expected error when no assignment")
	}
}

func TestActiveByWorkUnit_Found(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	_, _ = m.Assign("wu1", "agent_a", "first")
	a := m.ActiveByWorkUnit("wu1")
	if a == nil {
		t.Fatal("expected assignment")
	}
	if a.AgentProfile != "agent_a" {
		t.Errorf("expected agent_a, got %s", a.AgentProfile)
	}
}

func TestActiveByWorkUnit_NotFound(t *testing.T) {
	t.Parallel()
	m := assignment.NewManager()
	a := m.ActiveByWorkUnit("wu_nonexistent")
	if a != nil {
		t.Error("expected nil for nonexistent")
	}
}
