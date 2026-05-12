package trigger

const (
	QueryInsert = `
		INSERT INTO triggers (id, run_id, task_id, agent_session_id, trigger_type, status, anomaly_type, threshold_value, current_value, triggered_at, resolved_at, resolution_action, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	QueryGetByID = `
		SELECT id, run_id, task_id, agent_session_id, trigger_type, status, anomaly_type, threshold_value, current_value, triggered_at, resolved_at, resolution_action, created_at
		FROM triggers WHERE id = $1`

	QueryListActive = `
		SELECT id, run_id, task_id, agent_session_id, trigger_type, status, anomaly_type, threshold_value, current_value, triggered_at, resolved_at, resolution_action, created_at
		FROM triggers WHERE status IN ('active', 'triggered') ORDER BY created_at DESC`

	QueryListByRun = `
		SELECT id, run_id, task_id, agent_session_id, trigger_type, status, anomaly_type, threshold_value, current_value, triggered_at, resolved_at, resolution_action, created_at
		FROM triggers WHERE run_id = $1 ORDER BY created_at DESC`

	QueryUpdateStatus = `
		UPDATE triggers
		SET status = $2,
		    triggered_at = COALESCE($3, triggered_at),
		    resolved_at = COALESCE($4, resolved_at),
		    resolution_action = COALESCE($5, resolution_action)
		WHERE id = $1`
)
