package monitor

import (
	"sync"
	"time"
)

// SilenceWindow defines a time range during which alerts are suppressed.
type SilenceWindow struct {
	Start time.Time
	End   time.Time
}

// Active returns true if the given time falls within the silence window.
func (sw SilenceWindow) Active(t time.Time) bool {
	return !t.Before(sw.Start) && t.Before(sw.End)
}

// SilenceStore manages per-job silence windows.
type SilenceStore struct {
	mu      sync.RWMutex
	windows map[string][]SilenceWindow
	now     func() time.Time
}

// NewSilenceStore creates a new SilenceStore.
func NewSilenceStore(now func() time.Time) *SilenceStore {
	if now == nil {
		now = time.Now
	}
	return &SilenceStore{
		windows: make(map[string][]SilenceWindow),
		now:     now,
	}
}

// Add registers a silence window for the given job key.
func (s *SilenceStore) Add(key string, start, end time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[key] = append(s.windows[key], SilenceWindow{Start: start, End: end})
}

// IsSilenced returns true if the job is currently within any silence window.
func (s *SilenceStore) IsSilenced(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	for _, w := range s.windows[key] {
		if w.Active(now) {
			return true
		}
	}
	return false
}

// Purge removes expired silence windows for all keys.
func (s *SilenceStore) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	for key, windows := range s.windows {
		var active []SilenceWindow
		for _, w := range windows {
			if now.Before(w.End) {
				active = append(active, w)
			}
		}
		if len(active) == 0 {
			delete(s.windows, key)
		} else {
			s.windows[key] = active
		}
	}
}
