package monitor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/config"
	"github.com/yourorg/cronaudit/internal/monitor"
)

// TestScheduler_IntegrationWithHistory verifies that the Scheduler and History
// work together: a job recorded as recently finished must not trigger drift,
// but the same job recorded as stale must.
func TestScheduler_IntegrationWithHistory(t *testing.T) {
	h := monitor.NewHistory(20)

	// Simulate a fresh run.
	h.Record(monitor.Record{
		JobName:    "sync",
		FinishedAt: time.Now(),
		Success:    true,
	})

	var mu sync.Mutex
	var driftFired bool

	onDrift := func(_ config.Job, _ time.Time) {
		mu.Lock()
		defer mu.Unlock()
		driftFired = true
	}

	jobs := []config.Job{{
		Name:                  "sync",
		IntervalSeconds:       3600,
		DriftThresholdSeconds: 300,
	}}

	s := monitor.NewScheduler(jobs, h, 25*time.Millisecond, onDrift)
	s.Start()
	time.Sleep(80 * time.Millisecond)
	s.Stop()

	mu.Lock()
	if driftFired {
		t.Error("drift must not fire for a recently completed job")
	}
	mu.Unlock()

	// Now simulate a stale run and re-run the scheduler.
	h.Record(monitor.Record{
		JobName:    "sync",
		FinishedAt: time.Now().Add(-2 * time.Hour),
		Success:    true,
	})

	mu.Lock()
	driftFired = false
	mu.Unlock()

	s2 := monitor.NewScheduler(jobs, h, 25*time.Millisecond, onDrift)
	s2.Start()
	time.Sleep(80 * time.Millisecond)
	s2.Stop()

	mu.Lock()
	defer mu.Unlock()
	if !driftFired {
		t.Error("drift must fire for a stale job")
	}
}
