package monitor

import (
	"sync"
	"time"
)

// DedupKey uniquely identifies an alert event.
type DedupKey struct {
	Job    string
	Reason string
}

// DedupEntry holds the last time an alert was emitted for a given key.
type DedupEntry struct {
	LastSent time.Time
	Count    int
}

// DedupStore suppresses duplicate alerts within a configurable window.
type DedupStore struct {
	mu      sync.Mutex
	entries map[DedupKey]*DedupEntry
	window  time.Duration
	now     func() time.Time
}

// NewDedupStore creates a DedupStore with the given deduplication window.
func NewDedupStore(window time.Duration) *DedupStore {
	return &DedupStore{
		entries: make(map[DedupKey]*DedupEntry),
		window:  window,
		now:     time.Now,
	}
}

// IsDuplicate returns true if an identical alert was already sent within the window.
// If not a duplicate, it records the alert and returns false.
func (d *DedupStore) IsDuplicate(job, reason string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := DedupKey{Job: job, Reason: reason}
	now := d.now()

	if entry, ok := d.entries[key]; ok {
		if now.Sub(entry.LastSent) < d.window {
			entry.Count++
			return true
		}
	}

	d.entries[key] = &DedupEntry{LastSent: now, Count: 1}
	return false
}

// Reset clears the dedup state for a specific job and reason.
func (d *DedupStore) Reset(job, reason string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, DedupKey{Job: job, Reason: reason})
}

// SuppressedCount returns how many times an alert was suppressed for a key.
func (d *DedupStore) SuppressedCount(job, reason string) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	if entry, ok := d.entries[DedupKey{Job: job, Reason: reason}]; ok {
		return entry.Count - 1
	}
	return 0
}
