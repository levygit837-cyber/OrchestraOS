package coordination

import (
	"context"
	"database/sql"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// TransitionRunWithWorkUnit synchronizes a run transition with its associated work unit.
func TransitionRunWithWorkUnit(ctx context.Context, tx *sql.Tx, run *runmod.Run, target runmod.Status, input transition.TransitionInput) error {
	if run.WorkUnitID == "" {
		return nil
	}
	wu, err := workunitmod.RequireByID(ctx, tx, run.WorkUnitID)
	if err != nil {
		return err
	}
	var wuTarget workunitmod.Status
	switch target {
	case runmod.StatusRunning:
		wuTarget = workunitmod.StatusRunning
	case runmod.StatusValidating:
		wuTarget = workunitmod.StatusValidating
	case runmod.StatusCompleted:
		wuTarget = workunitmod.StatusCompleted
	case runmod.StatusFailed:
		wuTarget = workunitmod.StatusFailed
	case runmod.StatusCancelled:
		wuTarget = workunitmod.StatusCancelled
	default:
		return nil
	}
	if wu.Status == wuTarget {
		return nil
	}
	if wuTarget == workunitmod.StatusRunning {
		if err := dbcore.AcquireAdvisoryTxLock(ctx, tx, "work_unit_paths:"+wu.TaskID, "coordination.work_unit_path_lock"); err != nil {
			return err
		}
		if err := workunitmod.ValidateDependenciesCompleted(ctx, tx, wu); err != nil {
			return err
		}
		if err := workunitmod.ValidateOwnedPathAvailability(ctx, tx, wu); err != nil {
			return err
		}
	}
	if wuTarget == workunitmod.StatusCompleted && len(wu.AcceptanceCriteria) == 0 && input.Justification == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "coordination.run_workunit_sync", "related work unit completion requires acceptance criteria or explicit justification")
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(wuTarget), transition.TransitionContext(input)); err != nil {
		if wuTarget == workunitmod.StatusFailed && wu.Status == workunitmod.StatusCreated {
			return nil
		}
		return err
	}
	if _, _, err := transition.AppendTransition(ctx, tx, "", workUnitEventTypeForStatus(wuTarget), run.TaskID, run.ID, wu.ID, input.AgentID, transition.TransitionPayload(wu.Status, wuTarget, input)); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, workunitmod.QueryUpdateStatus, wu.ID, wuTarget, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "coordination.run_workunit_sync.update_work_unit", err)
	}
	return dbcore.EnsureRowsAffected(res, "work unit", "coordination.run_workunit_sync.update_work_unit")
}

func workUnitEventTypeForStatus(status workunitmod.Status) string {
	if status == workunitmod.StatusRunning {
		return "work_unit.started"
	}
	return "work_unit." + string(status)
}
