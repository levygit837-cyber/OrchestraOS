package workunit

const (
	QueryInsert = `
		INSERT INTO work_units (id, task_id, task_graph_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`

	QueryGetByID = `
		SELECT id, task_id, task_graph_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at
		FROM work_units WHERE id = $1`

	QueryListByTask = `
		SELECT id, task_id, task_graph_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at
		FROM work_units WHERE task_id = $1 ORDER BY created_at ASC`

	QueryListByTaskGraph = `
		SELECT id, task_id, task_graph_id, title, objective, assigned_agent_profile, status, owned_paths, read_paths, acceptance_criteria, validation_plan, depends_on, created_at, updated_at
		FROM work_units WHERE task_graph_id = $1 ORDER BY created_at ASC`

	QueryUpdateStatus = `
		UPDATE work_units SET status = $2, updated_at = $3 WHERE id = $1`

	QueryUpdateAssignment = `
		UPDATE work_units SET assigned_agent_profile = $2, updated_at = $3 WHERE id = $1`
)
