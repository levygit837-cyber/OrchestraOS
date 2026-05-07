-- +goose Up
-- +goose StatementBegin
-- Compatibility migration for databases that applied the initial M3 prompt
-- composition schema before canonical categories and snapshot deduplication.
ALTER TABLE prompt_fragments ADD COLUMN IF NOT EXISTS category VARCHAR(100);
ALTER TABLE prompt_fragments ADD COLUMN IF NOT EXISTS metadata_hash VARCHAR(80);

UPDATE prompt_fragments
SET category = COALESCE(NULLIF(category, ''), kind),
    metadata_hash = COALESCE(NULLIF(metadata_hash, ''), body_hash)
WHERE category IS NULL
   OR category = ''
   OR metadata_hash IS NULL
   OR metadata_hash = '';

ALTER TABLE prompt_fragments ALTER COLUMN category SET NOT NULL;
ALTER TABLE prompt_fragments ALTER COLUMN metadata_hash SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_prompt_fragments_category ON prompt_fragments(category);

ALTER TABLE prompt_snapshots ADD COLUMN IF NOT EXISTS composition_hash VARCHAR(80);
ALTER TABLE prompt_snapshots ADD COLUMN IF NOT EXISTS category_signature VARCHAR(80);
ALTER TABLE prompt_snapshots ADD COLUMN IF NOT EXISTS count_used INTEGER NOT NULL DEFAULT 1;
ALTER TABLE prompt_snapshots ADD COLUMN IF NOT EXISTS first_used_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE prompt_snapshots ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP WITH TIME ZONE;

UPDATE prompt_snapshots
SET composition_hash = COALESCE(NULLIF(composition_hash, ''), 'sha256:' || replace(id::text, '-', '') || replace(id::text, '-', '')),
    category_signature = COALESCE(NULLIF(category_signature, ''), 'sha256:' || replace(id::text, '-', '') || replace(id::text, '-', '')),
    first_used_at = COALESCE(first_used_at, created_at),
    last_used_at = COALESCE(last_used_at, created_at)
WHERE composition_hash IS NULL
   OR composition_hash = ''
   OR category_signature IS NULL
   OR category_signature = ''
   OR first_used_at IS NULL
   OR last_used_at IS NULL;

ALTER TABLE prompt_snapshots ALTER COLUMN composition_hash SET NOT NULL;
ALTER TABLE prompt_snapshots ALTER COLUMN category_signature SET NOT NULL;
ALTER TABLE prompt_snapshots ALTER COLUMN first_used_at SET NOT NULL;
ALTER TABLE prompt_snapshots ALTER COLUMN last_used_at SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'prompt_snapshots_count_used_positive'
    ) THEN
        ALTER TABLE prompt_snapshots
            ADD CONSTRAINT prompt_snapshots_count_used_positive CHECK (count_used > 0);
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_prompt_snapshots_composition_hash ON prompt_snapshots(composition_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_prompt_snapshots_composition_hash;
ALTER TABLE prompt_snapshots DROP CONSTRAINT IF EXISTS prompt_snapshots_count_used_positive;
ALTER TABLE prompt_snapshots DROP COLUMN IF EXISTS last_used_at;
ALTER TABLE prompt_snapshots DROP COLUMN IF EXISTS first_used_at;
ALTER TABLE prompt_snapshots DROP COLUMN IF EXISTS count_used;
ALTER TABLE prompt_snapshots DROP COLUMN IF EXISTS category_signature;
ALTER TABLE prompt_snapshots DROP COLUMN IF EXISTS composition_hash;
DROP INDEX IF EXISTS idx_prompt_fragments_category;
ALTER TABLE prompt_fragments DROP COLUMN IF EXISTS metadata_hash;
ALTER TABLE prompt_fragments DROP COLUMN IF EXISTS category;
-- +goose StatementEnd
