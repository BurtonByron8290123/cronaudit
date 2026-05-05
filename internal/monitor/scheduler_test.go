package monitor

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/config"
)

func testJob(name string, intervalSec, driftSec int) config.Job {
	return config.Job{
		Name:                  name,
		IntervalSeconds:       intervalSec,
		DriftThresholdSeconds: driftSec,
	}
}

func TestScheduler_TriggersDriftCallback(t *testing.T) {
	h := NewHistory(10)
	// Record a run that finished long ago so it is overdue.
	h.Record(Record{
		JobName:    "backup",
		FinishedAt: time.Now().Add(-10 * time.Minute),
		Success:    true,
	})

	var mu sync.Mutex
	var triggered []string

	onDrift := func(job config.Job, _ time.Time) {
		mu.Lock()
		defer mu.Unlock()
		triggered = append(triggered, job.Name)
	}

	jobs := []config.Job{testJob("backup", 60, 0)}
	s := NewScheduler(jobs, h, 20*time.Millisecond, onDrift)
	s.Start()
	time.Sleep(60 * time.Millisecond)
	s.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(triggered) == 0 {
		t.Error("expected drift callback to be triggered at least once")
	}
	if triggered[0] != "backup" {
		t.Errorf("expected job name 'backup', got %q", triggered[0])
	}
}

func TestScheduler_NoCallbackWhenWithinInterval(t *testing.T) {
	h := NewHistory(10)
	h.Record(Record{
		JobName:    "healthcheck",
		FinishedAt: time.Now(),
		Success:    true,
	})

	called := false
	onDrift := func(_ config.Job, _ time.Time) { called = true }

	jobs := []config.Job{testJob("healthcheck", 3600, 60)}
	s := NewScheduler(jobs, h, 20*time.Millisecond, onDrift)
	s.Start()
	time.Sleep(60 * time.Millisecond)
	s.Stop()

	if called {
		t.Error("drift callback should not fire when job is within its interval")
	}
}

func TestScheduler_NeverSeenJobTriggersDrift(t *testing.T) {
	h := NewHistory(10)

	var mu sync.Mutex
	var names []string
	onDrift := func(job config.Job, _ time.Time) {
		mu.Lock()
		defer mu.Unlock()
		names = append(names, job.Name)
	}

	jobs := []config.Job{testJob("newjob", 60, 0)}
	s := NewScheduler(jobs, h, 20*time.Millisecond, onDrift)
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(names) == 0 {
		t.Error("expected drift callback for job with no history")
	}
}
