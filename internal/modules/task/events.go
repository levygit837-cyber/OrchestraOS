package task

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

func EventTypeForStatus(status domain.TaskStatus) string {
	if status == domain.TaskStatusRunning {
		return "task.started"
	}
	return "task." + string(status)
}
