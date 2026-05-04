-- +goose Up
-- +goose StatementBegin
ALTER TABLE agent_sessions
    ADD COLUMN last_seen_event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    ADD COLUMN recoverable_state JSONB;

CREATE INDEX idx_agent_sessions_last_seen_event_id ON agent_sessions(last_seen_event_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE agent_sessions
    DROP COLUMN IF EXISTS recoverable_state,
    DROP COLUMN IF EXISTS last_seen_event_id;
-- +goose StatementEnd
