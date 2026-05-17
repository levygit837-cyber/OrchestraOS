package coordination

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
)

// ValidateRunForSessionCreation checks that a run is not in a terminal state before creating a session.
func ValidateRunForSessionCreation(ctx context.Context, tx *sql.Tx, runID string) error {
	run, err := runmod.RequireByID(ctx, tx, runID)
	if err != nil {
		return err
	}
	if run.Status == domain.RunStatusCompleted || run.Status == domain.RunStatusFailed || run.Status == domain.RunStatusCancelled {
		return apperrors.New(apperrors.CodeInvalidTransition, "coordination.validate_run_for_session", "cannot create session for terminal run")
	}
	return nil
}

// AgentSessionTimeout coordinates session timeout with run pause.
func AgentSessionTimeout(ctx context.Context, tx *sql.Tx, session *domain.AgentSession, recoverableState json.RawMessage, input transition.TransitionInput) (*domain.EventEnvelope, bool, error) {
	if len(recoverableState) > 0 {
		if err := agentsessionmod.NewRepository(tx).UpdateRecoverableState(session.ID, recoverableState); err != nil {
			return nil, false, apperrors.Wrap(apperrors.CodePersistence, "coordination.agent_session_timeout.update_state", err)
		}
	}

	run, err := runmod.RequireByID(ctx, tx, session.RunID)
	if err != nil {
		return nil, false, err
	}
	if run.Status == domain.RunStatusRunning || run.Status == domain.RunStatusWaitingApproval {
		if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(domain.RunStatusPaused), transition.TransitionContext(input)); err != nil {
			return nil, false, err
		}
		if _, _, err := transition.AppendTransition(ctx, tx, "", "run.paused", run.TaskID, run.ID, run.WorkUnitID, session.AgentID, transition.TransitionPayload(run.Status, domain.RunStatusPaused, input)); err != nil {
			return nil, false, err
		}
		if err := UpdateRunProjection(ctx, tx, run.ID, domain.RunStatusPaused, nil, nil); err != nil {
			return nil, false, err
		}
	}
	return nil, false, nil
}
