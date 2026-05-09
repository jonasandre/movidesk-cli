package movidesk

import (
	"context"
	"sync"
	"time"
)

// Limiter is a sliding-window rate limiter sized for Movidesk's 10 req/min cap.
//
// It is intentionally simple: it tracks the last N request timestamps and
// blocks when the oldest is still within the window.
type Limiter struct {
	mu       sync.Mutex
	capacity int
	window   time.Duration
	stamps   []time.Time
	now      func() time.Time
}

func NewLimiter(capacity int, window time.Duration) *Limiter {
	return &Limiter{
		capacity: capacity,
		window:   window,
		stamps:   make([]time.Time, 0, capacity),
		now:      time.Now,
	}
}

// Wait blocks until a slot is available or ctx is canceled.
//
// Note: when ctx is canceled while waiting for a slot, no stamp has been
// reserved yet, so there is nothing to roll back.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		wait := l.reserve()
		if wait <= 0 {
			return nil
		}
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reserve registers a new stamp if a slot is free, else returns the duration
// until one frees up.
func (l *Limiter) reserve() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	cutoff := now.Add(-l.window)

	// Drop expired stamps from the head.
	i := 0
	for i < len(l.stamps) && !l.stamps[i].After(cutoff) {
		i++
	}
	l.stamps = l.stamps[i:]

	if len(l.stamps) < l.capacity {
		l.stamps = append(l.stamps, now)
		return 0
	}
	return l.stamps[0].Add(l.window).Sub(now)
}
