package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// RunRepository handles run persistence
type RunRepository struct {
	db DBTX
}

// NewRunRepository creates a new run repository
func NewRunRepository(db DBTX) *RunRepository {
	return &RunRepository{db: db}
}

// Create inserts a new run
func (r *RunRepository) Create(run *domain.Run) error {
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
		db.QueryRunInsert,
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
func (r *RunRepository) GetByID(id string) (*domain.Run, error) {
	row := r.db.QueryRow(db.QueryRunGetByID, id)
	return r.scanRun(row)
}

// List retrieves all runs
func (r *RunRepository) List() ([]domain.Run, error) {
	rows, err := r.db.Query(db.QueryRunList)
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
func (r *RunRepository) ListByTask(taskID string) ([]domain.Run, error) {
	rows, err := r.db.Query(db.QueryRunListByTask, taskID)
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
func (r *RunRepository) UpdateStatus(id string, status domain.RunStatus, result *domain.RunResult, failureReason *string) error {
	now := time.Now()

	var startedAt, finishedAt *time.Time
	if status == domain.RunStatusRunning {
		startedAt = &now
	}
	if status == domain.RunStatusCompleted || status == domain.RunStatusFailed || status == domain.RunStatusCancelled {
		finishedAt = &now
	}

	var resultStr *string
	if result != nil {
		r := string(*result)
		resultStr = &r
	}

	_, err := r.db.Exec(
		db.QueryRunUpdateStatus,
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

func (r *RunRepository) scanRun(scanner interface {
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
		r := domain.RunResult(*result)
		run.Result = &r
	}
	if startedAt.Valid {
		run.StartedAt = startedAt.Time
	}

	return &run, nil
}
