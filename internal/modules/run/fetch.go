package run

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// RequireByID retrieves a run by ID within a transaction, returning a not-found error if absent.
func RequireByID(ctx context.Context, tx *sql.Tx, id string) (*Run, error) {
	_ = ctx
	r, err := NewRepository(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "run.get", err)
	}
	if r == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "run.get", "run not found")
	}
	return r, nil
}
