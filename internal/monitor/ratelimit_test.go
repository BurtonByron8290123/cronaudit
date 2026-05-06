package monitor

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRateLimiter_AllowsFirstAlert(t *testing.T) {
	rl := NewRateLimiter(5 * time.Minute)
	if !rl.Allow("job-a") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestRateLimiter_SuppressesWithinCooldown(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5 * time.Minute)
	rl.nowFn = fixedClock(now)

	rl.Allow("job-a") // first — allowed

	// Advance time by less than cooldown
	rl.nowFn = fixedClock(now.Add(2 * time.Minute))
	if rl.Allow("job-a") {
		t.Fatal("expected alert to be suppressed within cooldown")
	}
}

func TestRateLimiter_AllowsAfterCooldown(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5 * time.Minute)
	rl.nowFn = fixedClock(now)

	rl.Allow("job-a")

	// Advance time beyond cooldown
	rl.nowFn = fixedClock(now.Add(6 * time.Minute))
	if !rl.Allow("job-a") {
		t.Fatal("expected alert to be allowed after cooldown expires")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(5 * time.Minute)
	now := time.Now()
	rl.nowFn = fixedClock(now)

	rl.Allow("job-a")

	// job-b has never been seen — should be allowed
	if !rl.Allow("job-b") {
		t.Fatal("expected independent key to be allowed")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5 * time.Minute)
	rl.nowFn = fixedClock(now)

	rl.Allow("job-a")
	rl.Reset("job-a")

	// After reset, should be allowed again even within cooldown window
	if !rl.Allow("job-a") {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestRateLimiter_LastSent(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	rl := NewRateLimiter(5 * time.Minute)
	rl.nowFn = fixedClock(now)

	if _, ok := rl.LastSent("job-a"); ok {
		t.Fatal("expected no last-sent record before any alert")
	}

	rl.Allow("job-a")

	got, ok := rl.LastSent("job-a")
	if !ok {
		t.Fatal("expected last-sent record after alert")
	}
	if !got.Equal(now) {
		t.Fatalf("expected last-sent %v, got %v", now, got)
	}
}
