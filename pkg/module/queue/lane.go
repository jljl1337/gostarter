package queue

import "time"

type laneConfig struct {
	maxAttempts int
	baseBackoff time.Duration
}

type LaneOption func(*laneConfig)

// WithMaxAttempts sets the total number of attempts (first run included).
func WithMaxAttempts(n int) LaneOption {
	return func(c *laneConfig) { c.maxAttempts = n }
}

// WithBaseBackoff sets the delay before the first retry; it doubles on every retry.
func WithBaseBackoff(d time.Duration) LaneOption {
	return func(c *laneConfig) { c.baseBackoff = d }
}

type laneState struct {
	name    string
	handler Handler
	cfg     laneConfig
	wake    chan struct{} // capacity 1, in-memory signal, replaces polling
	hasWork bool          // lane is known to have unfinished tasks
	started bool          // a worker goroutine is running
}

func (l *laneState) signal() {
	select {
	case l.wake <- struct{}{}:
	default: // a wake-up is already pending
	}
}
