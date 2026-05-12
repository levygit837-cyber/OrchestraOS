-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_reviews_one_active_per_gate
ON reviews(work_unit_id, gate_type)
WHERE status NOT IN ('approved', 'changes_requested', 'needs_discussion');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reviews_one_active_per_gate;
-- +goose StatementEnd
