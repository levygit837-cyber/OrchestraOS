-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_agents_unique_active_profile_runtime
ON agents(profile, runtime_type)
WHERE status = 'active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_agents_unique_active_profile_runtime;
-- +goose StatementEnd
