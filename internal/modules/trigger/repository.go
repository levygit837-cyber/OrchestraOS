// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, boundary rules
// Ignoring these files will cause architecture test failures.

package trigger

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

// Repository handles trigger persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new trigger repository
func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Create inserts a new trigger
func (r *Repository) Create(trigger *Trigger) error {
	if trigger.ID == "" {
		trigger.ID = uuid.New().String()
	}
	now := time.Now()
	if trigger.CreatedAt.IsZero() {
		trigger.CreatedAt = now
	}

	_, err := r.db.Exec(
		QueryInsert,
		trigger.ID,
		trigger.RunID,
		trigger.TaskID,
		trigger.AgentSessionID,
		trigger.TriggerType,
		trigger.Status,
		trigger.AnomalyType,
		trigger.ThresholdValue,
		trigger.CurrentValue,
		trigger.TriggeredAt,
		trigger.ResolvedAt,
		trigger.ResolutionAction,
		trigger.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}
	return nil
}

// GetByID retrieves a trigger by ID
func (r *Repository) GetByID(id string) (*Trigger, error) {
	row := r.db.QueryRow(QueryGetByID, id)
	return r.scanTrigger(row)
}

// ListActive retrieves all active or triggered triggers
func (r *Repository) ListActive(ctx context.Context) ([]*Trigger, error) {
	rows, err := r.db.QueryContext(ctx, QueryListActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list active triggers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var triggers []*Trigger
	for rows.Next() {
		trigger, err := r.scanTrigger(rows)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, trigger)
	}
	return triggers, rows.Err()
}

// ListByRun retrieves all triggers for a run
func (r *Repository) ListByRun(ctx context.Context, runID string) ([]*Trigger, error) {
	rows, err := r.db.QueryContext(ctx, QueryListByRun, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to list triggers by run: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var triggers []*Trigger
	for rows.Next() {
		trigger, err := r.scanTrigger(rows)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, trigger)
	}
	return triggers, rows.Err()
}

// ExistsActiveSimilar checks if an active/triggered trigger already exists with the same
// trigger type, run/session and anomaly type.
func (r *Repository) ExistsActiveSimilar(triggerType Type, runID, agentSessionID, anomalyType *string) (bool, error) {
	var runVal, sessionVal, anomalyVal string
	if runID != nil {
		runVal = *runID
	}
	if agentSessionID != nil {
		sessionVal = *agentSessionID
	}
	if anomalyType != nil {
		anomalyVal = *anomalyType
	}
	var exists bool
	err := r.db.QueryRow(QueryExistsActiveSimilar, triggerType, runVal, sessionVal, anomalyVal).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check similar trigger existence: %w", err)
	}
	return exists, nil
}

// UpdateStatus updates trigger status and related timestamps
func (r *Repository) UpdateStatus(id string, status Status, triggeredAt, resolvedAt *time.Time, resolutionAction *ResolutionAction) error {
	_, err := r.db.Exec(
		QueryUpdateStatus,
		id,
		status,
		triggeredAt,
		resolvedAt,
		resolutionAction,
	)
	if err != nil {
		return fmt.Errorf("failed to update trigger status: %w", err)
	}
	return nil
}

func (r *Repository) scanTrigger(scanner interface {
	Scan(dest ...interface{}) error
}) (*Trigger, error) {
	var trigger Trigger
	var runID, taskID, agentSessionID *string
	var anomalyType *string
	var triggeredAt, resolvedAt sql.NullTime
	var resolutionAction *string

	err := scanner.Scan(
		&trigger.ID,
		&runID,
		&taskID,
		&agentSessionID,
		&trigger.TriggerType,
		&trigger.Status,
		&anomalyType,
		&trigger.ThresholdValue,
		&trigger.CurrentValue,
		&triggeredAt,
		&resolvedAt,
		&resolutionAction,
		&trigger.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan trigger: %w", err)
	}

	trigger.RunID = runID
	trigger.TaskID = taskID
	trigger.AgentSessionID = agentSessionID
	if anomalyType != nil {
		a := AnomalyType(*anomalyType)
		trigger.AnomalyType = &a
	}
	if triggeredAt.Valid {
		trigger.TriggeredAt = &triggeredAt.Time
	}
	if resolvedAt.Valid {
		trigger.ResolvedAt = &resolvedAt.Time
	}
	if resolutionAction != nil {
		ra := ResolutionAction(*resolutionAction)
		trigger.ResolutionAction = &ra
	}

	return &trigger, nil
}
