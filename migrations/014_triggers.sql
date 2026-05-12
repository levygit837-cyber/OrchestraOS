-- +goose Up
-- +goose StatementBegin

CREATE TABLE triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES runs(id) ON DELETE SET NULL,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    agent_session_id UUID REFERENCES agent_sessions(id) ON DELETE SET NULL,
    trigger_type VARCHAR(32) NOT NULL CHECK (trigger_type IN ('threshold', 'anomaly', 'heartbeat_timeout', 'policy')),
    status VARCHAR(32) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'triggered', 'resolved', 'dismissed')),
    anomaly_type VARCHAR(32) CHECK (anomaly_type IN ('stall', 'loop', 'drift', 'path_violation', 'token_exceeded', 'steps_exceeded', 'time_exceeded')),
    threshold_value JSONB,
    current_value JSONB,
    triggered_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    resolution_action VARCHAR(32) CHECK (resolution_action IN ('pause', 'cancel', 'notify', 'escalate')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_triggers_run_id ON triggers(run_id);
CREATE INDEX idx_triggers_task_id ON triggers(task_id);
CREATE INDEX idx_triggers_status ON triggers(status);
CREATE INDEX idx_triggers_anomaly_type ON triggers(anomaly_type);
CREATE INDEX idx_triggers_trigger_type ON triggers(trigger_type);
CREATE INDEX idx_triggers_agent_session_id ON triggers(agent_session_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS triggers;

-- +goose StatementEnd
