// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package run

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Repository handles run persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new run repository
func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Create inserts a new run
func (r *Repository) Create(run *domain.Run) error {
	if run.ID == "" {
		run.ID = uuid.New().String()
	}
	if run.Attempt == 0 {
		run.Attempt = 1
	}

	now := time.Now()

	var result *string
	if run.Result != nil {
		r := string(*run.Result)
		result = &r
	}
	var startedAt *time.Time
	if !run.StartedAt.IsZero() {
		startedAt = &run.StartedAt
	}

	_, err := r.db.Exec(
		QueryInsert,
		run.ID,
		run.TaskID,
		run.WorkUnitID,
		run.Status,
		run.Attempt,
		startedAt,
		run.FinishedAt,
		result,
		run.FailureReason,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	return nil
}

// GetByID retrieves a run by ID
func (r *Repository) GetByID(id string) (*domain.Run, error) {
	row := r.db.QueryRow(QueryGetByID, id)
	return r.scanRun(row)
}

// List retrieves all runs
func (r *Repository) List() ([]domain.Run, error) {
	rows, err := r.db.Query(QueryList)
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.Run
	for rows.Next() {
		run, err := r.scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}

	return runs, rows.Err()
}

// ListByTask retrieves all runs for a task
func (r *Repository) ListByTask(taskID string) ([]domain.Run, error) {
	rows, err := r.db.Query(QueryListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}
	defer rows.Close()

	var runs []domain.Run
	for rows.Next() {
		run, err := r.scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}

	return runs, rows.Err()
}

// UpdateStatus updates run status and result
func (r *Repository) UpdateStatus(id string, status Status, result *Result, failureReason *string) error {
	now := time.Now()

	var startedAt, finishedAt *time.Time
	if status == StatusRunning {
		startedAt = &now
	}
	if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
		finishedAt = &now
	}

	var resultStr *string
	if result != nil {
		r := string(*result)
		resultStr = &r
	}

	_, err := r.db.Exec(
		QueryUpdateStatus,
		id,
		status,
		startedAt,
		finishedAt,
		resultStr,
		failureReason,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	return nil
}

func (r *Repository) scanRun(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Run, error) {
	var run domain.Run
	var result *string
	var startedAt sql.NullTime
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&run.ID,
		&run.TaskID,
		&run.WorkUnitID,
		&run.Status,
		&run.Attempt,
		&startedAt,
		&run.FinishedAt,
		&result,
		&run.FailureReason,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan run: %w", err)
	}

	if result != nil {
		r := Result(*result)
		run.Result = &r
	}
	if startedAt.Valid {
		run.StartedAt = startedAt.Time
	}

	return &run, nil
}
