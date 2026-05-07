package monitor

import (
	"testing"
	"time"
)

var fixedCBClock = func() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

func defaultCBPolicy() CircuitBreakerPolicy {
	return CircuitBreakerPolicy{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		OpenDuration:     5 * time.Minute,
	}
}

func TestCircuitBreaker_InitiallyAllows(t *testing.T) {
	cb := NewCircuitBreakerStore(defaultCBPolicy(), fixedCBClock)
	if !cb.Allow("job1") {
		t.Error("expected Allow=true for new job")
	}
	if cb.State("job1") != CircuitClosed {
		t.Errorf("expected closed, got %s", cb.State("job1"))
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreakerStore(defaultCBPolicy(), fixedCBClock)
	for i := 0; i < 3; i++ {
		cb.RecordFailure("job1")
	}
	if cb.State("job1") != CircuitOpen {
		t.Errorf("expected open after threshold, got %s", cb.State("job1"))
	}
	if cb.Allow("job1") {
		t.Error("expected Allow=false when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpenAfterDuration(t *testing.T) {
	var now time.Time
	now = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreakerStore(defaultCBPolicy(), func() time.Time { return now })

	for i := 0; i < 3; i++ {
		cb.RecordFailure("job1")
	}
	now = now.Add(6 * time.Minute)
	if !cb.Allow("job1") {
		t.Error("expected Allow=true after open duration elapsed (half-open probe)")
	}
	if cb.State("job1") != CircuitHalfOpen {
		t.Errorf("expected half-open, got %s", cb.State("job1"))
	}
}

func TestCircuitBreaker_ClosesAfterSuccessThreshold(t *testing.T) {
	var now time.Time
	now = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreakerStore(defaultCBPolicy(), func() time.Time { return now })

	for i := 0; i < 3; i++ {
		cb.RecordFailure("job1")
	}
	now = now.Add(6 * time.Minute)
	cb.Allow("job1") // transitions to half-open

	cb.RecordSuccess("job1")
	if cb.State("job1") != CircuitHalfOpen {
		t.Errorf("expected still half-open after 1 success, got %s", cb.State("job1"))
	}
	cb.RecordSuccess("job1")
	if cb.State("job1") != CircuitClosed {
		t.Errorf("expected closed after success threshold, got %s", cb.State("job1"))
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	var now time.Time
	now = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	cb := NewCircuitBreakerStore(defaultCBPolicy(), func() time.Time { return now })

	for i := 0; i < 3; i++ {
		cb.RecordFailure("job1")
	}
	now = now.Add(6 * time.Minute)
	cb.Allow("job1") // transitions to half-open
	cb.RecordFailure("job1")
	if cb.State("job1") != CircuitOpen {
		t.Errorf("expected open after failure in half-open, got %s", cb.State("job1"))
	}
}

func TestCircuitBreaker_IndependentKeys(t *testing.T) {
	cb := NewCircuitBreakerStore(defaultCBPolicy(), fixedCBClock)
	for i := 0; i < 3; i++ {
		cb.RecordFailure("job1")
	}
	if !cb.Allow("job2") {
		t.Error("expected job2 unaffected by job1 failures")
	}
}

func TestCircuitState_String(t *testing.T) {
	cases := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("state %d: got %q, want %q", tc.state, got, tc.want)
		}
	}
}
