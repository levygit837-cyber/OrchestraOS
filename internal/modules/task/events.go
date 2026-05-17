package task

func EventTypeForStatus(status Status) string {
	if status == StatusRunning {
		return "task.started"
	}
	return "task." + string(status)
}
