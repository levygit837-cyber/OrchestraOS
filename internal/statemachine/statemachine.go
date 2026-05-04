package statemachine

import (
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
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
	return status == string(domain.TaskStatusCompleted) ||
		status == string(domain.TaskStatusFailed) ||
		status == string(domain.TaskStatusCancelled) ||
		status == string(domain.WorkUnitStatusCompleted) ||
		status == string(domain.WorkUnitStatusFailed) ||
		status == string(domain.WorkUnitStatusCancelled) ||
		status == string(domain.RunStatusCompleted) ||
		status == string(domain.RunStatusFailed) ||
		status == string(domain.RunStatusCancelled) ||
		status == string(domain.AgentSessionStatusStopped) ||
		status == string(domain.AgentSessionStatusFailed)
}

func isCompleted(status string) bool {
	return status == string(domain.TaskStatusCompleted) ||
		status == string(domain.WorkUnitStatusCompleted) ||
		status == string(domain.RunStatusCompleted)
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
		string(domain.TaskStatusCreated):          {string(domain.TaskStatusTriaged), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusTriaged):          {string(domain.TaskStatusPlanned), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusPlanned):          {string(domain.TaskStatusScheduled), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusScheduled):        {string(domain.TaskStatusSandboxPreparing), string(domain.TaskStatusRunning), string(domain.TaskStatusPaused), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusSandboxPreparing): {string(domain.TaskStatusRunning), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusRunning):          {string(domain.TaskStatusWaitingApproval), string(domain.TaskStatusPaused), string(domain.TaskStatusValidating), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusWaitingApproval):  {string(domain.TaskStatusRunning), string(domain.TaskStatusPaused), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusPaused):           {string(domain.TaskStatusRunning), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
		string(domain.TaskStatusValidating):       {string(domain.TaskStatusCompleted), string(domain.TaskStatusRunning), string(domain.TaskStatusFailed), string(domain.TaskStatusCancelled)},
	},
	AggregateWorkUnit: {
		string(domain.WorkUnitStatusCreated):         {string(domain.WorkUnitStatusPlanned), string(domain.WorkUnitStatusScheduled), string(domain.WorkUnitStatusBlocked), string(domain.WorkUnitStatusRunning), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusPlanned):         {string(domain.WorkUnitStatusScheduled), string(domain.WorkUnitStatusBlocked), string(domain.WorkUnitStatusFailed), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusScheduled):       {string(domain.WorkUnitStatusRunning), string(domain.WorkUnitStatusBlocked), string(domain.WorkUnitStatusPaused), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusBlocked):         {string(domain.WorkUnitStatusScheduled), string(domain.WorkUnitStatusRunning), string(domain.WorkUnitStatusFailed), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusRunning):         {string(domain.WorkUnitStatusWaitingApproval), string(domain.WorkUnitStatusPaused), string(domain.WorkUnitStatusValidating), string(domain.WorkUnitStatusFailed), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusWaitingApproval): {string(domain.WorkUnitStatusRunning), string(domain.WorkUnitStatusPaused), string(domain.WorkUnitStatusFailed), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusPaused):          {string(domain.WorkUnitStatusRunning), string(domain.WorkUnitStatusFailed), string(domain.WorkUnitStatusCancelled)},
		string(domain.WorkUnitStatusValidating):      {string(domain.WorkUnitStatusCompleted), string(domain.WorkUnitStatusRunning), string(domain.WorkUnitStatusFailed), string(domain.WorkUnitStatusCancelled)},
	},
	AggregateRun: {
		string(domain.RunStatusCreated):         {string(domain.RunStatusRunning), string(domain.RunStatusFailed), string(domain.RunStatusCancelled)},
		string(domain.RunStatusRunning):         {string(domain.RunStatusWaitingApproval), string(domain.RunStatusPaused), string(domain.RunStatusValidating), string(domain.RunStatusFailed), string(domain.RunStatusCancelled)},
		string(domain.RunStatusWaitingApproval): {string(domain.RunStatusRunning), string(domain.RunStatusPaused), string(domain.RunStatusFailed), string(domain.RunStatusCancelled)},
		string(domain.RunStatusPaused):          {string(domain.RunStatusRunning), string(domain.RunStatusFailed), string(domain.RunStatusCancelled)},
		string(domain.RunStatusValidating):      {string(domain.RunStatusCompleted), string(domain.RunStatusRunning), string(domain.RunStatusFailed), string(domain.RunStatusCancelled)},
	},
	AggregateAgentSession: {
		string(domain.AgentSessionStatusStarting):        {string(domain.AgentSessionStatusRunning), string(domain.AgentSessionStatusFailed)},
		string(domain.AgentSessionStatusRunning):         {string(domain.AgentSessionStatusWaitingApproval), string(domain.AgentSessionStatusPaused), string(domain.AgentSessionStatusDisconnected), string(domain.AgentSessionStatusStopping), string(domain.AgentSessionStatusFailed)},
		string(domain.AgentSessionStatusWaitingApproval): {string(domain.AgentSessionStatusRunning), string(domain.AgentSessionStatusPaused), string(domain.AgentSessionStatusStopping), string(domain.AgentSessionStatusFailed)},
		string(domain.AgentSessionStatusPaused):          {string(domain.AgentSessionStatusRunning), string(domain.AgentSessionStatusStopping), string(domain.AgentSessionStatusFailed)},
		string(domain.AgentSessionStatusDisconnected):    {string(domain.AgentSessionStatusRunning), string(domain.AgentSessionStatusStopping), string(domain.AgentSessionStatusFailed)},
		string(domain.AgentSessionStatusStopping):        {string(domain.AgentSessionStatusStopped), string(domain.AgentSessionStatusFailed)},
	},
}
