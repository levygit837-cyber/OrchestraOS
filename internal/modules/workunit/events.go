package workunit

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

func EventTypeForStatus(status domain.WorkUnitStatus) string {
	if status == domain.WorkUnitStatusRunning {
		return "work_unit.started"
	}
	return "work_unit." + string(status)
}
