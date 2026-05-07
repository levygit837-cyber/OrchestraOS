-- +goose Up
-- +goose StatementBegin
CREATE TABLE prompt_fragments (
    id VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    kind VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    exclusive_group VARCHAR(255) NOT NULL DEFAULT '',
    body_hash VARCHAR(80) NOT NULL,
    metadata_hash VARCHAR(80) NOT NULL,
    body TEXT NOT NULL,
    applies_when JSONB NOT NULL DEFAULT '{}'::jsonb,
    requires JSONB NOT NULL DEFAULT '[]'::jsonb,
    conflicts_with JSONB NOT NULL DEFAULT '[]'::jsonb,
    allows JSONB NOT NULL DEFAULT '[]'::jsonb,
    denies JSONB NOT NULL DEFAULT '[]'::jsonb,
    approval_required JSONB NOT NULL DEFAULT '[]'::jsonb,
    autonomy_level INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, version),
    CONSTRAINT prompt_fragments_priority_non_negative CHECK (priority >= 0),
    CONSTRAINT prompt_fragments_autonomy_non_negative CHECK (autonomy_level >= 0),
    CONSTRAINT prompt_fragments_applies_when_object CHECK (jsonb_typeof(applies_when) = 'object'),
    CONSTRAINT prompt_fragments_requires_array CHECK (jsonb_typeof(requires) = 'array'),
    CONSTRAINT prompt_fragments_conflicts_array CHECK (jsonb_typeof(conflicts_with) = 'array'),
    CONSTRAINT prompt_fragments_allows_array CHECK (jsonb_typeof(allows) = 'array'),
    CONSTRAINT prompt_fragments_denies_array CHECK (jsonb_typeof(denies) = 'array'),
    CONSTRAINT prompt_fragments_approval_array CHECK (jsonb_typeof(approval_required) = 'array')
);

CREATE INDEX idx_prompt_fragments_kind ON prompt_fragments(kind);
CREATE INDEX idx_prompt_fragments_category ON prompt_fragments(category);
CREATE INDEX idx_prompt_fragments_exclusive_group ON prompt_fragments(exclusive_group);

CREATE TABLE prompt_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    work_unit_id UUID NOT NULL REFERENCES work_units(id) ON DELETE CASCADE,
    agent_session_id UUID NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    system_prompt TEXT NOT NULL,
    task_prompt TEXT NOT NULL,
    combined_prompt TEXT NOT NULL,
    system_prompt_hash VARCHAR(80) NOT NULL,
    task_prompt_hash VARCHAR(80) NOT NULL,
    combined_prompt_hash VARCHAR(80) NOT NULL,
    composition_hash VARCHAR(80) NOT NULL,
    category_signature VARCHAR(80) NOT NULL,
    fragment_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    assembly_order JSONB NOT NULL DEFAULT '[]'::jsonb,
    variables_applied JSONB NOT NULL DEFAULT '{}'::jsonb,
    count_used INTEGER NOT NULL DEFAULT 1,
    first_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT prompt_snapshots_fragment_refs_array CHECK (jsonb_typeof(fragment_refs) = 'array'),
    CONSTRAINT prompt_snapshots_assembly_order_array CHECK (jsonb_typeof(assembly_order) = 'array'),
    CONSTRAINT prompt_snapshots_variables_object CHECK (jsonb_typeof(variables_applied) = 'object'),
    CONSTRAINT prompt_snapshots_count_used_positive CHECK (count_used > 0)
);

CREATE INDEX idx_prompt_snapshots_run_id ON prompt_snapshots(run_id);
CREATE INDEX idx_prompt_snapshots_work_unit_id ON prompt_snapshots(work_unit_id);
CREATE INDEX idx_prompt_snapshots_agent_session_id ON prompt_snapshots(agent_session_id);
CREATE INDEX idx_prompt_snapshots_combined_hash ON prompt_snapshots(combined_prompt_hash);
CREATE UNIQUE INDEX idx_prompt_snapshots_composition_hash ON prompt_snapshots(composition_hash);

CREATE TABLE toolset_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    agent_session_id UUID NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    tools JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT toolset_snapshots_tools_array CHECK (jsonb_typeof(tools) = 'array')
);

CREATE INDEX idx_toolset_snapshots_run_id ON toolset_snapshots(run_id);
CREATE INDEX idx_toolset_snapshots_agent_session_id ON toolset_snapshots(agent_session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS toolset_snapshots;
DROP TABLE IF EXISTS prompt_snapshots;
DROP TABLE IF EXISTS prompt_fragments;
-- +goose StatementEnd
