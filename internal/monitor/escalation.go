package monitor

import (
	"sync"
	"time"
)

// EscalationPolicy defines thresholds for escalating alerts.
type EscalationPolicy struct {
	// After this many consecutive failures, escalate.
	Threshold int
	// Cooldown between escalation alerts.
	Cooldown time.Duration
}

type escalationEntry struct {
	consecutiveFailures int
	lastEscalated       time.Time
	escalated           bool
}

// EscalationStore tracks consecutive failures per job and decides
// whether an alert should be escalated.
type EscalationStore struct {
	mu     sync.Mutex
	policy EscalationPolicy
	entries map[string]*escalationEntry
	now    func() time.Time
}

// NewEscalationStore creates a new EscalationStore with the given policy.
func NewEscalationStore(policy EscalationPolicy, now func() time.Time) *EscalationStore {
	if now == nil {
		now = time.Now
	}
	return &EscalationStore{
		policy:  policy,
		entries: make(map[string]*escalationEntry),
		now:     now,
	}
}

// RecordFailure increments the consecutive failure count for key.
// Returns true if the failure count has reached the escalation threshold
// and enough time has passed since the last escalation.
func (e *EscalationStore) RecordFailure(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	ent, ok := e.entries[key]
	if !ok {
		ent = &escalationEntry{}
		e.entries[key] = ent
	}
	ent.consecutiveFailures++

	if ent.consecutiveFailures < e.policy.Threshold {
		return false
	}

	now := e.now()
	if ent.escalated && now.Sub(ent.lastEscalated) < e.policy.Cooldown {
		return false
	}

	ent.escalated = true
	ent.lastEscalated = now
	return true
}

// RecordSuccess resets the consecutive failure count for key.
func (e *EscalationStore) RecordSuccess(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.entries, key)
}

// ConsecutiveFailures returns the current failure count for key.
func (e *EscalationStore) ConsecutiveFailures(key string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ent, ok := e.entries[key]; ok {
		return ent.consecutiveFailures
	}
	return 0
}
