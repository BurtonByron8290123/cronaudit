package monitor

import (
	"testing"
	"time"
)

// TestPipeline_CircuitBreakerSuppressesAlertsWhenOpen verifies that once the
// circuit breaker opens for a job, subsequent failure alerts are suppressed
// until the open duration has elapsed.
func TestPipeline_CircuitBreakerSuppressesAlertsWhenOpen(t *testing.T) {
	var now time.Time
	now = time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	cbPolicy := CircuitBreakerPolicy{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		OpenDuration:     10 * time.Minute,
	}
	cb := NewCircuitBreakerStore(cbPolicy, func() time.Time { return now })

	alertCount := 0
	notifier := &recordingNotifier{fn: func() { alertCount++ }}

	cfg := makeConfig()
	state := NewStateStore()
	pipeline := NewPipeline(cfg, state, notifier)

	// Patch pipeline to gate alerts through the circuit breaker.
	// We simulate this by wrapping the notifier call manually.
	sendIfAllowed := func(key string, success bool) {
		if success {
			cb.RecordSuccess(key)
			return
		}
		cb.RecordFailure(key)
		if cb.Allow(key) {
			notifier.fn()
		}
	}

	_ = pipeline // pipeline wiring tested elsewhere; here we test CB logic directly

	// First failure: circuit closed, alert allowed.
	sendIfAllowed("backup", false)
	if alertCount != 1 {
		t.Fatalf("expected 1 alert after first failure, got %d", alertCount)
	}

	// Second failure: hits threshold, circuit opens.
	sendIfAllowed("backup", false)
	if alertCount != 2 {
		t.Fatalf("expected 2 alerts at threshold, got %d", alertCount)
	}
	if cb.State("backup") != CircuitOpen {
		t.Fatalf("expected circuit open, got %s", cb.State("backup"))
	}

	// Additional failures while open: suppressed.
	for i := 0; i < 5; i++ {
		sendIfAllowed("backup", false)
	}
	if alertCount != 2 {
		t.Errorf("expected alerts suppressed while open, got %d", alertCount)
	}

	// Advance time past open duration; next failure probes (half-open).
	now = now.Add(11 * time.Minute)
	sendIfAllowed("backup", false)
	if alertCount != 3 {
		t.Errorf("expected probe alert in half-open, got %d", alertCount)
	}
	if cb.State("backup") != CircuitOpen {
		t.Errorf("expected circuit re-opened after half-open failure, got %s", cb.State("backup"))
	}
}

// TestPipeline_CircuitBreakerClosesOnRecovery verifies that successful
// executions after the open duration close the circuit.
func TestPipeline_CircuitBreakerClosesOnRecovery(t *testing.T) {
	var now time.Time
	now = time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	cbPolicy := CircuitBreakerPolicy{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		OpenDuration:     5 * time.Minute,
	}
	cb := NewCircuitBreakerStore(cbPolicy, func() time.Time { return now })

	// Open the circuit.
	cb.RecordFailure("sync")
	cb.RecordFailure("sync")
	if cb.State("sync") != CircuitOpen {
		t.Fatalf("expected open, got %s", cb.State("sync"))
	}

	// Advance past duration and record success.
	now = now.Add(6 * time.Minute)
	cb.Allow("sync") // probe: transitions to half-open
	cb.RecordSuccess("sync")

	if cb.State("sync") != CircuitClosed {
		t.Errorf("expected circuit closed after recovery, got %s", cb.State("sync"))
	}
	if !cb.Allow("sync") {
		t.Error("expected Allow=true after recovery")
	}
}

// recordingNotifier is a minimal Notifier for integration tests.
type recordingNotifier struct {
	fn func()
}

func (r *recordingNotifier) Send(_ string) error {
	if r.fn != nil {
		r.fn()
	}
	return nil
}
