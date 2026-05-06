-- +goose Up
-- +goose StatementBegin
CREATE TABLE task_graphs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    planner_strategy VARCHAR(100) NOT NULL,
    rationale TEXT,
    created_by VARCHAR(255),
    node_count INTEGER NOT NULL DEFAULT 0,
    edge_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT task_graphs_version_positive CHECK (version > 0),
    CONSTRAINT task_graphs_node_count_non_negative CHECK (node_count >= 0),
    CONSTRAINT task_graphs_edge_count_non_negative CHECK (edge_count >= 0),
    CONSTRAINT task_graphs_status_valid CHECK (status IN ('active', 'superseded'))
);

CREATE UNIQUE INDEX idx_task_graphs_task_version ON task_graphs(task_id, version);
CREATE UNIQUE INDEX idx_task_graphs_one_active_per_task ON task_graphs(task_id) WHERE status = 'active';
CREATE INDEX idx_task_graphs_task_id ON task_graphs(task_id);
CREATE INDEX idx_task_graphs_status ON task_graphs(status);

ALTER TABLE work_units ADD COLUMN task_graph_id UUID;

INSERT INTO task_graphs (id, task_id, version, status, planner_strategy, rationale, created_by, node_count, edge_count, created_at, updated_at)
SELECT
    gen_random_uuid(),
    task_id,
    1,
    'active',
    'legacy_manual',
    'Legacy graph created for work units that existed before task graph persistence.',
    'migration:007',
    COUNT(*),
    0,
    NOW(),
    NOW()
FROM work_units
GROUP BY task_id;

UPDATE work_units
SET task_graph_id = task_graphs.id
FROM task_graphs
WHERE work_units.task_id = task_graphs.task_id;

ALTER TABLE work_units ALTER COLUMN task_graph_id SET NOT NULL;
ALTER TABLE work_units ADD CONSTRAINT fk_work_units_task_graph_id FOREIGN KEY (task_graph_id) REFERENCES task_graphs(id) ON DELETE CASCADE;
CREATE INDEX idx_work_units_task_graph_id ON work_units(task_graph_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_work_units_task_graph_id;
ALTER TABLE work_units DROP CONSTRAINT IF EXISTS fk_work_units_task_graph_id;
ALTER TABLE work_units DROP COLUMN IF EXISTS task_graph_id;
DROP TABLE IF EXISTS task_graphs;
-- +goose StatementEnd
