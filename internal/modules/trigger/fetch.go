package trigger

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// RequireByID retrieves a trigger by ID within a transaction, returning a not-found error if absent.
func RequireByID(ctx context.Context, tx *sql.Tx, id string) (*domain.Trigger, error) {
	_ = ctx
	t, err := NewRepository(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "trigger.get", err)
	}
	if t == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "trigger.get", "trigger not found")
	}
	return t, nil
}
