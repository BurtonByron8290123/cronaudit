package monitor

import (
	"testing"
	"time"
)

func TestMetricsStore_RecordSuccess(t *testing.T) {
	s := NewMetricsStore()
	now := time.Now()
	s.Record("backup", true, 2*time.Second, now)

	m, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected metrics to exist")
	}
	if m.TotalRuns != 1 {
		t.Errorf("TotalRuns: got %d, want 1", m.TotalRuns)
	}
	if m.SuccessCount != 1 {
		t.Errorf("SuccessCount: got %d, want 1", m.SuccessCount)
	}
	if m.FailureCount != 0 {
		t.Errorf("FailureCount: got %d, want 0", m.FailureCount)
	}
	if m.LastDuration != 2*time.Second {
		t.Errorf("LastDuration: got %v, want 2s", m.LastDuration)
	}
}

func TestMetricsStore_RecordFailure(t *testing.T) {
	s := NewMetricsStore()
	s.Record("cleanup", false, 500*time.Millisecond, time.Now())

	m, _ := s.Get("cleanup")
	if m.FailureCount != 1 {
		t.Errorf("FailureCount: got %d, want 1", m.FailureCount)
	}
	if m.SuccessCount != 0 {
		t.Errorf("SuccessCount: got %d, want 0", m.SuccessCount)
	}
}

func TestMetricsStore_AvgDuration(t *testing.T) {
	s := NewMetricsStore()
	now := time.Now()
	s.Record("job", true, 2*time.Second, now)
	s.Record("job", true, 4*time.Second, now)

	m, _ := s.Get("job")
	if m.AvgDuration != 3*time.Second {
		t.Errorf("AvgDuration: got %v, want 3s", m.AvgDuration)
	}
	if m.TotalRuns != 2 {
		t.Errorf("TotalRuns: got %d, want 2", m.TotalRuns)
	}
}

func TestMetricsStore_Get_Missing(t *testing.T) {
	s := NewMetricsStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected false for missing job")
	}
}

func TestMetricsStore_All_ReturnsAllJobs(t *testing.T) {
	s := NewMetricsStore()
	now := time.Now()
	s.Record("job-a", true, time.Second, now)
	s.Record("job-b", false, time.Second, now)

	all := s.All()
	if len(all) != 2 {
		t.Errorf("All: got %d entries, want 2", len(all))
	}
}

func TestMetricsStore_LastRunAt(t *testing.T) {
	s := NewMetricsStore()
	runAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	s.Record("sync", true, time.Second, runAt)

	m, _ := s.Get("sync")
	if !m.LastRunAt.Equal(runAt) {
		t.Errorf("LastRunAt: got %v, want %v", m.LastRunAt, runAt)
	}
}
