// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package prompt

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// Prompt Fragment CRUD
// ============================================================================

func (r *Repository) CreateOrVerifyFragment(fragment *PromptFragment) error {
	appliesWhen := jsonOrEmptyObject(fragment.AppliesWhen)
	requires, err := marshalStringList(fragment.Requires)
	if err != nil {
		return fmt.Errorf("marshal fragment requires: %w", err)
	}
	conflictsWith, err := marshalStringList(fragment.ConflictsWith)
	if err != nil {
		return fmt.Errorf("marshal fragment conflicts: %w", err)
	}
	allows, err := marshalStringList(fragment.Allows)
	if err != nil {
		return fmt.Errorf("marshal fragment allows: %w", err)
	}
	denies, err := marshalStringList(fragment.Denies)
	if err != nil {
		return fmt.Errorf("marshal fragment denies: %w", err)
	}
	approvalRequired, err := marshalStringList(fragment.ApprovalRequired)
	if err != nil {
		return fmt.Errorf("marshal fragment approval required: %w", err)
	}

	_, err = r.db.Exec(
		QueryFragmentInsert,
		fragment.ID,
		fragment.Version,
		fragment.Category,
		fragment.Kind,
		fragment.Title,
		fragment.Priority,
		fragment.ExclusiveGroup,
		fragment.BodyHash,
		fragment.MetadataHash,
		fragment.Body,
		appliesWhen,
		requires,
		conflictsWith,
		allows,
		denies,
		approvalRequired,
		fragment.AutonomyLevel,
		fragment.CreatedAt,
		fragment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create prompt fragment: %w", err)
	}
	return nil
}

func (r *Repository) GetFragment(id, version string) (*PromptFragment, error) {
	row := r.db.QueryRow(QueryFragmentGetByIDVersion, id, version)
	return r.scanPromptFragment(row)
}

func (r *Repository) scanPromptFragment(scanner interface {
	Scan(dest ...interface{}) error
}) (*PromptFragment, error) {
	var fragment PromptFragment
	var appliesWhen, requires, conflictsWith, allows, denies, approvalRequired []byte

	err := scanner.Scan(
		&fragment.ID,
		&fragment.Version,
		&fragment.Category,
		&fragment.Kind,
		&fragment.Title,
		&fragment.Priority,
		&fragment.ExclusiveGroup,
		&fragment.BodyHash,
		&fragment.MetadataHash,
		&fragment.Body,
		&appliesWhen,
		&requires,
		&conflictsWith,
		&allows,
		&denies,
		&approvalRequired,
		&fragment.AutonomyLevel,
		&fragment.CreatedAt,
		&fragment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan prompt fragment: %w", err)
	}
	fragment.AppliesWhen = json.RawMessage(appliesWhen)
	if err := json.Unmarshal(requires, &fragment.Requires); err != nil {
		return nil, fmt.Errorf("unmarshal fragment requires: %w", err)
	}
	if err := json.Unmarshal(conflictsWith, &fragment.ConflictsWith); err != nil {
		return nil, fmt.Errorf("unmarshal fragment conflicts: %w", err)
	}
	if err := json.Unmarshal(allows, &fragment.Allows); err != nil {
		return nil, fmt.Errorf("unmarshal fragment allows: %w", err)
	}
	if err := json.Unmarshal(denies, &fragment.Denies); err != nil {
		return nil, fmt.Errorf("unmarshal fragment denies: %w", err)
	}
	if err := json.Unmarshal(approvalRequired, &fragment.ApprovalRequired); err != nil {
		return nil, fmt.Errorf("unmarshal fragment approval required: %w", err)
	}
	return &fragment, nil
}

// ============================================================================
// Prompt Snapshot CRUD
// ============================================================================

func (r *Repository) CreatePromptSnapshot(snapshot *PromptSnapshot) error {
	return r.CreateOrReferencePromptSnapshot(snapshot)
}

func (r *Repository) CreateOrReferencePromptSnapshot(snapshot *PromptSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	if snapshot.FirstUsedAt.IsZero() {
		snapshot.FirstUsedAt = snapshot.CreatedAt
	}
	if snapshot.LastUsedAt.IsZero() {
		snapshot.LastUsedAt = snapshot.CreatedAt
	}

	fragmentRefs := snapshot.FragmentRefs
	if fragmentRefs == nil {
		fragmentRefs = []PromptFragmentRef{}
	}
	fragmentRefsJSON, err := json.Marshal(fragmentRefs)
	if err != nil {
		return fmt.Errorf("marshal prompt snapshot fragment refs: %w", err)
	}
	assemblyOrder := snapshot.AssemblyOrder
	if assemblyOrder == nil {
		assemblyOrder = []string{}
	}
	assemblyOrderJSON, err := json.Marshal(assemblyOrder)
	if err != nil {
		return fmt.Errorf("marshal prompt snapshot assembly order: %w", err)
	}
	variablesApplied := jsonOrEmptyObject(snapshot.VariablesApplied)

	row := r.db.QueryRow(
		QuerySnapshotInsert,
		snapshot.ID,
		snapshot.RunID,
		snapshot.WorkUnitID,
		snapshot.AgentSessionID,
		snapshot.SystemPrompt,
		snapshot.TaskPrompt,
		snapshot.CombinedPrompt,
		snapshot.SystemPromptHash,
		snapshot.TaskPromptHash,
		snapshot.CombinedPromptHash,
		snapshot.CompositionHash,
		snapshot.CategorySignature,
		fragmentRefsJSON,
		assemblyOrderJSON,
		variablesApplied,
		snapshot.LastUsedAt,
	)
	persisted, err := r.scanPromptSnapshot(row)
	if err != nil {
		return fmt.Errorf("failed to create or reference prompt snapshot: %w", err)
	}
	*snapshot = *persisted
	return nil
}

