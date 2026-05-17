// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

// Repository handles task persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new task repository
func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Create inserts a new task
func (r *Repository) Create(task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = task.CreatedAt
	}

	acceptanceCriteria, err := json.Marshal(task.AcceptanceCriteria)
	if err != nil {
		return fmt.Errorf("failed to marshal acceptance criteria: %w", err)
	}

	_, err = r.db.Exec(
		QueryInsert,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.RiskLevel,
		task.CreatedFromMessageID,
		acceptanceCriteria,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetByID retrieves a task by ID
func (r *Repository) GetByID(id string) (*Task, error) {
	row := r.db.QueryRow(QueryGetByID, id)

	return r.scanTask(row)
}

// List retrieves all tasks
func (r *Repository) List() ([]Task, error) {
	rows, err := r.db.Query(QueryList)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tasks []Task
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

// Update updates a task
func (r *Repository) Update(task *Task) error {
	task.UpdatedAt = time.Now()

	acceptanceCriteria, err := json.Marshal(task.AcceptanceCriteria)
	if err != nil {
		return fmt.Errorf("failed to marshal acceptance criteria: %w", err)
	}

	_, err = r.db.Exec(
		QueryUpdate,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.RiskLevel,
		acceptanceCriteria,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (r *Repository) scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*Task, error) {
	var task Task
	var acceptanceCriteriaJSON []byte

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.RiskLevel,
		&task.CreatedFromMessageID,
		&acceptanceCriteriaJSON,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	if len(acceptanceCriteriaJSON) > 0 {
		if err := json.Unmarshal(acceptanceCriteriaJSON, &task.AcceptanceCriteria); err != nil {
			return nil, fmt.Errorf("failed to unmarshal acceptance criteria: %w", err)
		}
	}

	return &task, nil
}
