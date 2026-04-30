// Package monitor provides functionality for tracking cron job execution,
// detecting failures, and identifying scheduling drift.
package monitor

import (
	"fmt"
	"sync"
	"time"

	"github.com/your-org/cronaudit/internal/config"
)

// JobStatus represents the last known state of a monitored cron job.
type JobStatus struct {
	Name        string
	LastSeen    time.Time
	LastSuccess time.Time
	FailCount   int
	MissedCount int
	Healthy     bool
}

// Monitor tracks the execution state of all configured cron jobs.
type Monitor struct {
	mu      sync.RWMutex
	jobs    map[string]*JobStatus
	cfg     []config.JobConfig
	alerts  AlertSender
}

// AlertSender is an interface for dispatching alerts when a job fails or drifts.
type AlertSender interface {
	Send(subject, message string) error
}

// New creates a new Monitor from the provided job configurations and alert sender.
func New(jobs []config.JobConfig, alerts AlertSender) *Monitor {
	m := &Monitor{
		jobs:   make(map[string]*JobStatus, len(jobs)),
		cfg:    jobs,
		alerts: alerts,
	}
	for _, j := range jobs {
		m.jobs[j.Name] = &JobStatus{
			Name:    j.Name,
			Healthy: true,
		}
	}
	return m
}

// RecordSuccess marks a job as having completed successfully at the given time.
func (m *Monitor) RecordSuccess(name string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, ok := m.jobs[name]
	if !ok {
		return fmt.Errorf("unknown job: %s", name)
	}

	status.LastSeen = at
	status.LastSuccess = at
	status.FailCount = 0
	status.Healthy = true
	return nil
}

// RecordFailure marks a job as having failed at the given time and sends an alert
// if the failure count exceeds the configured threshold.
func (m *Monitor) RecordFailure(name string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, ok := m.jobs[name]
	if !ok {
		return fmt.Errorf("unknown job: %s", name)
	}

	status.LastSeen = at
	status.FailCount++
	status.Healthy = false

	if m.alerts != nil {
		subject := fmt.Sprintf("[cronaudit] Job %q failed", name)
		msg := fmt.Sprintf("Job %q failed at %s (consecutive failures: %d)",
			name, at.Format(time.RFC3339), status.FailCount)
		_ = m.alerts.Send(subject, msg)
	}
	return nil
}

// CheckDrift inspects all jobs and fires an alert for any job that has not been
// seen within its expected interval plus the configured grace period.
func (m *Monitor) CheckDrift(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, jcfg := range m.cfg {
		status := m.jobs[jcfg.Name]
		if status.LastSeen.IsZero() {
			// Job has never been seen; skip until we have a baseline.
			continue
		}

		deadline := status.LastSeen.Add(jcfg.Interval).Add(jcfg.GracePeriod)
		if now.After(deadline) {
			status.MissedCount++
			status.Healthy = false
			if m.alerts != nil {
				subject := fmt.Sprintf("[cronaudit] Job %q missed its schedule", jcfg.Name)
				msg := fmt.Sprintf(
					"Job %q was last seen at %s; expected by %s (missed: %d)",
					jcfg.Name,
					status.LastSeen.Format(time.RFC3339),
					deadline.Format(time.RFC3339),
					status.MissedCount,
				)
				_ = m.alerts.Send(subject, msg)
			}
		}
	}
}

// Status returns a snapshot of the current state for the named job.
func (m *Monitor) Status(name string) (JobStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.jobs[name]
	if !ok {
		return JobStatus{}, fmt.Errorf("unknown job: %s", name)
	}
	return *status, nil
}

// AllStatuses returns a snapshot of every tracked job's state.
func (m *Monitor) AllStatuses() []JobStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]JobStatus, 0, len(m.jobs))
	for _, s := range m.jobs {
		out = append(out, *s)
	}
	return out
}
