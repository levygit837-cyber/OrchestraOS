package orchestrator

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

// CancelTaskDependents cancels all non-terminal runs and work units belonging to a task.
// Located in the orchestrator module because it coordinates across run and workunit modules.
// task cannot import run (run already imports task for TaskReader), so this cross-module
// logic lives in the orchestrator which is the designated cross-module coordinator.
func CancelTaskDependents(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error {
	runRepo := runmod.NewRepository(tx)
	runs, err := runRepo.ListByTask(taskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestrator.cancel_task_dependents.list_runs", err)
	}
	for _, run := range runs {
		if transition.IsFinalStatus(string(run.Status)) {
			continue
		}
		if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(runmod.StatusCancelled), transition.TransitionContext(input)); err != nil {
			return err
		}
		if _, _, err := transition.AppendTransition(ctx, tx, "", "run.cancelled", taskID, run.ID, run.WorkUnitID, input.AgentID, transition.TransitionPayload(run.Status, runmod.StatusCancelled, input)); err != nil {
			return err
		}
		result := runmod.ResultForStatus(runmod.StatusCancelled)
		if err := runmod.UpdateRunProjection(ctx, tx, run.ID, runmod.StatusCancelled, result, nil); err != nil {
			return err
		}
	}

	wuRepo := workunitmod.NewRepository(tx)
	workUnits, err := wuRepo.ListByTask(taskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestrator.cancel_task_dependents.list_work_units", err)
	}
	for _, wu := range workUnits {
		if transition.IsFinalStatus(string(wu.Status)) {
			continue
		}
		if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(workunitmod.StatusCancelled), transition.TransitionContext(input)); err != nil {
			return err
		}
		if _, _, err := transition.AppendTransition(ctx, tx, "", "work_unit.cancelled", taskID, "", wu.ID, input.AgentID, transition.TransitionPayload(wu.Status, workunitmod.StatusCancelled, input)); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, workunitmod.QueryUpdateStatus, wu.ID, workunitmod.StatusCancelled, time.Now().UTC())
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "orchestrator.cancel_task_dependents.update_work_unit", err)
		}
		if err := dbcore.EnsureRowsAffected(res, "work unit", "orchestrator.cancel_task_dependents.update_work_unit"); err != nil {
			return err
		}
	}
	return nil
}
