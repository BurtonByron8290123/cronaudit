package monitor

import (
	"testing"
	"time"
)

func fixedThrottleClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func defaultThrottlePolicy() ThrottlePolicy {
	return ThrottlePolicy{MaxAlerts: 3, Window: 10 * time.Minute}
}

func TestThrottleStore_AllowsUpToMax(t *testing.T) {
	now := time.Now()
	s := NewThrottleStore(defaultThrottlePolicy(), fixedThrottleClock(now))

	for i := 0; i < 3; i++ {
		if !s.Allow("job-a") {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
	if s.Allow("job-a") {
		t.Fatal("expected deny after max alerts reached")
	}
}

func TestThrottleStore_ResetsAfterWindow(t *testing.T) {
	base := time.Now()
	clock := fixedThrottleClock(base)
	var currentTime time.Time = base
	s := NewThrottleStore(defaultThrottlePolicy(), func() time.Time { return currentTime })

	for i := 0; i < 3; i++ {
		s.Allow("job-b")
	}
	if s.Allow("job-b") {
		t.Fatal("expected deny within window")
	}

	currentTime = base.Add(11 * time.Minute)
	if !s.Allow("job-b") {
		t.Fatal("expected allow after window expired")
	}
	_ = clock
}

func TestThrottleStore_IndependentKeys(t *testing.T) {
	now := time.Now()
	s := NewThrottleStore(ThrottlePolicy{MaxAlerts: 1, Window: time.Minute}, fixedThrottleClock(now))

	if !s.Allow("job-x") {
		t.Fatal("expected allow for job-x")
	}
	if s.Allow("job-x") {
		t.Fatal("expected deny for job-x after max")
	}
	if !s.Allow("job-y") {
		t.Fatal("expected allow for job-y (independent key)")
	}
}

func TestThrottleStore_ResetClearsState(t *testing.T) {
	now := time.Now()
	s := NewThrottleStore(ThrottlePolicy{MaxAlerts: 1, Window: time.Minute}, fixedThrottleClock(now))

	s.Allow("job-c")
	if s.Allow("job-c") {
		t.Fatal("expected deny before reset")
	}
	s.Reset("job-c")
	if !s.Allow("job-c") {
		t.Fatal("expected allow after reset")
	}
}

func TestThrottleStore_CountReflectsAlerts(t *testing.T) {
	now := time.Now()
	s := NewThrottleStore(defaultThrottlePolicy(), fixedThrottleClock(now))

	if s.Count("job-d") != 0 {
		t.Fatal("expected count 0 for unseen key")
	}
	s.Allow("job-d")
	s.Allow("job-d")
	if s.Count("job-d") != 2 {
		t.Fatalf("expected count 2, got %d", s.Count("job-d"))
	}
}
