// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package run

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/orchestration"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// TaskReader abstracts task reads to avoid cyclic imports.
type TaskReader interface {
	GetByID(id string) (*domain.Task, error)
}

// WorkUnitReader abstracts work-unit reads to avoid cyclic imports.
type WorkUnitReader interface {
	GetByID(id string) (*domain.WorkUnit, error)
}

type RunService struct {
	db                *sql.DB
	newTaskReader     func(*sql.Tx) TaskReader
	newWorkUnitReader func(*sql.Tx) WorkUnitReader
}

type CreateRunInput struct {
	ID         string
	EventID    string
	TaskID     string
	WorkUnitID string
	Attempt    int
}

func NewRunService(database *sql.DB, newTaskReader func(*sql.Tx) TaskReader, newWorkUnitReader func(*sql.Tx) WorkUnitReader) *RunService {
	return &RunService{db: database, newTaskReader: newTaskReader, newWorkUnitReader: newWorkUnitReader}
}

func (s *RunService) Create(ctx context.Context, input CreateRunInput) (*orchestration.OperationResult[*domain.Run], error) {
	if input.ID == "" {
		input.ID = uuid.New().String()
	}
	if input.Attempt == 0 {
		input.Attempt = 1
	}
	if err := validateCreateRunInput(input); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "run_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	task, err := s.requireTaskByID(tx, input.TaskID)
	if err != nil {
		return nil, err
	}
	wu, err := s.requireWorkUnitByID(tx, input.WorkUnitID)
	if err != nil {
		return nil, err
	}
	if wu.TaskID != task.ID {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "run_service.validate_refs", "work_unit_id does not belong to task_id")
	}

	run := &domain.Run{
		ID:         input.ID,
		TaskID:     task.ID,
		WorkUnitID: wu.ID,
		Status:     domain.RunStatusCreated,
		Attempt:    input.Attempt,
	}
	if err := NewRepository(tx).Create(run); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "run_service.create_projection", err)
	}

	payload, err := serialization.MarshalPayload("run_service.create_payload", map[string]interface{}{
		"run_id":       run.ID,
		"task_id":      run.TaskID,
		"work_unit_id": run.WorkUnitID,
		"status":       run.Status,
		"attempt":      run.Attempt,
	})
	if err != nil {
		return nil, err
	}
	appendResult, err := orchestration.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "run.created",
		Version:     orchestration.EventVersionV1,
		TaskID:      run.TaskID,
		RunID:       run.ID,
		WorkUnitID:  run.WorkUnitID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if err := dbcore.CommitTx(tx, "run_service.commit_create"); err != nil {
		return nil, err
	}
	return &orchestration.OperationResult[*domain.Run]{Value: run, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *RunService) Start(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusRunning, input, true)
}

func (s *RunService) Resume(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusRunning, input, false)
}

func (s *RunService) Validate(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusValidating, input, true)
}

func (s *RunService) Complete(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusCompleted, input, true)
}

func (s *RunService) Fail(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusFailed, input, true)
}

func (s *RunService) Cancel(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusCancelled, input, true)
}

func (s *RunService) Timeout(ctx context.Context, runID string, input orchestration.TransitionInput) (*orchestration.OperationResult[*domain.Run], error) {
	if input.FailureReason == "" {
		input.FailureReason = "run timed out"
	}
	if input.Justification == "" {
		input.Justification = "timeout reached before run completed"
	}
	if input.Extra == nil {
		input.Extra = map[string]interface{}{}
	}
	input.Extra["timeout"] = true
	return s.transition(ctx, runID, domain.RunStatusFailed, input, true)
}

func (s *RunService) transition(ctx context.Context, runID string, target domain.RunStatus, input orchestration.TransitionInput, updateWorkUnit bool) (*orchestration.OperationResult[*domain.Run], error) {
	op := "run_service.transition"
	if err := validation.RequiredUUID(runID, "run_id", op); err != nil {
		return nil, err
	}
	if err := validation.Runtime(input.Runtime, op); err != nil {
		return nil, err
	}
	if err := orchestration.RequireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "run_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	run, err := RequireByID(ctx, tx, runID)
	if err != nil {
		return nil, err
	}
	task, err := s.requireTaskByID(tx, run.TaskID)
	if err != nil {
		return nil, err
	}
	if target == domain.RunStatusRunning {
		if err := validateRunStartPolicy(task, input); err != nil {
			return nil, err
		}
	}
	if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(target), orchestration.TransitionContext(input)); err != nil {
		return nil, err
	}

	if updateWorkUnit && run.WorkUnitID != "" {
		if err := s.transitionRelatedWorkUnit(ctx, tx, run, target, input); err != nil {
			return nil, err
		}
	}

	payload := orchestration.TransitionPayload(run.Status, target, input)
	payload["attempt"] = run.Attempt
	payload["run_id"] = run.ID
	if target == domain.RunStatusCompleted {
		payload["result"] = domain.RunResultSucceeded
	}
	event, duplicate, err := orchestration.AppendTransition(ctx, tx, input.EventID, EventTypeForStatus(target), run.TaskID, run.ID, run.WorkUnitID, input.AgentID, payload)
	if err != nil {
		return nil, err
	}
	if !duplicate {
		result := ResultForStatus(target)
		var failureReason *string
		if input.FailureReason != "" {
			failureReason = &input.FailureReason
		}
		if err := orchestration.UpdateRunProjection(ctx, tx, run.ID, target, result, failureReason); err != nil {
			return nil, err
		}
		run.Status = target
		if target == domain.RunStatusRunning && run.StartedAt.IsZero() {
			run.StartedAt = time.Now().UTC()
		}
	}
	if err := dbcore.CommitTx(tx, "run_service.commit_transition"); err != nil {
		return nil, err
	}
	return &orchestration.OperationResult[*domain.Run]{Value: run, Event: event, Duplicate: duplicate}, nil
}

