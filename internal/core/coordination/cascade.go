package coordination

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

// CancelTaskDependents cancels all non-terminal runs and work units belonging to a task.
func CancelTaskDependents(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error {
	runRepo := runmod.NewRepository(tx)
	runs, err := runRepo.ListByTask(taskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "coordination.cancel_task_dependents.list_runs", err)
	}
	for _, run := range runs {
		if transition.IsFinalStatus(string(run.Status)) {
			continue
		}
		if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(domain.RunStatusCancelled), transition.TransitionContext(input)); err != nil {
			return err
		}
		if _, _, err := transition.AppendTransition(ctx, tx, "", "run.cancelled", taskID, run.ID, run.WorkUnitID, input.AgentID, transition.TransitionPayload(run.Status, domain.RunStatusCancelled, input)); err != nil {
			return err
		}
		if err := UpdateRunProjection(ctx, tx, run.ID, domain.RunStatusCancelled, runmod.ResultForStatus(domain.RunStatusCancelled), nil); err != nil {
			return err
		}
	}

	wuRepo := workunitmod.NewRepository(tx)
	workUnits, err := wuRepo.ListByTask(taskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "coordination.cancel_task_dependents.list_work_units", err)
	}
	for _, wu := range workUnits {
		if transition.IsFinalStatus(string(wu.Status)) {
			continue
		}
		if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(domain.WorkUnitStatusCancelled), transition.TransitionContext(input)); err != nil {
			return err
		}
		if _, _, err := transition.AppendTransition(ctx, tx, "", "work_unit.cancelled", taskID, "", wu.ID, input.AgentID, transition.TransitionPayload(wu.Status, domain.WorkUnitStatusCancelled, input)); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, workunitmod.QueryUpdateStatus, wu.ID, domain.WorkUnitStatusCancelled, time.Now().UTC())
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "coordination.cancel_task_dependents.update_work_unit", err)
		}
		if err := dbcore.EnsureRowsAffected(res, "work unit", "coordination.cancel_task_dependents.update_work_unit"); err != nil {
			return err
		}
	}
	return nil
}
