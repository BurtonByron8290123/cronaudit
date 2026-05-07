package monitor

import (
	"testing"
	"time"
)

var baseTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedRetryClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func defaultPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   10 * time.Second,
		MaxDelay:    60 * time.Second,
	}
}

func TestRetryStore_FirstAttemptAllowed(t *testing.T) {
	store := NewRetryStore(defaultPolicy(), fixedRetryClock(baseTime))
	if !store.ShouldRetry("job1") {
		t.Fatal("expected first attempt to be allowed")
	}
}

func TestRetryStore_RecordAdvancesAttempts(t *testing.T) {
	store := NewRetryStore(defaultPolicy(), fixedRetryClock(baseTime))
	store.Record("job1")
	if store.Attempts("job1") != 1 {
		t.Fatalf("expected 1 attempt, got %d", store.Attempts("job1"))
	}
}

func TestRetryStore_SuppressesBeforeDelay(t *testing.T) {
	store := NewRetryStore(defaultPolicy(), fixedRetryClock(baseTime))
	store.Record("job1")
	// next retry is baseTime + 10s; clock is still at baseTime
	if store.ShouldRetry("job1") {
		t.Fatal("expected retry to be suppressed before delay")
	}
}

func TestRetryStore_AllowsAfterDelay(t *testing.T) {
	now := baseTime
	store := NewRetryStore(defaultPolicy(), func() time.Time { return now })
	store.Record("job1")
	now = baseTime.Add(15 * time.Second)
	if !store.ShouldRetry("job1") {
		t.Fatal("expected retry to be allowed after delay")
	}
}

func TestRetryStore_ExponentialBackoff(t *testing.T) {
	now := baseTime
	store := NewRetryStore(defaultPolicy(), func() time.Time { return now })
	store.Record("job1") // attempt 1, next delay = 10s
	now = now.Add(15 * time.Second)
	store.Record("job1") // attempt 2, next delay = 20s
	now = now.Add(10 * time.Second)
	if store.ShouldRetry("job1") {
		t.Fatal("expected retry suppressed; 20s delay not yet elapsed")
	}
	now = now.Add(15 * time.Second)
	if !store.ShouldRetry("job1") {
		t.Fatal("expected retry allowed after 20s delay")
	}
}

func TestRetryStore_MaxAttemptsExhausted(t *testing.T) {
	now := baseTime
	store := NewRetryStore(defaultPolicy(), func() time.Time { return now })
	for i := 0; i < 3; i++ {
		store.Record("job1")
		now = now.Add(120 * time.Second)
	}
	if store.ShouldRetry("job1") {
		t.Fatal("expected retries exhausted after MaxAttempts")
	}
}

func TestRetryStore_ResetClearsState(t *testing.T) {
	store := NewRetryStore(defaultPolicy(), fixedRetryClock(baseTime))
	for i := 0; i < 3; i++ {
		store.Record("job1")
	}
	store.Reset("job1")
	if store.Attempts("job1") != 0 {
		t.Fatal("expected attempts to be 0 after reset")
	}
	if !store.ShouldRetry("job1") {
		t.Fatal("expected retry allowed after reset")
	}
}

func TestRetryStore_MaxDelayCapped(t *testing.T) {
	now := baseTime
	policy := RetryPolicy{MaxAttempts: 10, BaseDelay: 30 * time.Second, MaxDelay: 60 * time.Second}
	store := NewRetryStore(policy, func() time.Time { return now })
	for i := 0; i < 5; i++ {
		store.Record("job1")
		now = now.Add(120 * time.Second)
	}
	// delay should be capped at 60s; advance only 30s — should be suppressed
	now = now.Add(30 * time.Second)
	if store.ShouldRetry("job1") {
		t.Fatal("expected delay to be capped and retry suppressed")
	}
}
