package taskgraph

const (
	QueryInsert = `
		INSERT INTO task_graphs (id, task_id, version, status, planner_strategy, rationale, created_by, node_count, edge_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	QueryGetByID = `
		SELECT id, task_id, version, status, planner_strategy, rationale, created_by, node_count, edge_count, created_at, updated_at
		FROM task_graphs WHERE id = $1`

	QueryGetActiveByTask = `
		SELECT id, task_id, version, status, planner_strategy, rationale, created_by, node_count, edge_count, created_at, updated_at
		FROM task_graphs WHERE task_id = $1 AND status = 'active'
		ORDER BY version DESC LIMIT 1`

	QueryListByTask = `
		SELECT id, task_id, version, status, planner_strategy, rationale, created_by, node_count, edge_count, created_at, updated_at
		FROM task_graphs WHERE task_id = $1 ORDER BY version DESC`

	QueryGetNextVersion = `
		SELECT COALESCE(MAX(version), 0) + 1 FROM task_graphs WHERE task_id = $1`

	QueryUpdateActiveToSupersededByTask = `
		UPDATE task_graphs SET status = 'superseded', updated_at = $2 WHERE task_id = $1 AND status = 'active'`
)
