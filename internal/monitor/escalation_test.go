package monitor

import (
	"testing"
	"time"
)

func fixedEscalationClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func defaultEscalationPolicy() EscalationPolicy {
	return EscalationPolicy{
		Threshold: 3,
		Cooldown:  5 * time.Minute,
	}
}

func TestEscalation_NoEscalationBelowThreshold(t *testing.T) {
	now := time.Now()
	es := NewEscalationStore(defaultEscalationPolicy(), fixedEscalationClock(now))

	if es.RecordFailure("job1") {
		t.Error("expected no escalation on first failure")
	}
	if es.RecordFailure("job1") {
		t.Error("expected no escalation on second failure")
	}
}

func TestEscalation_EscalatesAtThreshold(t *testing.T) {
	now := time.Now()
	es := NewEscalationStore(defaultEscalationPolicy(), fixedEscalationClock(now))

	es.RecordFailure("job1")
	es.RecordFailure("job1")
	if !es.RecordFailure("job1") {
		t.Error("expected escalation at threshold")
	}
}

func TestEscalation_SuppressesWithinCooldown(t *testing.T) {
	now := time.Now()
	es := NewEscalationStore(defaultEscalationPolicy(), fixedEscalationClock(now))

	es.RecordFailure("job1")
	es.RecordFailure("job1")
	es.RecordFailure("job1") // escalates

	// Fourth failure within cooldown — should be suppressed.
	if es.RecordFailure("job1") {
		t.Error("expected suppression within cooldown")
	}
}

func TestEscalation_AllowsAfterCooldown(t *testing.T) {
	now := time.Now()
	var current time.Time = now
	es := NewEscalationStore(defaultEscalationPolicy(), func() time.Time { return current })

	es.RecordFailure("job1")
	es.RecordFailure("job1")
	es.RecordFailure("job1") // escalates

	current = now.Add(6 * time.Minute)
	if !es.RecordFailure("job1") {
		t.Error("expected re-escalation after cooldown")
	}
}

func TestEscalation_ResetOnSuccess(t *testing.T) {
	now := time.Now()
	es := NewEscalationStore(defaultEscalationPolicy(), fixedEscalationClock(now))

	es.RecordFailure("job1")
	es.RecordFailure("job1")
	es.RecordSuccess("job1")

	if es.ConsecutiveFailures("job1") != 0 {
		t.Errorf("expected 0 failures after reset, got %d", es.ConsecutiveFailures("job1"))
	}
}

func TestEscalation_IndependentKeys(t *testing.T) {
	now := time.Now()
	es := NewEscalationStore(defaultEscalationPolicy(), fixedEscalationClock(now))

	es.RecordFailure("job1")
	es.RecordFailure("job1")
	es.RecordFailure("job1") // job1 escalates

	// job2 should start fresh.
	if es.RecordFailure("job2") {
		t.Error("expected no escalation for independent key job2")
	}
}

func TestEscalation_ConsecutiveFailures(t *testing.T) {
	now := time.Now()
	es := NewEscalationStore(defaultEscalationPolicy(), fixedEscalationClock(now))

	es.RecordFailure("job1")
	es.RecordFailure("job1")

	if got := es.ConsecutiveFailures("job1"); got != 2 {
		t.Errorf("expected 2 consecutive failures, got %d", got)
	}
}
