package monitor

import (
	"log"
	"sync"
	"time"

	"github.com/yourorg/cronaudit/internal/config"
)

// SchedulerFunc is called when a job is expected but has not reported in.
type SchedulerFunc func(job config.Job, lastSeen time.Time)

// Scheduler periodically checks whether jobs have run within their expected
// interval and invokes the provided callback on drift or absence.
type Scheduler struct {
	cfg      []config.Job
	history  *History
	onDrift  SchedulerFunc
	stopCh   chan struct{}
	wg       sync.WaitGroup
	tickRate time.Duration
}

// NewScheduler creates a Scheduler that polls at the given tickRate.
func NewScheduler(jobs []config.Job, h *History, tickRate time.Duration, onDrift SchedulerFunc) *Scheduler {
	return &Scheduler{
		cfg:      jobs,
		history:  h,
		onDrift:  onDrift,
		stopCh:   make(chan struct{}),
		tickRate: tickRate,
	}
}

// Start begins the background polling loop.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.tickRate)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.check(time.Now())
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// deadline returns the absolute time by which a job must have finished,
// given its last completion time, expected interval, and allowed drift.
func deadline(job config.Job, lastFinished time.Time) time.Time {
	interval := time.Duration(job.IntervalSeconds) * time.Second
	drift := time.Duration(job.DriftThresholdSeconds) * time.Second
	return lastFinished.Add(interval).Add(drift)
}

// check evaluates each configured job against its expected interval.
func (s *Scheduler) check(now time.Time) {
	for _, job := range s.cfg {
		rec, err := s.history.Last(job.Name)
		if err != nil {
			// Job has never been seen; treat epoch as last seen.
			log.Printf("[scheduler] no history for job %q, treating as never run", job.Name)
			s.onDrift(job, time.Time{})
			continue
		}
		if now.After(deadline(job, rec.FinishedAt)) {
			s.onDrift(job, rec.FinishedAt)
		}
	}
}
