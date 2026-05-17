package run

func EventTypeForStatus(status Status) string {
	if status == StatusRunning {
		return "run.started"
	}
	return "run." + string(status)
}

func ResultForStatus(status Status) *Result {
	switch status {
	case StatusCompleted:
		result := ResultSucceeded
		return &result
	case StatusFailed:
		result := ResultFailed
		return &result
	case StatusCancelled:
		result := ResultCancelled
		return &result
	default:
		return nil
	}
}
