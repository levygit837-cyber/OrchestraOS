package agentsession

import (
	"encoding/json"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

// ValidateStatus checks if the status is valid.
func ValidateStatus(s Status, op string) error {
	switch s {
	case StatusStarting, StatusRunning, StatusWaitingApproval, StatusPaused,
		StatusStopping, StatusStopped, StatusDisconnected, StatusFailed:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid agent session status")
	}
}

// ValidateCreateAgentSessionInput validates the input for creating a session.
func ValidateCreateAgentSessionInput(input CreateAgentSessionInput) error {
	op := "agentsession.validate_create"
	if err := validation.RequiredUUID(input.ID, "agent_session_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validation.RequiredText(input.AgentID, "agent_id", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(input.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.LastSeenEventID, "last_seen_event_id", op); err != nil {
		return err
	}
	if len(input.RecoverableState) > 0 && !json.Valid(input.RecoverableState) {
		return apperrors.New(apperrors.CodeValidation, op, "recoverable_state must be valid JSON")
	}
	return nil
}
