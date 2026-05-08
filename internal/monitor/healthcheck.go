package monitor

import (
	"fmt"
	"sync"
	"time"
)

// HealthStatus represents the overall health of the cronaudit daemon.
type HealthStatus struct {
	Healthy   bool              `json:"healthy"`
	CheckedAt time.Time         `json:"checked_at"`
	Jobs      map[string]string `json:"jobs"`
}

// HealthChecker aggregates job states and reports daemon health.
type HealthChecker struct {
	mu      sync.RWMutex
	state   *StateStore
	clock   func() time.Time
	staleTTL time.Duration
}

// NewHealthChecker creates a HealthChecker backed by the given StateStore.
func NewHealthChecker(state *StateStore, staleTTL time.Duration) *HealthChecker {
	return &HealthChecker{
		state:    state,
		clock:    time.Now,
		staleTTL: staleTTL,
	}
}

// Check returns a HealthStatus snapshot across all known jobs.
func (h *HealthChecker) Check() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	now := h.clock()
	status := HealthStatus{
		Healthy:   true,
		CheckedAt: now,
		Jobs:      make(map[string]string),
	}

	for _, entry := range h.state.All() {
		label := entry.Status.String()
		if entry.Status == StatusFailed {
			status.Healthy = false
			label = fmt.Sprintf("failed (since %s)", entry.LastChange.Format(time.RFC3339))
		} else if now.Sub(entry.LastChange) > h.staleTTL {
			status.Healthy = false
			label = "stale"
		}
		status.Jobs[entry.Name] = label
	}

	return status
}
