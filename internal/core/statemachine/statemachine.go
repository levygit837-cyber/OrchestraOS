package statemachine

import (
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

type Aggregate string

const (
	AggregateTask         Aggregate = "task"
	AggregateWorkUnit     Aggregate = "work_unit"
	AggregateRun          Aggregate = "run"
	AggregateAgentSession Aggregate = "agent_session"
)

type TransitionContext struct {
	EvidenceRefs      []string
	ValidationEventID string
	Justification     string
}

func CanTransition(aggregate Aggregate, from, to string, ctx TransitionContext) error {
	if from == to {
		return nil
	}

	table, ok := transitions[aggregate]
	if !ok {
		return apperrors.New(apperrors.CodeInvalidInput, "statemachine.transition", fmt.Sprintf("unknown aggregate %q", aggregate))
	}
	if isTerminal(from) {
		return invalidTransition(aggregate, from, to)
	}

	allowed := table[from]
	for _, candidate := range allowed {
		if candidate == to {
			if isCompleted(to) && !hasCompletionEvidence(ctx) {
				return apperrors.New(
					apperrors.CodeInvalidTransition,
					"statemachine.completion",
					"completed transition requires validation evidence or justification",
				)
			}
			return nil
		}
	}

	return invalidTransition(aggregate, from, to)
}

func isTerminal(status string) bool {
	return status == "completed" ||
		status == "failed" ||
		status == "cancelled" ||
		status == "stopped"
}

func isCompleted(status string) bool {
	return status == "completed"
}

func hasCompletionEvidence(ctx TransitionContext) bool {
	return len(ctx.EvidenceRefs) > 0 || ctx.ValidationEventID != "" || ctx.Justification != ""
}

func invalidTransition(aggregate Aggregate, from, to string) error {
	return apperrors.New(
		apperrors.CodeInvalidTransition,
		"statemachine.transition",
		fmt.Sprintf("%s cannot transition from %q to %q", aggregate, from, to),
	)
}

var transitions = map[Aggregate]map[string][]string{
	AggregateTask: {
		"created":          {"triaged", "failed", "cancelled"},
		"triaged":          {"planned", "failed", "cancelled"},
		"planned":          {"scheduled", "failed", "cancelled"},
		"scheduled":        {"sandbox_preparing", "running", "paused", "cancelled"},
		"sandbox_preparing": {"running", "failed", "cancelled"},
		"running":          {"waiting_approval", "paused", "validating", "failed", "cancelled"},
		"waiting_approval": {"running", "paused", "failed", "cancelled"},
		"paused":           {"running", "failed", "cancelled"},
		"validating":       {"completed", "running", "failed", "cancelled"},
	},
	AggregateWorkUnit: {
		"created":         {"planned", "scheduled", "blocked", "running", "cancelled"},
		"planned":         {"scheduled", "blocked", "failed", "cancelled"},
		"scheduled":       {"running", "blocked", "paused", "cancelled"},
		"blocked":         {"scheduled", "running", "failed", "cancelled"},
		"running":         {"waiting_approval", "paused", "validating", "failed", "cancelled"},
		"waiting_approval": {"running", "paused", "failed", "cancelled"},
		"paused":          {"running", "failed", "cancelled"},
		"validating":      {"completed", "running", "failed", "cancelled"},
	},
	AggregateRun: {
		"created":         {"running", "failed", "cancelled"},
		"running":         {"waiting_approval", "paused", "validating", "failed", "cancelled"},
		"waiting_approval": {"running", "paused", "failed", "cancelled"},
		"paused":          {"running", "failed", "cancelled"},
		"validating":      {"completed", "running", "failed", "cancelled"},
	},
	AggregateAgentSession: {
		"starting":        {"running", "failed"},
		"running":         {"waiting_approval", "paused", "disconnected", "stopping", "failed"},
		"waiting_approval": {"running", "paused", "stopping", "failed"},
		"paused":          {"running", "stopping", "failed"},
		"disconnected":    {"running", "stopping", "failed"},
		"stopping":        {"stopped", "failed"},
	},
}
