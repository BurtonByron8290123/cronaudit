package monitor

import (
	"sync"
	"time"
)

// ExecutionRecord stores the result of a single cron job execution.
type ExecutionRecord struct {
	JobName   string
	StartedAt time.Time
	FinishedAt time.Time
	Success   bool
	ExitCode  int
}

// Duration returns how long the job took to execute.
func (r ExecutionRecord) Duration() time.Duration {
	return r.FinishedAt.Sub(r.StartedAt)
}

// History maintains a bounded in-memory log of job execution records.
type History struct {
	mu      sync.RWMutex
	records map[string][]ExecutionRecord
	maxPerJob int
}

// NewHistory creates a History that retains at most maxPerJob records per job.
func NewHistory(maxPerJob int) *History {
	if maxPerJob <= 0 {
		maxPerJob = 10
	}
	return &History{
		records:   make(map[string][]ExecutionRecord),
		maxPerJob: maxPerJob,
	}
}

// Record appends an ExecutionRecord for the given job, evicting the oldest
// entry when the per-job limit is exceeded.
func (h *History) Record(rec ExecutionRecord) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entries := h.records[rec.JobName]
	entries = append(entries, rec)
	if len(entries) > h.maxPerJob {
		entries = entries[len(entries)-h.maxPerJob:]
	}
	h.records[rec.JobName] = entries
}

// Last returns the most recent ExecutionRecord for a job and whether one exists.
func (h *History) Last(jobName string) (ExecutionRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := h.records[jobName]
	if len(entries) == 0 {
		return ExecutionRecord{}, false
	}
	return entries[len(entries)-1], true
}

// All returns a copy of all records for a job.
func (h *History) All(jobName string) []ExecutionRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries := h.records[jobName]
	result := make([]ExecutionRecord, len(entries))
	copy(result, entries)
	return result
}
