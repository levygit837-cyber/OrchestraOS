package run

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

type RetryPolicy struct {
	MaxAttempts       int
	AttemptTimeout    time.Duration
	OperationTimeout  time.Duration
	InitialBackoff    time.Duration
	BackoffMultiplier int
}

func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       3,
		AttemptTimeout:    5 * time.Second,
		OperationTimeout:  30 * time.Second,
		InitialBackoff:    100 * time.Millisecond,
		BackoffMultiplier: 2,
	}
}

func RetryPolicyFromInput(extra map[string]interface{}, op string) (RetryPolicy, error) {
	policy := DefaultRetryPolicy()
	if extra == nil {
		return policy, nil
	}
	if value, ok, err := IntExtra(extra, "max_attempts"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.MaxAttempts = value
	}
	if value, ok, err := IntExtra(extra, "attempt_timeout_seconds"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.AttemptTimeout = time.Duration(value) * time.Second
	}
	if value, ok, err := IntExtra(extra, "operation_timeout_seconds"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.OperationTimeout = time.Duration(value) * time.Second
	}
	if value, ok, err := IntExtra(extra, "initial_backoff_millis"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.InitialBackoff = time.Duration(value) * time.Millisecond
	}
	if value, ok, err := IntExtra(extra, "backoff_multiplier"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.BackoffMultiplier = value
	}
	if policy.MaxAttempts < 1 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "max_attempts must be greater than zero")
	}
	if policy.AttemptTimeout <= 0 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "attempt_timeout_seconds must be greater than zero")
	}
	if policy.OperationTimeout <= 0 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "operation_timeout_seconds must be greater than zero")
	}
	if policy.InitialBackoff < 0 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "initial_backoff_millis must not be negative")
	}
	if policy.BackoffMultiplier < 1 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "backoff_multiplier must be greater than zero")
	}
	return policy, nil
}

func IntExtra(extra map[string]interface{}, key string) (int, bool, error) {
	raw, ok := extra[key]
	if !ok {
		return 0, false, nil
	}
	switch value := raw.(type) {
	case int:
		return value, true, nil
	case int8:
		return int(value), true, nil
	case int16:
		return int(value), true, nil
	case int32:
		return int(value), true, nil
	case int64:
		return int(value), true, nil
	case uint:
		return int(value), true, nil
	case uint8:
		return int(value), true, nil
	case uint16:
		return int(value), true, nil
	case uint32:
		return int(value), true, nil
	case uint64:
		return int(value), true, nil
	case float64:
		converted := int(value)
		if value != float64(converted) {
			return 0, true, fmt.Errorf("%s must be an integer", key)
		}
		return converted, true, nil
	case json.Number:
		converted, err := strconv.Atoi(value.String())
		if err != nil {
			return 0, true, fmt.Errorf("%s must be an integer: %w", key, err)
		}
		return converted, true, nil
	default:
		return 0, true, fmt.Errorf("%s must be an integer", key)
	}
}

func (p RetryPolicy) BackoffDelayForAttempt(attempt int) time.Duration {
	if attempt <= 1 || p.InitialBackoff == 0 {
		return 0
	}
	delay := p.InitialBackoff
	for i := 2; i < attempt; i++ {
		delay *= time.Duration(p.BackoffMultiplier)
	}
	return delay
}

func (p RetryPolicy) Payload(delay time.Duration) map[string]interface{} {
	return map[string]interface{}{
		"max_attempts":              p.MaxAttempts,
		"attempt_timeout_seconds":   int(p.AttemptTimeout / time.Second),
		"operation_timeout_seconds": int(p.OperationTimeout / time.Second),
		"initial_backoff_millis":    int(p.InitialBackoff / time.Millisecond),
		"backoff_multiplier":        p.BackoffMultiplier,
		"applied_backoff_millis":    int(delay / time.Millisecond),
	}
}

func WaitForRetryBackoff(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return apperrors.Wrap(apperrors.CodeTimeout, "services.retry_backoff", ctx.Err())
	case <-timer.C:
		return nil
	}
}
