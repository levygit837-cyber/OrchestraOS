package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Repository handles review persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new review repository
func NewRepository(database db.DBTX) *Repository {
	return &Repository{db: database}
}

// Create inserts a new review
func (r *Repository) Create(review *domain.Review) error {
	if review.ID == "" {
		review.ID = uuid.New().String()
	}

	now := time.Now().UTC()
	review.CreatedAt = now
	review.UpdatedAt = now

	criteriaChecked, err := json.Marshal(review.CriteriaChecked)
	if err != nil {
		return fmt.Errorf("failed to marshal criteria_checked: %w", err)
	}

	_, err = r.db.Exec(
		QueryInsert,
		review.ID,
		sqlNullString(review.RunID),
		sqlNullString(review.WorkUnitID),
		sqlNullString(review.TaskID),
		sqlNullString(review.AgentSessionID),
		sqlNullString(review.ReviewerAgentID),
		review.GateType,
		review.Status,
		sqlNullStringPtr(&review.VerdictReason),
		pqArray(review.EvidenceRefs),
		criteriaChecked,
		review.CreatedAt,
		review.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

// GetByID retrieves a review by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	row := r.db.QueryRowContext(ctx, QueryGetByID, id)
	return r.scanReview(row)
}

// ListByTask retrieves all reviews for a task
func (r *Repository) ListByTask(ctx context.Context, taskID string) ([]*domain.Review, error) {
	rows, err := r.db.QueryContext(ctx, QueryListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews by task: %w", err)
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		review, err := r.scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

// ListPending retrieves all pending or in_progress reviews
func (r *Repository) ListPending(ctx context.Context) ([]*domain.Review, error) {
	rows, err := r.db.QueryContext(ctx, QueryListPending)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		review, err := r.scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

// ExistsActiveByWorkUnitAndGate checks if an active review exists for a work unit + gate
func (r *Repository) ExistsActiveByWorkUnitAndGate(workUnitID string, gate domain.ValidationGate) (bool, error) {
	var exists bool
	err := r.db.QueryRow(QueryExistsActiveByWorkUnitAndGate, workUnitID, gate).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active review existence: %w", err)
	}
	return exists, nil
}

// ExistsActiveByRunAndGate checks if an active review exists for a run + gate
func (r *Repository) ExistsActiveByRunAndGate(runID string, gate domain.ValidationGate) (bool, error) {
	var exists bool
	err := r.db.QueryRow(QueryExistsActiveByRunAndGate, runID, gate).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active review existence by run: %w", err)
	}
	return exists, nil
}

// ExistsActiveByTaskAndGate checks if an active review exists for a task + gate
func (r *Repository) ExistsActiveByTaskAndGate(taskID string, gate domain.ValidationGate) (bool, error) {
	var exists bool
	err := r.db.QueryRow(QueryExistsActiveByTaskAndGate, taskID, gate).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active review existence by task: %w", err)
	}
	return exists, nil
}

// UpdateStatus updates the status and optional fields of a review
func (r *Repository) UpdateStatus(review *domain.Review) error {
	now := time.Now().UTC()
	review.UpdatedAt = now

	criteriaChecked, err := json.Marshal(review.CriteriaChecked)
	if err != nil {
		return fmt.Errorf("failed to marshal criteria_checked: %w", err)
	}

	var completedAt sql.NullTime
	if review.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *review.CompletedAt, Valid: true}
	}

	_, err = r.db.Exec(
		QueryUpdateStatus,
		review.ID,
		review.Status,
		review.UpdatedAt,
		completedAt,
		sqlNullStringPtr(&review.VerdictReason),
		pqArray(review.EvidenceRefs),
		criteriaChecked,
	)
	if err != nil {
		return fmt.Errorf("failed to update review status: %w", err)
	}
	return nil
}

func (r *Repository) scanReview(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Review, error) {
	var review domain.Review
	var runID, workUnitID, taskID, agentSessionID, reviewerAgentID sql.NullString
	var verdictReason sql.NullString
	var evidenceRefs []byte
	var criteriaCheckedJSON []byte
	var createdAt, updatedAt time.Time
	var completedAt sql.NullTime

	err := scanner.Scan(
		&review.ID,
		&runID,
		&workUnitID,
		&taskID,
		&agentSessionID,
		&reviewerAgentID,
		&review.GateType,
		&review.Status,
		&verdictReason,
		&evidenceRefs,
		&criteriaCheckedJSON,
		&createdAt,
		&updatedAt,
		&completedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan review: %w", err)
	}

	if runID.Valid {
		review.RunID = &runID.String
	}
	if workUnitID.Valid {
		review.WorkUnitID = &workUnitID.String
	}
	if taskID.Valid {
		review.TaskID = &taskID.String
	}
	if agentSessionID.Valid {
		review.AgentSessionID = &agentSessionID.String
	}
	if reviewerAgentID.Valid {
		review.ReviewerAgentID = &reviewerAgentID.String
	}
	if verdictReason.Valid {
		review.VerdictReason = verdictReason.String
	}
	if completedAt.Valid {
		review.CompletedAt = &completedAt.Time
	}
	review.CreatedAt = createdAt
	review.UpdatedAt = updatedAt

	if len(evidenceRefs) > 0 {
		var refs []string
		if err := json.Unmarshal(evidenceRefs, &refs); err == nil {
			review.EvidenceRefs = refs
		}
	}
	if len(criteriaCheckedJSON) > 0 {
		var criteria []domain.ReviewCriteriaChecked
		if err := json.Unmarshal(criteriaCheckedJSON, &criteria); err == nil {
			review.CriteriaChecked = criteria
		}
	}

	return &review, nil
}

func sqlNullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func sqlNullStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func pqArray(ss []string) interface{} {
	if len(ss) == 0 {
		return nil
	}
	return ss
}
