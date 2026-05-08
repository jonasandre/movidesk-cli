package movidesk

import (
	"errors"
	"net/http"
	"time"
)

type RetryPolicy struct {
	MaxAttempts int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	Disabled    bool
}

func DefaultRetry() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseBackoff: time.Second,
		MaxBackoff:  5 * time.Minute,
	}
}

// ShouldRetry returns true for retryable errors below the attempt cap.
func (p RetryPolicy) ShouldRetry(err error, attempt int) bool {
	if attempt >= p.MaxAttempts {
		return false
	}
	var ae *APIError
	if errors.As(err, &ae) {
		switch ae.Status {
		case http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		}
		return false
	}
	// Network/transport errors: retry.
	return true
}

// Backoff returns the wait duration before the next attempt. If the server
// supplied Retry-After we honor that; otherwise exponential backoff.
func (p RetryPolicy) Backoff(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		if retryAfter > p.MaxBackoff {
			return p.MaxBackoff
		}
		return retryAfter
	}
	d := p.BaseBackoff << (attempt - 1)
	if d > p.MaxBackoff {
		return p.MaxBackoff
	}
	return d
}
