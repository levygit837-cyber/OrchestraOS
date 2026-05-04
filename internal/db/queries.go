package db

// Queries holds SQL queries used across the application
// Using prepared statements is recommended for production

const (
	// Task queries
	QueryTaskInsert = `
		INSERT INTO tasks (id, title, description, status, priority, risk_level, created_from_message_id, acceptance_criteria, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	QueryTaskGetByID = `
		SELECT id, title, description, status, priority, risk_level, created_from_message_id, acceptance_criteria, created_at, updated_at
		FROM tasks WHERE id = $1`

	QueryTaskList = `
		SELECT id, title, description, status, priority, risk_level, created_from_message_id, acceptance_criteria, created_at, updated_at
		FROM tasks ORDER BY created_at DESC`

	QueryTaskUpdate = `
		UPDATE tasks SET title = $2, description = $3, status = $4, priority = $5, risk_level = $6, 
		acceptance_criteria = $7, updated_at = $8 WHERE id = $1`

	// WorkUnit queries
	QueryWorkUnitInsert = `
		INSERT INTO work_units (id, task_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	QueryWorkUnitGetByID = `
		SELECT id, task_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at
		FROM work_units WHERE id = $1`

	QueryWorkUnitListByTask = `
		SELECT id, task_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at
		FROM work_units WHERE task_id = $1 ORDER BY created_at ASC`

	QueryWorkUnitUpdateStatus = `
		UPDATE work_units SET status = $2, updated_at = $3 WHERE id = $1`

	// Run queries
	QueryRunInsert = `
		INSERT INTO runs (id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	QueryRunGetByID = `
		SELECT id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at
		FROM runs WHERE id = $1`

	QueryRunList = `
		SELECT id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at
		FROM runs ORDER BY created_at DESC`

	QueryRunListByTask = `
		SELECT id, task_id, work_unit_id, status, attempt, started_at, finished_at, result, failure_reason, created_at, updated_at
		FROM runs WHERE task_id = $1 ORDER BY created_at DESC`

	QueryRunUpdateStatus = `
		UPDATE runs
		SET status = $2,
		    started_at = COALESCE(started_at, $3),
		    finished_at = COALESCE($4, finished_at),
		    result = COALESCE($5, result),
		    failure_reason = COALESCE($6, failure_reason),
		    updated_at = $7
		WHERE id = $1`

	// AgentSession queries
	QueryAgentSessionInsert = `
		INSERT INTO agent_sessions (id, agent_id, run_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`

	QueryAgentSessionGetByID = `
		SELECT id, agent_id, run_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at
		FROM agent_sessions WHERE id = $1`

	QueryAgentSessionGetByRunID = `
		SELECT id, agent_id, run_id, sandbox_id, connection_id, status, last_heartbeat_at, last_checkpoint_at, last_seen_event_id, recoverable_state, created_at, updated_at
		FROM agent_sessions WHERE run_id = $1 ORDER BY created_at DESC LIMIT 1`

	QueryAgentSessionUpdateStatus = `
		UPDATE agent_sessions
		SET status = $2,
		    last_heartbeat_at = COALESCE($3, last_heartbeat_at),
		    last_checkpoint_at = COALESCE($4, last_checkpoint_at),
		    updated_at = $5
		WHERE id = $1`

	// Event queries
	QueryEventInsert = `
		INSERT INTO events (id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO NOTHING`

	QueryEventGetByID = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE id = $1`

	QueryEventList = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events ORDER BY sequence ASC`

	QueryEventListByTask = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE task_id = $1 ORDER BY sequence ASC`

	QueryEventListByRun = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE run_id = $1 ORDER BY sequence ASC`

	QueryEventListByWorkUnit = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events WHERE work_unit_id = $1 ORDER BY sequence ASC`

	QueryEventLastCheckpointByRun = `
		SELECT id, type, version, task_id, run_id, work_unit_id, agent_id, trace_id, span_id, parent_span_id, sequence, priority, requires_ack, created_at, payload
		FROM events
		WHERE run_id = $1 AND type = 'agent.checkpoint_reached'
		ORDER BY sequence DESC
		LIMIT 1`

	QueryEventNextSequence = `SELECT nextval('events_sequence')`
)
