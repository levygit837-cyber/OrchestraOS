package run

const (
	QueryInsert = `
		INSERT INTO runs (id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	QueryGetByID = `
		SELECT id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at
		FROM runs WHERE id = $1`

	QueryList = `
		SELECT id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at
		FROM runs ORDER BY created_at DESC`

	QueryListByTask = `
		SELECT id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at
		FROM runs WHERE task_id = $1 ORDER BY created_at DESC`

	QueryUpdateStatus = `
		UPDATE runs
		SET status = $2,
		    started_at = COALESCE(started_at, $3),
		    finished_at = COALESCE($4, finished_at),
		    result = COALESCE($5, result),
		    failure_reason = COALESCE($6, failure_reason),
		    updated_at = $7
		WHERE id = $1`
)
