-- +goose Up
-- +goose StatementBegin
CREATE TABLE agent_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    run_id UUID NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
    sandbox_id VARCHAR(255),
    connection_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'starting',
    last_heartbeat_at TIMESTAMP WITH TIME ZONE,
    last_checkpoint_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_sessions_run_id ON agent_sessions(run_id);
CREATE INDEX idx_agent_sessions_status ON agent_sessions(status);
CREATE INDEX idx_agent_sessions_agent_id ON agent_sessions(agent_id);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS agent_sessions;
