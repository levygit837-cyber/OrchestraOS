package transition

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
)

// TransitionPayload builds a standard transition payload from input metadata.
func TransitionPayload(from, to interface{}, input TransitionInput) map[string]interface{} {
	payload := map[string]interface{}{
		"from_status": from,
		"to_status":   to,
	}
	if input.Runtime != "" {
		payload["runtime"] = input.Runtime
	}
	if len(input.EvidenceRefs) > 0 {
		payload["evidence_refs"] = input.EvidenceRefs
	}
	if input.ValidationEventID != "" {
		payload["validation_event_id"] = input.ValidationEventID
	}
	if input.Justification != "" {
		payload["justification"] = input.Justification
	}
	if input.FailureReason != "" {
		payload["failure_reason"] = input.FailureReason
	}
	for key, value := range input.Extra {
		payload[key] = value
	}
	return payload
}

// TransitionContext adapts TransitionInput to the statemachine context type.
func TransitionContext(input TransitionInput) statemachine.TransitionContext {
	return statemachine.TransitionContext{
		EvidenceRefs:      input.EvidenceRefs,
		ValidationEventID: input.ValidationEventID,
		Justification:     input.Justification,
	}
}

// RequireFinalAudit enforces that final states carry audit metadata.
func RequireFinalAudit(target string, input TransitionInput, op string) error {
	if !IsFinalStatus(target) {
		return nil
	}
	if len(input.EvidenceRefs) > 0 || input.ValidationEventID != "" || input.Justification != "" || input.FailureReason != "" {
		return nil
	}
	return apperrors.New(apperrors.CodeInvalidInput, op, "final state requires evidence, validation event, failure reason, or justification")
}

// IsFinalStatus reports whether the given status represents a terminal state.
func IsFinalStatus(status string) bool {
	switch status {
	case "completed", "failed", "cancelled", "stopped":
		return true
	default:
		return false
	}
}
