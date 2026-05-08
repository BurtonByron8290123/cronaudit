package monitor

import (
	"strings"
	"testing"
	"time"
)

func TestPipeline_DigestAccumulatesFailures(t *testing.T) {
	cfg := makeConfig()
	pipeline := NewPipeline(cfg)
	ds := NewDigestStore(func() time.Time { return time.Now() })

	// Simulate two failing job runs and record to digest.
	for i := 0; i < 2; i++ {
		result := pipeline.RunJob("test-job", func() error {
			return fmt.Errorf("job failed")
		})
		if result == nil {
			ds.Record("test-job", "failure", "job failed")
		}
	}

	entries := ds.Flush()
	if len(entries) != 2 {
		t.Fatalf("expected 2 digest entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Status != "failure" {
			t.Errorf("expected failure status, got %s", e.Status)
		}
	}
}

func TestPipeline_DigestFormatIncludesAllJobs(t *testing.T) {
	ds := NewDigestStore(func() time.Time {
		return time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)
	})
	ds.Record("backup", "failure", "timeout")
	ds.Record("cleanup", "success", "ok")
	ds.Record("report", "failure", "exit 2")

	out := FormatDigest(ds.Flush())
	for _, name := range []string{"backup", "cleanup", "report"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in digest output", name)
		}
	}
	if !strings.Contains(out, "3 event") {
		t.Errorf("expected 3 events in digest summary, got: %s", out)
	}
}
