package monitor

import (
	"testing"
	"time"
)

func TestCheckDrift_WithinLimit(t *testing.T) {
	now := time.Now()
	expected := now
	actual := now.Add(30 * time.Second)
	limit := 1 * time.Minute

	result := CheckDrift("backup-job", expected, actual, limit)

	if result.JobName != "backup-job" {
		t.Errorf("expected job name 'backup-job', got '%s'", result.JobName)
	}
	if result.Drift != 30*time.Second {
		t.Errorf("expected drift 30s, got %s", result.Drift)
	}
	if result.ExceededLimit {
		t.Error("expected drift to be within limit, but ExceededLimit is true")
	}
}

func TestCheckDrift_ExceedsLimit(t *testing.T) {
	now := time.Now()
	expected := now
	actual := now.Add(5 * time.Minute)
	limit := 2 * time.Minute

	result := CheckDrift("cleanup-job", expected, actual, limit)

	if !result.ExceededLimit {
		t.Error("expected drift to exceed limit, but ExceededLimit is false")
	}
	if result.Drift != 5*time.Minute {
		t.Errorf("expected drift 5m, got %s", result.Drift)
	}
}

func TestCheckDrift_NegativeDriftAbsoluteValue(t *testing.T) {
	now := time.Now()
	expected := now
	// actual is before expected (early execution)
	actual := now.Add(-90 * time.Second)
	limit := 1 * time.Minute

	result := CheckDrift("early-job", expected, actual, limit)

	if result.Drift != 90*time.Second {
		t.Errorf("expected absolute drift 90s, got %s", result.Drift)
	}
	if !result.ExceededLimit {
		t.Error("expected drift to exceed limit for early execution")
	}
}

func TestCheckDrift_ZeroDrift(t *testing.T) {
	now := time.Now()
	result := CheckDrift("exact-job", now, now, 10*time.Second)

	if result.Drift != 0 {
		t.Errorf("expected zero drift, got %s", result.Drift)
	}
	if result.ExceededLimit {
		t.Error("zero drift should not exceed limit")
	}
}

func TestDriftResult_String(t *testing.T) {
	now := time.Now()
	result := DriftResult{
		JobName:       "test-job",
		ExpectedAt:    now,
		ActualAt:      now.Add(45 * time.Second),
		Drift:         45 * time.Second,
		ExceededLimit: false,
	}

	s := result.String()
	if s == "" {
		t.Error("expected non-empty string from DriftResult.String()")
	}
}
