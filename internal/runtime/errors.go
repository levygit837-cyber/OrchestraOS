package runtime

import (
	"context"
	"fmt"
	"net"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
)

func classifyHTTPError(op string, err error) *apperrors.Error {
	if err == nil {
		return nil
	}
	if ctx, ok := err.(interface{ Err() error }); ok {
		if ctx.Err() == context.DeadlineExceeded {
			return apperrors.Wrap(apperrors.KindTimeout, op, err)
		}
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return apperrors.Wrap(apperrors.KindTimeout, op, err)
	}
	return apperrors.Wrap(apperrors.KindProviderDown, op, err)
}

func classifyStatusCode(op string, status int, body []byte) *apperrors.Error {
	switch {
	case status >= 200 && status < 300:
		return nil
	case status == 401 || status == 403:
		return apperrors.New(apperrors.KindAuthFailure, op, fmt.Sprintf("authentication failed (HTTP %d): %s", status, truncate(body)))
	case status == 429:
		return apperrors.New(apperrors.KindRateLimit, op, fmt.Sprintf("rate limited (HTTP %d): %s", status, truncate(body)))
	case status >= 500:
		return apperrors.New(apperrors.KindProviderDown, op, fmt.Sprintf("provider error (HTTP %d): %s", status, truncate(body)))
	default:
		return apperrors.New(apperrors.KindInternal, op, fmt.Sprintf("unexpected status %d: %s", status, truncate(body)))
	}
}

func truncate(b []byte) string {
	const maxLen = 200
	if len(b) <= maxLen {
		return string(b)
	}
	return string(b[:maxLen]) + "..."
}
