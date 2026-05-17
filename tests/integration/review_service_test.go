package integration

import (
	"context"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	reviewmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/review"
)

func TestReviewServiceCreateAndGet(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	result, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	if result.Value.Status != reviewmod.StatusPending {
		t.Fatalf("expected status pending, got %s", result.Value.Status)
	}
	if result.Value.GateType != reviewmod.GateHard {
		t.Fatalf("expected gate hard, got %s", result.Value.GateType)
	}

	fetched, err := reviewService.GetByID(ctx, result.Value.ID)
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected review to exist")
	}
	if fetched.ID != result.Value.ID {
		t.Fatalf("expected review id %s, got %s", result.Value.ID, fetched.ID)
	}
}

func TestReviewServiceStartAndSubmitVerdict(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	created, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateSoft,
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	started, err := reviewService.Start(ctx, created.Value.ID, reviewmod.StartReviewInput{
		AgentID: "agent-reviewer",
	})
	if err != nil {
		t.Fatalf("start review: %v", err)
	}
	if started.Value.Status != reviewmod.StatusInProgress {
		t.Fatalf("expected status in_progress, got %s", started.Value.Status)
	}

	verdicts := []reviewmod.Decision{
		reviewmod.StatusApproved,
		reviewmod.StatusChangesRequested,
		reviewmod.StatusNeedsDiscussion,
	}
	for _, verdict := range verdicts {
		// Create a fresh review for each verdict to avoid immutability conflict
		review, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
			TaskID:     taskID,
			WorkUnitID: workUnitID,
			GateType:   reviewmod.GateSoft,
		})
		if err != nil {
			t.Fatalf("create review for verdict %s: %v", verdict, err)
		}
		if _, err := reviewService.Start(ctx, review.Value.ID, reviewmod.StartReviewInput{AgentID: "agent-reviewer"}); err != nil {
			t.Fatalf("start review for verdict %s: %v", verdict, err)
		}

		result, err := reviewService.SubmitVerdict(ctx, review.Value.ID, reviewmod.SubmitVerdictInput{
			AgentID:      "agent-reviewer",
			Verdict:      verdict,
			Reason:       "test verdict",
			EvidenceRefs: []string{"evidence:1"},
			CriteriaChecked: []reviewmod.CriteriaChecked{
				{Criterion: "tests pass", Passed: true},
			},
		})
		if err != nil {
			t.Fatalf("submit verdict %s: %v", verdict, err)
		}
		if result.Value.Status != verdict {
			t.Fatalf("expected status %s, got %s", verdict, result.Value.Status)
		}
		if result.Value.VerdictReason != "test verdict" {
			t.Fatalf("expected verdict reason, got %s", result.Value.VerdictReason)
		}
		if result.Value.CompletedAt == nil {
			t.Fatal("expected completed_at to be set")
		}
	}
}

func TestReviewServiceVerdictImmutable(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	created, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	if _, err := reviewService.Start(ctx, created.Value.ID, reviewmod.StartReviewInput{AgentID: "agent-reviewer"}); err != nil {
		t.Fatalf("start review: %v", err)
	}

	_, err = reviewService.SubmitVerdict(ctx, created.Value.ID, reviewmod.SubmitVerdictInput{
		AgentID: "agent-reviewer",
		Verdict: reviewmod.StatusApproved,
		Reason:  "first verdict",
	})
	if err != nil {
		t.Fatalf("submit first verdict: %v", err)
	}

	_, err = reviewService.SubmitVerdict(ctx, created.Value.ID, reviewmod.SubmitVerdictInput{
		AgentID: "agent-reviewer",
		Verdict: reviewmod.StatusChangesRequested,
		Reason:  "second verdict",
	})
	if err == nil {
		t.Fatal("expected second verdict to be rejected")
	}
	appErr, ok := err.(*apperrors.Error)
	if !ok || appErr.Code != apperrors.CodeInvalidTransition {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}

func TestReviewServiceListPending(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	pending1, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create review 1: %v", err)
	}
	pending2, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateSoft,
	})
	if err != nil {
		t.Fatalf("create review 2: %v", err)
	}
	completed, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GatePolicy,
	})
	if err != nil {
		t.Fatalf("create review 3: %v", err)
	}
	if _, err := reviewService.Start(ctx, completed.Value.ID, reviewmod.StartReviewInput{AgentID: "agent-reviewer"}); err != nil {
		t.Fatalf("start completed review: %v", err)
	}
	if _, err := reviewService.SubmitVerdict(ctx, completed.Value.ID, reviewmod.SubmitVerdictInput{
		AgentID: "agent-reviewer",
		Verdict: reviewmod.StatusApproved,
		Reason:  "done",
	}); err != nil {
		t.Fatalf("submit verdict: %v", err)
	}

	pending, err := reviewService.ListPending(ctx)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) < 2 {
		t.Fatalf("expected at least 2 pending reviews, got %d", len(pending))
	}
	found1, found2 := false, false
	for _, r := range pending {
		if r.ID == pending1.Value.ID {
			found1 = true
		}
		if r.ID == pending2.Value.ID {
			found2 = true
		}
		if r.ID == completed.Value.ID {
			t.Fatal("expected completed review not to be in pending list")
		}
	}
	if !found1 || !found2 {
		t.Fatalf("expected pending reviews to be listed")
	}
}

