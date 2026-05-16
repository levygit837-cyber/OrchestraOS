package orchestration

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

// PromptOrchestrator coordinates cross-module data gathering for prompt preparation.
type PromptOrchestrator struct {
	db            *sql.DB
	promptService *promptmod.PromptService
}

// NewPromptOrchestrator creates a new prompt orchestrator.
func NewPromptOrchestrator(db *sql.DB, promptService *promptmod.PromptService) *PromptOrchestrator {
	return &PromptOrchestrator{
		db:            db,
		promptService: promptService,
	}
}

// PrepareRunPrompt gathers run, work unit, task and session data and prepares the prompt.
func (o *PromptOrchestrator) PrepareRunPrompt(ctx context.Context, input promptmod.PrepareRunPromptInput) (*promptmod.PreparedRunPrompt, error) {
	const op = "prompt_orchestrator.prepare_run_prompt"
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

	tx, err := dbcore.BeginTx(ctx, o.db, "prompt_orchestrator.begin_prepare")
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

	return o.promptService.PrepareAndPersistPrompt(ctx, tx, promptmod.PrepareAndPersistInput{
		Run:      run,
		WorkUnit: wu,
		// TODO: remove when prompt.PrepareAndPersistInput.Task uses *task.Task directly.
		Task: &domain.Task{
			ID:                   task.ID,
			Title:                task.Title,
			Description:          task.Description,
			Status:               domain.TaskStatus(task.Status),
			Priority:             domain.Priority(task.Priority),
			RiskLevel:            domain.RiskLevel(task.RiskLevel),
			CreatedFromMessageID: task.CreatedFromMessageID,
			AcceptanceCriteria:   task.AcceptanceCriteria,
			CreatedAt:            task.CreatedAt,
			UpdatedAt:            task.UpdatedAt,
		},
		Session:                session,
		PromptSnapshotID:       input.PromptSnapshotID,
		ToolsetSnapshotID:      input.ToolsetSnapshotID,
		PromptSnapshotEventID:  input.PromptSnapshotEventID,
		ToolsetSnapshotEventID: input.ToolsetSnapshotEventID,
	})
}
