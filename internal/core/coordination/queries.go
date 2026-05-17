package coordination

const (
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
