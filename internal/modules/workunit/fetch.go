package workunit

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// RequireByID retrieves a work unit by ID within a transaction, returning a not-found error if absent.
func RequireByID(ctx context.Context, tx *sql.Tx, id string) (*WorkUnit, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	wu, err := NewRepository(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "workunit.get", err)
	}
	if wu == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "workunit.get", "work unit not found")
	}
	return wu, nil
}
