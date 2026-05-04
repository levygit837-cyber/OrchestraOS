-- +goose Up
-- +goose StatementBegin
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    priority VARCHAR(10) NOT NULL DEFAULT 'P2',
    risk_level VARCHAR(20) NOT NULL DEFAULT 'low',
    created_from_message_id VARCHAR(255),
    acceptance_criteria JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS tasks;
