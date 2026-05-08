package monitor

import (
	"strings"
	"testing"
	"time"
)

var fixedDigestClock = func() time.Time {
	return time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
}

func TestDigestStore_RecordAndFlush(t *testing.T) {
	ds := NewDigestStore(fixedDigestClock)
	ds.Record("backup", "failure", "exit code 1")
	ds.Record("cleanup", "success", "ok")

	entries := ds.Flush()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].JobName != "backup" {
		t.Errorf("expected backup, got %s", entries[0].JobName)
	}
	if entries[1].Status != "success" {
		t.Errorf("expected success, got %s", entries[1].Status)
	}
}

func TestDigestStore_FlushClearsBuffer(t *testing.T) {
	ds := NewDigestStore(fixedDigestClock)
	ds.Record("job1", "failure", "timeout")
	ds.Flush()

	if ds.Len() != 0 {
		t.Errorf("expected empty buffer after flush, got %d", ds.Len())
	}
}

func TestDigestStore_Len(t *testing.T) {
	ds := NewDigestStore(fixedDigestClock)
	if ds.Len() != 0 {
		t.Errorf("expected 0, got %d", ds.Len())
	}
	ds.Record("job", "failure", "err")
	if ds.Len() != 1 {
		t.Errorf("expected 1, got %d", ds.Len())
	}
}

func TestFormatDigest_Empty(t *testing.T) {
	out := FormatDigest(nil)
	if !strings.Contains(out, "No events") {
		t.Errorf("expected no-events message, got: %s", out)
	}
}

func TestFormatDigest_WithEntries(t *testing.T) {
	entries := []DigestEntry{
		{JobName: "backup", Status: "failure", OccuredAt: fixedDigestClock(), Message: "exit 1"},
	}
	out := FormatDigest(entries)
	if !strings.Contains(out, "backup") {
		t.Errorf("expected job name in output: %s", out)
	}
	if !strings.Contains(out, "failure") {
		t.Errorf("expected status in output: %s", out)
	}
	if !strings.Contains(out, "1 event") {
		t.Errorf("expected event count in output: %s", out)
	}
}

func TestDigestStore_DefaultClock(t *testing.T) {
	ds := NewDigestStore(nil)
	ds.Record("job", "success", "ok")
	entries := ds.Flush()
	if entries[0].OccuredAt.IsZero() {
		t.Error("expected non-zero timestamp with default clock")
	}
}
