// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package prompt

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
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

type PrepareAndPersistInput struct {
	Run *domain.Run
	// TODO[ADR-0022]: migrar para *workunit.WorkUnit
	WorkUnit               *domain.WorkUnit
	Task                   *domain.Task
	Session                *domain.AgentSession
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

func (s *PromptService) PrepareAndPersistPrompt(ctx context.Context, tx *sql.Tx, input PrepareAndPersistInput) (*PreparedRunPrompt, error) {
	const op = "prompt_service.prepare_and_persist"
	run := input.Run
	wu := input.WorkUnit
	task := input.Task
	session := input.Session

	toolset, err := SelectToolset(wu.AssignedAgentProfile)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	}
	composed, err := Compose(TaskContext{
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

	repo := NewRepository(tx)
	for _, fragment := range composed.Fragments {
		domainFragment, err := promptFragmentToDomain(fragment)
		if err != nil {
			return nil, err
		}
		if err := repo.CreateOrVerifyFragment(domainFragment); err != nil {
			return nil, apperrors.Wrap(apperrors.CodeConflict, "prompt_service.persist_fragment", err)
		}
	}

	variablesApplied, err := MarshalVariables(composed.VariablesApplied)
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

	promptPayload, err := serialization.MarshalPayload("prompt_service.prompt_snapshot_payload", map[string]interface{}{
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
	if _, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.PromptSnapshotEventID,
		Type:        "prompt.snapshot_created",
		Version:     transition.EventVersionV1,
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

	toolsetPayload, err := serialization.MarshalPayload("prompt_service.toolset_snapshot_payload", map[string]interface{}{
		"toolset_snapshot_id": toolsetSnapshot.ID,
		"agent_session_id":    session.ID,
		"run_id":              run.ID,
		"tool_count":          len(toolsetSnapshot.Tools),
	})
	if err != nil {
		return nil, err
	}
	if _, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.ToolsetSnapshotEventID,
		Type:        "toolset.snapshot_created",
		Version:     transition.EventVersionV1,
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

	if err := dbcore.CommitTx(tx, "prompt_service.commit_prepare"); err != nil {
		return nil, err
	}

	return &PreparedRunPrompt{
		PromptSnapshot:  promptSnapshot,
		ToolsetSnapshot: toolsetSnapshot,
		SystemPrompt:    composed.SystemPrompt,
		TaskPrompt:      composed.TaskPrompt,
		CombinedPrompt:  composed.CombinedPrompt,
		PromptHash:      composed.CombinedPromptHash,
		Toolset:         ToolNames(toolset.Tools),
	}, nil
}

func promptFragmentToDomain(fragment Fragment) (*domain.PromptFragment, error) {
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

func promptFragmentRefsToDomain(refs []FragmentRef) []domain.PromptFragmentRef {
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

func toolsetToolsToDomain(tools []Tool) []domain.ToolsetTool {
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
