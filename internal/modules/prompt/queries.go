package prompt

const (
	QueryFragmentInsert = `
		INSERT INTO prompt_fragments (id, version, category, kind, title, priority, exclusive_group, body_hash, metadata_hash, body, applies_when, requires, conflicts_with, allows, denies, approval_required, autonomy_level, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	QueryFragmentGetByIDVersion = `
		SELECT id, version, category, kind, title, priority, exclusive_group, body_hash, metadata_hash, body, applies_when, requires, conflicts_with, allows, denies, approval_required, autonomy_level, created_at, updated_at
		FROM prompt_fragments WHERE id = $1 AND version = $2`

	QuerySnapshotInsert = `
		INSERT INTO prompt_snapshots (id, run_id, work_unit_id, agent_session_id, system_prompt, task_prompt, combined_prompt, system_prompt_hash, task_prompt_hash, combined_prompt_hash, composition_hash, category_signature, fragment_refs, assembly_order, variables_applied, count_used, first_used_at, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 1, $16, $16, $16)
		ON CONFLICT (composition_hash) DO UPDATE
		SET count_used = prompt_snapshots.count_used + 1,
		    last_used_at = EXCLUDED.last_used_at
		RETURNING id, run_id, work_unit_id, agent_session_id, system_prompt, task_prompt, combined_prompt, system_prompt_hash, task_prompt_hash, combined_prompt_hash, composition_hash, category_signature, fragment_refs, assembly_order, variables_applied, count_used, first_used_at, last_used_at, created_at`

	QuerySnapshotGetByID = `
		SELECT id, run_id, work_unit_id, agent_session_id, system_prompt, task_prompt, combined_prompt, system_prompt_hash, task_prompt_hash, combined_prompt_hash, composition_hash, category_signature, fragment_refs, assembly_order, variables_applied, count_used, first_used_at, last_used_at, created_at
		FROM prompt_snapshots WHERE id = $1`

	QuerySnapshotLatestByRun = `
		SELECT id, run_id, work_unit_id, agent_session_id, system_prompt, task_prompt, combined_prompt, system_prompt_hash, task_prompt_hash, combined_prompt_hash, composition_hash, category_signature, fragment_refs, assembly_order, variables_applied, count_used, first_used_at, last_used_at, created_at
		FROM prompt_snapshots WHERE run_id = $1 ORDER BY created_at DESC LIMIT 1`

	QueryToolsetInsert = `
		INSERT INTO toolset_snapshots (id, run_id, agent_session_id, tools, created_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	QueryToolsetGetByID = `
		SELECT id, run_id, agent_session_id, tools, created_reason, created_at
		FROM toolset_snapshots WHERE id = $1`

	QueryToolsetLatestByAgentSession = `
		SELECT id, run_id, agent_session_id, tools, created_reason, created_at
		FROM toolset_snapshots WHERE agent_session_id = $1 ORDER BY created_at DESC LIMIT 1`
)
