package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// TaskRepository handles task persistence
type TaskRepository struct {
	db DBTX
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db DBTX) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create inserts a new task
func (r *TaskRepository) Create(task *domain.Task) error {
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
		db.QueryTaskInsert,
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
func (r *TaskRepository) GetByID(id string) (*domain.Task, error) {
	row := r.db.QueryRow(db.QueryTaskGetByID, id)

	return r.scanTask(row)
}

// List retrieves all tasks
func (r *TaskRepository) List() ([]domain.Task, error) {
	rows, err := r.db.Query(db.QueryTaskList)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
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
func (r *TaskRepository) Update(task *domain.Task) error {
	task.UpdatedAt = time.Now()

	acceptanceCriteria, err := json.Marshal(task.AcceptanceCriteria)
	if err != nil {
		return fmt.Errorf("failed to marshal acceptance criteria: %w", err)
	}

	_, err = r.db.Exec(
		db.QueryTaskUpdate,
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

func (r *TaskRepository) scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Task, error) {
	var task domain.Task
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
