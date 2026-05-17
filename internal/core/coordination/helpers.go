package coordination

import (
	"context"
	"database/sql"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
)

// runToDomain converts a local run.Run to domain.Run for cross-module compatibility.
// TODO[ADR-0022]: remover quando todos os consumidores usarem *run.Run diretamente.
func runToDomain(r *runmod.Run) *domain.Run {
	if r == nil {
		return nil
	}
	var result *domain.RunResult
	if r.Result != nil {
		rr := domain.RunResult(*r.Result)
		result = &rr
	}
	return &domain.Run{
		ID:            r.ID,
		TaskID:        r.TaskID,
		WorkUnitID:    r.WorkUnitID,
		Status:        domain.RunStatus(r.Status),
		Attempt:       r.Attempt,
		StartedAt:     r.StartedAt,
		FinishedAt:    r.FinishedAt,
		Result:        result,
		FailureReason: r.FailureReason,
	}
}

// UpdateRunProjection updates the runs table projection.
// TODO: move to run module once cross-module calls are refactored.
// TODO[ADR-0022]: migrar para run.Status e run.Result quando run module for totalmente desacoplado.
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
		return apperrors.Wrap(apperrors.CodePersistence, "coordination.update_run_projection", err)
	}
	return dbcore.EnsureRowsAffected(res, "run", "coordination.update_run_projection")
}
