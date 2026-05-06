package monitor

import (
	"sync"
	"time"
)

// JobStatus represents the last known status of a monitored cron job.
type JobStatus int

const (
	StatusUnknown JobStatus = iota
	StatusOK
	StatusFailed
	StatusDrifted
)

func (s JobStatus) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusFailed:
		return "failed"
	case StatusDrifted:
		return "drifted"
	default:
		return "unknown"
	}
}

// JobState holds the current runtime state for a single cron job.
type JobState struct {
	JobName    string
	Status     JobStatus
	LastSeen   time.Time
	LastChange time.Time
	Message    string
}

// StateStore tracks the live status of all monitored jobs.
type StateStore struct {
	mu     sync.RWMutex
	states map[string]*JobState
}

// NewStateStore creates an empty StateStore.
func NewStateStore() *StateStore {
	return &StateStore{
		states: make(map[string]*JobState),
	}
}

// Set updates the state for a given job.
func (s *StateStore) Set(name string, status JobStatus, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	existing, ok := s.states[name]
	if !ok || existing.Status != status {
		s.states[name] = &JobState{
			JobName:    name,
			Status:     status,
			LastSeen:   now,
			LastChange: now,
			Message:    msg,
		}
		return
	}
	existing.LastSeen = now
	existing.Message = msg
}

// Get returns the state for a job, or nil if not tracked.
func (s *StateStore) Get(name string) *JobState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.states[name]
	if !ok {
		return nil
	}
	copy := *st
	return &copy
}

// All returns a snapshot of all job states.
func (s *StateStore) All() []JobState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]JobState, 0, len(s.states))
	for _, st := range s.states {
		out = append(out, *st)
	}
	return out
}
