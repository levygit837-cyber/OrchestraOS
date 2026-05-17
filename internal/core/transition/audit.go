package transition

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

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
