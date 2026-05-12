package review

const (
	QueryInsert = `
		INSERT INTO reviews (id, run_id, work_unit_id, task_id, agent_session_id, reviewer_agent_id, gate_type, status, verdict_reason, evidence_refs, criteria_checked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	QueryGetByID = `
		SELECT id, run_id, work_unit_id, task_id, agent_session_id, reviewer_agent_id, gate_type, status, verdict_reason, evidence_refs, criteria_checked, created_at, updated_at, completed_at
		FROM reviews WHERE id = $1`

	QueryListByTask = `
		SELECT id, run_id, work_unit_id, task_id, agent_session_id, reviewer_agent_id, gate_type, status, verdict_reason, evidence_refs, criteria_checked, created_at, updated_at, completed_at
		FROM reviews WHERE task_id = $1 ORDER BY created_at ASC`

	QueryListPending = `
		SELECT id, run_id, work_unit_id, task_id, agent_session_id, reviewer_agent_id, gate_type, status, verdict_reason, evidence_refs, criteria_checked, created_at, updated_at, completed_at
		FROM reviews WHERE status IN ('pending', 'in_progress') ORDER BY created_at ASC`

	QueryUpdateStatus = `
		UPDATE reviews SET status = $2, updated_at = $3, completed_at = $4, verdict_reason = $5, evidence_refs = $6, criteria_checked = $7
		WHERE id = $1`

	QueryExistsActiveByWorkUnitAndGate = `
		SELECT EXISTS(
			SELECT 1 FROM reviews
			WHERE work_unit_id = $1 AND gate_type = $2
			AND status NOT IN ('approved', 'changes_requested', 'needs_discussion')
		)`
)
