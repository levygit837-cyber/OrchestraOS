package assignment

import (
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Manager handles agent assignment lifecycle for work units.
type Manager struct {
	assignments map[string]*domain.AgentAssignment
}

// NewManager creates a Manager with empty state.
func NewManager() *Manager {
	return &Manager{assignments: make(map[string]*domain.AgentAssignment)}
}

// Assign couples an agent profile to a work unit.
func (m *Manager) Assign(wuID, agentProfile, reason string) (*domain.AgentAssignment, error) {
	if wuID == "" || agentProfile == "" {
		return nil, apperrors.New(apperrors.KindValidation, "assignment.assign", "work unit ID and agent profile are required")
	}

	if existing, ok := m.assignments[wuID]; ok && existing.Status == domain.AssignmentStatusActive {
		return nil, apperrors.New(apperrors.KindConflict, "assignment.assign", "work unit already has an active assignment; use Replace")
	}

	a := &domain.AgentAssignment{
		ID:           uuid.New().String(),
		WorkUnitID:   wuID,
		AgentProfile: agentProfile,
		Status:       domain.AssignmentStatusActive,
		Reason:       reason,
		AssignedAt:   time.Now(),
	}
	m.assignments[wuID] = a
	return a, nil
}

// Replace swaps the current agent for a work unit with a new one.
func (m *Manager) Replace(wuID, newAgent, reason string) (*domain.AgentAssignment, error) {
	if wuID == "" || newAgent == "" {
		return nil, apperrors.New(apperrors.KindValidation, "assignment.replace", "work unit ID and new agent profile are required")
	}

	existing, ok := m.assignments[wuID]
	if !ok || existing.Status != domain.AssignmentStatusActive {
		return nil, apperrors.New(apperrors.KindNotFound, "assignment.replace", "no active assignment for work unit: "+wuID)
	}

	now := time.Now()
	existing.Status = domain.AssignmentStatusReplaced
	existing.RemovedAt = &now

	a := &domain.AgentAssignment{
		ID:           uuid.New().String(),
		WorkUnitID:   wuID,
		AgentProfile: newAgent,
		Status:       domain.AssignmentStatusActive,
		Reason:       reason,
		AssignedAt:   now,
	}
	m.assignments[wuID] = a
	return a, nil
}

// Remove deactivates the agent assignment for a work unit.
func (m *Manager) Remove(wuID, reason string) error {
	existing, ok := m.assignments[wuID]
	if !ok || existing.Status != domain.AssignmentStatusActive {
		return apperrors.New(apperrors.KindNotFound, "assignment.remove", "no active assignment for work unit: "+wuID)
	}

	now := time.Now()
	existing.Status = domain.AssignmentStatusRemoved
	existing.RemovedAt = &now
	existing.Reason = reason
	return nil
}

// Get returns the current assignment for a work unit.
func (m *Manager) Get(wuID string) (*domain.AgentAssignment, error) {
	a, ok := m.assignments[wuID]
	if !ok {
		return nil, apperrors.New(apperrors.KindNotFound, "assignment.get", "no assignment for work unit: "+wuID)
	}
	return a, nil
}

// ActiveByWorkUnit returns the active assignment for a work unit, or nil.
func (m *Manager) ActiveByWorkUnit(wuID string) *domain.AgentAssignment {
	a, ok := m.assignments[wuID]
	if !ok || a.Status != domain.AssignmentStatusActive {
		return nil
	}
	return a
}
