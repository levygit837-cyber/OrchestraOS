-- +goose Up
-- +goose StatementBegin
-- Compatibility migration for local databases that applied the first M3 prompt
-- snapshot table before combined prompt text was made explicit.
ALTER TABLE prompt_snapshots ADD COLUMN IF NOT EXISTS combined_prompt TEXT;

UPDATE prompt_snapshots
SET combined_prompt = system_prompt || E'\n\n--- TASK PROMPT ---\n\n' || task_prompt
WHERE combined_prompt IS NULL OR combined_prompt = '';

ALTER TABLE prompt_snapshots ALTER COLUMN combined_prompt SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No-op: combined_prompt is part of the canonical 008 prompt_snapshots
-- table definition for fresh databases.
SELECT 1;
-- +goose StatementEnd
