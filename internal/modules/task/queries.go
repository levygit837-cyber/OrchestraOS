package task

const (
	QueryInsert = `
		INSERT INTO tasks (id, title, description, status, priority, risk_level, created_from_message_id, acceptance_criteria, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	QueryGetByID = `
		SELECT id, title, description, status, priority, risk_level, created_from_message_id, acceptance_criteria, created_at, updated_at
		FROM tasks WHERE id = $1`

	QueryList = `
		SELECT id, title, description, status, priority, risk_level, created_from_message_id, acceptance_criteria, created_at, updated_at
		FROM tasks ORDER BY created_at DESC`

	QueryUpdate = `
		UPDATE tasks SET title = $2, description = $3, status = $4, priority = $5, risk_level = $6,
		acceptance_criteria = $7, updated_at = $8 WHERE id = $1`
)
