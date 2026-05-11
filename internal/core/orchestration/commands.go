package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type Commander struct {
	db *sql.DB
}

type TransitionOptions struct {
	AgentID           string
	Runtime           string
	EvidenceRefs      []string
	ValidationEventID string
	Justification     string
	Result            *domain.RunResult
	FailureReason     *string
	Extra             map[string]interface{}
}

func NewCommander(database *sql.DB) *Commander {
	return &Commander{db: database}
}

func (c *Commander) TransitionTask(ctx context.Context, taskID string, target domain.TaskStatus, options TransitionOptions) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.begin_task_transition", err)
	}
	defer tx.Rollback()

	current, err := getTaskStatus(ctx, tx, taskID)
	if err != nil {
		return err
	}
	if err := statemachine.CanTransition(statemachine.AggregateTask, string(current), string(target), transitionContext(options)); err != nil {
		return err
	}

	if err := appendTransitionEvent(tx, taskTransitionEvent(target), taskID, "", "", options.AgentID, current, target, options); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE tasks SET status = $2, updated_at = $3 WHERE id = $1`, taskID, target, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.update_task_projection", err)
	}
	if err := ensureUpdated(result, "task"); err != nil {
		return err
	}

	return commit(tx, "orchestration.commit_task_transition")
}

func (c *Commander) TransitionWorkUnit(ctx context.Context, workUnitID string, target domain.WorkUnitStatus, options TransitionOptions) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.begin_work_unit_transition", err)
	}
	defer tx.Rollback()

	taskID, current, err := getWorkUnitStatus(ctx, tx, workUnitID)
	if err != nil {
		return err
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(current), string(target), transitionContext(options)); err != nil {
		return err
	}

	if err := appendTransitionEvent(tx, workUnitTransitionEvent(target), taskID, "", workUnitID, options.AgentID, current, target, options); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, QueryWorkUnitUpdateStatus, workUnitID, target, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.update_work_unit_projection", err)
	}
	if err := ensureUpdated(result, "work unit"); err != nil {
		return err
	}

	return commit(tx, "orchestration.commit_work_unit_transition")
}

func (c *Commander) TransitionRun(ctx context.Context, runID string, target domain.RunStatus, options TransitionOptions) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.begin_run_transition", err)
	}
	defer tx.Rollback()

	taskID, workUnitID, current, err := getRunStatus(ctx, tx, runID)
	if err != nil {
		return err
	}
	if err := statemachine.CanTransition(statemachine.AggregateRun, string(current), string(target), transitionContext(options)); err != nil {
		return err
	}

	if err := appendTransitionEvent(tx, runTransitionEvent(target), taskID, runID, workUnitID, options.AgentID, current, target, options); err != nil {
		return err
	}

	result := options.Result
	if result == nil {
		switch target {
		case domain.RunStatusCompleted:
			defaultResult := domain.RunResultSucceeded
			result = &defaultResult
		case domain.RunStatusFailed:
			defaultResult := domain.RunResultFailed
			result = &defaultResult
		case domain.RunStatusCancelled:
			defaultResult := domain.RunResultCancelled
			result = &defaultResult
		}
	}

	if err := UpdateRunProjection(ctx, tx, runID, target, result, options.FailureReason); err != nil {
		return err
	}

	return commit(tx, "orchestration.commit_run_transition")
}

func (c *Commander) TransitionAgentSession(ctx context.Context, sessionID string, target domain.AgentSessionStatus, options TransitionOptions) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.begin_agent_session_transition", err)
	}
	defer tx.Rollback()

	session, err := getAgentSessionStatus(ctx, tx, sessionID)
	if err != nil {
		return err
	}
	if err := statemachine.CanTransition(statemachine.AggregateAgentSession, string(session.status), string(target), transitionContext(options)); err != nil {
		return err
	}
	run, err := getRunContext(ctx, tx, session.runID)
	if err != nil {
		return err
	}

	agentID := options.AgentID
	if agentID == "" {
		agentID = session.agentID
	}
	if err := appendTransitionEvent(tx, agentSessionTransitionEvent(target), run.taskID, session.runID, run.workUnitID, agentID, session.status, target, options); err != nil {
		return err
	}

	var heartbeatAt, checkpointAt *time.Time
	now := time.Now().UTC()
	if target == domain.AgentSessionStatusRunning {
		heartbeatAt = &now
	}
	result, err := tx.ExecContext(ctx, QueryAgentSessionUpdateStatus, sessionID, target, heartbeatAt, checkpointAt, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.update_agent_session_projection", err)
	}
	if err := ensureUpdated(result, "agent session"); err != nil {
		return err
	}

	return commit(tx, "orchestration.commit_agent_session_transition")
}

func getTaskStatus(ctx context.Context, tx *sql.Tx, taskID string) (domain.TaskStatus, error) {
	var status domain.TaskStatus
	if err := tx.QueryRowContext(ctx, `SELECT status FROM tasks WHERE id = $1`, taskID).Scan(&status); err != nil {
		if err == sql.ErrNoRows {
			return "", apperrors.New(apperrors.CodeNotFound, "orchestration.get_task", "task not found")
		}
		return "", apperrors.Wrap(apperrors.CodePersistence, "orchestration.get_task", err)
	}
	return status, nil
}

func getWorkUnitStatus(ctx context.Context, tx *sql.Tx, workUnitID string) (string, domain.WorkUnitStatus, error) {
	var taskID string
	var status domain.WorkUnitStatus
	if err := tx.QueryRowContext(ctx, `SELECT task_id, status FROM work_units WHERE id = $1`, workUnitID).Scan(&taskID, &status); err != nil {
		if err == sql.ErrNoRows {
			return "", "", apperrors.New(apperrors.CodeNotFound, "orchestration.get_work_unit", "work unit not found")
		}
		return "", "", apperrors.Wrap(apperrors.CodePersistence, "orchestration.get_work_unit", err)
	}
	return taskID, status, nil
}

func getRunStatus(ctx context.Context, tx *sql.Tx, runID string) (string, string, domain.RunStatus, error) {
	var taskID string
	var workUnitID sql.NullString
	var status domain.RunStatus
	if err := tx.QueryRowContext(ctx, `SELECT task_id, work_unit_id, status FROM runs WHERE id = $1`, runID).Scan(&taskID, &workUnitID, &status); err != nil {
		if err == sql.ErrNoRows {
			return "", "", "", apperrors.New(apperrors.CodeNotFound, "orchestration.get_run", "run not found")
		}
		return "", "", "", apperrors.Wrap(apperrors.CodePersistence, "orchestration.get_run", err)
	}
	return taskID, workUnitID.String, status, nil
}

type sessionStatus struct {
	agentID string
	runID   string
	status  domain.AgentSessionStatus
}

func getAgentSessionStatus(ctx context.Context, tx *sql.Tx, sessionID string) (sessionStatus, error) {
	var session sessionStatus
	if err := tx.QueryRowContext(ctx, `SELECT agent_id, run_id, status FROM agent_sessions WHERE id = $1`, sessionID).Scan(&session.agentID, &session.runID, &session.status); err != nil {
		if err == sql.ErrNoRows {
			return sessionStatus{}, apperrors.New(apperrors.CodeNotFound, "orchestration.get_agent_session", "agent session not found")
		}
		return sessionStatus{}, apperrors.Wrap(apperrors.CodePersistence, "orchestration.get_agent_session", err)
	}
	return session, nil
}

type runContext struct {
	taskID     string
	workUnitID string
}

func getRunContext(ctx context.Context, tx *sql.Tx, runID string) (runContext, error) {
	var run runContext
	var workUnitID sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT task_id, work_unit_id FROM runs WHERE id = $1`, runID).Scan(&run.taskID, &workUnitID); err != nil {
		if err == sql.ErrNoRows {
			return runContext{}, apperrors.New(apperrors.CodeNotFound, "orchestration.get_run_context", "run not found")
		}
		return runContext{}, apperrors.Wrap(apperrors.CodePersistence, "orchestration.get_run_context", err)
	}
	run.workUnitID = workUnitID.String
	return run, nil
}

