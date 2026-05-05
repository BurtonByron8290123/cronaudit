package monitor

import (
	"testing"
	"time"
)

func makeRecord(job string, success bool, start time.Time, dur time.Duration) ExecutionRecord {
	return ExecutionRecord{
		JobName:    job,
		StartedAt:  start,
		FinishedAt: start.Add(dur),
		Success:    success,
		ExitCode:   0,
	}
}

func TestHistory_RecordAndLast(t *testing.T) {
	h := NewHistory(5)
	now := time.Now()

	rec := makeRecord("backup", true, now, 2*time.Second)
	h.Record(rec)

	got, ok := h.Last("backup")
	if !ok {
		t.Fatal("expected a record, got none")
	}
	if got.JobName != "backup" || !got.Success {
		t.Errorf("unexpected record: %+v", got)
	}
}

func TestHistory_Last_MissingJob(t *testing.T) {
	h := NewHistory(5)
	_, ok := h.Last("nonexistent")
	if ok {
		t.Error("expected no record for unknown job")
	}
}

func TestHistory_BoundedSize(t *testing.T) {
	const max = 3
	h := NewHistory(max)
	now := time.Now()

	for i := 0; i < 6; i++ {
		h.Record(makeRecord("cleanup", true, now.Add(time.Duration(i)*time.Minute), time.Second))
	}

	all := h.All("cleanup")
	if len(all) != max {
		t.Errorf("expected %d records, got %d", max, len(all))
	}
	// Oldest entries should have been evicted; last entry should be the 6th.
	expected := now.Add(5 * time.Minute)
	if !all[max-1].StartedAt.Equal(expected) {
		t.Errorf("expected last StartedAt %v, got %v", expected, all[max-1].StartedAt)
	}
}

func TestHistory_All_ReturnsCopy(t *testing.T) {
	h := NewHistory(5)
	now := time.Now()
	h.Record(makeRecord("sync", false, now, time.Second))

	all := h.All("sync")
	all[0].Success = true // mutate the copy

	got, _ := h.Last("sync")
	if got.Success {
		t.Error("All() should return a copy; internal state was mutated")
	}
}

func TestExecutionRecord_Duration(t *testing.T) {
	now := time.Now()
	rec := makeRecord("job", true, now, 5*time.Second)
	if rec.Duration() != 5*time.Second {
		t.Errorf("expected 5s duration, got %v", rec.Duration())
	}
}

func TestNewHistory_DefaultMax(t *testing.T) {
	h := NewHistory(0) // should default to 10
	if h.maxPerJob != 10 {
		t.Errorf("expected default maxPerJob=10, got %d", h.maxPerJob)
	}
}
