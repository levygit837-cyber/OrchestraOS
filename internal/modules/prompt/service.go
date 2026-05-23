// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package prompt

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

func NewPromptService(database *sql.DB) *PromptService {
	return &PromptService{db: database}
}

// SelectToolset selects the minimum toolset for a given agent profile.
func (s *PromptService) SelectToolset(profile string) (domain.ToolsetSelection, error) {
	selection, err := SelectToolset(profile)
	if err != nil {
		return domain.ToolsetSelection{}, err
	}
	tools := make([]domain.ToolsetTool, 0, len(selection.Tools))
	for _, t := range selection.Tools {
		tools = append(tools, domain.ToolsetTool{
			Name:   t.Name,
			Scope:  t.Scope,
			Risk:   string(t.Risk),
			Reason: t.Reason,
		})
	}
	return domain.ToolsetSelection{
		Profile:       selection.Profile,
		Tools:         tools,
		CreatedReason: selection.CreatedReason,
	}, nil
}

// PreparePrompt selects a toolset, composes a prompt, and persists it in one operation.
func (s *PromptService) PreparePrompt(ctx context.Context, input domain.PromptComposeInput, metadata domain.PersistMetadata) (*domain.PreparedRunPrompt, error) {
	composed, err := s.Compose(ctx, input)
	if err != nil {
		return nil, err
	}
	return s.PersistComposedPrompt(ctx, composed, metadata)
}

