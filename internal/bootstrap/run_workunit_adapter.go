package bootstrap

import (
	"context"
	"database/sql"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// TransitionRunWithWorkUnit synchronizes a run transition with its associated work unit.
// Lives in bootstrap to avoid run→workunit cross-module import (ADR-0019).
func TransitionRunWithWorkUnit(ctx context.Context, tx *sql.Tx, run *runmod.Run, target runmod.Status, input transition.TransitionInput) error {
	if run.WorkUnitID == "" {
		return nil
	}
	wu, err := workunitmod.RequireByID(ctx, tx, run.WorkUnitID)
	if err != nil {
		return err
	}
	var wuTarget domain.WorkUnitStatus
	switch target {
	case runmod.StatusRunning:
		wuTarget = domain.WorkUnitStatusRunning
	case runmod.StatusValidating:
		wuTarget = domain.WorkUnitStatusValidating
	case runmod.StatusCompleted:
		wuTarget = domain.WorkUnitStatusCompleted
	case runmod.StatusFailed:
		wuTarget = domain.WorkUnitStatusFailed
	case runmod.StatusCancelled:
		wuTarget = domain.WorkUnitStatusCancelled
	default:
		return nil
	}
	if wu.Status == wuTarget {
		return nil
	}
	if wuTarget == domain.WorkUnitStatusRunning {
		if err := dbcore.AcquireAdvisoryTxLock(ctx, tx, "work_unit_paths:"+wu.TaskID, "run.work_unit_path_lock"); err != nil {
			return err
		}
		if err := workunitmod.ValidateDependenciesCompleted(ctx, tx, wu); err != nil {
			return err
		}
		if err := workunitmod.ValidateOwnedPathAvailability(ctx, tx, wu); err != nil {
			return err
		}
	}
	if wuTarget == domain.WorkUnitStatusCompleted && len(wu.AcceptanceCriteria) == 0 && input.Justification == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "run.run_workunit_sync", "related work unit completion requires acceptance criteria or explicit justification")
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(wuTarget), transition.TransitionContext(input)); err != nil {
		if wuTarget == domain.WorkUnitStatusFailed && wu.Status == domain.WorkUnitStatusCreated {
			return nil
		}
		return err
	}
	if _, _, err := transition.AppendTransition(ctx, tx, "", workUnitEventTypeForStatus(wuTarget), run.TaskID, run.ID, wu.ID, input.AgentID, transition.TransitionPayload(wu.Status, wuTarget, input)); err != nil {
		return err
	}
	res, err := workunitmod.NewRepository(tx).UpdateStatus(wu.ID, wuTarget, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "run.run_workunit_sync.update_work_unit", err)
	}
	return dbcore.EnsureRowsAffected(res, "work unit", "run.run_workunit_sync.update_work_unit")
}

func workUnitEventTypeForStatus(status domain.WorkUnitStatus) string {
	if status == domain.WorkUnitStatusRunning {
		return "work_unit.started"
	}
	return "work_unit." + string(status)
}
