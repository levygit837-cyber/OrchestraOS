-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(100) NOT NULL,
    version VARCHAR(20) NOT NULL,
    task_id UUID REFERENCES tasks(id) ON DELETE CASCADE,
    run_id UUID REFERENCES runs(id) ON DELETE CASCADE,
    work_unit_id UUID REFERENCES work_units(id) ON DELETE SET NULL,
    agent_id VARCHAR(255),
    trace_id VARCHAR(255),
    span_id VARCHAR(255),
    parent_span_id VARCHAR(255),
    sequence BIGINT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'background',
    requires_ack BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    payload JSONB NOT NULL
);

CREATE INDEX idx_events_task_id ON events(task_id);
CREATE INDEX idx_events_run_id ON events(run_id);
CREATE INDEX idx_events_work_unit_id ON events(work_unit_id);
CREATE INDEX idx_events_agent_id ON events(agent_id);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_sequence ON events(sequence);
CREATE INDEX idx_events_created_at ON events(created_at);

-- Sequence for event ordering
CREATE SEQUENCE events_sequence START 1;

-- +goose StatementEnd

-- +goose Down
DROP SEQUENCE IF EXISTS events_sequence;
DROP TABLE IF EXISTS events;
