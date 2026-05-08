package monitor

import (
	"sync"
	"time"
)

// JobMetrics holds runtime statistics for a single cron job.
type JobMetrics struct {
	JobName      string
	TotalRuns    int64
	SuccessCount int64
	FailureCount int64
	LastDuration time.Duration
	AvgDuration  time.Duration
	LastRunAt    time.Time
}

// MetricsStore collects and exposes per-job execution metrics.
type MetricsStore struct {
	mu      sync.RWMutex
	records map[string]*jobMetricsRecord
}

type jobMetricsRecord struct {
	JobMetrics
	totalDuration time.Duration
}

// NewMetricsStore creates an empty MetricsStore.
func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		records: make(map[string]*jobMetricsRecord),
	}
}

// Record updates metrics for the given job after an execution.
func (m *MetricsStore) Record(jobName string, success bool, duration time.Duration, runAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.records[jobName]
	if !ok {
		r = &jobMetricsRecord{}
		r.JobName = jobName
		m.records[jobName] = r
	}

	r.TotalRuns++
	if success {
		r.SuccessCount++
	} else {
		r.FailureCount++
	}
	r.LastDuration = duration
	r.totalDuration += duration
	r.AvgDuration = time.Duration(int64(r.totalDuration) / r.TotalRuns)
	r.LastRunAt = runAt
}

// Get returns a snapshot of metrics for the given job.
// Returns false if the job has not been recorded.
func (m *MetricsStore) Get(jobName string) (JobMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.records[jobName]
	if !ok {
		return JobMetrics{}, false
	}
	return r.JobMetrics, true
}

// All returns a snapshot of metrics for every recorded job.
func (m *MetricsStore) All() []JobMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]JobMetrics, 0, len(m.records))
	for _, r := range m.records {
		out = append(out, r.JobMetrics)
	}
	return out
}
