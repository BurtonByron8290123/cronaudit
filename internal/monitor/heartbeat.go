package monitor

import (
	"sync"
	"time"
)

// HeartbeatStatus represents the liveness state of a monitored job.
type HeartbeatStatus struct {
	JobName   string
	LastSeen  time.Time
	Deadline  time.Duration
	Alerted   bool
}

// HeartbeatMonitor tracks the last-seen time for jobs and detects
// jobs that have not reported within their expected deadline.
type HeartbeatMonitor struct {
	mu       sync.Mutex
	statuses map[string]*HeartbeatStatus
	now      func() time.Time
}

// NewHeartbeatMonitor creates a HeartbeatMonitor. If nowFn is nil,
// time.Now is used.
func NewHeartbeatMonitor(nowFn func() time.Time) *HeartbeatMonitor {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &HeartbeatMonitor{
		statuses: make(map[string]*HeartbeatStatus),
		now:      nowFn,
	}
}

// Register registers a job with the given deadline tolerance.
func (h *HeartbeatMonitor) Register(jobName string, deadline time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.statuses[jobName] = &HeartbeatStatus{
		JobName:  jobName,
		Deadline: deadline,
		LastSeen: time.Time{},
	}
}

// Ping records that a job has been seen at the current time.
func (h *HeartbeatMonitor) Ping(jobName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, ok := h.statuses[jobName]; ok {
		s.LastSeen = h.now()
		s.Alerted = false
	}
}

// Stale returns a list of jobs whose last-seen time exceeds their
// deadline. Jobs that have never been seen are included if their
// registration time is not tracked — callers should Ping on first
// successful run. Only jobs not yet alerted are returned; call
// MarkAlerted to suppress repeat alerts.
func (h *HeartbeatMonitor) Stale() []HeartbeatStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := h.now()
	var stale []HeartbeatStatus
	for _, s := range h.statuses {
		if s.Alerted {
			continue
		}
		if s.LastSeen.IsZero() {
			continue // never seen; skip until first ping
		}
		if now.Sub(s.LastSeen) > s.Deadline {
			stale = append(stale, *s)
		}
	}
	return stale
}

// MarkAlerted suppresses further alerts for a job until the next Ping.
func (h *HeartbeatMonitor) MarkAlerted(jobName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, ok := h.statuses[jobName]; ok {
		s.Alerted = true
	}
}
