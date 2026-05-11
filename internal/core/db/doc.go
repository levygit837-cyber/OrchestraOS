// Package db provides low-level database infrastructure helpers.
//
// # Responsibility
// Transaction management, row-affect checks, and advisory locking.
// This package contains NO domain logic — only reusable SQL helpers.
//
// # Key Types
//   - DBTX: interface abstracting *sql.DB and *sql.Tx
//   - BeginTx, CommitTx, RollbackTx: transaction lifecycle
//   - EnsureRowsAffected: verifies that an UPDATE/DELETE touched at least one row
//   - AcquireAdvisoryTxLock: PostgreSQL advisory transaction-level lock
//
// # Dependencies
//   - core/apperrors: error wrapping
//
// # Related Packages
//   - All modules/: use BeginTx/EnsureRowsAffected in repository and service code
package db
