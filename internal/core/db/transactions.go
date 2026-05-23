package db

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// BeginTx starts a new transaction.
func BeginTx(ctx context.Context, database *sql.DB, op string) (*sql.Tx, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return tx, nil
}

// CommitTx commits a transaction.
func CommitTx(tx *sql.Tx, op string) error {
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return nil
}

// RollbackTx rolls back a transaction. It is safe to call on a committed or rolled-back tx.
func RollbackTx(tx *sql.Tx) {
	_ = tx.Rollback() // ignore: Rollback on committed/rolled-back tx is always safe
}

// EnsureRowsAffected checks that a SQL execution affected at least one row.
func EnsureRowsAffected(result sql.Result, entity, op string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if rows == 0 {
		return apperrors.New(apperrors.CodeConflict, op, fmt.Sprintf("%s projection was not updated", entity))
	}
	return nil
}

// AcquireAdvisoryTxLock acquires a PostgreSQL advisory transaction-level lock.
func AcquireAdvisoryTxLock(ctx context.Context, tx *sql.Tx, key, op string) error {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(key))
	lockID := int64(hasher.Sum64())
	if _, err := tx.ExecContext(ctx, QueryAdvisoryLock, lockID); err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return nil
}
