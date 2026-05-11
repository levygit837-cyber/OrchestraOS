package run

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

func EventTypeForStatus(status domain.RunStatus) string {
	if status == domain.RunStatusRunning {
		return "run.started"
	}
	return "run." + string(status)
}

func ResultForStatus(status domain.RunStatus) *domain.RunResult {
	switch status {
	case domain.RunStatusCompleted:
		result := domain.RunResultSucceeded
		return &result
	case domain.RunStatusFailed:
		result := domain.RunResultFailed
		return &result
	case domain.RunStatusCancelled:
		result := domain.RunResultCancelled
		return &result
	default:
		return nil
	}
}