func (s *RunService) transitionRelatedWorkUnit(ctx context.Context, tx *sql.Tx, run *domain.Run, target domain.RunStatus, input orchestration.TransitionInput) error {
	wu, err := s.requireWorkUnitByID(tx, run.WorkUnitID)
	if err != nil {
		return err
	}
	var wuTarget domain.WorkUnitStatus
	switch target {
	case domain.RunStatusRunning:
		wuTarget = domain.WorkUnitStatusRunning
	case domain.RunStatusValidating:
		wuTarget = domain.WorkUnitStatusValidating
	case domain.RunStatusCompleted:
		wuTarget = domain.WorkUnitStatusCompleted
	case domain.RunStatusFailed:
		wuTarget = domain.WorkUnitStatusFailed
	case domain.RunStatusCancelled:
		wuTarget = domain.WorkUnitStatusCancelled
	default:
		return nil
	}
	if wu.Status == wuTarget {
		return nil
	}
	if wuTarget == domain.WorkUnitStatusRunning {
		if err := dbcore.AcquireAdvisoryTxLock(ctx, tx, "work_unit_paths:"+wu.TaskID, "run_service.work_unit_path_lock"); err != nil {
			return err
		}
		if err := workunitmod.ValidateDependenciesCompleted(ctx, tx, wu); err != nil {
			return err
		}
		if err := workunitmod.ValidateOwnedPathAvailability(ctx, tx, wu); err != nil {
			return err
		}
	}
	if wuTarget == domain.WorkUnitStatusCompleted && len(wu.AcceptanceCriteria) == 0 && input.Justification == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "run_service.work_unit_completion", "related work unit completion requires acceptance criteria or explicit justification")
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(wuTarget), orchestration.TransitionContext(input)); err != nil {
		if wuTarget == domain.WorkUnitStatusFailed && wu.Status == domain.WorkUnitStatusCreated {
			return nil
		}
		return err
	}
	if _, _, err := orchestration.AppendTransition(ctx, tx, "", workUnitEventTypeForStatus(wuTarget), run.TaskID, run.ID, wu.ID, input.AgentID, orchestration.TransitionPayload(wu.Status, wuTarget, input)); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE work_units SET status = $2, updated_at = $3 WHERE id = $1`, wu.ID, wuTarget, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "run_service.update_work_unit_projection", err)
	}
	return dbcore.EnsureRowsAffected(res, "work unit", "run_service.update_work_unit_projection")
}

func validateCreateRunInput(input CreateRunInput) error {
	op := "run_service.validate_create"
	if err := validation.RequiredUUID(input.ID, "run_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(input.WorkUnitID, "work_unit_id", op); err != nil {
		return err
	}
	if input.Attempt < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, op, "attempt must be greater than zero")
	}
	return nil
}

func validateRunStartPolicy(task *domain.Task, input orchestration.TransitionInput) error {
	if task.RiskLevel == domain.RiskLevelHigh || task.RiskLevel == domain.RiskLevelCritical {
		if input.Justification == "" {
			return apperrors.New(apperrors.CodePolicy, "run_service.policy", "starting high or critical risk tasks requires explicit justification")
		}
	}
	return nil
}

func (s *RunService) requireTaskByID(tx *sql.Tx, id string) (*domain.Task, error) {
	task, err := s.newTaskReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task.get", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "task.get", "task not found")
	}
	return task, nil
}

func (s *RunService) requireWorkUnitByID(tx *sql.Tx, id string) (*domain.WorkUnit, error) {
	wu, err := s.newWorkUnitReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "workunit.get", err)
	}
	if wu == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "workunit.get", "work unit not found")
	}
	return wu, nil
}

func workUnitEventTypeForStatus(status domain.WorkUnitStatus) string {
	if status == domain.WorkUnitStatusRunning {
		return "work_unit.started"
	}
	return "work_unit." + string(status)
}
