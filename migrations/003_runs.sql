-- +goose Up
-- +goose StatementBegin
CREATE TABLE runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    work_unit_id UUID REFERENCES work_units(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    attempt INTEGER NOT NULL DEFAULT 1,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    result VARCHAR(50),
    failure_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_runs_task_id ON runs(task_id);
CREATE INDEX idx_runs_work_unit_id ON runs(work_unit_id);
CREATE INDEX idx_runs_status ON runs(status);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS runs;