func (r *Repository) GetPromptSnapshot(id string) (*PromptSnapshot, error) {
	row := r.db.QueryRow(QuerySnapshotGetByID, id)
	return r.scanPromptSnapshot(row)
}

func (r *Repository) GetLatestPromptSnapshotByRun(runID string) (*PromptSnapshot, error) {
	row := r.db.QueryRow(QuerySnapshotLatestByRun, runID)
	return r.scanPromptSnapshot(row)
}

func (r *Repository) scanPromptSnapshot(scanner interface {
	Scan(dest ...interface{}) error
}) (*PromptSnapshot, error) {
	var snapshot PromptSnapshot
	var fragmentRefs, assemblyOrder, variablesApplied []byte
	err := scanner.Scan(
		&snapshot.ID,
		&snapshot.RunID,
		&snapshot.WorkUnitID,
		&snapshot.AgentSessionID,
		&snapshot.SystemPrompt,
		&snapshot.TaskPrompt,
		&snapshot.CombinedPrompt,
		&snapshot.SystemPromptHash,
		&snapshot.TaskPromptHash,
		&snapshot.CombinedPromptHash,
		&snapshot.CompositionHash,
		&snapshot.CategorySignature,
		&fragmentRefs,
		&assemblyOrder,
		&variablesApplied,
		&snapshot.CountUsed,
		&snapshot.FirstUsedAt,
		&snapshot.LastUsedAt,
		&snapshot.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan prompt snapshot: %w", err)
	}
	if err := json.Unmarshal(fragmentRefs, &snapshot.FragmentRefs); err != nil {
		return nil, fmt.Errorf("unmarshal prompt snapshot fragment refs: %w", err)
	}
	if err := json.Unmarshal(assemblyOrder, &snapshot.AssemblyOrder); err != nil {
		return nil, fmt.Errorf("unmarshal prompt snapshot assembly order: %w", err)
	}
	snapshot.VariablesApplied = json.RawMessage(variablesApplied)
	return &snapshot, nil
}

// ============================================================================
// Toolset Snapshot CRUD
// ============================================================================

func (r *Repository) CreateToolsetSnapshot(snapshot *ToolsetSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}

	tools := snapshot.Tools
	if tools == nil {
		tools = []ToolsetTool{}
	}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		return fmt.Errorf("marshal toolset snapshot tools: %w", err)
	}
	_, err = r.db.Exec(
		QueryToolsetInsert,
		snapshot.ID,
		snapshot.RunID,
		snapshot.AgentSessionID,
		toolsJSON,
		snapshot.CreatedReason,
		snapshot.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create toolset snapshot: %w", err)
	}
	return nil
}

func (r *Repository) GetToolsetSnapshot(id string) (*ToolsetSnapshot, error) {
	row := r.db.QueryRow(QueryToolsetGetByID, id)
	return r.scanToolsetSnapshot(row)
}

func (r *Repository) GetLatestToolsetSnapshotByAgentSession(agentSessionID string) (*ToolsetSnapshot, error) {
	row := r.db.QueryRow(QueryToolsetLatestByAgentSession, agentSessionID)
	return r.scanToolsetSnapshot(row)
}

func (r *Repository) scanToolsetSnapshot(scanner interface {
	Scan(dest ...interface{}) error
}) (*ToolsetSnapshot, error) {
	var snapshot ToolsetSnapshot
	var tools []byte
	err := scanner.Scan(
		&snapshot.ID,
		&snapshot.RunID,
		&snapshot.AgentSessionID,
		&tools,
		&snapshot.CreatedReason,
		&snapshot.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan toolset snapshot: %w", err)
	}
	if err := json.Unmarshal(tools, &snapshot.Tools); err != nil {
		return nil, fmt.Errorf("unmarshal toolset snapshot tools: %w", err)
	}
	return &snapshot, nil
}

// ============================================================================
// Helpers
// ============================================================================

func jsonOrEmptyObject(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return raw
}

func marshalStringList(values []string) ([]byte, error) {
	if values == nil {
		values = []string{}
	}
	return json.Marshal(values)
}
