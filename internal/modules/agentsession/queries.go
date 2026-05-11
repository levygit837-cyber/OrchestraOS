package agentsession

const (
	QueryInsert = `
		INSERT INTO agent_sessions (id, agent_id, run_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	QueryGetByID = `
		SELECT id, agent_id, run_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at
		FROM agent_sessions WHERE id = $1`

	QueryGetByRunID = `
		SELECT id, agent_id, run_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at
		FROM agent_sessions WHERE run_id = $1 ORDER BY created_at DESC LIMIT 1`

	QueryUpdateStatus = `
		UPDATE agent_sessions
		SET status = $2,
		    last_heartbeat_at = COALESCE($3, last_heartbeat_at),
		    last_checkpoint_at = COALESCE($4, last_checkpoint_at),
		    updated_at = $5
		WHERE id = $1`
)
