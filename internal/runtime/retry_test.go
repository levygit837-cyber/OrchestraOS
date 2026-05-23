package runtime_test

import (
	"context"
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/retry"
)

func fastRetryConfig(maxRetries int) retry.Config {
	return retry.Config{
		MaxRetries:    maxRetries,
		InitialDelay:  time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
	}
}

func TestRetryDo_SucceedsFirstAttempt(t *testing.T) {
	t.Parallel()
	calls := 0
	cfg := fastRetryConfig(3)

	err := retry.Do(context.Background(), func(_ context.Context) error {
		calls++
		return nil
	}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryDo_SucceedsAfterRetries(t *testing.T) {
	t.Parallel()
	calls := 0
	cfg := fastRetryConfig(5)

	err := retry.Do(context.Background(), func(_ context.Context) error {
		calls++
		if calls <= 3 {
			return apperrors.New(apperrors.KindProviderDown, "test", "down")
		}
		return nil
	}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 4 {
		t.Errorf("expected 4 calls, got %d", calls)
	}
}

func TestRetryDo_ExhaustsRetries(t *testing.T) {
	t.Parallel()
	calls := 0
	cfg := fastRetryConfig(3)

	err := retry.Do(context.Background(), func(_ context.Context) error {
		calls++
		return apperrors.New(apperrors.KindTimeout, "test", "timeout")
	}, cfg)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 4 {
		t.Errorf("expected 4 calls (1 + 3 retries), got %d", calls)
	}
}

func TestRetryDo_NonRetryableStopsImmediately(t *testing.T) {
	t.Parallel()
	calls := 0
	cfg := fastRetryConfig(5)

	err := retry.Do(context.Background(), func(_ context.Context) error {
		calls++
		return apperrors.New(apperrors.KindAuthFailure, "test", "auth failed")
	}, cfg)
	if err == nil {
		t.Fatal("expected error for non-retryable")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry), got %d", calls)
	}
}

func TestRetryDo_RespectsContextCancel(t *testing.T) {
	t.Parallel()
	cfg := retry.Config{
		MaxRetries:    10,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      time.Second,
		BackoffFactor: 2.0,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := retry.Do(ctx, func(_ context.Context) error {
		return apperrors.New(apperrors.KindProviderDown, "test", "down")
	}, cfg)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestRetryDelay_Exponential(t *testing.T) {
	t.Parallel()
	cfg := retry.Config{
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
	d0 := retry.Delay(0, cfg)
	d1 := retry.Delay(1, cfg)
	d2 := retry.Delay(2, cfg)

	if d0 != 100*time.Millisecond {
		t.Errorf("d0: expected 100ms, got %v", d0)
	}
	if d1 != 200*time.Millisecond {
		t.Errorf("d1: expected 200ms, got %v", d1)
	}
	if d2 != 400*time.Millisecond {
		t.Errorf("d2: expected 400ms, got %v", d2)
	}
}

func TestRetryDelay_CapsAtMax(t *testing.T) {
	t.Parallel()
	cfg := retry.Config{
		InitialDelay:  time.Second,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 10.0,
	}
	d := retry.Delay(3, cfg)
	if d != 5*time.Second {
		t.Errorf("expected cap at 5s, got %v", d)
	}
}

func TestRetryDefaultConfig_Values(t *testing.T) {
	t.Parallel()
	cfg := retry.DefaultConfig()
	if cfg.MaxRetries != 10 {
		t.Errorf("MaxRetries: expected 10, got %d", cfg.MaxRetries)
	}
	if cfg.InitialDelay != 100*time.Millisecond {
		t.Errorf("InitialDelay: expected 100ms, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay: expected 30s, got %v", cfg.MaxDelay)
	}
	if cfg.BackoffFactor != 2.0 {
		t.Errorf("BackoffFactor: expected 2.0, got %v", cfg.BackoffFactor)
	}
}
