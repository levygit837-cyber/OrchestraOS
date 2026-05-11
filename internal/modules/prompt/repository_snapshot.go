// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package prompt

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func (r *Repository) CreatePromptSnapshot(snapshot *domain.PromptSnapshot) error {
	_, err := r.CreateOrReferencePromptSnapshot(snapshot)
	return err
}

func (r *Repository) CreateOrReferencePromptSnapshot(snapshot *domain.PromptSnapshot) (bool, error) {
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
		return false, fmt.Errorf("failed to create or reference prompt snapshot: %w", err)
	}
	*snapshot = *persisted
	return snapshot.ID != requestedID || snapshot.CountUsed > 1, nil
}

func (r *Repository) GetPromptSnapshot(id string) (*domain.PromptSnapshot, error) {
	row := r.db.QueryRow(QuerySnapshotGetByID, id)
	return r.scanPromptSnapshot(row)
}

func (r *Repository) LatestPromptSnapshotByRun(runID string) (*domain.PromptSnapshot, error) {
	row := r.db.QueryRow(QuerySnapshotLatestByRun, runID)
	return r.scanPromptSnapshot(row)
}

func (r *Repository) CreateToolsetSnapshot(snapshot *domain.ToolsetSnapshot) error {
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

func (r *Repository) GetToolsetSnapshot(id string) (*domain.ToolsetSnapshot, error) {
	row := r.db.QueryRow(QueryToolsetGetByID, id)
	return r.scanToolsetSnapshot(row)
}

func (r *Repository) LatestToolsetSnapshotByAgentSession(agentSessionID string) (*domain.ToolsetSnapshot, error) {
	row := r.db.QueryRow(QueryToolsetLatestByAgentSession, agentSessionID)
	return r.scanToolsetSnapshot(row)
}

func (r *Repository) scanPromptSnapshot(scanner interface {
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

func (r *Repository) scanToolsetSnapshot(scanner interface {
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
