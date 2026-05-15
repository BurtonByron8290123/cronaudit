package monitor

import (
	"testing"
	"time"
)

func fixedSuppressionClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSuppressionLog_RecordAndAll(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	log := NewSuppressionLog(100, fixedSuppressionClock(now))

	log.Record("backup", SuppressionRateLimit, "cooldown active")
	log.Record("sync", SuppressionDedup, "duplicate alert")

	events := log.All()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].JobName != "backup" {
		t.Errorf("expected backup, got %s", events[0].JobName)
	}
	if events[1].Reason != SuppressionDedup {
		t.Errorf("expected dedup reason, got %s", events[1].Reason)
	}
	if !events[0].SuppressedAt.Equal(now) {
		t.Errorf("expected timestamp %v, got %v", now, events[0].SuppressedAt)
	}
}

func TestSuppressionLog_BoundedSize(t *testing.T) {
	now := time.Now()
	log := NewSuppressionLog(3, fixedSuppressionClock(now))

	for i := 0; i < 5; i++ {
		log.Record("job", SuppressionThrottle, "throttled")
	}

	events := log.All()
	if len(events) != 3 {
		t.Errorf("expected 3 events after bounding, got %d", len(events))
	}
}

func TestSuppressionLog_CountByReason(t *testing.T) {
	log := NewSuppressionLog(100, nil)

	log.Record("a", SuppressionRateLimit, "")
	log.Record("b", SuppressionRateLimit, "")
	log.Record("c", SuppressionSilence, "")
	log.Record("d", SuppressionCircuitBreak, "")

	counts := log.CountByReason()
	if counts[SuppressionRateLimit] != 2 {
		t.Errorf("expected 2 rate_limit, got %d", counts[SuppressionRateLimit])
	}
	if counts[SuppressionSilence] != 1 {
		t.Errorf("expected 1 silence, got %d", counts[SuppressionSilence])
	}
	if counts[SuppressionCircuitBreak] != 1 {
		t.Errorf("expected 1 circuit_breaker, got %d", counts[SuppressionCircuitBreak])
	}
}

func TestSuppressionLog_Flush(t *testing.T) {
	log := NewSuppressionLog(100, nil)

	log.Record("job1", SuppressionDedup, "dup")
	log.Record("job2", SuppressionThrottle, "throttle")

	flushed := log.Flush()
	if len(flushed) != 2 {
		t.Fatalf("expected 2 flushed events, got %d", len(flushed))
	}

	remaining := log.All()
	if len(remaining) != 0 {
		t.Errorf("expected empty log after flush, got %d events", len(remaining))
	}
}

func TestSuppressionLog_AllReturnsCopy(t *testing.T) {
	log := NewSuppressionLog(100, nil)
	log.Record("job", SuppressionRateLimit, "msg")

	events := log.All()
	events[0].JobName = "mutated"

	original := log.All()
	if original[0].JobName == "mutated" {
		t.Error("All() should return a copy, not a reference")
	}
}

func TestSuppressionLog_DefaultMaxLen(t *testing.T) {
	log := NewSuppressionLog(0, nil)
	if log.maxLen != 500 {
		t.Errorf("expected default maxLen 500, got %d", log.maxLen)
	}
}
