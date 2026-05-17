package prompt

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// PrepareRunPrompt gathers run, work unit, task and session data and prepares the prompt.
func (s *PromptService) PrepareRunPrompt(ctx context.Context, input PrepareRunPromptInput) (*PreparedRunPrompt, error) {
	const op = "prompt_service.prepare_run_prompt"
	if err := validation.RequiredUUID(input.RunID, "run_id", op); err != nil {
		return nil, err
	}
	if err := validation.RequiredUUID(input.AgentSessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.PromptSnapshotID, "prompt_snapshot_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.ToolsetSnapshotID, "toolset_snapshot_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.PromptSnapshotEventID, "prompt_snapshot_event_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.ToolsetSnapshotEventID, "toolset_snapshot_event_id", op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "prompt_service.begin_prepare")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	run, err := runmod.RequireByID(ctx, tx, input.RunID)
	if err != nil {
		return nil, err
	}
	wu, err := workunitmod.RequireByID(ctx, tx, run.WorkUnitID)
	if err != nil {
		return nil, err
	}
	task, err := taskmod.RequireByID(ctx, tx, run.TaskID)
	if err != nil {
		return nil, err
	}
	session, err := agentsessionmod.RequireByID(ctx, tx, input.AgentSessionID)
	if err != nil {
		return nil, err
	}
	if session.RunID != run.ID {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "agent_session_id does not belong to run_id")
	}
	if wu.TaskID != task.ID {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "work_unit_id does not belong to task_id")
	}

	return s.PrepareAndPersistPrompt(ctx, tx, PrepareAndPersistInput{
		Run:                    run,
		WorkUnit:               wu,
		Task:                   task,
		Session:                session,
		PromptSnapshotID:       input.PromptSnapshotID,
		ToolsetSnapshotID:      input.ToolsetSnapshotID,
		PromptSnapshotEventID:  input.PromptSnapshotEventID,
		ToolsetSnapshotEventID: input.ToolsetSnapshotEventID,
	})
}
