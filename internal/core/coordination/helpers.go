package coordination

import (
	"context"
	"database/sql"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
)

// UpdateRunProjection updates the runs table projection.
// TODO: move to run module once cross-module calls are refactored.
func UpdateRunProjection(ctx context.Context, tx *sql.Tx, runID string, status runmod.Status, result *runmod.Result, failureReason *string) error {
	now := time.Now().UTC()
	var startedAt, finishedAt *time.Time
	if status == runmod.StatusRunning {
		startedAt = &now
	}
	if status == runmod.StatusCompleted || status == runmod.StatusFailed || status == runmod.StatusCancelled {
		finishedAt = &now
	}

	var resultStr *string
	if result != nil {
		r := string(*result)
		resultStr = &r
	}

	res, err := tx.ExecContext(ctx, QueryRunUpdateStatus, runID, status, startedAt, finishedAt, resultStr, failureReason, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "coordination.update_run_projection", err)
	}
	return dbcore.EnsureRowsAffected(res, "run", "coordination.update_run_projection")
}