func appendTransitionEvent(tx *sql.Tx, eventType, taskID, runID, workUnitID, agentID string, from, to interface{}, options TransitionOptions) error {
	payload := map[string]interface{}{
		"from_status": from,
		"to_status":   to,
	}
	if options.Runtime != "" {
		payload["runtime"] = options.Runtime
	}
	if len(options.EvidenceRefs) > 0 {
		payload["evidence_refs"] = options.EvidenceRefs
	}
	if options.ValidationEventID != "" {
		payload["validation_event_id"] = options.ValidationEventID
	}
	if options.Justification != "" {
		payload["justification"] = options.Justification
	}
	if options.Result != nil {
		payload["result"] = *options.Result
	}
	if options.FailureReason != nil {
		payload["failure_reason"] = *options.FailureReason
	}
	for key, value := range options.Extra {
		payload[key] = value
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeValidation, "orchestration.marshal_transition_payload", err)
	}
	store, err := eventstore.NewStoreWithExecutor(tx)
	if err != nil {
		return err
	}
	return store.Append(&domain.EventEnvelope{
		Type:        eventType,
		Version:     "v1",
		TaskID:      taskID,
		RunID:       runID,
		WorkUnitID:  workUnitID,
		AgentID:     agentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payloadBytes,
	})
}

func ensureUpdated(result sql.Result, entity string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.rows_affected", err)
	}
	if rows == 0 {
		return apperrors.New(apperrors.CodeConflict, "orchestration.update_projection", fmt.Sprintf("%s projection was not updated", entity))
	}
	return nil
}

func commit(tx *sql.Tx, op string) error {
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return nil
}

func transitionContext(options TransitionOptions) statemachine.TransitionContext {
	return statemachine.TransitionContext{
		EvidenceRefs:      options.EvidenceRefs,
		ValidationEventID: options.ValidationEventID,
		Justification:     options.Justification,
	}
}

func taskTransitionEvent(status domain.TaskStatus) string {
	switch status {
	case domain.TaskStatusRunning:
		return "task.started"
	default:
		return "task." + string(status)
	}
}

func workUnitTransitionEvent(status domain.WorkUnitStatus) string {
	switch status {
	case domain.WorkUnitStatusRunning:
		return "work_unit.started"
	default:
		return "work_unit." + string(status)
	}
}

func runTransitionEvent(status domain.RunStatus) string {
	switch status {
	case domain.RunStatusRunning:
		return "run.started"
	default:
		return "run." + string(status)
	}
}

func agentSessionTransitionEvent(status domain.AgentSessionStatus) string {
	return "agent.session_" + string(status)
}

func UpdateRunProjection(ctx context.Context, tx *sql.Tx, runID string, status domain.RunStatus, result *domain.RunResult, failureReason *string) error {
	now := time.Now().UTC()
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

	res, err := tx.ExecContext(ctx, QueryRunUpdateStatus, runID, status, startedAt, finishedAt, resultStr, failureReason, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "orchestration.update_run_projection", err)
	}
	return dbcore.EnsureRowsAffected(res, "run", "orchestration.update_run_projection")
}