func TestReviewServiceListByTask(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	_, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}

	reviews, err := reviewService.ListByTask(ctx, taskID)
	if err != nil {
		t.Fatalf("list by task: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}

	otherTaskID := createTestTask(t, db)
	otherReviews, err := reviewService.ListByTask(ctx, otherTaskID)
	if err != nil {
		t.Fatalf("list by other task: %v", err)
	}
	if len(otherReviews) != 0 {
		t.Fatalf("expected 0 reviews for other task, got %d", len(otherReviews))
	}
}

func TestReviewServiceInvalidGateType(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	_, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.ValidationGate("invalid_gate"),
	})
	if err == nil {
		t.Fatal("expected invalid gate type to be rejected")
	}
	appErr, ok := err.(*apperrors.Error)
	if !ok || appErr.Code != apperrors.CodeInvalidInput {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestReviewServiceInvalidVerdict(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	created, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	if _, err := reviewService.Start(ctx, created.Value.ID, reviewmod.StartReviewInput{AgentID: "agent-reviewer"}); err != nil {
		t.Fatalf("start review: %v", err)
	}

	_, err = reviewService.SubmitVerdict(ctx, created.Value.ID, reviewmod.SubmitVerdictInput{
		AgentID: "agent-reviewer",
		Verdict: reviewmod.Decision("invalid_verdict"),
	})
	if err == nil {
		t.Fatal("expected invalid verdict to be rejected")
	}
	appErr, ok := err.(*apperrors.Error)
	if !ok || appErr.Code != apperrors.CodeInvalidInput {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestReviewServiceStartRequiresPending(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	created, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	if _, err := reviewService.Start(ctx, created.Value.ID, reviewmod.StartReviewInput{AgentID: "agent-reviewer"}); err != nil {
		t.Fatalf("first start: %v", err)
	}
	_, err = reviewService.Start(ctx, created.Value.ID, reviewmod.StartReviewInput{AgentID: "agent-reviewer"})
	if err == nil {
		t.Fatal("expected second start to be rejected")
	}
	appErr, ok := err.(*apperrors.Error)
	if !ok || appErr.Code != apperrors.CodeInvalidTransition {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}

func TestReviewServiceRejectsDuplicateActiveReview(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)

	reviewService := reviewmod.NewReviewService(db)
	_, err := reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err != nil {
		t.Fatalf("create first review: %v", err)
	}

	_, err = reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateHard,
	})
	if err == nil {
		t.Fatal("expected duplicate active review to be rejected")
	}
	appErr, ok := err.(*apperrors.Error)
	if !ok || appErr.Code != apperrors.CodeConflict {
		t.Fatalf("expected conflict error, got %v", err)
	}

	// Different gate for same work unit should be allowed
	_, err = reviewService.Create(ctx, reviewmod.CreateReviewInput{
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		GateType:   reviewmod.GateSoft,
	})
	if err != nil {
		t.Fatalf("create review with different gate: %v", err)
	}
}
