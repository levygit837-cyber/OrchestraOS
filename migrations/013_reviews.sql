-- +goose Up
-- +goose StatementBegin
CREATE TYPE review_status AS ENUM ('pending', 'in_progress', 'approved', 'changes_requested', 'needs_discussion');
CREATE TYPE validation_gate AS ENUM ('hard', 'soft', 'policy');

CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES runs(id) ON DELETE SET NULL,
    work_unit_id UUID REFERENCES work_units(id) ON DELETE SET NULL,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    agent_session_id UUID REFERENCES agent_sessions(id) ON DELETE SET NULL,
    reviewer_agent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    gate_type validation_gate NOT NULL,
    status review_status NOT NULL DEFAULT 'pending',
    verdict_reason TEXT,
    evidence_refs TEXT[],
    criteria_checked JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_reviews_run_id ON reviews(run_id);
CREATE INDEX idx_reviews_work_unit_id ON reviews(work_unit_id);
CREATE INDEX idx_reviews_task_id ON reviews(task_id);
CREATE INDEX idx_reviews_status ON reviews(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reviews_status;
DROP INDEX IF EXISTS idx_reviews_task_id;
DROP INDEX IF EXISTS idx_reviews_work_unit_id;
DROP INDEX IF EXISTS idx_reviews_run_id;
DROP TABLE IF EXISTS reviews;
DROP TYPE IF EXISTS validation_gate;
DROP TYPE IF EXISTS review_status;
-- +goose StatementEnd
