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

	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateOrVerifyFragment(fragment *PromptFragment) error {
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