// Compose builds a composed prompt from the given input.
func (s *PromptService) Compose(ctx context.Context, input domain.PromptComposeInput) (*domain.ComposedPrompt, error) {
	composed, err := Compose(TaskContext{
		TaskID:             input.TaskID,
		TaskTitle:          input.TaskTitle,
		TaskDescription:    input.TaskDescription,
		RunID:              input.RunID,
		WorkUnitID:         input.WorkUnitID,
		TaskGraphID:        input.TaskGraphID,
		WorkUnitTitle:      input.WorkUnitTitle,
		WorkUnitObjective:  input.WorkUnitObjective,
		AgentProfile:       input.AgentProfile,
		OwnedPaths:         input.OwnedPaths,
		ReadPaths:          input.ReadPaths,
		DependsOn:          input.DependsOn,
		AcceptanceCriteria: input.AcceptanceCriteria,
		ValidationPlan:     input.ValidationPlan,
		Toolset: func() ToolsetSelection {
			tools := make([]Tool, 0, len(input.Toolset.Tools))
			for _, t := range input.Toolset.Tools {
				tools = append(tools, Tool{
					Name:   t.Name,
					Scope:  t.Scope,
					Risk:   ToolRisk(t.Risk),
					Reason: t.Reason,
				})
			}
			return ToolsetSelection{
				Profile:       input.Toolset.Profile,
				Tools:         tools,
				CreatedReason: input.Toolset.CreatedReason,
			}
		}(),
	})
	if err != nil {
		return nil, err
	}
	fragments := make([]domain.PromptFragment, 0, len(composed.Fragments))
	for _, f := range composed.Fragments {
		fragments = append(fragments, domain.PromptFragment{
			ID:               f.ID,
			Version:          f.Version,
			Category:         string(f.Category),
			Kind:             string(f.Kind),
			Title:            f.Title,
			Priority:         f.Priority,
			ExclusiveGroup:   f.ExclusiveGroup,
			BodyHash:         f.BodyHash,
			MetadataHash:     f.MetadataHash,
			Body:             f.Body,
			Requires:         f.Requires,
			ConflictsWith:    f.ConflictsWith,
			Allows:           f.Allows,
			Denies:           f.Denies,
			ApprovalRequired: f.ApprovalRequired,
			AutonomyLevel:    f.AutonomyLevel,
		})
	}
	fragmentRefs := make([]domain.PromptFragmentRef, 0, len(composed.FragmentRefs))
	for _, ref := range composed.FragmentRefs {
		fragmentRefs = append(fragmentRefs, domain.PromptFragmentRef{
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
	return &domain.ComposedPrompt{
		SystemPrompt:       composed.SystemPrompt,
		TaskPrompt:         composed.TaskPrompt,
		CombinedPrompt:     composed.CombinedPrompt,
		SystemPromptHash:   composed.SystemPromptHash,
		TaskPromptHash:     composed.TaskPromptHash,
		CombinedPromptHash: composed.CombinedPromptHash,
		CompositionHash:    composed.CompositionHash,
		CategorySignature:  composed.CategorySignature,
		SystemProfile: func() domain.SystemProfile {
			categories := make(map[string]domain.PromptFragmentRef, len(composed.SystemProfile.Categories))
			for k, v := range composed.SystemProfile.Categories {
				categories[k] = domain.PromptFragmentRef{
					ID:           v.ID,
					Version:      v.Version,
					Category:     string(v.Category),
					Kind:         string(v.Kind),
					Order:        v.Order,
					BodyHash:     v.BodyHash,
					MetadataHash: v.MetadataHash,
					Title:        v.Title,
				}
			}
			return domain.SystemProfile{
				Persona:               composed.SystemProfile.Persona,
				OperatingMode:         composed.SystemProfile.OperatingMode,
				TechnicalDomain:       composed.SystemProfile.TechnicalDomain,
				OutputContract:        composed.SystemProfile.OutputContract,
				ToolNames:             composed.SystemProfile.ToolNames,
				Allows:                composed.SystemProfile.Allows,
				Denies:                composed.SystemProfile.Denies,
				ApprovalRequired:      composed.SystemProfile.ApprovalRequired,
				Categories:            categories,
				CategorySignature:     composed.SystemProfile.CategorySignature,
				TaskExecutionFocus:    composed.SystemProfile.TaskExecutionFocus,
				CanonicalAgentProfile: composed.SystemProfile.CanonicalAgentProfile,
			}
		}(),
		Fragments:        fragments,
		FragmentRefs:     fragmentRefs,
		AssemblyOrder:    composed.AssemblyOrder,
		VariablesApplied: composed.VariablesApplied,
		Toolset: func() domain.ToolsetSelection {
			tools := make([]domain.ToolsetTool, 0, len(composed.Toolset.Tools))
			for _, t := range composed.Toolset.Tools {
				tools = append(tools, domain.ToolsetTool{
					Name:   t.Name,
					Scope:  t.Scope,
					Risk:   string(t.Risk),
					Reason: t.Reason,
				})
			}
			return domain.ToolsetSelection{
				Profile:       composed.Toolset.Profile,
				Tools:         tools,
				CreatedReason: composed.Toolset.CreatedReason,
			}
		}(),
	}, nil
}

func (s *PromptService) PersistComposedPrompt(ctx context.Context, composed *domain.ComposedPrompt, metadata domain.PersistMetadata) (*domain.PreparedRunPrompt, error) {
	const op = "prompt_service.persist_composed"

	tx, err := dbcore.BeginTx(ctx, s.db, "prompt_service.persist_composed")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	repo := NewRepository(tx)
	for _, fragment := range composed.Fragments {
		appliesWhen, err := json.Marshal(fragment.AppliesWhen)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodeValidation, "prompt_service.fragment_applies_when", err)
		}
		localFragment := &PromptFragment{
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
		}
		existing, err := repo.GetFragment(localFragment.ID, localFragment.Version)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "prompt_service.verify_fragment", err)
		}
		if existing != nil && existing.MetadataHash != localFragment.MetadataHash {
			return nil, apperrors.New(apperrors.CodeConflict, "prompt_service.verify_fragment",
				fmt.Sprintf("prompt fragment %s@%s already exists with metadata hash %s, got %s",
					localFragment.ID, localFragment.Version, existing.MetadataHash, localFragment.MetadataHash))
		}
		if existing == nil {
			if err := repo.CreateOrVerifyFragment(localFragment); err != nil {
				return nil, apperrors.Wrap(apperrors.CodeConflict, "prompt_service.persist_fragment", err)
			}
		}
	}

	variablesApplied, err := MarshalVariables(composed.VariablesApplied)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op, err)
	}
	fragmentRefs := make([]PromptFragmentRef, 0, len(composed.FragmentRefs))
	for _, ref := range composed.FragmentRefs {
		fragmentRefs = append(fragmentRefs, PromptFragmentRef{
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
	promptSnapshot := &PromptSnapshot{
		ID:                 valueOrNewUUID(metadata.PromptSnapshotID),
		RunID:              metadata.RunID,
		WorkUnitID:         metadata.WorkUnitID,
		AgentSessionID:     metadata.AgentSessionID,
		SystemPrompt:       composed.SystemPrompt,
		TaskPrompt:         composed.TaskPrompt,
		CombinedPrompt:     composed.CombinedPrompt,
		SystemPromptHash:   composed.SystemPromptHash,
		TaskPromptHash:     composed.TaskPromptHash,
		CombinedPromptHash: composed.CombinedPromptHash,
		CompositionHash:    composed.CompositionHash,
		CategorySignature:  composed.CategorySignature,
		FragmentRefs:       fragmentRefs,
		AssemblyOrder:      composed.AssemblyOrder,
		VariablesApplied:   variablesApplied,
		CreatedAt:          time.Now().UTC(),
	}
	if err := repo.CreateOrReferencePromptSnapshot(promptSnapshot); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "prompt_service.create_prompt_snapshot", err)
	}

	toolset := composed.Toolset
	tools := make([]ToolsetTool, 0, len(toolset.Tools))
	for _, t := range toolset.Tools {
		tools = append(tools, ToolsetTool{
			Name:   t.Name,
			Scope:  t.Scope,
			Risk:   string(t.Risk),
			Reason: t.Reason,
		})
	}
	toolsetSnapshot := &ToolsetSnapshot{
		ID:             valueOrNewUUID(metadata.ToolsetSnapshotID),
		RunID:          metadata.RunID,
		AgentSessionID: metadata.AgentSessionID,
		Tools:          tools,
		CreatedReason:  toolset.CreatedReason,
		CreatedAt:      time.Now().UTC(),
	}
	if err := repo.CreateToolsetSnapshot(toolsetSnapshot); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "prompt_service.create_toolset_snapshot", err)
	}

	promptPayload, err := serialization.MarshalPayload("prompt_service.prompt_snapshot_payload", map[string]interface{}{
		"prompt_snapshot_id": promptSnapshot.ID,
		"hash":               promptSnapshot.CombinedPromptHash,
		"run_id":             metadata.RunID,
		"work_unit_id":       metadata.WorkUnitID,
		"agent_session_id":   metadata.AgentSessionID,
		"system_prompt_hash": promptSnapshot.SystemPromptHash,
		"task_prompt_hash":   promptSnapshot.TaskPromptHash,
		"composition_hash":   promptSnapshot.CompositionHash,
		"category_signature": promptSnapshot.CategorySignature,
		"fragment_count":     len(promptSnapshot.FragmentRefs),
		"reused":             false,
		"count_used":         promptSnapshot.CountUsed,
	})
	if err != nil {
		return nil, err
	}
	if _, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          metadata.PromptSnapshotEventID,
		Type:        "prompt.snapshot_created",
		Version:     transition.EventVersionV1,
		TaskID:      metadata.TaskID,
		RunID:       metadata.RunID,
		WorkUnitID:  metadata.WorkUnitID,
		AgentID:     metadata.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     promptPayload,
	}); err != nil {
		return nil, err
	}

	toolsetPayload, err := serialization.MarshalPayload("prompt_service.toolset_snapshot_payload", map[string]interface{}{
		"toolset_snapshot_id": toolsetSnapshot.ID,
		"agent_session_id":    metadata.AgentSessionID,
		"run_id":              metadata.RunID,
		"tool_count":          len(toolsetSnapshot.Tools),
	})
	if err != nil {
		return nil, err
	}
	if _, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          metadata.ToolsetSnapshotEventID,
		Type:        "toolset.snapshot_created",
		Version:     transition.EventVersionV1,
		TaskID:      metadata.TaskID,
		RunID:       metadata.RunID,
		WorkUnitID:  metadata.WorkUnitID,
		AgentID:     metadata.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     toolsetPayload,
	}); err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "prompt_service.commit_prepare"); err != nil {
		return nil, err
	}

	return &domain.PreparedRunPrompt{
		PromptSnapshot: &domain.PromptSnapshot{
			ID:                 promptSnapshot.ID,
			RunID:              promptSnapshot.RunID,
			WorkUnitID:         promptSnapshot.WorkUnitID,
			AgentSessionID:     promptSnapshot.AgentSessionID,
			SystemPrompt:       promptSnapshot.SystemPrompt,
			TaskPrompt:         promptSnapshot.TaskPrompt,
			CombinedPrompt:     promptSnapshot.CombinedPrompt,
			SystemPromptHash:   promptSnapshot.SystemPromptHash,
			TaskPromptHash:     promptSnapshot.TaskPromptHash,
			CombinedPromptHash: promptSnapshot.CombinedPromptHash,
			CompositionHash:    promptSnapshot.CompositionHash,
			CategorySignature:  promptSnapshot.CategorySignature,
			FragmentRefs:       toDomainFragmentRefs(promptSnapshot.FragmentRefs),
			AssemblyOrder:      promptSnapshot.AssemblyOrder,
			VariablesApplied:   promptSnapshot.VariablesApplied,
			CountUsed:          promptSnapshot.CountUsed,
			FirstUsedAt:        promptSnapshot.FirstUsedAt,
			LastUsedAt:         promptSnapshot.LastUsedAt,
			CreatedAt:          promptSnapshot.CreatedAt,
		},
		ToolsetSnapshot: &domain.ToolsetSnapshot{
			ID:             toolsetSnapshot.ID,
			RunID:          toolsetSnapshot.RunID,
			AgentSessionID: toolsetSnapshot.AgentSessionID,
			Tools: func() []domain.ToolsetTool {
				out := make([]domain.ToolsetTool, 0, len(toolsetSnapshot.Tools))
				for _, t := range toolsetSnapshot.Tools {
					out = append(out, domain.ToolsetTool{
						Name:   t.Name,
						Scope:  t.Scope,
						Risk:   string(t.Risk),
						Reason: t.Reason,
					})
				}
				return out
			}(),
			CreatedReason: toolsetSnapshot.CreatedReason,
			CreatedAt:     toolsetSnapshot.CreatedAt,
		},
		SystemPrompt:   composed.SystemPrompt,
		TaskPrompt:     composed.TaskPrompt,
		CombinedPrompt: composed.CombinedPrompt,
		PromptHash:     composed.CombinedPromptHash,
		Toolset:        toolNamesFromDomain(toolset.Tools),
	}, nil
}

func toDomainFragmentRefs(refs []PromptFragmentRef) []domain.PromptFragmentRef {
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

func toolNamesFromDomain(tools []domain.ToolsetTool) []string {
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name)
	}
	return names
}

func valueOrNewUUID(value string) string {
	if value != "" {
		return value
	}
	return uuid.New().String()
}
