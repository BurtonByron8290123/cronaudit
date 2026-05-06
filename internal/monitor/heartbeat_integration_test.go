package monitor_test

import (
	"testing"
	"time"

	"github.com/cronaudit/internal/monitor"
)

// TestHeartbeat_IntegrationWithPipeline verifies that HeartbeatMonitor
// integrates correctly: jobs pinged via a simulated pipeline run are
// not reported stale, while a missing job eventually triggers a stale
// entry that can be acted upon by an alert dispatcher.
func TestHeartbeat_IntegrationWithPipeline(t *testing.T) {
	var tick time.Time
	tick = time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)
	nowFn := func() time.Time { return tick }

	hm := monitor.NewHeartbeatMonitor(nowFn)
	hm.Register("etl-import", 30*time.Minute)
	hm.Register("db-backup", 60*time.Minute)

	// Simulate both jobs running at t=0
	hm.Ping("etl-import")
	hm.Ping("db-backup")

	// t+20m: etl-import runs again; db-backup does not
	tick = tick.Add(20 * time.Minute)
	hm.Ping("etl-import")

	// t+55m: etl-import still within deadline (last seen at t+20)
	// db-backup last seen at t=0, deadline=60m, so not yet stale
	tick = tick.Add(35 * time.Minute)
	stale := hm.Stale()
	if len(stale) != 0 {
		t.Fatalf("at t+55m expected no stale jobs, got %v", stale)
	}

	// t+90m: db-backup last seen at t=0, now 90m elapsed > 60m deadline
	// etl-import last seen at t+20m, now 70m elapsed > 30m deadline
	tick = tick.Add(35 * time.Minute)
	stale = hm.Stale()
	if len(stale) != 2 {
		t.Fatalf("at t+90m expected 2 stale jobs, got %v", stale)
	}

	// Alert and verify suppression
	for _, s := range stale {
		hm.MarkAlerted(s.JobName)
	}
	if s2 := hm.Stale(); len(s2) != 0 {
		t.Fatalf("expected no stale after MarkAlerted, got %v", s2)
	}

	// etl-import recovers; db-backup remains silent
	tick = tick.Add(5 * time.Minute)
	hm.Ping("etl-import")

	tick = tick.Add(35 * time.Minute)
	stale = hm.Stale()
	if len(stale) != 1 || stale[0].JobName != "db-backup" {
		t.Fatalf("expected only db-backup stale after etl-import recovery, got %v", stale)
	}
}
