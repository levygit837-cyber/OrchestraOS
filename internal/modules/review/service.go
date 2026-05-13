// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package review

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type ReviewService struct {
	db *sql.DB
}

type CreateReviewInput struct {
	ID              string
	EventID         string
	RunID           string
	WorkUnitID      string
	TaskID          string
	AgentSessionID  string
	ReviewerAgentID string
	GateType        domain.ValidationGate
	EvidenceRefs    []string
	CriteriaChecked []domain.ReviewCriteriaChecked
}

type StartReviewInput struct {
	EventID string
	AgentID string
}

type SubmitVerdictInput struct {
	EventID         string
	AgentID         string
	Verdict         domain.ReviewDecision
	Reason          string
	EvidenceRefs    []string
	CriteriaChecked []domain.ReviewCriteriaChecked
}

func NewReviewService(database *sql.DB) *ReviewService {
	return &ReviewService{db: database}
}

func (s *ReviewService) Create(ctx context.Context, input CreateReviewInput) (*transition.OperationResult[*domain.Review], error) {
	if err := validateCreateReviewInput(input); err != nil {
		return nil, err
	}
	if input.ID == "" {
		input.ID = uuid.New().String()
	}

	var runID, workUnitID, taskID, agentSessionID, reviewerAgentID *string
	if input.RunID != "" {
		runID = &input.RunID
	}
	if input.WorkUnitID != "" {
		workUnitID = &input.WorkUnitID
	}
	if input.TaskID != "" {
		taskID = &input.TaskID
	}
	if input.AgentSessionID != "" {
		agentSessionID = &input.AgentSessionID
	}
	if input.ReviewerAgentID != "" {
		reviewerAgentID = &input.ReviewerAgentID
	}

	review := &domain.Review{
		ID:              input.ID,
		RunID:           runID,
		WorkUnitID:      workUnitID,
		TaskID:          taskID,
		AgentSessionID:  agentSessionID,
		ReviewerAgentID: reviewerAgentID,
		GateType:        input.GateType,
		Status:          domain.ReviewStatusPending,
		EvidenceRefs:    input.EvidenceRefs,
		CriteriaChecked: input.CriteriaChecked,
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "review_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	if input.WorkUnitID != "" {
		exists, err := NewRepository(tx).ExistsActiveByWorkUnitAndGate(input.WorkUnitID, input.GateType)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.check_duplicate", err)
		}
		if exists {
			return nil, apperrors.New(apperrors.CodeConflict, "review_service.create", "an active review already exists for this work unit and gate")
		}
	}
	if input.RunID != "" {
		exists, err := NewRepository(tx).ExistsActiveByRunAndGate(input.RunID, input.GateType)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.check_duplicate_run", err)
		}
		if exists {
			return nil, apperrors.New(apperrors.CodeConflict, "review_service.create", "an active review already exists for this run and gate")
		}
	}
	if input.TaskID != "" {
		exists, err := NewRepository(tx).ExistsActiveByTaskAndGate(input.TaskID, input.GateType)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.check_duplicate_task", err)
		}
		if exists {
			return nil, apperrors.New(apperrors.CodeConflict, "review_service.create", "an active review already exists for this task and gate")
		}
	}

	if err := NewRepository(tx).Create(review); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.create_projection", err)
	}

	payload, err := serialization.MarshalPayload("review_service.create_payload", map[string]interface{}{
		"review_id":         review.ID,
		"run_id":            review.RunID,
		"work_unit_id":      review.WorkUnitID,
		"task_id":           review.TaskID,
		"agent_session_id":  review.AgentSessionID,
		"reviewer_agent_id": review.ReviewerAgentID,
		"gate_type":         review.GateType,
		"status":            review.Status,
		"evidence_refs":     review.EvidenceRefs,
		"criteria_checked":  review.CriteriaChecked,
	})
	if err != nil {
		return nil, err
	}

	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        EventTypeCreated,
		Version:     transition.EventVersionV1,
		TaskID:      derefString(taskID),
		RunID:       derefString(runID),
		WorkUnitID:  derefString(workUnitID),
		AgentID:     input.ReviewerAgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "review_service.commit_create"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Review]{Value: review, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *ReviewService) Start(ctx context.Context, reviewID string, input StartReviewInput) (*transition.OperationResult[*domain.Review], error) {
	op := "review_service.start"
	if err := validation.RequiredUUID(reviewID, "review_id", op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "review_service.begin_start")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	review, err := RequireByID(ctx, tx, reviewID)
	if err != nil {
		return nil, err
	}
	if review.Status != domain.ReviewStatusPending {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "review can only be started from pending status")
	}

	review.Status = domain.ReviewStatusInProgress
	if err := NewRepository(tx).UpdateStatus(review); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.update_start", err)
	}

	payload, err := serialization.MarshalPayload("review_service.start_payload", map[string]interface{}{
		"review_id": review.ID,
		"status":    review.Status,
	})
	if err != nil {
		return nil, err
	}

	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        EventTypeStarted,
		Version:     transition.EventVersionV1,
		TaskID:      derefString(review.TaskID),
		RunID:       derefString(review.RunID),
		WorkUnitID:  derefString(review.WorkUnitID),
		AgentID:     input.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "review_service.commit_start"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Review]{Value: review, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *ReviewService) SubmitVerdict(ctx context.Context, reviewID string, input SubmitVerdictInput) (*transition.OperationResult[*domain.Review], error) {
	op := "review_service.submit_verdict"
	if err := validation.RequiredUUID(reviewID, "review_id", op); err != nil {
		return nil, err
	}
	if err := validateVerdict(input.Verdict, op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "review_service.begin_verdict")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	review, err := RequireByID(ctx, tx, reviewID)
	if err != nil {
		return nil, err
	}
	if isFinalReviewStatus(review.Status) {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "verdict is immutable after submission")
	}
	if review.Status != domain.ReviewStatusInProgress && review.Status != domain.ReviewStatusPending {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "review must be pending or in_progress to submit a verdict")
	}

	now := time.Now().UTC()
	review.Status = input.Verdict
	review.VerdictReason = input.Reason
	review.EvidenceRefs = input.EvidenceRefs
	review.CriteriaChecked = input.CriteriaChecked
	review.CompletedAt = &now

	if err := NewRepository(tx).UpdateStatus(review); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.update_verdict", err)
	}

	payload, err := serialization.MarshalPayload("review_service.verdict_payload", map[string]interface{}{
		"review_id":         review.ID,
		"status":            review.Status,
		"verdict_reason":    review.VerdictReason,
		"evidence_refs":     review.EvidenceRefs,
		"criteria_checked":  review.CriteriaChecked,
		"completed_at":      review.CompletedAt,
	})
	if err != nil {
		return nil, err
	}

	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        EventTypeVerdictSubmitted,
		Version:     transition.EventVersionV1,
		TaskID:      derefString(review.TaskID),
		RunID:       derefString(review.RunID),
		WorkUnitID:  derefString(review.WorkUnitID),
		AgentID:     input.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "review_service.commit_verdict"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Review]{Value: review, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *ReviewService) GetByID(ctx context.Context, id string) (*domain.Review, error) {
	op := "review_service.get_by_id"
	if err := validation.RequiredUUID(id, "review_id", op); err != nil {
		return nil, err
	}
	review, err := NewRepository(s.db).GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return review, nil
}

func (s *ReviewService) ListByTask(ctx context.Context, taskID string) ([]*domain.Review, error) {
	op := "review_service.list_by_task"
	if err := validation.RequiredUUID(taskID, "task_id", op); err != nil {
		return nil, err
	}
	reviews, err := NewRepository(s.db).ListByTask(ctx, taskID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return reviews, nil
}

func (s *ReviewService) ListPending(ctx context.Context) ([]*domain.Review, error) {
	reviews, err := NewRepository(s.db).ListPending(ctx)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "review_service.list_pending", err)
	}
	return reviews, nil
}

func RequireByID(ctx context.Context, tx *sql.Tx, id string) (*domain.Review, error) {
	review, err := NewRepository(tx).GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "review.require_by_id", err)
	}
	if review == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "review.require_by_id", "review not found")
	}
	return review, nil
}

func isFinalReviewStatus(status domain.ReviewStatus) bool {
	switch status {
	case domain.ReviewStatusApproved, domain.ReviewStatusChangesRequested, domain.ReviewStatusNeedsDiscussion:
		return true
	default:
		return false
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
