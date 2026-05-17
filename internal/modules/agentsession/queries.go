package agentsession

const (
	QueryInsert = `
		INSERT INTO agent_sessions (id, agent_id, run_id, task_id, work_unit_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	QueryGetByID = `
		SELECT id, agent_id, run_id, task_id, work_unit_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at
		FROM agent_sessions WHERE id = $1`

	QueryGetByRunID = `
		SELECT id, agent_id, run_id, task_id, work_unit_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at
		FROM agent_sessions WHERE run_id = $1 ORDER BY created_at DESC LIMIT 1`

	QueryUpdateStatus = `
		UPDATE agent_sessions
		SET status = $2,
		    last_heartbeat_at = COALESCE($3, last_heartbeat_at),
		    last_checkpoint_at = COALESCE($4, last_checkpoint_at),
		    updated_at = $5
		WHERE id = $1`

	QueryUpdateHeartbeat = `
		UPDATE agent_sessions SET last_heartbeat_at = $2, updated_at = $3 WHERE id = $1`

	QueryUpdateHeartbeatWithEvent = `
		UPDATE agent_sessions SET last_heartbeat_at = $2, last_seen_event_id = COALESCE($3, last_seen_event_id), updated_at = $4 WHERE id = $1`

	QueryUpdateCheckpoint = `
		UPDATE agent_sessions SET last_checkpoint_at = $2, updated_at = $3 WHERE id = $1`

	QueryUpdateCheckpointWithEvent = `
		UPDATE agent_sessions SET last_checkpoint_at = $2, last_seen_event_id = COALESCE($3, last_seen_event_id), updated_at = $4 WHERE id = $1`

	QueryUpdateRecoverableState = `
		UPDATE agent_sessions SET recoverable_state = $2, updated_at = $3 WHERE id = $1`

	QueryUpdateConnection = `
		UPDATE agent_sessions SET connection_id = $2, sandbox_id = $3, updated_at = $4 WHERE id = $1`
)
