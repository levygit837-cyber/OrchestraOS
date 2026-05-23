package task

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// RequireByID retrieves a task by ID within a transaction, returning a not-found error if absent.
func RequireByID(ctx context.Context, tx *sql.Tx, id string) (*Task, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	task, err := NewRepository(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task.get", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "task.get", "task not found")
	}
	return task, nil
}
