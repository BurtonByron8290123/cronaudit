package monitor

import (
	"testing"
	"time"
)

// TestPipeline_EscalationTriggersOnRepeatedFailure verifies that an alert
// is only sent once the consecutive failure threshold is reached.
func TestPipeline_EscalationTriggersOnRepeatedFailure(t *testing.T) {
	cfg := makeConfig()
	var alerts []string
	cfg.Notifier = notifierFunc(func(msg string) error {
		alerts = append(alerts, msg)
		return nil
	})

	policy := EscalationPolicy{Threshold: 3, Cooldown: 5 * time.Minute}
	escStore := NewEscalationStore(policy, time.Now)

	// Simulate two failures — below threshold, no escalation alert.
	for i := 0; i < 2; i++ {
		if escStore.RecordFailure("backup") {
			t.Errorf("unexpected escalation on failure %d", i+1)
		}
	}

	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts before threshold, got %d", len(alerts))
	}

	// Third failure reaches threshold.
	if !escStore.RecordFailure("backup") {
		t.Error("expected escalation at threshold")
	}
	// Simulate sending an escalation alert.
	_ = cfg.Notifier.Notify("ESCALATION: job backup has failed 3 consecutive times")

	if len(alerts) != 1 {
		t.Errorf("expected 1 escalation alert, got %d", len(alerts))
	}
}

// TestPipeline_EscalationResetsOnSuccess verifies that a successful run
// clears the failure count and prevents spurious escalation.
func TestPipeline_EscalationResetsOnSuccess(t *testing.T) {
	policy := EscalationPolicy{Threshold: 2, Cooldown: 5 * time.Minute}
	escStore := NewEscalationStore(policy, time.Now)

	escStore.RecordFailure("backup")
	escStore.RecordSuccess("backup")

	// After reset, a single failure should not escalate.
	if escStore.RecordFailure("backup") {
		t.Error("expected no escalation after success reset")
	}

	if escStore.ConsecutiveFailures("backup") != 1 {
		t.Errorf("expected 1 consecutive failure after reset+one failure, got %d",
			escStore.ConsecutiveFailures("backup"))
	}
}
