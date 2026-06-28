package vault

import (
	"sync"
	"time"
)

// Lockout throttles unlock attempts to blunt online PIN guessing. After a few
// free failures it imposes an exponentially growing delay before the next
// attempt is permitted. State is in-memory and single-user, so it resets on
// restart (which also re-locks the vault).
type Lockout struct {
	mu        sync.Mutex
	failures  int
	nextAllow time.Time

	freeAttempts int
	baseDelay    time.Duration
	maxDelay     time.Duration
	now          func() time.Time // injectable for tests
}

// NewLockout returns a Lockout with sane defaults: 3 free attempts, then 5s
// doubling up to a 15m cap.
func NewLockout() *Lockout {
	return &Lockout{
		freeAttempts: 3,
		baseDelay:    5 * time.Second,
		maxDelay:     15 * time.Minute,
		now:          time.Now,
	}
}

// Retry reports how long the caller must wait before another attempt is allowed.
// Zero means an attempt is permitted now.
func (l *Lockout) Retry() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	if d := l.nextAllow.Sub(l.now()); d > 0 {
		return d
	}
	return 0
}

// Fail records a failed attempt and arms the next backoff window.
func (l *Lockout) Fail() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures++
	if l.failures <= l.freeAttempts {
		return
	}
	shift := l.failures - l.freeAttempts - 1
	delay := l.baseDelay << shift
	if delay <= 0 || delay > l.maxDelay { // overflow or past the cap
		delay = l.maxDelay
	}
	l.nextAllow = l.now().Add(delay)
}

// Reset clears all failure state after a successful unlock.
func (l *Lockout) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures = 0
	l.nextAllow = time.Time{}
}
