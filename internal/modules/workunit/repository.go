// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package workunit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

// Repository handles work unit persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new work unit repository
func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Create inserts a new work unit
func (r *Repository) Create(wu *WorkUnit) error {
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
		QueryInsert,
		wu.ID,
		wu.TaskID,
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
func (r *Repository) GetByID(id string) (*WorkUnit, error) {
	row := r.db.QueryRow(QueryGetByID, id)
	return r.scanWorkUnit(row)
}

// ListByTask retrieves all work units for a task
func (r *Repository) ListByTask(taskID string) ([]WorkUnit, error) {
	rows, err := r.db.Query(QueryListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list work units: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var workUnits []WorkUnit
	for rows.Next() {
		wu, err := r.scanWorkUnit(rows)
		if err != nil {
			return nil, err
		}
		workUnits = append(workUnits, *wu)
	}

	return workUnits, rows.Err()
}

// ListByTaskGraph retrieves all work units for a task graph.
func (r *Repository) ListByTaskGraph(taskGraphID string) ([]WorkUnit, error) {
	rows, err := r.db.Query(QueryListByTaskGraph, taskGraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to list work units by task graph: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var workUnits []WorkUnit
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
func (r *Repository) UpdateStatus(id string, status Status, updatedAt time.Time) (sql.Result, error) {
	res, err := r.db.Exec(QueryUpdateStatus, id, status, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update work unit status: %w", err)
	}
	return res, nil
}

func (r *Repository) scanWorkUnit(scanner interface {
	Scan(dest ...interface{}) error
}) (*WorkUnit, error) {
	var wu WorkUnit
	var ownedPathsJSON, readPathsJSON, acceptanceCriteriaJSON, validationPlanJSON, dependsOnJSON []byte
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&wu.ID,
		&wu.TaskID,
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
		if err := json.Unmarshal(ownedPathsJSON, &wu.OwnedPaths); err != nil {
			return nil, fmt.Errorf("failed to unmarshal owned_paths: %w", err)
		}
	}
	if len(readPathsJSON) > 0 {
		if err := json.Unmarshal(readPathsJSON, &wu.ReadPaths); err != nil {
			return nil, fmt.Errorf("failed to unmarshal read_paths: %w", err)
		}
	}
	if len(acceptanceCriteriaJSON) > 0 {
		if err := json.Unmarshal(acceptanceCriteriaJSON, &wu.AcceptanceCriteria); err != nil {
			return nil, fmt.Errorf("failed to unmarshal acceptance_criteria: %w", err)
		}
	}
	if len(validationPlanJSON) > 0 {
		if err := json.Unmarshal(validationPlanJSON, &wu.ValidationPlan); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validation_plan: %w", err)
		}
	}
	if len(dependsOnJSON) > 0 {
		if err := json.Unmarshal(dependsOnJSON, &wu.DependsOn); err != nil {
			return nil, fmt.Errorf("failed to unmarshal depends_on: %w", err)
		}
	}

	return &wu, nil
}
