package orchestration

import (
	"context"
	"database/sql"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// UpdateRunProjection updates the runs table projection.
// TODO: move to run module once cross-module calls are refactored.
func UpdateRunProjection(ctx context.Context, tx *sql.Tx, runID string, status domain.RunStatus, result *domain.RunResult, failureReason *string) error {
	now := time.Now().UTC()
	var startedAt, finishedAt *time.Time
	if status == domain.RunStatusRunning {
		startedAt = &now
	}
	if status == domain.RunStatusCompleted || status == domain.RunStatusFailed || status == domain.RunStatusCancelled {
		finishedAt = &now
	}

	var resultStr *string
	if result != nil {
		r := string(*result)
		resultStr = &r
	}

	res, err := tx.ExecContext(ctx, QueryRunUpdateStatus, runID, status, startedAt, finishedAt, resultStr, failureReason, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.update_run_projection", err)
	}
	return dbcore.EnsureRowsAffected(res, "run", "orchestration.update_run_projection")
}
