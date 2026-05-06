package monitor

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestHeartbeat_StaleAfterDeadline(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	hm := NewHeartbeatMonitor(fixedNow(base))
	hm.Register("backup", 10*time.Minute)
	hm.Ping("backup")

	// Advance clock past deadline
	hm.now = fixedNow(base.Add(15 * time.Minute))
	stale := hm.Stale()
	if len(stale) != 1 || stale[0].JobName != "backup" {
		t.Fatalf("expected backup to be stale, got %v", stale)
	}
}

func TestHeartbeat_NotStaleWithinDeadline(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	hm := NewHeartbeatMonitor(fixedNow(base))
	hm.Register("sync", 10*time.Minute)
	hm.Ping("sync")

	hm.now = fixedNow(base.Add(5 * time.Minute))
	if stale := hm.Stale(); len(stale) != 0 {
		t.Fatalf("expected no stale jobs, got %v", stale)
	}
}

func TestHeartbeat_NeverSeenNotStale(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	hm := NewHeartbeatMonitor(fixedNow(base))
	hm.Register("nightly", time.Minute)

	hm.now = fixedNow(base.Add(1 * time.Hour))
	if stale := hm.Stale(); len(stale) != 0 {
		t.Fatalf("never-seen job should not appear in stale list")
	}
}

func TestHeartbeat_MarkAlerted_SuppressesRepeat(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	hm := NewHeartbeatMonitor(fixedNow(base))
	hm.Register("report", 5*time.Minute)
	hm.Ping("report")

	hm.now = fixedNow(base.Add(10 * time.Minute))
	stale := hm.Stale()
	if len(stale) != 1 {
		t.Fatalf("expected stale alert, got %v", stale)
	}
	hm.MarkAlerted("report")
	if stale2 := hm.Stale(); len(stale2) != 0 {
		t.Fatalf("expected suppressed alert after MarkAlerted")
	}
}

func TestHeartbeat_PingResetsAlert(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	hm := NewHeartbeatMonitor(fixedNow(base))
	hm.Register("cleanup", 5*time.Minute)
	hm.Ping("cleanup")

	hm.now = fixedNow(base.Add(10 * time.Minute))
	hm.MarkAlerted("cleanup")

	// Job runs again
	hm.now = fixedNow(base.Add(11 * time.Minute))
	hm.Ping("cleanup")

	// Should not be stale right after ping
	if stale := hm.Stale(); len(stale) != 0 {
		t.Fatalf("expected no stale after ping, got %v", stale)
	}
}

func TestHeartbeat_MultipleJobs(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	hm := NewHeartbeatMonitor(fixedNow(base))
	hm.Register("job-a", 5*time.Minute)
	hm.Register("job-b", 5*time.Minute)
	hm.Ping("job-a")
	hm.Ping("job-b")

	hm.now = fixedNow(base.Add(3 * time.Minute))
	hm.Ping("job-b") // job-b pings again

	hm.now = fixedNow(base.Add(8 * time.Minute))
	stale := hm.Stale()
	if len(stale) != 1 || stale[0].JobName != "job-a" {
		t.Fatalf("expected only job-a stale, got %v", stale)
	}
}
