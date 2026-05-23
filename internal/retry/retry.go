package retry

import (
	"context"
	"math"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
)

// Config controls exponential backoff settings.
type Config struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultConfig returns the standard retry settings.
func DefaultConfig() Config {
	return Config{
		MaxRetries:    10,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Func is called on each attempt. Return nil to stop retrying.
type Func func(ctx context.Context) error

// Do executes fn with exponential backoff for retryable errors.
func Do(ctx context.Context, fn Func, cfg Config) error {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := fn(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			if !apperrors.IsRetryable(err) {
				return err
			}
		}
		if attempt == cfg.MaxRetries {
			break
		}
		if err := sleepWithContext(ctx, Delay(attempt, cfg)); err != nil {
			return err
		}
	}
	return lastErr
}

// Delay computes the backoff duration for the given attempt.
func Delay(attempt int, cfg Config) time.Duration {
	d := float64(cfg.InitialDelay) * math.Pow(cfg.BackoffFactor, float64(attempt))
	if d > float64(cfg.MaxDelay) {
		d = float64(cfg.MaxDelay)
	}
	return time.Duration(d)
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
