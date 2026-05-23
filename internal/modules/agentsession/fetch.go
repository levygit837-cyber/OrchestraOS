package agentsession

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// RequireByID retrieves an agent session by ID within a transaction, returning a not-found error if absent.
func RequireByID(ctx context.Context, tx *sql.Tx, id string) (*AgentSession, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	session, err := NewRepository(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "agentsession.get", err)
	}
	if session == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "agentsession.get", "agent session not found")
	}
	return session, nil
}
