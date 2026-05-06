package monitor

import (
	"sync"
	"time"
)

// RateLimiter suppresses repeated alerts for the same job within a cooldown window.
type RateLimiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[string]time.Time
	nowFn    func() time.Time
}

// NewRateLimiter creates a RateLimiter with the given cooldown duration.
func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
		nowFn:    time.Now,
	}
}

// Allow returns true if an alert for the given key should be sent.
// It records the current time as the last-sent timestamp when allowed.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.nowFn()
	if last, ok := r.lastSent[key]; ok {
		if now.Sub(last) < r.cooldown {
			return false
		}
	}
	r.lastSent[key] = now
	return true
}

// Reset clears the rate-limit record for the given key, allowing the next
// alert through immediately.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lastSent, key)
}

// LastSent returns the time the last alert was sent for key, and whether
// a record exists.
func (r *RateLimiter) LastSent(key string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.lastSent[key]
	return t, ok
}
