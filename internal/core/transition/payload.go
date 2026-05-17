package transition

import (
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
