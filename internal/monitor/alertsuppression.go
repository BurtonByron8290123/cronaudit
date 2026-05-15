package monitor

import (
	"sync"
	"time"
)

// SuppressionReason describes why an alert was suppressed.
type SuppressionReason string

const (
	SuppressionRateLimit    SuppressionReason = "rate_limit"
	SuppressionDedup        SuppressionReason = "dedup"
	SuppressionCircuitBreak SuppressionReason = "circuit_breaker"
	SuppressionSilence      SuppressionReason = "silence_window"
	SuppressionThrottle     SuppressionReason = "throttle"
)

// SuppressionEvent records a single suppressed alert.
type SuppressionEvent struct {
	JobName   string
	Reason    SuppressionReason
	SuppressedAt time.Time
	Message   string
}

// SuppressionLog tracks suppressed alerts for observability and diagnostics.
type SuppressionLog struct {
	mu     sync.Mutex
	events []SuppressionEvent
	maxLen int
	clock  func() time.Time
}

// NewSuppressionLog creates a SuppressionLog that retains up to maxLen events.
func NewSuppressionLog(maxLen int, clock func() time.Time) *SuppressionLog {
	if clock == nil {
		clock = time.Now
	}
	if maxLen <= 0 {
		maxLen = 500
	}
	return &SuppressionLog{
		maxLen: maxLen,
		clock:  clock,
	}
}

// Record appends a suppression event to the log.
func (s *SuppressionLog) Record(jobName string, reason SuppressionReason, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event := SuppressionEvent{
		JobName:      jobName,
		Reason:       reason,
		SuppressedAt: s.clock(),
		Message:      message,
	}
	s.events = append(s.events, event)
	if len(s.events) > s.maxLen {
		s.events = s.events[len(s.events)-s.maxLen:]
	}
}

// All returns a copy of all recorded suppression events.
func (s *SuppressionLog) All() []SuppressionEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]SuppressionEvent, len(s.events))
	copy(out, s.events)
	return out
}

// CountByReason returns a map of suppression reason to count.
func (s *SuppressionLog) CountByReason() map[SuppressionReason]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	counts := make(map[SuppressionReason]int)
	for _, e := range s.events {
		counts[e.Reason]++
	}
	return counts
}

// Flush clears all recorded events and returns them.
func (s *SuppressionLog) Flush() []SuppressionEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]SuppressionEvent, len(s.events))
	copy(out, s.events)
	s.events = nil
	return out
}
