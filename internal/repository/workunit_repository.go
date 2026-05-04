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

// WorkUnitRepository handles work unit persistence
type WorkUnitRepository struct {
	db *sql.DB
}

// NewWorkUnitRepository creates a new work unit repository
func NewWorkUnitRepository(db *sql.DB) *WorkUnitRepository {
	return &WorkUnitRepository{db: db}
}

// Create inserts a new work unit
func (r *WorkUnitRepository) Create(wu *domain.WorkUnit) error {
	if wu.ID == "" {
		wu.ID = uuid.New().String()
	}

	now := time.Now()

	ownedPaths, err := json.Marshal(wu.OwnedPaths)
	if err != nil {
		return fmt.Errorf("failed to marshal owned paths: %w", err)
	}

	readPaths, err := json.Marshal(wu.ReadPaths)
	if err != nil {
		return fmt.Errorf("failed to marshal read paths: %w", err)
	}

	acceptanceCriteria, err := json.Marshal(wu.AcceptanceCriteria)
	if err != nil {
		return fmt.Errorf("failed to marshal acceptance criteria: %w", err)
	}

	validationPlan, err := json.Marshal(wu.ValidationPlan)
	if err != nil {
		return fmt.Errorf("failed to marshal validation plan: %w", err)
	}

	dependsOn, err := json.Marshal(wu.DependsOn)
	if err != nil {
		return fmt.Errorf("failed to marshal depends on: %w", err)
	}

	_, err = r.db.Exec(
		db.QueryWorkUnitInsert,
		wu.ID,
		wu.TaskGraphID,
		wu.Title,
		wu.Objective,
		wu.AssignedAgentProfile,
		wu.Status,
		ownedPaths,
		readPaths,
		acceptanceCriteria,
		validationPlan,
		dependsOn,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create work unit: %w", err)
	}

	return nil
}

// GetByID retrieves a work unit by ID
func (r *WorkUnitRepository) GetByID(id string) (*domain.WorkUnit, error) {
	row := r.db.QueryRow(db.QueryWorkUnitGetByID, id)
	return r.scanWorkUnit(row)
}

// ListByTask retrieves all work units for a task
func (r *WorkUnitRepository) ListByTask(taskID string) ([]domain.WorkUnit, error) {
	rows, err := r.db.Query(db.QueryWorkUnitListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list work units: %w", err)
	}
	defer rows.Close()

	var workUnits []domain.WorkUnit
	for rows.Next() {
		wu, err := r.scanWorkUnit(rows)
		if err != nil {
			return nil, err
		}
		workUnits = append(workUnits, *wu)
	}

	return workUnits, rows.Err()
}

// UpdateStatus updates the status of a work unit
func (r *WorkUnitRepository) UpdateStatus(id string, status domain.WorkUnitStatus) error {
	_, err := r.db.Exec(db.QueryWorkUnitUpdateStatus, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update work unit status: %w", err)
	}
	return nil
}

func (r *WorkUnitRepository) scanWorkUnit(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.WorkUnit, error) {
	var wu domain.WorkUnit
	var ownedPathsJSON, readPathsJSON, acceptanceCriteriaJSON, validationPlanJSON, dependsOnJSON []byte
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&wu.ID,
		&wu.TaskGraphID,
		&wu.Title,
		&wu.Objective,
		&wu.AssignedAgentProfile,
		&wu.Status,
		&ownedPathsJSON,
		&readPathsJSON,
		&acceptanceCriteriaJSON,
		&validationPlanJSON,
		&dependsOnJSON,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan work unit: %w", err)
	}

	if len(ownedPathsJSON) > 0 {
		json.Unmarshal(ownedPathsJSON, &wu.OwnedPaths)
	}
	if len(readPathsJSON) > 0 {
		json.Unmarshal(readPathsJSON, &wu.ReadPaths)
	}
	if len(acceptanceCriteriaJSON) > 0 {
		json.Unmarshal(acceptanceCriteriaJSON, &wu.AcceptanceCriteria)
	}
	if len(validationPlanJSON) > 0 {
		json.Unmarshal(validationPlanJSON, &wu.ValidationPlan)
	}
	if len(dependsOnJSON) > 0 {
		json.Unmarshal(dependsOnJSON, &wu.DependsOn)
	}

	return &wu, nil
}
