-- +goose Up
-- +goose StatementBegin
ALTER TABLE agent_sessions
ADD COLUMN task_id UUID REFERENCES tasks(id),
ADD COLUMN work_unit_id UUID REFERENCES work_units(id);

CREATE INDEX idx_agent_sessions_task_id ON agent_sessions(task_id);
CREATE INDEX idx_agent_sessions_work_unit_id ON agent_sessions(work_unit_id);
-- +goose StatementEnd

-- +goose Down
ALTER TABLE agent_sessions DROP COLUMN IF EXISTS task_id;
ALTER TABLE agent_sessions DROP COLUMN IF EXISTS work_unit_id;
