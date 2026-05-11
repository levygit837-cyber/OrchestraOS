package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/prompting"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
)

type PromptService struct {
	db *sql.DB
}

type PrepareRunPromptInput struct {
	RunID                  string
	AgentSessionID         string
	PromptSnapshotID       string
	ToolsetSnapshotID      string
	PromptSnapshotEventID  string
	ToolsetSnapshotEventID string
}

type PreparedRunPrompt struct {
	PromptSnapshot  *domain.PromptSnapshot
	ToolsetSnapshot *domain.ToolsetSnapshot
	SystemPrompt    string
	TaskPrompt      string
	CombinedPrompt  string
	PromptHash      string
	Toolset         []string
}

func NewPromptService(database *sql.DB) *PromptService {
	return &PromptService{db: database}
}

func (s *PromptService) PrepareRunPrompt(ctx context.Context, input PrepareRunPromptInput) (*PreparedRunPrompt, error) {
	const op = "prompt_service.prepare_run_prompt"
	if err := validateRequiredUUID(input.RunID, "run_id", op); err != nil {
		return nil, err
	}
	if err := validateRequiredUUID(input.AgentSessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.PromptSnapshotID, "prompt_snapshot_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.ToolsetSnapshotID, "toolset_snapshot_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.PromptSnapshotEventID, "prompt_snapshot_event_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.ToolsetSnapshotEventID, "toolset_snapshot_event_id", op); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "prompt_service.begin_prepare")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	run, err := getRun(ctx, tx, input.RunID)
	if err != nil {
		return nil, err
	}
	wu, err := getWorkUnit(ctx, tx, run.WorkUnitID)
	if err != nil {
		return nil, err
	}
	task, err := getTask(ctx, tx, run.TaskID)
	if err != nil {
		return nil, err
	}
	session, err := getAgentSession(ctx, tx, input.AgentSessionID)
	if err != nil {
		return nil, err
	}
	if session.RunID != run.ID {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "agent_session_id does not belong to run_id")
	}
	if wu.TaskID != task.ID {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "work_unit_id does not belong to task_id")
	}

	toolset, err := prompting.SelectToolset(wu.AssignedAgentProfile)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	}
	composed, err := prompting.Compose(prompting.TaskContext{
		TaskID:             task.ID,
		TaskTitle:          task.Title,
		TaskDescription:    task.Description,
		RunID:              run.ID,
		WorkUnitID:         wu.ID,
		TaskGraphID:        wu.TaskGraphID,
		WorkUnitTitle:      wu.Title,
		WorkUnitObjective:  wu.Objective,
		AgentProfile:       wu.AssignedAgentProfile,
		OwnedPaths:         wu.OwnedPaths,
		ReadPaths:          wu.ReadPaths,
		DependsOn:          wu.DependsOn,
		AcceptanceCriteria: wu.AcceptanceCriteria,
		ValidationPlan:     wu.ValidationPlan,
		Toolset:            toolset,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op, err)
	}

	repo := repository.NewPromptRepository(tx)
	for _, fragment := range composed.Fragments {
		domainFragment, err := promptFragmentToDomain(fragment)
		if err != nil {
			return nil, err
		}
		if err := repo.CreateOrVerifyFragment(domainFragment); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeConflict, "prompt_service.persist_fragment", err)
		}
	}

	variablesApplied, err := prompting.MarshalVariables(composed.VariablesApplied)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op, err)
	}
	promptSnapshot := &domain.PromptSnapshot{
		ID:                 valueOrNewUUID(input.PromptSnapshotID),
		RunID:              run.ID,
		WorkUnitID:         wu.ID,
		AgentSessionID:     session.ID,
		SystemPrompt:       composed.SystemPrompt,
		TaskPrompt:         composed.TaskPrompt,
		CombinedPrompt:     composed.CombinedPrompt,
		SystemPromptHash:   composed.SystemPromptHash,
		TaskPromptHash:     composed.TaskPromptHash,
		CombinedPromptHash: composed.CombinedPromptHash,
		CompositionHash:    composed.CompositionHash,
		CategorySignature:  composed.CategorySignature,
		FragmentRefs:       promptFragmentRefsToDomain(composed.FragmentRefs),
		AssemblyOrder:      composed.AssemblyOrder,
		VariablesApplied:   variablesApplied,
		CreatedAt:          time.Now().UTC(),
	}
	reusedPromptSnapshot, err := repo.CreateOrReferencePromptSnapshot(promptSnapshot)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "prompt_service.create_prompt_snapshot", err)
	}
	toolsetSnapshot := &domain.ToolsetSnapshot{
		ID:             valueOrNewUUID(input.ToolsetSnapshotID),
		RunID:          run.ID,
		AgentSessionID: session.ID,
		Tools:          toolsetToolsToDomain(toolset.Tools),
		CreatedReason:  toolset.CreatedReason,
		CreatedAt:      time.Now().UTC(),
	}
	if err := repo.CreateToolsetSnapshot(toolsetSnapshot); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "prompt_service.create_toolset_snapshot", err)
	}

	promptPayload, err := marshalPayload("prompt_service.prompt_snapshot_payload", map[string]interface{}{
		"prompt_snapshot_id": promptSnapshot.ID,
		"hash":               promptSnapshot.CombinedPromptHash,
		"run_id":             run.ID,
		"work_unit_id":       wu.ID,
		"agent_session_id":   session.ID,
		"system_prompt_hash": promptSnapshot.SystemPromptHash,
		"task_prompt_hash":   promptSnapshot.TaskPromptHash,
		"composition_hash":   promptSnapshot.CompositionHash,
		"category_signature": promptSnapshot.CategorySignature,
		"fragment_count":     len(promptSnapshot.FragmentRefs),
		"reused":             reusedPromptSnapshot,
		"count_used":         promptSnapshot.CountUsed,
	})
	if err != nil {
		return nil, err
	}
	if _, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.PromptSnapshotEventID,
		Type:        "prompt.snapshot_created",
		Version:     eventVersionV1,
		TaskID:      task.ID,
		RunID:       run.ID,
		WorkUnitID:  wu.ID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     promptPayload,
	}); err != nil {
		return nil, err
	}

	toolsetPayload, err := marshalPayload("prompt_service.toolset_snapshot_payload", map[string]interface{}{
		"toolset_snapshot_id": toolsetSnapshot.ID,
		"agent_session_id":    session.ID,
		"run_id":              run.ID,
		"tool_count":          len(toolsetSnapshot.Tools),
	})
	if err != nil {
		return nil, err
	}
	if _, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.ToolsetSnapshotEventID,
		Type:        "toolset.snapshot_created",
		Version:     eventVersionV1,
		TaskID:      task.ID,
		RunID:       run.ID,
		WorkUnitID:  wu.ID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     toolsetPayload,
	}); err != nil {
		return nil, err
	}

	if err := commitTx(tx, "prompt_service.commit_prepare"); err != nil {
		return nil, err
	}

	return &PreparedRunPrompt{
		PromptSnapshot:  promptSnapshot,
		ToolsetSnapshot: toolsetSnapshot,
		SystemPrompt:    composed.SystemPrompt,
		TaskPrompt:      composed.TaskPrompt,
		CombinedPrompt:  composed.CombinedPrompt,
		PromptHash:      composed.CombinedPromptHash,
		Toolset:         prompting.ToolNames(toolset.Tools),
	}, nil
}

