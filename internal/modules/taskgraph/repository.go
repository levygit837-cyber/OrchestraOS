// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package taskgraph

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(graph *domain.TaskGraph) error {
	if graph.ID == "" {
		graph.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if graph.CreatedAt.IsZero() {
		graph.CreatedAt = now
	}
	if graph.UpdatedAt.IsZero() {
		graph.UpdatedAt = graph.CreatedAt
	}
	_, err := r.db.Exec(
		QueryInsert,
		graph.ID,
		graph.TaskID,
		graph.Version,
		graph.Status,
		graph.PlannerStrategy,
		graph.Rationale,
		graph.CreatedBy,
		graph.NodeCount,
		graph.EdgeCount,
		graph.CreatedAt,
		graph.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task graph: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(id string) (*domain.TaskGraph, error) {
	row := r.db.QueryRow(QueryGetByID, id)
	return r.scanTaskGraph(row)
}

func (r *Repository) GetActiveByTask(taskID string) (*domain.TaskGraph, error) {
	row := r.db.QueryRow(QueryGetActiveByTask, taskID)
	return r.scanTaskGraph(row)
}

func (r *Repository) ListByTask(taskID string) ([]domain.TaskGraph, error) {
	rows, err := r.db.Query(QueryListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list task graphs: %w", err)
	}
	defer rows.Close()

	var graphs []domain.TaskGraph
	for rows.Next() {
		graph, err := r.scanTaskGraph(rows)
		if err != nil {
			return nil, err
		}
		graphs = append(graphs, *graph)
	}
	return graphs, rows.Err()
}

func (r *Repository) NextVersion(taskID string) (int, error) {
	var version int
	if err := r.db.QueryRow(QueryNextVersion, taskID).Scan(&version); err != nil {
		return 0, fmt.Errorf("failed to get next task graph version: %w", err)
	}
	return version, nil
}

func (r *Repository) SupersedeActiveByTask(taskID string, updatedAt time.Time) error {
	_, err := r.db.Exec(QuerySupersedeActiveByTask, taskID, updatedAt)
	if err != nil {
		return fmt.Errorf("failed to supersede active task graph: %w", err)
	}
	return nil
}

func (r *Repository) scanTaskGraph(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.TaskGraph, error) {
	var graph domain.TaskGraph
	var rationale, createdBy sql.NullString
	err := scanner.Scan(
		&graph.ID,
		&graph.TaskID,
		&graph.Version,
		&graph.Status,
		&graph.PlannerStrategy,
		&rationale,
		&createdBy,
		&graph.NodeCount,
		&graph.EdgeCount,
		&graph.CreatedAt,
		&graph.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan task graph: %w", err)
	}
	if rationale.Valid {
		graph.Rationale = rationale.String
	}
	if createdBy.Valid {
		graph.CreatedBy = createdBy.String
	}
	return &graph, nil
}
