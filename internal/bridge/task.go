package bridge

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
)

// TaskToDomain converts a local task.Task to the legacy domain.Task for external consumers.
func TaskToDomain(t *taskmod.Task) *domain.Task {
	if t == nil {
		return nil
	}
	return &domain.Task{
		ID:                   t.ID,
		Title:                t.Title,
		Description:          t.Description,
		Status:               domain.TaskStatus(t.Status),
		Priority:             domain.Priority(t.Priority),
		RiskLevel:            domain.RiskLevel(t.RiskLevel),
		CreatedFromMessageID: t.CreatedFromMessageID,
		AcceptanceCriteria:   t.AcceptanceCriteria,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

// TaskFromDomain converts a legacy domain.Task to the local task.Task.
func TaskFromDomain(t *domain.Task) *taskmod.Task {
	if t == nil {
		return nil
	}
	return &taskmod.Task{
		ID:                   t.ID,
		Title:                t.Title,
		Description:          t.Description,
		Status:               taskmod.Status(t.Status),
		Priority:             taskmod.Priority(t.Priority),
		RiskLevel:            taskmod.RiskLevel(t.RiskLevel),
		CreatedFromMessageID: t.CreatedFromMessageID,
		AcceptanceCriteria:   t.AcceptanceCriteria,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}
