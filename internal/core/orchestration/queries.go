package orchestration

const (
	QueryRunUpdateStatus = `
		UPDATE runs
		SET status = $2,
		    started_at = COALESCE(started_at, $3),
		    finished_at = COALESCE($4, finished_at),
		    result = COALESCE($5, result),
		    failure_reason = COALESCE($6, failure_reason),
		    updated_at = $7
		WHERE id = $1`

	QueryWorkUnitUpdateStatus = `
		UPDATE work_units SET status = $2, updated_at = $3 WHERE id = $1`

	QueryAgentSessionUpdateStatus = `
		UPDATE agent_sessions
		SET status = $2,
		    last_heartbeat_at = COALESCE($3, last_heartbeat_at),
		    last_checkpoint_at = COALESCE($4, last_checkpoint_at),
		    updated_at = $5
		WHERE id = $1`
)
