-- +goose Up
-- +goose StatementBegin
CREATE TABLE work_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    objective TEXT,
    assigned_agent_profile VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    owned_paths JSONB,
    read_paths JSONB,
    acceptance_criteria JSONB,
    validation_plan JSONB,
    depends_on JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_work_units_task_id ON work_units(task_id);
CREATE INDEX idx_work_units_status ON work_units(status);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS work_units;
