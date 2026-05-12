package agent

const (
	QueryInsert = `
		INSERT INTO agents (id, name, profile, capabilities, allowed_tools, default_prompt_fragments, runtime_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	QueryGetByID = `
		SELECT id, name, profile, capabilities, allowed_tools, default_prompt_fragments, runtime_type, status, created_at, updated_at
		FROM agents WHERE id = $1`

	QueryFindByProfileAndRuntime = `
		SELECT id, name, profile, capabilities, allowed_tools, default_prompt_fragments, runtime_type, status, created_at, updated_at
		FROM agents WHERE profile = $1 AND runtime_type = $2 AND status = 'active'
		ORDER BY created_at ASC LIMIT 1`

	QueryList = `
		SELECT id, name, profile, capabilities, allowed_tools, default_prompt_fragments, runtime_type, status, created_at, updated_at
		FROM agents ORDER BY created_at DESC`
)
