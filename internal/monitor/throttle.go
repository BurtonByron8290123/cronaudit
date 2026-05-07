package monitor

import (
	"sync"
	"time"
)

// ThrottlePolicy defines the maximum number of alerts allowed within a window.
type ThrottlePolicy struct {
	MaxAlerts int
	Window    time.Duration
}

type throttleEntry struct {
	count     int
	windowEnd time.Time
}

// ThrottleStore tracks per-key alert counts within a sliding window.
type ThrottleStore struct {
	mu     sync.Mutex
	entries map[string]*throttleEntry
	policy ThrottlePolicy
	now    func() time.Time
}

// NewThrottleStore creates a ThrottleStore with the given policy.
func NewThrottleStore(policy ThrottlePolicy, now func() time.Time) *ThrottleStore {
	if now == nil {
		now = time.Now
	}
	return &ThrottleStore{
		entries: make(map[string]*throttleEntry),
		policy:  policy,
		now:     now,
	}
}

// Allow returns true if an alert for the given key is permitted under the policy.
// It increments the counter if allowed.
func (t *ThrottleStore) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	e, ok := t.entries[key]
	if !ok || now.After(e.windowEnd) {
		t.entries[key] = &throttleEntry{
			count:     1,
			windowEnd: now.Add(t.policy.Window),
		}
		return true
	}

	if e.count >= t.policy.MaxAlerts {
		return false
	}

	e.count++
	return true
}

// Reset clears the throttle state for the given key.
func (t *ThrottleStore) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, key)
}

// Count returns the current alert count for the given key within the active window.
func (t *ThrottleStore) Count(key string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.entries[key]
	if !ok || t.now().After(e.windowEnd) {
		return 0
	}
	return e.count
}
