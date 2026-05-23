package db

// QueryAdvisoryLock is the SQL for acquiring a PostgreSQL advisory transaction-level lock.
const QueryAdvisoryLock = `SELECT pg_advisory_xact_lock($1)`
