-- +goose Up
-- +goose StatementBegin
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    profile VARCHAR(100) NOT NULL,
    capabilities TEXT[],
    allowed_tools TEXT[],
    default_prompt_fragments TEXT[],
    runtime_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT agents_runtime_type_check CHECK (runtime_type IN ('fake', 'gemini', 'codex_cli', 'external')),
    CONSTRAINT agents_profile_check CHECK (profile IN ('code_worker', 'docs_writer', 'reviewer', 'debugger', 'default'))
);

CREATE INDEX idx_agents_profile ON agents(profile);
CREATE INDEX idx_agents_runtime_type ON agents(runtime_type);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_profile_runtime ON agents(profile, runtime_type) WHERE status = 'active';

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS agents;
