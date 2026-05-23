package eventstore

const (
	QueryInsert = `
		INSERT INTO events (id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO NOTHING`

	QueryGetByID = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE id = $1`

	QueryList = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events ORDER BY sequence ASC`

	QueryListByTask = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE task_id = $1 ORDER BY sequence ASC`

	QueryListByRun = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE run_id = $1 ORDER BY sequence ASC`

	QueryListByWorkUnit = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE work_unit_id = $1 ORDER BY sequence ASC`

	QueryGetLastCheckpointByRun = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events
		WHERE run_id = $1 AND type = 'agent.checkpoint_reached'
		ORDER BY sequence DESC
		LIMIT 1`

	QueryNextSequence = `SELECT nextval('events_sequence')`
)
