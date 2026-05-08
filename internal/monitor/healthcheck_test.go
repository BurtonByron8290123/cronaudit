package monitor

import (
	"testing"
	"time"
)

func fixedHealthClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestHealthChecker_AllHealthy(t *testing.T) {
	store := NewStateStore()
	now := time.Now()
	store.Set("backup", StatusOK)
	store.Set("cleanup", StatusOK)

	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(now)

	status := hc.Check()
	if !status.Healthy {
		t.Errorf("expected healthy, got unhealthy: %v", status.Jobs)
	}
	if len(status.Jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(status.Jobs))
	}
}

func TestHealthChecker_FailedJobMakesUnhealthy(t *testing.T) {
	store := NewStateStore()
	store.Set("backup", StatusOK)
	store.Set("cleanup", StatusFailed)

	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	status := hc.Check()
	if status.Healthy {
		t.Error("expected unhealthy due to failed job")
	}
}

func TestHealthChecker_StaleJobMakesUnhealthy(t *testing.T) {
	store := NewStateStore()
	store.Set("backup", StatusOK)

	hc := NewHealthChecker(store, 5*time.Minute)
	// advance clock well past staleTTL
	hc.clock = fixedHealthClock(time.Now().Add(20 * time.Minute))

	status := hc.Check()
	if status.Healthy {
		t.Error("expected unhealthy due to stale job")
	}
	if got := status.Jobs["backup"]; got != "stale" {
		t.Errorf("expected stale label, got %q", got)
	}
}

func TestHealthChecker_EmptyStoreIsHealthy(t *testing.T) {
	store := NewStateStore()
	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	status := hc.Check()
	if !status.Healthy {
		t.Error("expected empty store to be healthy")
	}
	if len(status.Jobs) != 0 {
		t.Errorf("expected no jobs, got %d", len(status.Jobs))
	}
}

func TestHealthChecker_CheckedAtIsSet(t *testing.T) {
	store := NewStateStore()
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(now)

	status := hc.Check()
	if !status.CheckedAt.Equal(now) {
		t.Errorf("expected CheckedAt=%v, got %v", now, status.CheckedAt)
	}
}