func promptFragmentToDomain(fragment prompting.Fragment) (*domain.PromptFragment, error) {
	appliesWhen, err := json.Marshal(fragment.AppliesWhen)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, "prompt_service.fragment_applies_when", err)
	}
	return &domain.PromptFragment{
		ID:               fragment.ID,
		Version:          fragment.Version,
		Category:         string(fragment.Category),
		Kind:             string(fragment.Kind),
		Title:            fragment.Title,
		Priority:         fragment.Priority,
		ExclusiveGroup:   fragment.ExclusiveGroup,
		BodyHash:         fragment.BodyHash,
		MetadataHash:     fragment.MetadataHash,
		Body:             fragment.Body,
		AppliesWhen:      appliesWhen,
		Requires:         fragment.Requires,
		ConflictsWith:    fragment.ConflictsWith,
		Allows:           fragment.Allows,
		Denies:           fragment.Denies,
		ApprovalRequired: fragment.ApprovalRequired,
		AutonomyLevel:    fragment.AutonomyLevel,
	}, nil
}

func promptFragmentRefsToDomain(refs []prompting.FragmentRef) []domain.PromptFragmentRef {
	out := make([]domain.PromptFragmentRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, domain.PromptFragmentRef{
			ID:           ref.ID,
			Version:      ref.Version,
			Category:     string(ref.Category),
			Kind:         string(ref.Kind),
			Order:        ref.Order,
			BodyHash:     ref.BodyHash,
			MetadataHash: ref.MetadataHash,
			Title:        ref.Title,
		})
	}
	return out
}

func toolsetToolsToDomain(tools []prompting.Tool) []domain.ToolsetTool {
	out := make([]domain.ToolsetTool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, domain.ToolsetTool{
			Name:   tool.Name,
			Scope:  tool.Scope,
			Risk:   string(tool.Risk),
			Reason: tool.Reason,
		})
	}
	return out
}

func valueOrNewUUID(value string) string {
	if value != "" {
		return value
	}
	return uuid.New().String()
}
