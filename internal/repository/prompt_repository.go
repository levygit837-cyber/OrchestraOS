package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type PromptRepository struct {
	db DBTX
}

func NewPromptRepository(db DBTX) *PromptRepository {
	return &PromptRepository{db: db}
}

func (r *PromptRepository) CreateOrVerifyFragment(fragment *domain.PromptFragment) error {
	existing, err := r.GetFragment(fragment.ID, fragment.Version)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.MetadataHash != fragment.MetadataHash {
			return fmt.Errorf("prompt fragment %s@%s already exists with metadata hash %s, got %s", fragment.ID, fragment.Version, existing.MetadataHash, fragment.MetadataHash)
		}
		return nil
	}

	now := time.Now().UTC()
	if fragment.CreatedAt.IsZero() {
		fragment.CreatedAt = now
	}
	if fragment.UpdatedAt.IsZero() {
		fragment.UpdatedAt = fragment.CreatedAt
	}

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
		db.QueryPromptFragmentInsert,
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

func (r *PromptRepository) GetFragment(id, version string) (*domain.PromptFragment, error) {
	row := r.db.QueryRow(db.QueryPromptFragmentGetByIDVersion, id, version)
	return r.scanPromptFragment(row)
}

func (r *PromptRepository) CreatePromptSnapshot(snapshot *domain.PromptSnapshot) error {
	_, err := r.CreateOrReferencePromptSnapshot(snapshot)
	return err
}

func (r *PromptRepository) CreateOrReferencePromptSnapshot(snapshot *domain.PromptSnapshot) (bool, error) {
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	if snapshot.FirstUsedAt.IsZero() {
		snapshot.FirstUsedAt = snapshot.CreatedAt
	}
	if snapshot.LastUsedAt.IsZero() {
		snapshot.LastUsedAt = snapshot.CreatedAt
	}

	fragmentRefs := snapshot.FragmentRefs
	if fragmentRefs == nil {
		fragmentRefs = []domain.PromptFragmentRef{}
	}
	fragmentRefsJSON, err := json.Marshal(fragmentRefs)
	if err != nil {
		return false, fmt.Errorf("marshal prompt snapshot fragment refs: %w", err)
	}
	assemblyOrder := snapshot.AssemblyOrder
	if assemblyOrder == nil {
		assemblyOrder = []string{}
	}
	assemblyOrderJSON, err := json.Marshal(assemblyOrder)
	if err != nil {
		return false, fmt.Errorf("marshal prompt snapshot assembly order: %w", err)
	}
	variablesApplied := jsonOrEmptyObject(snapshot.VariablesApplied)

	requestedID := snapshot.ID
	row := r.db.QueryRow(
		db.QueryPromptSnapshotInsert,
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
		return false, fmt.Errorf("failed to create or reference prompt snapshot: %w", err)
	}
	*snapshot = *persisted
	return snapshot.ID != requestedID || snapshot.CountUsed > 1, nil
}

func (r *PromptRepository) GetPromptSnapshot(id string) (*domain.PromptSnapshot, error) {
	row := r.db.QueryRow(db.QueryPromptSnapshotGetByID, id)
	return r.scanPromptSnapshot(row)
}

func (r *PromptRepository) LatestPromptSnapshotByRun(runID string) (*domain.PromptSnapshot, error) {
	row := r.db.QueryRow(db.QueryPromptSnapshotLatestByRun, runID)
	return r.scanPromptSnapshot(row)
}

func (r *PromptRepository) CreateToolsetSnapshot(snapshot *domain.ToolsetSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = uuid.New().String()
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	tools := snapshot.Tools
	if tools == nil {
		tools = []domain.ToolsetTool{}
	}
	toolsJSON, err := json.Marshal(tools)
	if err != nil {
		return fmt.Errorf("marshal toolset snapshot tools: %w", err)
	}
	_, err = r.db.Exec(
		db.QueryToolsetSnapshotInsert,
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

func (r *PromptRepository) GetToolsetSnapshot(id string) (*domain.ToolsetSnapshot, error) {
	row := r.db.QueryRow(db.QueryToolsetSnapshotGetByID, id)
	return r.scanToolsetSnapshot(row)
}

func (r *PromptRepository) LatestToolsetSnapshotByAgentSession(agentSessionID string) (*domain.ToolsetSnapshot, error) {
	row := r.db.QueryRow(db.QueryToolsetSnapshotLatestByAgentSession, agentSessionID)
	return r.scanToolsetSnapshot(row)
}

func (r *PromptRepository) scanPromptFragment(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.PromptFragment, error) {
	var fragment domain.PromptFragment
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

func (r *PromptRepository) scanPromptSnapshot(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.PromptSnapshot, error) {
	var snapshot domain.PromptSnapshot
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

func (r *PromptRepository) scanToolsetSnapshot(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.ToolsetSnapshot, error) {
	var snapshot domain.ToolsetSnapshot
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
