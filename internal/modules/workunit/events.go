package workunit

func EventTypeForStatus(status Status) string {
	if status == StatusRunning {
		return "work_unit.started"
	}
	return "work_unit." + string(status)
}
